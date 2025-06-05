package symboldag

import (
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/annotations"
)

type SymbolGraph struct {
	nodes map[any]*SymbolNode // keyed by ast node
}

type SymbolNode struct {
	Id          any // Key, like *ast.FuncDecl, *ast.TypeSpec, *ast.Ident
	Kind        definitions.SymKind
	Version     *definitions.FileVersion
	Value       any // Actual metadata: RouteMetadata, TypeMetadata, etc.
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
	ctrlNode := &SymbolNode{
		Id:          request.Decl,
		Kind:        definitions.SymKindStruct,
		Version:     request.Data.FVersion,
		Value:       request.Data,
		Annotations: request.Annotations,
		Deps:        map[*SymbolNode]struct{}{},
		RevDeps:     map[*SymbolNode]struct{}{},
	}

	for _, route := range request.Data.Routes {
		depNode := g.addRoute(&route, ctrlNode)
		ctrlNode.Deps[depNode] = struct{}{} // Add the route as a dependency for the controller
	}

	g.addNode(ctrlNode)
	return ctrlNode
}

func (g *SymbolGraph) addRoute(route *definitions.RouteMetadata, ctrlNode *SymbolNode) *SymbolNode {
	routeNode := NewRouteNode(route.Decl, route)
	g.addNode(routeNode)
	g.AddDep(ctrlNode, routeNode)

	for _, param := range route.FuncParams {
		paramNode := NewParamNode(param.Ident, param)
		g.addNode(paramNode)
		g.AddDep(routeNode, paramNode)

		if typeNode := g.AddTypeRecursive(param.TypeMeta); typeNode != nil {
			g.AddDep(paramNode, typeNode)
		}
	}

	for _, ret := range route.Responses {
		retNode := NewRetValNode(ret.Ident, ret)
		g.addNode(retNode)
		g.AddDep(routeNode, retNode)

		if typeNode := g.AddTypeRecursive(ret.TypeMetadata); typeNode != nil {
			g.AddDep(retNode, typeNode)
		}
	}

	return routeNode
}

func (g *SymbolGraph) AddTypeRecursive(meta *definitions.TypeMetadata) *SymbolNode {
	if meta == nil || meta.Decl == nil {
		return nil
	}

	if existing, ok := g.nodes[meta.Decl]; ok {
		return existing
	}

	typeNode := NewTypeNode(meta.Decl, meta)
	g.addNode(typeNode)

	for _, field := range meta.Fields {
		if field.TypeMeta != nil {
			if dep := g.AddTypeRecursive(field.TypeMeta); dep != nil {
				g.AddDep(typeNode, dep)
			}
		}
	}

	return typeNode
}
