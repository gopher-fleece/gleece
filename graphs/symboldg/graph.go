package symboldg

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/gast"
)

type SymbolGraphBuilder interface {
	AddController(request CreateControllerNode) *SymbolNode
	AddRoute(request CreateRouteNode) *SymbolNode
	AddRouteParam(request CreateParameterNode) *SymbolNode
	AddRouteRetVal(request CreateReturnValueNode) *SymbolNode
	AddType(request CreateTypeNode) *SymbolNode
}

type SymbolGraph struct {
	nodes map[SymbolKey]*SymbolNode // keyed by ast node
}

func (g *SymbolGraph) NewSymbolGraph() {
	g.nodes = make(map[SymbolKey]*SymbolNode)
}

type SymbolKey string

type SymbolNode struct {
	Id          SymbolKey
	Kind        common.SymKind
	Version     *gast.FileVersion
	Data        any // Actual metadata: RouteMetadata, TypeMetadata, etc.
	Annotations *annotations.AnnotationHolder

	Deps    map[*SymbolNode]struct{} // Points to symbols this one depends on
	RevDeps map[*SymbolNode]struct{} // Who depends on this symbol
}

func (g *SymbolGraph) addNode(n *SymbolNode) {
	g.nodes[n.Id] = n
}

func (g *SymbolGraph) AddDep(from, to *SymbolNode) {
	from.Deps[to] = struct{}{}
	to.RevDeps[from] = struct{}{}
}

func (g *SymbolGraph) AddController(request CreateControllerNode) *SymbolNode {
	g.idempotencyGuard(request.Decl, request.Data.FVersion)

	ctrlNode := &SymbolNode{
		Id:          SymbolKeyFor(request.Decl, request.Data.FVersion),
		Kind:        common.SymKindStruct,
		Version:     request.Data.FVersion,
		Data:        request.Data,
		Annotations: request.Annotations,
		Deps:        make(map[*SymbolNode]struct{}),
		RevDeps:     make(map[*SymbolNode]struct{}),
	}
	g.addNode(ctrlNode)
	return ctrlNode
}

func (g *SymbolGraph) AddRoute(request CreateRouteNode) *SymbolNode {
	g.idempotencyGuard(request.Decl, request.Data.FVersion)

	routeNode := &SymbolNode{
		Id:          SymbolKeyFor(request.Decl, request.Data.FVersion),
		Kind:        common.SymKindFunction,
		Version:     request.Data.FVersion,
		Data:        request.Data,
		Annotations: request.Annotations,
		Deps:        make(map[*SymbolNode]struct{}),
		RevDeps:     make(map[*SymbolNode]struct{}),
	}

	g.addNode(routeNode)
	g.AddDep(request.ParentController, routeNode)
	return routeNode
}

func (g *SymbolGraph) AddRouteParam(request CreateParameterNode) *SymbolNode {
	g.idempotencyGuard(request.Decl, request.ParentRoute.Version)

	paramNode := &SymbolNode{
		Id:      SymbolKeyFor(request.Decl, request.Data.FVersion),
		Kind:    common.SymKindParameter,
		Version: request.ParentRoute.Version,
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
		Annotations: request.Annotations,
		Deps:        make(map[*SymbolNode]struct{}),
		RevDeps:     make(map[*SymbolNode]struct{}),
	}

	g.addNode(paramNode)
	g.AddDep(request.ParentRoute, paramNode)

	g.AddType(CreateTypeNode{
		Data:        request.Data.TypeMetadataWithAst,
		Annotations: request.Data.Annotations,
	})

	return paramNode
}

func (g *SymbolGraph) AddRouteRetVal(request CreateReturnValueNode) *SymbolNode {
	g.idempotencyGuard(request.Decl, request.ParentRoute.Version)

	retValNode := &SymbolNode{
		Id:          SymbolKeyFor(request.Decl, request.Data.FVersion),
		Kind:        common.SymKindVariable,
		Version:     request.ParentRoute.Version,
		Data:        request,
		Annotations: request.Annotations,
		Deps:        make(map[*SymbolNode]struct{}),
		RevDeps:     make(map[*SymbolNode]struct{}),
	}
	g.addNode(retValNode)
	g.AddDep(request.ParentRoute, retValNode)
	return retValNode
}

func (g *SymbolGraph) AddType(request CreateTypeNode) *SymbolNode {
	g.idempotencyGuard(request.Data.Expr, request.Data.FVersion)

	node := &SymbolNode{
		Id:          SymbolKeyFor(request.Data.Expr, request.Data.FVersion),
		Kind:        request.Data.SymbolKind,
		Version:     request.Data.FVersion,
		Data:        request.Data,
		Annotations: request.Annotations,
		Deps:        make(map[*SymbolNode]struct{}),
		RevDeps:     make(map[*SymbolNode]struct{}),
	}

	g.addNode(node)
	return node
}

// idempotencyGuard checks if the given node with the given version exists in the graph.
// If the node exists but has a different FVersion, the old node will be evicted, alongside its dependents.
func (g *SymbolGraph) idempotencyGuard(decl any, version *gast.FileVersion) *SymbolNode {
	key := SymbolKeyFor(decl, version)
	if existing := g.nodes[key]; existing != nil {
		if existing.Version.Equals(version) {
			return existing
		}
		g.evictNode(existing)
	}
	return nil
}

func (g *SymbolGraph) Dump() string {
	var sb strings.Builder
	sb.WriteString("=== SymbolGraph Dump ===\n")

	for _, node := range g.nodes {
		// Basic node info
		sb.WriteString(fmt.Sprintf("- [%s] %T | Version: %s\n", node.Kind, node.Id, node.Version))

		// Dependencies
		if len(node.Deps) > 0 {
			sb.WriteString("    -> Deps:\n")
			for dep := range node.Deps {
				sb.WriteString(fmt.Sprintf("        - [%s] %T | Version: %s\n", dep.Kind, dep.Id, dep.Version))
			}
		}

		// Reverse dependencies (who depends on this)
		if len(node.RevDeps) > 0 {
			sb.WriteString("    <- RevDeps:\n")
			for rev := range node.RevDeps {
				sb.WriteString(fmt.Sprintf("        - [%s] %T | Version: %s\n", rev.Kind, rev.Id, rev.Version))
			}
		}
	}

	sb.WriteString("=== End SymbolGraph ===\n")
	return sb.String()
}

func (g *SymbolGraph) evictNode(n *SymbolNode) {
	delete(g.nodes, n.Id)
	for dep := range n.Deps {
		delete(dep.RevDeps, n)
	}
	for rev := range n.RevDeps {
		delete(rev.Deps, n)
	}
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
