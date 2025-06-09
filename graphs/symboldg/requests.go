package symboldg

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/extractor/arbitrators"
	"github.com/gopher-fleece/gleece/gast"
)

type KeyableNodeMeta struct {
	// Decl is the node's AST decl (*ast.Field, *ast.StructType)
	Decl any

	// FVersion is the version of the AST file the decl was found on
	FVersion gast.FileVersion
}

func (k KeyableNodeMeta) SymbolKey() SymbolKey {
	return SymbolKeyFor(k.Decl, &k.FVersion)
}

type CreateControllerNode struct {
	Data        ControllerSymbolicMetadata
	Decl        *ast.TypeSpec
	Annotations *annotations.AnnotationHolder
}

type CreateRouteNode struct {
	Data             RouteSymbolicMetadata
	Decl             *ast.FuncDecl
	Annotations      *annotations.AnnotationHolder
	ParentController KeyableNodeMeta
}

type CreateParameterNode struct {
	Data        arbitrators.FuncParamWithAst
	Decl        ast.Expr
	ParentRoute KeyableNodeMeta
}

type CreateReturnValueNode struct {
	Data        arbitrators.FuncReturnValueWithAst
	Decl        ast.Expr
	ParentRoute KeyableNodeMeta
}

type CreateTypeNode struct {
	Data        arbitrators.TypeMetadataWithAst // full type info
	Annotations *annotations.AnnotationHolder   // comments on that type, if any
}
