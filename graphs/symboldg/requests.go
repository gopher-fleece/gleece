package symboldag

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/annotations"
)

type CreateControllerNode struct {
	Data        *definitions.ControllerMetadata
	Decl        *ast.FuncDecl
	Annotations *annotations.AnnotationHolder
}
