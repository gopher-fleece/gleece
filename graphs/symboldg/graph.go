package symboldg

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/dot"
)

type SymbolGraphBuilder interface {
	AddController(request CreateControllerNode) (*SymbolNode, error)
	AddRoute(request CreateRouteNode) (*SymbolNode, error)
	AddRouteParam(request CreateParameterNode) (*SymbolNode, error)
	AddRouteRetVal(request CreateReturnValueNode) (*SymbolNode, error)
	AddStruct(request CreateStructNode) (*SymbolNode, error)
	AddEnum(request CreateEnumNode) (*SymbolNode, error)
	AddField(request CreateFieldNode) (*SymbolNode, error)
	AddConst(request CreateConstNode) (*SymbolNode, error)
	AddPrimitive(kind PrimitiveType) *SymbolNode
	AddSpecial(special SpecialType) *SymbolNode
	AddEdge(from, to graphs.SymbolKey, kind SymbolEdgeKind, meta map[string]string)

	Structs() []metadata.StructMeta
	Enums() []metadata.EnumMeta
	FindByKind(kind common.SymKind) []*SymbolNode

	IsPrimitivePresent(primitive PrimitiveType) bool
	IsSpecialPresent(special SpecialType) bool

	// Children returns direct outward SymbolNode dependencies from the given node,
	// applying the given traversal filter if non-nil.
	Children(node *SymbolNode, filter *TraversalFilter) []*SymbolNode

	// Parents returns nodes that have edges pointing to the given node,
	// applying the given traversal filter if non-nil.
	Parents(node *SymbolNode, filter *TraversalFilter) []*SymbolNode

	// Descendants returns all transitive children reachable from root,
	// applying the filter at each step to decide traversal and inclusion.
	Descendants(root *SymbolNode, filter *TraversalFilter) []*SymbolNode

	ToDot(theme *dot.DotTheme) string
	String() string
}

type SymbolGraph struct {
	// A map for retrieving full symbol keys from their comparable, non-versioned IDs
	lookupKeys map[string]graphs.SymbolKey

	nodes map[string]*SymbolNode // keyed by ast node

	edges map[string][]SymbolEdge // Node relations

	deps    map[string]map[graphs.SymbolKey]struct{} // from → set of to
	revDeps map[string]map[graphs.SymbolKey]struct{} // to   → set of from
}

func NewSymbolGraph() SymbolGraph {
	return SymbolGraph{
		lookupKeys: map[string]graphs.SymbolKey{},
		edges:      make(map[string][]SymbolEdge),
		nodes:      make(map[string]*SymbolNode),
		deps:       make(map[string]map[graphs.SymbolKey]struct{}),
		revDeps:    make(map[string]map[graphs.SymbolKey]struct{}),
	}
}

func (g *SymbolGraph) addNode(n *SymbolNode) {
	baseId := n.Id.BaseId()
	g.nodes[baseId] = n
	g.lookupKeys[baseId] = n.Id
}

// AddEdge adds a semantic relationship FROM → TO.
// For example: AddEdge(structKey, receiverKey, EdgeKindReceiver, nil)
// means "struct has receiver".
func (g *SymbolGraph) AddEdge(from, to graphs.SymbolKey, kind SymbolEdgeKind, meta map[string]string) {
	fromBaseId := from.BaseId()
	toBaseId := to.BaseId()

	// Always update dependency graphs
	if g.deps[fromBaseId] == nil {
		g.deps[fromBaseId] = make(map[graphs.SymbolKey]struct{})
	}
	g.deps[fromBaseId][to] = struct{}{}

	if g.revDeps[toBaseId] == nil {
		g.revDeps[toBaseId] = make(map[graphs.SymbolKey]struct{})
	}
	g.revDeps[toBaseId][from] = struct{}{}

	// Check for duplicate edges before appending
	existingEdges := g.edges[fromBaseId]
	for _, edge := range existingEdges {
		if edge.To.Equals(to) && edge.Kind == kind {
			return // Duplicate edge — skip
		}
	}

	// Append the new edge
	g.edges[fromBaseId] = append(existingEdges, SymbolEdge{
		To:       to,
		Kind:     kind,
		Metadata: meta,
	})

	g.lookupKeys[fromBaseId] = from
}

