package arbitrators

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/metadata"
	"golang.org/x/tools/go/packages"
)

type FieldVisitor interface {
	VisitField(
		pkg *packages.Package,
		file *ast.File,
		field *ast.Field,
		kind common.SymKind,
	) ([]metadata.FieldMeta, error)
}
