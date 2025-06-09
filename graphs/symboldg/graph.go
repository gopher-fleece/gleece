package symboldg

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/gast"
)

type SymbolGraphBuilder interface {
	AddController(request CreateControllerNode) (*SymbolNode, error)
	AddRoute(request CreateRouteNode) (*SymbolNode, error)
	AddRouteParam(request CreateParameterNode) (*SymbolNode, error)
	AddRouteRetVal(request CreateReturnValueNode) (*SymbolNode, error)
	AddType(request CreateTypeNode) (*SymbolNode, error)
}

type SymbolGraph struct {
	nodes map[SymbolKey]*SymbolNode // keyed by ast node

	deps    map[SymbolKey]map[SymbolKey]struct{} // from → set of to
	revDeps map[SymbolKey]map[SymbolKey]struct{} // to   → set of from
}

func NewSymbolGraph() SymbolGraph {
	return SymbolGraph{
		nodes:   make(map[SymbolKey]*SymbolNode),
		deps:    make(map[SymbolKey]map[SymbolKey]struct{}),
		revDeps: make(map[SymbolKey]map[SymbolKey]struct{}),
	}
}

type SymbolKey string

func (g *SymbolGraph) addNode(n *SymbolNode) {
	g.nodes[n.Id] = n
}

func (g *SymbolGraph) AddDep(from, to SymbolKey) {
	if g.deps[from] == nil {
		g.deps[from] = make(map[SymbolKey]struct{})
	}
	g.deps[from][to] = struct{}{}

	if g.revDeps[to] == nil {
		g.revDeps[to] = make(map[SymbolKey]struct{})
	}
	g.revDeps[to][from] = struct{}{}
}

func (g *SymbolGraph) AddController(request CreateControllerNode) (*SymbolNode, error) {
	existing, err := g.idempotencyGuard(request.Decl, &request.Data.FVersion)
	if existing != nil || err != nil {
		return existing, err
	}

	ctrlNode := &SymbolNode{
		Id:          SymbolKeyFor(request.Decl, &request.Data.FVersion),
		Kind:        common.SymKindStruct,
		Version:     &request.Data.FVersion,
		Data:        request.Data,
		Annotations: request.Annotations,
	}
	g.addNode(ctrlNode)
	return ctrlNode, nil
}

func (g *SymbolGraph) AddRoute(request CreateRouteNode) (*SymbolNode, error) {
	existing, err := g.idempotencyGuard(request.Decl, request.Data.FVersion)
	if existing != nil || err != nil {
		return existing, err
	}

	routeNode := &SymbolNode{
		Id:          SymbolKeyFor(request.Decl, request.Data.FVersion),
		Kind:        common.SymKindFunction,
		Version:     request.Data.FVersion,
		Data:        request.Data,
		Annotations: request.Annotations,
	}

	g.addNode(routeNode)
	g.AddDep(request.ParentController.SymbolKey(), routeNode.Id)
	return routeNode, nil
}

func (g *SymbolGraph) AddRouteParam(request CreateParameterNode) (*SymbolNode, error) {
	existing, err := g.idempotencyGuard(request.Decl, &request.ParentRoute.FVersion)
	if existing != nil || err != nil {
		return existing, err
	}

	paramNode := &SymbolNode{
		Id:      SymbolKeyFor(request.Decl, &request.ParentRoute.FVersion),
		Kind:    common.SymKindParameter,
		Version: &request.ParentRoute.FVersion,
		Data: &FuncParamSymbolicMetadata{
			OrderedIdent:       request.Data.OrderedIdent,
			Name:               request.Data.Name,
			IsContext:          request.Data.IsContext,
			PassedIn:           request.Data.PassedIn,
			NameInSchema:       request.Data.NameInSchema,
			Description:        request.Data.Description,
			UniqueImportSerial: request.Data.UniqueImportSerial,
			Validator:          request.Data.Validator,
			Deprecation:        request.Data.Deprecation,
		},
	}

	g.addNode(paramNode)
	g.AddDep(request.ParentRoute.SymbolKey(), paramNode.Id)

	g.AddType(CreateTypeNode{
		Data:        request.Data.TypeMetadataWithAst,
		Annotations: request.Data.Annotations,
	})

	return paramNode, nil
}