func (g *SymbolGraph) AddController(request CreateControllerNode) (*SymbolNode, error) {
	symNode, err := g.createAndAddSymNode(
		request.Data.Node,
		common.SymKindController,
		request.Data.FVersion,
		request.Annotations,
		request.Data,
	)

	if err != nil {
		return nil, err
	}

	// Add the controller to the graph
	g.addNode(symNode)

	return symNode, nil
}

func (g *SymbolGraph) AddRoute(request CreateRouteNode) (*SymbolNode, error) {
	symNode, err := g.createAndAddSymNode(
		request.Data.Node,
		common.SymKindReceiver,
		request.Data.FVersion,
		request.Annotations,
		request.Data,
	)

	if err != nil {
		return nil, err
	}

	// Add the route to the graph
	g.addNode(symNode)

	// Add a dependency FROM the parent controller TO the route
	g.AddEdge(request.ParentController.SymbolKey(), symNode.Id, EdgeKindReceiver, nil)

	return symNode, nil
}

func (g *SymbolGraph) AddRouteParam(request CreateParameterNode) (*SymbolNode, error) {
	symNode, err := g.createAndAddSymNode(
		request.Data.Node,
		common.SymKindParameter,
		&request.ParentRoute.FVersion,
		nil,
		request.Data,
	)

	if err != nil {
		return nil, err
	}

	// Add a dependency FROM the route TO the parameter node itself
	g.AddEdge(request.ParentRoute.SymbolKey(), symNode.Id, EdgeKindParam, nil)

	// Add a dependency FROM the PARAM TO the PARAM TYPE node - DEPRECATED?
	// g.AddEdge(symNode.Id, request.Data.Type.TypeRefKey, EdgeKindReference, nil)

	return symNode, nil
}

func (g *SymbolGraph) AddRouteRetVal(request CreateReturnValueNode) (*SymbolNode, error) {
	symNode, err := g.createAndAddSymNode(
		request.Data.Node,
		common.SymKindReturnType,
		&request.ParentRoute.FVersion,
		nil,
		request.Data,
	)

	if err != nil {
		return nil, err
	}

	g.AddEdge(request.ParentRoute.SymbolKey(), symNode.Id, EdgeKindRetVal, nil)

	// Add a dependency FROM the RETVAL TO the RETVAL TYPE node - DEPRECATED?
	// g.AddEdge(symNode.Id, request.Data.Type.TypeRefKey, EdgeKindReference, nil)

	return symNode, nil
}

func (g *SymbolGraph) AddStruct(request CreateStructNode) (*SymbolNode, error) {
	symNode, err := g.createAndAddSymNode(
		request.Data.Node,
		common.SymKindStruct,
		request.Data.FVersion,
		request.Annotations,
		request.Data,
	)

	if err != nil {
		return nil, err
	}

	// Register the fields to the struct
	for _, field := range request.Data.Fields {
		fieldKey := graphs.NewSymbolKey(field.Node, field.FVersion)
		g.AddEdge(symNode.Id, fieldKey, EdgeKindField, nil)
	}

	return symNode, nil
}

