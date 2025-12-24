package symboldg

import (
	"cmp"
	"fmt"
	"go/ast"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/common/linq"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/metadata/typeref"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/dot"
)

type SymbolGraph struct {
	// A map for retrieving full symbol keys from their comparable, non-versioned IDs
	lookupKeys map[string]graphs.SymbolKey

	nodes map[string]*SymbolNode // keyed by ast node

	// A counter used to assign ordinals to new outgoing edges
	nextEdgeSeq uint32
	edges       map[string]map[string]SymbolEdgeDescriptor // Node relations

	deps    map[string]map[graphs.SymbolKey]struct{} // from -> set of to
	revDeps map[string]map[graphs.SymbolKey]struct{} // to   -> set of from
}

func NewSymbolGraph() SymbolGraph {
	return SymbolGraph{
		lookupKeys: map[string]graphs.SymbolKey{},
		edges:      make(map[string]map[string]SymbolEdgeDescriptor),
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

func (g *SymbolGraph) getAndIncrementNextEdgeOrdinal() uint32 {
	current := g.nextEdgeSeq
	g.nextEdgeSeq++
	return current
}

// AddEdge adds a semantic relationship FROM → TO.
// For example: AddEdge(structKey, receiverKey, EdgeKindReceiver, nil)
// means "struct has receiver".
func (g *SymbolGraph) AddEdge(from, to graphs.SymbolKey, kind SymbolEdgeKind, meta map[string]string) {
	fromBaseId := from.BaseId()
	toBaseId := to.BaseId()

	// ensure deps / revDeps
	if g.deps[fromBaseId] == nil {
		g.deps[fromBaseId] = make(map[graphs.SymbolKey]struct{})
	}
	g.deps[fromBaseId][to] = struct{}{}

	if g.revDeps[toBaseId] == nil {
		g.revDeps[toBaseId] = make(map[graphs.SymbolKey]struct{})
	}
	g.revDeps[toBaseId][from] = struct{}{}

	inner := g.edges[fromBaseId]
	if inner == nil {
		inner = make(map[string]SymbolEdgeDescriptor)
		g.edges[fromBaseId] = inner
	}

	k := edgeMapKey(kind, toBaseId)
	if _, exists := inner[k]; exists {
		return // duplicate (same from, to base, kind)
	}

	inner[k] = SymbolEdgeDescriptor{
		Edge:    SymbolEdge{From: from, To: to, Kind: kind, Metadata: meta},
		Ordinal: g.getAndIncrementNextEdgeOrdinal(),
	}
	g.lookupKeys[fromBaseId] = from
}

func (g *SymbolGraph) RemoveEdge(from, to graphs.SymbolKey, kind *SymbolEdgeKind) {
	fromBase := from.BaseId()
	toBase := to.BaseId()

	if inner, ok := g.edges[fromBase]; ok {
		if kind != nil {
			// Remove edges with the specified kind
			delete(inner, edgeMapKey(*kind, toBase))
		} else {
			// Remove all edges -> toBase entries
			suffix := "::" + toBase
			for k := range inner {
				if strings.HasSuffix(k, suffix) {
					delete(inner, k)
				}
			}
		}
		if len(inner) == 0 {
			delete(g.edges, fromBase)
		}
	}

	if depsMap, ok := g.deps[fromBase]; ok {
		delete(depsMap, to)
		if len(depsMap) == 0 {
			delete(g.deps, fromBase)
		}
	}

	if revMap, ok := g.revDeps[toBase]; ok {
		delete(revMap, from)
		if len(revMap) == 0 {
			delete(g.revDeps, toBase)
		}
	}
}

func (g *SymbolGraph) AddController(request CreateControllerNode) (*SymbolNode, error) {
	symNode, err := g.createAndAddSymNode(
		request.Data.Struct.Node,
		common.SymKindController,
		request.Data.Struct.FVersion,
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

	// Add a dependency FROM the parent controller TO the route.
	// Note - maybe add a check on ParentController validity, node wise? (FVersion and such)
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
		primitive, isPrimitive := common.ToPrimitiveType(string(request.Data.ValueKind))
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
		request.Data.SymbolKind,
		request.Data.FVersion,
		request.Annotations,
		request.Data,
	)
	if err != nil {
		return nil, err
	}

	typeRef, err := g.getKeyForUsage(request.Data.Type)
	if err != nil {
		return nil, err
	}

	edgeKind := EdgeKindType
	if request.Data.Type.Root.Kind() == metadata.TypeRefKindParam {
		edgeKind = EdgeKindTypeParameter
	}

	g.AddEdge(symNode.Id, typeRef, edgeKind, nil)
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

func (g *SymbolGraph) AddAlias(request CreateAliasNode) (*SymbolNode, error) {
	return g.createAndAddSymNode(
		request.Data.Node,
		common.SymKindAlias,
		request.Data.FVersion,
		request.Annotations,
		request.Data,
	)
}

func (g *SymbolGraph) Get(key graphs.SymbolKey) *SymbolNode {
	return g.nodes[key.BaseId()]
}

// GetEdges returns a map of edge descriptors involving 'key'.
// It includes both outgoing edges (key -> X) and incoming edges (Y -> key).
// If 'kinds' is non-empty, only edges whose kind matches one of the kinds are returned.
// Returned map keys are stable and unique: "<fromBaseId>::<innerEdgeKey>".
func (g *SymbolGraph) GetEdges(key graphs.SymbolKey, kinds []SymbolEdgeKind) map[string]SymbolEdgeDescriptor {
	mapKey := key.BaseId()
	out := make(map[string]SymbolEdgeDescriptor)

	requestedKinds := mapset.NewSet(kinds...)

	// 1) Outgoing edges from `key`
	if outgoing, ok := g.edges[mapKey]; ok {
		for innerKey, desc := range outgoing {
			if !requestedKinds.IsEmpty() && !requestedKinds.ContainsOne(desc.Edge.Kind) {
				continue
			}
			// Use fromBaseId prefix to guarantee uniqueness when aggregating from many sources
			out[mapKey+"::"+innerKey] = desc
		}
	}

	// 2) Incoming edges (parents -> key). Use revDeps to find parents.
	if parents, ok := g.revDeps[mapKey]; ok {
		for parentKey := range parents {
			parentBase := parentKey.BaseId()
			if parentEdges, ok := g.edges[parentBase]; ok {
				for innerKey, desc := range parentEdges {
					// we only want those edges whose 'To' is our key
					if desc.Edge.To.BaseId() != mapKey {
						continue
					}
					if !requestedKinds.IsEmpty() && !requestedKinds.ContainsOne(desc.Edge.Kind) {
						continue
					}
					out[parentBase+"::"+innerKey] = desc
				}
			}
		}
	}

	return out
}

func (g *SymbolGraph) Exists(key graphs.SymbolKey) bool {
	return g.Get(key) != nil
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

func (g *SymbolGraph) IsPrimitivePresent(primitive common.PrimitiveType) bool {
	return g.builtinSymbolExists(string(primitive), true)
}

func (g *SymbolGraph) AddPrimitive(p common.PrimitiveType) *SymbolNode {
	// Primitives are always 'universe' types
	return g.addBuiltinSymbol(string(p), common.SymKindBuiltin, true)
}

func (g *SymbolGraph) IsSpecialPresent(special common.SpecialType) bool {
	return g.builtinSymbolExists(string(special), special.IsUniverse())
}

func (g *SymbolGraph) AddSpecial(special common.SpecialType) *SymbolNode {
	return g.addBuiltinSymbol(string(special), common.SymKindSpecialBuiltin, special.IsUniverse())
}

// addComposite creates (or returns existing) a composite-type SymbolNode and wires edges to operands.
// It's idempotent.
func (g *SymbolGraph) addComposite(req CreateCompositeNode) (*SymbolNode, error) {
	// Return existing node if present
	if node, exists := g.nodes[req.Key.BaseId()]; exists {
		return node, nil
	}

	node := &SymbolNode{
		Id:   req.Key,
		Kind: common.SymKindComposite,
		Data: &metadata.CompositeMeta{
			Canonical: req.Canonical,
			Operands:  req.Operands,
		},
	}

	g.addNode(node)

	// Add edges from composite -> operand nodes (type-parameter edges)
	for _, op := range req.Operands {
		g.AddEdge(node.Id, op, EdgeKindTypeParameter, nil)
	}

	// If this composite is an instantiation of a declared base, create the instantiation edge.
	if !req.Base.Equals(graphs.SymbolKey{}) {
		// we expect the base to be present (ensureInstantiatedBasePresent enforces this).
		if g.Exists(req.Base) {
			g.AddEdge(node.Id, req.Base, EdgeKindInstantiates, nil)
		} else {
			// defensive: this should not happen if ensureInstantiatedBasePresent ran.
			return nil, fmt.Errorf("addComposite: base missing for instantiation: %s", req.Base.Id())
		}
	}

	return node, nil
}

func (g *SymbolGraph) Children(node *SymbolNode, behavior *TraversalBehavior) []*SymbolNode {
	if behavior != nil && behavior.Sorting != TraversalSortingNone {
		return g.childrenSorted(node, behavior)
	}

	return g.childrenUnsorted(node, behavior)
}

func (g *SymbolGraph) childrenUnsorted(node *SymbolNode, behavior *TraversalBehavior) []*SymbolNode {
	var result []*SymbolNode
	for _, edgeDescriptor := range g.edges[node.Id.BaseId()] { // iterate map values now
		if !shouldIncludeEdge(edgeDescriptor.Edge, behavior) {
			continue
		}
		child, ok := g.nodes[edgeDescriptor.Edge.To.BaseId()]
		if !ok || !shouldIncludeNode(child, behavior) {
			continue
		}
		result = append(result, child)
	}

	return result
}

func (g *SymbolGraph) childrenSorted(node *SymbolNode, behavior *TraversalBehavior) []*SymbolNode {
	var results []SymbolNodeWithOrdinal
	for _, edgeDescriptor := range g.edges[node.Id.BaseId()] { // iterate map values now
		if !shouldIncludeEdge(edgeDescriptor.Edge, behavior) {
			continue
		}
		child, ok := g.nodes[edgeDescriptor.Edge.To.BaseId()]
		if !ok || !shouldIncludeNode(child, behavior) {
			continue
		}
		results = append(results, SymbolNodeWithOrdinal{Node: child, Ordinal: edgeDescriptor.Ordinal})
	}

	slices.SortFunc(results, func(a, b SymbolNodeWithOrdinal) int {
		if behavior.Sorting == TraversalSortingOrdinalDesc {
			return cmp.Compare(b.Ordinal, a.Ordinal)
		}
		return cmp.Compare(a.Ordinal, b.Ordinal)
	})

	return linq.Map(results, func(v SymbolNodeWithOrdinal) *SymbolNode {
		return v.Node
	})
}

func (g *SymbolGraph) Parents(node *SymbolNode, behavior *TraversalBehavior) []*SymbolNode {
	if behavior != nil && behavior.Sorting != TraversalSortingNone {
		return g.parentsSorted(node, behavior)
	}

	return g.parentsUnsorted(node, behavior)
}

func (g *SymbolGraph) parentsUnsorted(node *SymbolNode, behavior *TraversalBehavior) []*SymbolNode {
	var result []*SymbolNode
	for parentKey := range g.revDeps[node.Id.BaseId()] {
		edges := g.edges[parentKey.BaseId()]
		for _, edgeDescriptor := range edges {
			if edgeDescriptor.Edge.To != node.Id {
				continue
			}
			if !shouldIncludeEdge(edgeDescriptor.Edge, behavior) {
				continue
			}
			parentNode := g.nodes[parentKey.BaseId()]
			if parentNode != nil && shouldIncludeNode(parentNode, behavior) {
				result = append(result, parentNode)
			}
		}
	}
	return result
}

func (g *SymbolGraph) parentsSorted(node *SymbolNode, behavior *TraversalBehavior) []*SymbolNode {
	var results []SymbolNodeWithOrdinal
	for parentKey := range g.revDeps[node.Id.BaseId()] {
		edges := g.edges[parentKey.BaseId()]
		for _, edgeDescriptor := range edges {
			if edgeDescriptor.Edge.To != node.Id {
				continue
			}
			if !shouldIncludeEdge(edgeDescriptor.Edge, behavior) {
				continue
			}
			parentNode := g.nodes[parentKey.BaseId()]
			if parentNode != nil && shouldIncludeNode(parentNode, behavior) {
				results = append(results, SymbolNodeWithOrdinal{Node: parentNode, Ordinal: edgeDescriptor.Ordinal})
			}
		}
	}

	slices.SortFunc(results, func(a, b SymbolNodeWithOrdinal) int {
		if behavior.Sorting == TraversalSortingOrdinalDesc {
			return cmp.Compare(b.Ordinal, a.Ordinal)
		}
		return cmp.Compare(a.Ordinal, b.Ordinal)
	})

	return linq.Map(results, func(v SymbolNodeWithOrdinal) *SymbolNode {
		return v.Node
	})
}

func (g *SymbolGraph) Descendants(root *SymbolNode, behavior *TraversalBehavior) []*SymbolNode {
	visited := make(map[graphs.SymbolKey]struct{})
	var result []*SymbolNode

	var walk func(*SymbolNode)
	walk = func(n *SymbolNode) {
		for _, child := range g.Children(n, behavior) {
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

func (g *SymbolGraph) Walk(node *SymbolNode, walker TreeWalker) {
	edges := g.edges[node.Id.BaseId()]
	for {
		next := walker(node, edges)
		if next == nil {
			break
		}
	}
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

	g.addNode(node)

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
		g.RemoveNode(existing.Id)
	}

	return nil, key, nil
}

// Evict removes the given node and conservatively evicts dependents that
// become orphaned as a result. It uses RemoveEdge to keep the graph indices
// (edges, deps, revDeps) consistent.
func (g *SymbolGraph) RemoveNode(key graphs.SymbolKey) {
	idToRemove := key.BaseId()

	// Short-circuit if node doesn't exist
	if _, exists := g.nodes[idToRemove]; !exists {
		return
	}

	// RemoveEdge mutates our internals state so we 'snapshot' dependents (nodes that had edges -> key)
	var dependents []graphs.SymbolKey
	if revs, ok := g.revDeps[idToRemove]; ok {
		for fromKey := range revs {
			dependents = append(dependents, fromKey)
		}
	}

	// For each dependent, remove the edge(s) from dependent -> key.
	// If the dependent becomes orphaned (no outgoing deps to existing nodes), evict it.
	for _, fromKey := range dependents {
		dependentBaseId := fromKey.BaseId()

		// Remove all edges from `fromKey` to `key` (all kinds).
		g.RemoveEdge(fromKey, key, nil)

		// Decide whether 'fromKey' is now orphaned:
		// it's an isOrphaned if it has no outgoing dependency to an existing node.
		isOrphaned := true
		if depsMap, ok := g.deps[dependentBaseId]; ok {
			for toKey := range depsMap {
				if _, exists := g.nodes[toKey.BaseId()]; exists {
					// There's an outbound dependency - not an orphan
					isOrphaned = false
					break
				}
			}
		}

		if isOrphaned {
			g.RemoveNode(fromKey)
		}
	}

	// As RemoveEdge actually mutates our state, we basically
	// 'snapshot' outgoingEdges edges from this node so we can remove them safely.
	var outgoingEdges []SymbolEdge
	if inner, ok := g.edges[idToRemove]; ok {
		outgoingEdges = make([]SymbolEdge, 0, len(inner))
		for _, edgeDescriptor := range inner {
			outgoingEdges = append(outgoingEdges, edgeDescriptor.Edge)
		}
	}

	// Now that we've a list of all outgoing edges, we can go ahead and remove them all
	for _, e := range outgoingEdges {
		// Remove the specific kind edge from key -> e.To - this is to accommodate
		// for the rather unusual scenario of two nodes having multiple, different-kind edges to each other
		g.RemoveEdge(key, e.To, &e.Kind)
	}

	// Finally clean up any remaining indices for this node and remove the node.
	delete(g.deps, idToRemove)
	delete(g.revDeps, idToRemove)
	delete(g.lookupKeys, idToRemove)
	delete(g.nodes, idToRemove)
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
			for _, edgeDescriptor := range g.edges[key] {
				toNode := g.nodes[edgeDescriptor.Edge.To.BaseId()]
				linkedPrettyKey := edgeDescriptor.Edge.To.PrettyPrint()

				if toNode == nil {
					sb.WriteString(fmt.Sprintf("    • [%s] (%s)\n", linkedPrettyKey, edgeDescriptor.Edge.Kind))
				} else {
					sb.WriteString(fmt.Sprintf("    • [%s] %s (%s)\n", toNode.Kind, linkedPrettyKey, edgeDescriptor.Edge.Kind))
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
		for _, edgeDescriptor := range edges {
			builder.AddEdge(
				g.lookupKeys[fromKey],
				edgeDescriptor.Edge.To,
				string(edgeDescriptor.Edge.Kind),
				g.getParamIndexesLabelSuffix(edgeDescriptor),
			)
		}
	}

	return builder.Finish()
}

func (g *SymbolGraph) getParamIndexesLabelSuffix(edgeDescriptor SymbolEdgeDescriptor) *string {
	if edgeDescriptor.Edge.Kind != EdgeKindTypeParameter {
		return nil
	}

	indices := g.getTypeParamIndex(edgeDescriptor.Edge.From, edgeDescriptor.Edge.To)
	if len(indices) <= 0 {
		return nil
	}

	indexStrings := linq.Map(indices, func(idx int) string {
		return fmt.Sprint(idx)
	})

	formattedIndices := fmt.Sprintf(
		" %s",
		strings.Join(indexStrings, ", "),
	)
	return &formattedIndices
}

func (g *SymbolGraph) getTypeParamIndex(from, to graphs.SymbolKey) []int {
	// Get the actual node that uses the 'to' key
	fromNode := g.nodes[from.BaseId()]
	if fromNode == nil {
		return nil
	}

	composite, isComposite := fromNode.Data.(*metadata.CompositeMeta)
	if !isComposite {
		// Shouldn't happen - we should get here only if we're inspecting a composite node
		return nil
	}

	paramIndices := []int{}

	for idx, typeParam := range composite.Operands {
		if to == typeParam {
			paramIndices = append(paramIndices, idx)
		}
	}

	if len(paramIndices) > 0 {
		return paramIndices
	}

	return nil
}

// getKeyForUsage is a thin wrapper used by callers (AddField, AddRouteRetVal, ...).
// It validates the input and delegates to ensureTypeNode.
func (g *SymbolGraph) getKeyForUsage(typeUsage metadata.TypeUsageMeta) (graphs.SymbolKey, error) {
	if typeUsage.Root == nil {
		return graphs.SymbolKey{}, fmt.Errorf("getKeyForUsage: TypeUsage '%s' missing Root TypeRef", typeUsage.Name)
	}
	return g.ensureTypeNode(typeUsage.Root, typeUsage.FVersion)
}

// ensureTypeNode ensures a graph node exists for the given metadata.TypeRef and returns its SymbolKey.
// - idempotent
// - creates primitive / special universe nodes when needed
// - creates composite nodes (ptr/slice/map/func/instantiation) and edges to operands
// - does NOT attempt automatic materialization of declared base types; it errors if those are missing
func (g *SymbolGraph) ensureTypeNode(root metadata.TypeRef, fileVersion *gast.FileVersion) (graphs.SymbolKey, error) {
	if root == nil {
		return graphs.SymbolKey{}, fmt.Errorf("ensureTypeNode: nil TypeRef")
	}

	// derive key for this usage
	key, err := root.ToSymKey(fileVersion)
	if err != nil {
		return graphs.SymbolKey{}, fmt.Errorf(
			"ensureTypeNode: ToSymKey failed for '%s': %w",
			root.CanonicalString(),
			err,
		)
	}

	// 1) Universe / builtin -> ensure primitive/special node and return that key immediately.
	if key.IsUniverse || key.IsBuiltIn {
		if err := g.conditionalEnsureBuiltInNode(key); err != nil {
			return graphs.SymbolKey{}, fmt.Errorf("ensureTypeNode: ensureUniverseNode failed '%s': %w",
				key.Name, err)
		}
		return key, nil
	}

	// 2) Named (plain) that points to a declared base: prefer the declared base node.
	//    This prevents creating a zero-operand composite node for plain namedRef types.
	namedRef, isNamedRef := root.(*typeref.NamedTypeRef)

	if isNamedRef && len(namedRef.TypeArgs) == 0 && !namedRef.Key.Empty() {
		// If declared base exists in graph - return it (preserve declared identity).
		if g.Exists(namedRef.Key) {
			return namedRef.Key, nil
		}
		// Declared base missing -> that's a policy error (we don't auto-materialize).
		return graphs.SymbolKey{}, fmt.Errorf(
			"ensureTypeNode: declared base not present for named type usage %s -> base %s",
			root.CanonicalString(),
			namedRef.Key.Id(),
		)
	}

	// 3) If node already exists for this exact key, return it.
	if g.Exists(key) {
		return key, nil
	}

	// 4) Type parameter nodes: produce a TypeParam node (idempotent helper).
	if root.Kind() == metadata.TypeRefKindParam {
		castRef := root.(*typeref.ParamTypeRef)
		tNode, err := g.conditionalEnsureTypeParamNode(key, castRef.Index)
		if err != nil {
			return graphs.SymbolKey{}, fmt.Errorf("ensureTypeNode: ensure type-param node failed: %w", err)
		}
		return tNode.Id, nil
	}

	// 5) For instantiated named types ensure the declared base exists (no auto-materialize).
	if err := g.conditionalEnsureInstantiatedBase(root); err != nil {
		return graphs.SymbolKey{}, err
	}

	// 6) Now ensure composite placement (recursively ensures operand nodes and creates composite).
	if err := g.conditionalEnsureComposite(root, fileVersion, key); err != nil {
		return graphs.SymbolKey{}, fmt.Errorf("ensureTypeNode: failed to ensure composite '%s' placement in graph - %w", key.Id(), err)
	}

	// final: key should now be present (addComposite is responsible for inserting node)
	return key, nil
}

func (g *SymbolGraph) conditionalEnsureComposite(
	root metadata.TypeRef,
	fileVersion *gast.FileVersion,
	rootKey graphs.SymbolKey,
) error {

	// Ensure operand nodes recursively
	operands := gatherOperandTypeRefs(root)
	operandKeys := make([]graphs.SymbolKey, 0, len(operands))
	for _, op := range operands {
		opKey, err := g.ensureTypeNode(op, fileVersion)
		if err != nil {
			return fmt.Errorf("ensureTypeNode: ensuring operand '%s' failed: %w",
				op.CanonicalString(),
				err,
			)
		}
		operandKeys = append(operandKeys, opKey)
	}

	// Create composite node and wire operands
	create := CreateCompositeNode{
		Key:       rootKey,
		Canonical: root.CanonicalString(),
		Operands:  operandKeys,
	}

	// attach declared base if root is NamedTypeRef with base key
	if named, ok := root.(*typeref.NamedTypeRef); ok {
		if !named.Key.Equals(graphs.SymbolKey{}) && len(named.TypeArgs) > 0 {
			create.Base = named.Key
		}
	}

	if _, err := g.addComposite(create); err != nil {
		return fmt.Errorf("ensureComposite: AddComposite failed for '%s': %w", rootKey.Id(), err)
	}

	return nil
}

// conditionalEnsureInstantiatedBase checks that for NamedTypeRef instantiations the base exists.
// It does NOT attempt to auto-materialize declared bases; it returns an error when a declared base is absent.
func (g *SymbolGraph) conditionalEnsureInstantiatedBase(root metadata.TypeRef) error {
	// We only care about instantiated NamedTypeRef with a declared base key.
	named, ok := root.(*typeref.NamedTypeRef)
	if !ok {
		return nil
	}

	// If there's no base key or no type args -> nothing to check.
	if named.Key.Equals(graphs.SymbolKey{}) || len(named.TypeArgs) == 0 {
		return nil
	}

	// Universe/builtin bases are always fine.
	if named.Key.IsUniverse || named.Key.IsBuiltIn {
		return nil
	}

	// Declared base must already exist in graph (no auto materialize policy).
	if !g.Exists(named.Key) {
		return fmt.Errorf("ensureInstantiatedBasePresent: declared base not present for instantiation: %s", named.Key.Id())
	}

	return nil
}

// conditionalEnsureBuiltInNode creates primitives/special nodes (idempotent).
func (g *SymbolGraph) conditionalEnsureBuiltInNode(key graphs.SymbolKey) error {
	if prim, ok := common.ToPrimitiveType(key.Name); ok {
		g.AddPrimitive(prim)
		return nil
	}
	if sp, ok := common.ToSpecialType(key.Name); ok {
		g.AddSpecial(sp)
		return nil
	}
	return fmt.Errorf("ensureUniverseNode: unknown universe type '%s'", key.Name)
}

func (g *SymbolGraph) conditionalEnsureTypeParamNode(
	key graphs.SymbolKey,
	index int,
) (*SymbolNode, error) {
	if existing := g.Get(key); existing != nil {
		return existing, nil
	}

	node := &SymbolNode{
		Id:   key,
		Kind: common.SymKindTypeParam,
		Data: metadata.TypeParamDeclMeta{
			Name:  key.Name,
			Index: index,
		},
	}

	g.addNode(node)
	return node, nil
}

// gatherOperandTypeRefs returns deterministic operand TypeRefs for a composite root.
func gatherOperandTypeRefs(root metadata.TypeRef) []metadata.TypeRef {
	switch t := root.(type) {
	case *typeref.PtrTypeRef:
		return []metadata.TypeRef{t.Elem}
	case *typeref.SliceTypeRef:
		return []metadata.TypeRef{t.Elem}
	case *typeref.ArrayTypeRef:
		return []metadata.TypeRef{t.Elem}
	case *typeref.MapTypeRef:
		return []metadata.TypeRef{t.Key, t.Value}
	case *typeref.FuncTypeRef:
		out := make([]metadata.TypeRef, 0, len(t.Params)+len(t.Results))
		out = append(out, t.Params...)
		out = append(out, t.Results...)
		return out
	case *typeref.NamedTypeRef:
		// For instantiation, the explicit TypeArgs are operands (base handled separately)
		return t.TypeArgs
	case *typeref.InlineStructTypeRef:
		out := make([]metadata.TypeRef, 0, len(t.Fields))
		for _, f := range t.Fields {
			if f.Type.Root != nil {
				out = append(out, f.Type.Root)
			}
		}
		return out
	default:
		return nil
	}
}

func edgeMapKey(kind SymbolEdgeKind, toBaseId string) string {
	return string(kind) + "::" + toBaseId
}
