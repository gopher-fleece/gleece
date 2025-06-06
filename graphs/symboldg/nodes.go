package symboldag

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/gast"
)

func NewControllerNode(id *ast.GenDecl, version *gast.FileVersion, holder *annotations.AnnotationHolder) *SymbolNode {
	return &SymbolNode{
		Id:      id,
		Kind:    common.SymKindPackage, // Or maybe SymKindUnknown if not exposed?
		Version: version,
		Value:   holder,
		Deps:    make(map[*SymbolNode]struct{}),
		RevDeps: make(map[*SymbolNode]struct{}),
	}
}

func NewRouteNode(id *ast.FuncDecl, meta *definitions.RouteMetadata) *SymbolNode {
	return &SymbolNode{
		Id:      id,
		Kind:    common.SymKindFunction,
		Version: meta.FVersion,
		Value:   meta,
		Deps:    make(map[*SymbolNode]struct{}),
		RevDeps: make(map[*SymbolNode]struct{}),
	}
}

func NewParamNode(id *ast.Ident, param definitions.FuncParam) *SymbolNode {
	return &SymbolNode{
		Id:      id,
		Kind:    common.SymKindParameter,
		Version: param.TypeMeta.FVersion,
		Value:   param,
		Deps:    make(map[*SymbolNode]struct{}),
		RevDeps: make(map[*SymbolNode]struct{}),
	}
}

func NewRetValNode(id *ast.Ident, retVal definitions.FuncReturnValue) *SymbolNode {
	return &SymbolNode{
		Id:      id,
		Kind:    common.SymKindVariable, // or maybe SymKindField?
		Version: retVal.TypeMetadata.FVersion,
		Value:   retVal,
		Deps:    make(map[*SymbolNode]struct{}),
		RevDeps: make(map[*SymbolNode]struct{}),
	}
}

func NewTypeNode(id *ast.TypeSpec, meta *definitions.TypeMetadata) *SymbolNode {
	return &SymbolNode{
		Id:      id,
		Kind:    meta.SymbolKind,
		Version: meta.FVersion,
		Value:   meta,
		Deps:    make(map[*SymbolNode]struct{}),
		RevDeps: make(map[*SymbolNode]struct{}),
	}
}

func NewAliasNode(id *ast.TypeSpec, alias *definitions.AliasMetadata, version *gast.FileVersion) *SymbolNode {
	return &SymbolNode{
		Id:      id,
		Kind:    common.SymKindAlias,
		Version: version,
		Value:   alias,
		Deps:    make(map[*SymbolNode]struct{}),
		RevDeps: make(map[*SymbolNode]struct{}),
	}
}