func (g *SymbolGraph) AddEnum(request CreateEnumNode) (*SymbolNode, error) {
	// Add the enum node itself
	symNode, err := g.createAndAddSymNode(
		request.Data.Node,
		common.SymKindEnum,
		request.Data.FVersion,
		request.Annotations,
		request.Data,
	)

	if err != nil {
		return nil, err
	}

	// Link each enum value
	typeRef := graphs.NewUniverseSymbolKey(string(request.Data.ValueKind))

	// Add the underlying primitive, if it does not exist.
	// This is not strictly necessary due to how the visitor code is built but it serves as a bit of
	// an extra layer of 'protection' against malformed graphs.
	if g.nodes[typeRef.BaseId()] == nil {
		primitive, isPrimitive := ToPrimitiveType(string(request.Data.ValueKind))
		if !isPrimitive {
			return nil, fmt.Errorf(
				"value kind for enum '%s' is '%s' which is unexpected",
				request.Data.Name,
				request.Data.ValueKind,
			)
		}
		g.AddPrimitive(primitive)
	}

	for _, valueDef := range request.Data.Values {
		valueNode, err := g.createAndAddSymNode(
			valueDef.Node,
			common.SymKindConstant,
			valueDef.FVersion,
			valueDef.Annotations,
			valueDef,
		)

		if err != nil {
			return nil, err
		}

		// Link the ENUM TYPE to its VALUE
		g.AddEdge(symNode.Id, valueNode.Id, EdgeKindValue, nil)

		// Link the enum VALUE to its underlying type
		g.AddEdge(valueNode.Id, typeRef, EdgeKindReference, nil)
	}

	return symNode, nil
}

func (g *SymbolGraph) AddField(request CreateFieldNode) (*SymbolNode, error) {
	symNode, err := g.createAndAddSymNode(
		request.Data.Node,
		common.SymKindField,
		request.Data.FVersion,
		request.Annotations,
		request.Data,
	)
	if err != nil {
		return nil, err
	}

	typeRef, err := getTypeRef(request.Data.Type)
	if err != nil {
		return nil, err
	}

	g.AddEdge(symNode.Id, typeRef, EdgeKindType, nil)
	return symNode, err
}

func (g *SymbolGraph) AddConst(request CreateConstNode) (*SymbolNode, error) {
	return g.createAndAddSymNode(
		request.Data.Node,
		common.SymKindConstant,
		request.Data.FVersion, // For constants, the FVersion is the Constants'- not the underlying type's
		request.Annotations,
		request.Data,
	)
}

func (g *SymbolGraph) FindByKind(kind common.SymKind) []*SymbolNode {
	var results []*SymbolNode

	for _, node := range g.nodes {
		if node.Kind == kind {
			results = append(results, node)
		}
	}

	return results
}

func (g *SymbolGraph) Enums() []metadata.EnumMeta {
	nodes := g.FindByKind(common.SymKindEnum)

	var results []metadata.EnumMeta
	for _, node := range nodes {
		if enum, ok := node.Data.(metadata.EnumMeta); ok {
			results = append(results, enum)
		}
	}
	return results
}

func (g *SymbolGraph) Structs() []metadata.StructMeta {
	nodes := g.FindByKind(common.SymKindStruct)

	var results []metadata.StructMeta
	for _, node := range nodes {
		if structMeta, ok := node.Data.(metadata.StructMeta); ok {
			results = append(results, structMeta)
		}
	}
	return results
}

func (g *SymbolGraph) IsPrimitivePresent(primitive PrimitiveType) bool {
	return g.builtinSymbolExists(string(primitive), true)
}

func (g *SymbolGraph) AddPrimitive(p PrimitiveType) *SymbolNode {
	// Primitives are always 'universe' types
	return g.addBuiltinSymbol(string(p), common.SymKindBuiltin, true)
}

func (g *SymbolGraph) IsSpecialPresent(special SpecialType) bool {
	return g.builtinSymbolExists(string(special), special.IsUniverse())
}

func (g *SymbolGraph) AddSpecial(special SpecialType) *SymbolNode {
	return g.addBuiltinSymbol(string(special), common.SymKindSpecialBuiltin, special.IsUniverse())
}

