package arbitrators

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/graphs"
	"golang.org/x/tools/go/packages"
)

type TypeVisitor interface {
	Visit(node ast.Node) ast.Visitor
	VisitStructType(
		file *ast.File,
		nodeGenDecl *ast.GenDecl,
		node *ast.TypeSpec,
	) (*metadata.StructMeta, graphs.SymbolKey, error)
	VisitField(pkg *packages.Package, file *ast.File, field *ast.Field) ([]metadata.FieldMeta, error)
}