func (g *SymbolGraph) AddRouteRetVal(request CreateReturnValueNode) (*SymbolNode, error) {
	existing, err := g.idempotencyGuard(request.Decl, &request.ParentRoute.FVersion)
	if existing != nil || err != nil {
		return existing, err
	}

	retValNode := &SymbolNode{
		Id:      SymbolKeyFor(request.Decl, request.Data.FVersion),
		Kind:    common.SymKindVariable,
		Version: &request.ParentRoute.FVersion,
		Data:    request,
	}
	g.addNode(retValNode)
	g.AddDep(request.ParentRoute.SymbolKey(), retValNode.Id)
	return retValNode, nil
}

func (g *SymbolGraph) AddType(request CreateTypeNode) (*SymbolNode, error) {
	existing, err := g.idempotencyGuard(request.Data.TypeExpr, request.Data.FVersion)
	if existing != nil || err != nil {
		return existing, err
	}

	node := &SymbolNode{
		Id:          SymbolKeyFor(request.Data.TypeExpr, request.Data.FVersion),
		Kind:        request.Data.SymbolKind,
		Version:     request.Data.FVersion,
		Data:        request.Data,
		Annotations: request.Annotations,
	}

	g.addNode(node)
	return node, nil
}

// idempotencyGuard checks if the given node with the given version exists in the graph.
// If the node exists but has a different FVersion, the old node will be evicted, alongside its dependents.
func (g *SymbolGraph) idempotencyGuard(decl any, version *gast.FileVersion) (*SymbolNode, error) {
	if decl == nil {
		return nil, fmt.Errorf("idempotencyGuard received a nil decl parameter")
	}

	if version == nil {
		return nil, fmt.Errorf("idempotencyGuard received a nil version parameter")
	}

	key := SymbolKeyFor(decl, version)
	if existing := g.nodes[key]; existing != nil {
		if existing.Version.Equals(version) {
			return existing, nil
		}
		g.evict(existing.Id)
	}

	return nil, nil
}

// evict removes the given node *and* recursively evicts any nodes that depend on it.
func (g *SymbolGraph) evict(key SymbolKey) {
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
			for toKey := range deps {
				toNode := g.nodes[toKey]
				linkedPrettyKey := PrettyPrintSymbolKey(toKey)
				sb.WriteString(fmt.Sprintf("    • [%s] %s\n", toNode.Kind, linkedPrettyKey))
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

func SymbolKeyFor(decl any, version *gast.FileVersion) SymbolKey {
	switch node := decl.(type) {
	case *ast.FuncDecl:
		return SymbolKey(fmt.Sprintf("Func:%s@%s", node.Name.Name, version.String()))
	case *ast.TypeSpec:
		return SymbolKey(fmt.Sprintf("Type:%s@%s", node.Name.Name, version.String()))
	case *ast.Field:
		// Field has no name always (e.g., return values), fallback to position
		return SymbolKey(fmt.Sprintf("Field@%d:%s", node.Pos(), version.String()))
	case *ast.Ident:
		return SymbolKey(fmt.Sprintf("Ident:%s@%s", node.Name, version.String()))
	case ast.Expr:
		return SymbolKey(fmt.Sprintf("Expr:%T@%d:%s", node, node.Pos(), version.String()))
	case nil:
		return SymbolKey("nil")
	default:
		return SymbolKey(fmt.Sprintf("%T@%p", node, node)) // fallback on pointer identity
	}
}

func PrettyPrintSymbolKey(key SymbolKey) string {
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