func (g *SymbolGraph) Children(node *SymbolNode, filter *TraversalFilter) []*SymbolNode {
	var result []*SymbolNode
	for _, edge := range g.edges[node.Id.BaseId()] {
		// Edge kind match (if specified)
		if !shouldIncludeEdge(edge, filter) {
			continue
		}
		child, ok := g.nodes[edge.To.BaseId()]
		if !ok || !shouldIncludeNode(child, filter) {
			continue
		}
		result = append(result, child)
	}
	return result
}

func (g *SymbolGraph) Parents(node *SymbolNode, filter *TraversalFilter) []*SymbolNode {
	var result []*SymbolNode
	for parentKey := range g.revDeps[node.Id.BaseId()] {
		edges := g.edges[parentKey.BaseId()]
		for _, edge := range edges {
			if edge.To != node.Id {
				continue
			}
			if !shouldIncludeEdge(edge, filter) {
				continue
			}
			parentNode := g.nodes[parentKey.BaseId()]
			if parentNode != nil && shouldIncludeNode(parentNode, filter) {
				result = append(result, parentNode)
			}
		}
	}
	return result
}

func (g *SymbolGraph) Descendants(root *SymbolNode, filter *TraversalFilter) []*SymbolNode {
	visited := make(map[graphs.SymbolKey]struct{})
	var result []*SymbolNode

	var walk func(*SymbolNode)
	walk = func(n *SymbolNode) {
		for _, child := range g.Children(n, filter) {
			if _, seen := visited[child.Id]; seen {
				continue
			}
			visited[child.Id] = struct{}{}
			result = append(result, child)
			walk(child)
		}
	}

	walk(root)
	return result
}

func (g *SymbolGraph) builtinSymbolExists(name string, isUniverse bool) bool {
	var key graphs.SymbolKey
	if isUniverse {
		key = graphs.NewUniverseSymbolKey(name)
	} else {
		key = graphs.NewNonUniverseBuiltInSymbolKey(name)
	}

	if _, exists := g.nodes[key.BaseId()]; exists {
		return true
	}
	return false
}

func (g *SymbolGraph) createAndAddSymNode(
	node ast.Node,
	kind common.SymKind,
	fVersion *gast.FileVersion,
	annotations *annotations.AnnotationHolder,
	data any,
) (*SymbolNode, error) {
	existing, key, err := g.idempotencyGuard(node, fVersion)
	if existing != nil || err != nil {
		return existing, err
	}

	symNode := &SymbolNode{
		Id:          key,
		Kind:        kind,
		Version:     fVersion,
		Data:        data,
		Annotations: annotations,
	}

	g.addNode(symNode)
	return symNode, nil
}

func (g *SymbolGraph) addBuiltinSymbol(typeName string, kind common.SymKind, isUniverse bool) *SymbolNode {
	var key graphs.SymbolKey
	if isUniverse {
		key = graphs.NewUniverseSymbolKey(typeName)
	} else {
		key = graphs.NewNonUniverseBuiltInSymbolKey(typeName)
	}

	if node, exists := g.nodes[key.BaseId()]; exists {
		return node
	}

	node := &SymbolNode{
		Id:   key,
		Kind: kind,
	}
	g.nodes[key.BaseId()] = node

	return node
}

// idempotencyGuard checks if the given node with the given version exists in the graph.
// If the node exists but has a different FVersion, the old node will be evicted, alongside its dependents.
func (g *SymbolGraph) idempotencyGuard(decl ast.Node, version *gast.FileVersion) (*SymbolNode, graphs.SymbolKey, error) {
	if decl == nil {
		return nil, graphs.SymbolKey{}, fmt.Errorf("idempotencyGuard received a nil decl parameter")
	}

	if version == nil {
		return nil, graphs.SymbolKey{}, fmt.Errorf("idempotencyGuard received a nil version parameter")
	}

	key := graphs.NewSymbolKey(decl, version)
	if existing := g.nodes[key.BaseId()]; existing != nil {
		if existing.Version.Equals(version) {
			return existing, key, nil
		}
		g.evict(existing.Id)
	}

	return nil, key, nil
}

