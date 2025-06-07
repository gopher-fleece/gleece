package symboldg

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/extractor/arbitrators"
)

type CreateControllerNode struct {
	Data        ControllerSymbolicMetadata
	Decl        *ast.TypeSpec
	Annotations *annotations.AnnotationHolder
}

type CreateRouteNode struct {
	Data             RouteSymbolicMetadata
	Decl             *ast.FuncDecl
	Annotations      *annotations.AnnotationHolder
	ParentController *SymbolNode
}

type CreateParameterNode struct {
	Data        arbitrators.FuncParamWithAst
	Decl        *ast.Field
	Annotations *annotations.AnnotationHolder
	ParentRoute *SymbolNode
}

type CreateReturnValueNode struct {
	Data        arbitrators.FuncReturnValueWithAst
	Decl        *ast.Field
	Annotations *annotations.AnnotationHolder
	ParentRoute *SymbolNode
}

type CreateTypeNode struct {
	Data        arbitrators.TypeMetadataWithAst // full type info
	Annotations *annotations.AnnotationHolder   // comments on that type, if any
}
