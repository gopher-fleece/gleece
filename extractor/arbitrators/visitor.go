package arbitrators

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/extractor/metadata"
)

type TypeVisitor interface {
	Visit(node ast.Node) ast.Visitor
	VisitStructType(file *ast.File, node *ast.TypeSpec) (*metadata.StructMeta, error)
	VisitField(file *ast.File, field *ast.Field, pkgPath string) ([]metadata.FieldMeta, error)
}