// evict removes the given node *and* recursively evicts any nodes that depend on it.
func (g *SymbolGraph) evict(key graphs.SymbolKey) {
	baseId := key.BaseId()

	// if already gone, stop
	if _, exists := g.nodes[baseId]; !exists {
		return
	}

	// first, recursively evict all nodes that rev-depend on me
	if revs, ok := g.revDeps[baseId]; ok {
		for fromKey := range revs {
			g.evict(fromKey)
		}
	}

	// remove all outgoing edges
	delete(g.deps, baseId)

	// remove all reverse edges pointing to me
	delete(g.revDeps, baseId)

	// delete the lookup key
	delete(g.lookupKeys, baseId)

	// finally remove node itself
	delete(g.nodes, baseId)
}

func (g *SymbolGraph) String() string {
	var sb strings.Builder

	// Summary
	sb.WriteString("=== SymbolGraph Dump ===\n")
	sb.WriteString(fmt.Sprintf("Total nodes: %d\n\n", len(g.nodes)))

	// Per-node details
	for key, node := range g.nodes {
		prettyKey := g.lookupKeys[key].PrettyPrint()
		sb.WriteString(fmt.Sprintf("[%s] %s\n", node.Kind, prettyKey))
		// Outgoing dependencies
		if deps, ok := g.deps[key]; ok && len(deps) > 0 {
			sb.WriteString("  Dependencies:\n")
			for _, edge := range g.edges[key] {
				toNode := g.nodes[edge.To.BaseId()]
				linkedPrettyKey := edge.To.PrettyPrint()

				if toNode == nil {
					sb.WriteString(fmt.Sprintf("    • [%s] (%s)\n", linkedPrettyKey, edge.Kind))
				} else {
					sb.WriteString(fmt.Sprintf("    • [%s] %s (%s)\n", toNode.Kind, linkedPrettyKey, edge.Kind))
				}
			}

		} else {
			sb.WriteString("  Dependencies: (none)\n")
		}

		// Incoming (reverse) dependencies
		if revs, ok := g.revDeps[key]; ok && len(revs) > 0 {
			sb.WriteString("  Dependents:\n")
			for fromKey := range revs {
				fromNode := g.nodes[fromKey.BaseId()]
				linkedPrettyKey := fromKey.PrettyPrint()
				sb.WriteString(fmt.Sprintf("    • [%s] %s\n", fromNode.Kind, linkedPrettyKey))
			}
		} else {
			sb.WriteString("  Dependents: (none)\n")
		}

		sb.WriteString("\n")
	}

	sb.WriteString("=== End SymbolGraph ===\n")
	return sb.String()
}

func (g *SymbolGraph) ToDot(theme *dot.DotTheme) string {
	builder := dot.NewDotBuilder(theme)

	// Add all nodes
	for key, node := range g.nodes {
		symKey := g.lookupKeys[key]
		label := symKey.ShortLabel()
		builder.AddNode(symKey, node.Kind, label)
	}

	// Add all edges
	for fromKey, edges := range g.edges {
		for _, edge := range edges {
			builder.AddEdge(g.lookupKeys[fromKey], edge.To, string(edge.Kind))
		}
	}

	// Add legend if theme requests it
	builder.RenderLegend()

	return builder.Finish()
}

func getTypeRef(typeUsage metadata.TypeUsageMeta) (graphs.SymbolKey, error) {
	if !typeUsage.SymbolKind.IsBuiltin() {

		return typeUsage.GetBaseTypeRefKey()
	}

	if typeUsage.IsUniverseType() {
		return graphs.NewUniverseSymbolKey(typeUsage.Name), nil
	}

	return graphs.NewNonUniverseBuiltInSymbolKey(fmt.Sprintf("%s.%s", typeUsage.PkgPath, typeUsage.Name)), nil
}
