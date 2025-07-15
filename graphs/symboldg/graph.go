package symboldg

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
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
	AddEdge(from, to graphs.SymbolKey, kind SymbolEdgeKind, meta map[string]string)
}

type SymbolGraph struct {
	nodes map[graphs.SymbolKey]*SymbolNode // keyed by ast node

	edges map[graphs.SymbolKey][]SymbolEdge // Node relations

	deps    map[graphs.SymbolKey]map[graphs.SymbolKey]struct{} // from → set of to
	revDeps map[graphs.SymbolKey]map[graphs.SymbolKey]struct{} // to   → set of from
}

func NewSymbolGraph() SymbolGraph {
	return SymbolGraph{
		edges:   make(map[graphs.SymbolKey][]SymbolEdge),
		nodes:   make(map[graphs.SymbolKey]*SymbolNode),
		deps:    make(map[graphs.SymbolKey]map[graphs.SymbolKey]struct{}),
		revDeps: make(map[graphs.SymbolKey]map[graphs.SymbolKey]struct{}),
	}
}

func (g *SymbolGraph) addNode(n *SymbolNode) {
	g.nodes[n.Id] = n
}

func (g *SymbolGraph) AddPrimitive(p PrimitiveType) *SymbolNode {
	key := graphs.SymbolKeyForUniverseType(string(p))
	if node, exists := g.nodes[key]; exists {
		return node
	}

	node := &SymbolNode{
		Id:   key,
		Kind: common.SymKindBuiltin,
	}
	g.nodes[key] = node

	return node
}

// AddEdge adds a semantic relationship FROM → TO.
// For example: AddEdge(structKey, receiverKey, EdgeKindReceiver, nil)
// means "struct has receiver".
func (g *SymbolGraph) AddEdge(from, to graphs.SymbolKey, kind SymbolEdgeKind, meta map[string]string) {
	existingEdges, exist := g.edges[from]
	if exist {
		for _, edge := range existingEdges {
			if edge.To == to {
				panic(fmt.Sprintf("duplicate edge insertion detected:\nFrom: %s\nTo: %s", from, to))
			}
		}
	}
	// Maintain semantically-typed edge
	g.edges[from] = append(g.edges[from], SymbolEdge{
		To:       to,
		Kind:     kind,
		Metadata: meta,
	})

	if g.deps[from] == nil {
		g.deps[from] = make(map[graphs.SymbolKey]struct{})
	}
	g.deps[from][to] = struct{}{}

	if g.revDeps[to] == nil {
		g.revDeps[to] = make(map[graphs.SymbolKey]struct{})
	}
	g.revDeps[to][from] = struct{}{}
}

func (g *SymbolGraph) AddController(request CreateControllerNode) (*SymbolNode, error) {
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

	// Add a dependency FROM the PARAM TO the PARAM TYPE node
	g.AddEdge(symNode.Id, request.Data.Type.TypeRefKey, EdgeKindReference, nil)

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

	// Add a dependency FROM the RETVAL TO the RETVAL TYPE node
	g.AddEdge(symNode.Id, request.Data.Type.TypeRefKey, EdgeKindReference, nil)

	return symNode, nil
}

func (g *SymbolGraph) AddStruct(request CreateStructNode) (*SymbolNode, error) {
	return g.createAndAddSymNode(
		request.Data.Node,
		common.SymKindStruct,
		request.Data.FVersion,
		request.Annotations,
		request.Data,
	)
}

func (g *SymbolGraph) AddEnum(request CreateEnumNode) (*SymbolNode, error) {
	return g.createAndAddSymNode(
		request.Data.Node,
		common.SymKindEnum,
		request.Data.FVersion,
		request.Annotations,
		request.Data,
	)
}

func (g *SymbolGraph) AddField(request CreateFieldNode) (*SymbolNode, error) {
	return g.createAndAddSymNode(
		request.Data.Node,
		common.SymKindField,
		request.Data.FVersion,
		request.Annotations,
		request.Data,
	)
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

// idempotencyGuard checks if the given node with the given version exists in the graph.
// If the node exists but has a different FVersion, the old node will be evicted, alongside its dependents.
func (g *SymbolGraph) idempotencyGuard(decl ast.Node, version *gast.FileVersion) (*SymbolNode, graphs.SymbolKey, error) {
	if decl == nil {
		return nil, "", fmt.Errorf("idempotencyGuard received a nil decl parameter")
	}

	if version == nil {
		return nil, "", fmt.Errorf("idempotencyGuard received a nil version parameter")
	}

	key := graphs.SymbolKeyFor(decl, version)
	if existing := g.nodes[key]; existing != nil {
		if existing.Version.Equals(version) {
			return existing, key, nil
		}
		g.evict(existing.Id)
	}

	return nil, key, nil
}

// evict removes the given node *and* recursively evicts any nodes that depend on it.
func (g *SymbolGraph) evict(key graphs.SymbolKey) {
	// if already gone, stop
	if _, exists := g.nodes[key]; !exists {
		return
	}

	// first, recursively evict all nodes that rev-depend on me
	if revs, ok := g.revDeps[key]; ok {
		for fromKey := range revs {
			g.evict(fromKey)
		}
	}

	// remove all outgoing edges
	delete(g.deps, key)
	// remove all reverse edges pointing to me
	delete(g.revDeps, key)

	// finally remove node itself
	delete(g.nodes, key)
}

func (g *SymbolGraph) Dump() string {
	var sb strings.Builder

	// Summary
	sb.WriteString("=== SymbolGraph Dump ===\n")
	sb.WriteString(fmt.Sprintf("Total nodes: %d\n\n", len(g.nodes)))

	// Per-node details
	for key, node := range g.nodes {
		prettyKey := PrettyPrintSymbolKey(key)
		sb.WriteString(fmt.Sprintf("[%s] %s\n", node.Kind, prettyKey))
		// Outgoing dependencies
		if deps, ok := g.deps[key]; ok && len(deps) > 0 {
			sb.WriteString("  Dependencies:\n")
			for _, edge := range g.edges[key] {
				toNode := g.nodes[edge.To]
				linkedPrettyKey := PrettyPrintSymbolKey(edge.To)

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
				fromNode := g.nodes[fromKey]
				linkedPrettyKey := PrettyPrintSymbolKey(fromKey)
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

func (g *SymbolGraph) ToDot() string {
	ERROR_NODE_ID := "N_ERROR"

	var sb strings.Builder
	sb.WriteString("digraph SymbolGraph {\n")

	// Map full SymbolKey to short ID, e.g. N0, N1...
	idMap := make(map[graphs.SymbolKey]string)
	counter := 0

	// Assign short IDs
	for key := range g.nodes {
		idMap[key] = fmt.Sprintf("N%d", counter)
		counter++
	}

	// Write nodes
	for key, node := range g.nodes {
		id := idMap[key]
		label := key.ShortLabel() // short label: e.g. "Field@graph.controller.go"
		sb.WriteString(fmt.Sprintf("  %s [label=\"%s (%s)\"];\n", id, label, node.Kind))
	}

	// Write edges
	for fromKey, edges := range g.edges {
		fromID := idMap[fromKey]
		for _, edge := range edges {
			toID, ok := idMap[edge.To]
			if !ok || toID == "" {
				toID = ERROR_NODE_ID
			}
			sb.WriteString(fmt.Sprintf("  %s -> %s [label=\"%s\"];\n", fromID, toID, edge.Kind))
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}

func PrettyPrintSymbolKey(key graphs.SymbolKey) string {
	withoutPrefix, hasPrefix := strings.CutPrefix(string(key), graphs.UniverseTypeSymKeyPrefix)
	if hasPrefix {
		return withoutPrefix
	}

	keyParts := strings.Split(string(key), "@")
	// Expected 3-length
	fVerParts := strings.Split(keyParts[1], "|")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s\n", keyParts[0])) // Node Name

	for _, part := range fVerParts {
		sb.WriteString(fmt.Sprintf("    • %s\n", part))
	}

	return sb.String()
}
