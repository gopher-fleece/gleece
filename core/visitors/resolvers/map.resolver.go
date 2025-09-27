package resolvers

import (
	"fmt"
	"go/ast"

	"github.com/gopher-fleece/gleece/gast"
	"golang.org/x/tools/go/packages"
)

func ResolveMap(
	pkg *packages.Package,
	file *ast.File,
	exprIdent *ast.Ident,
	exprType *ast.MapType,
	getPkg func(pkgPath string) (*packages.Package, error),
) (gast.TypeSpecResolution, error) {
	resolvedKey, err := gast.ResolveTypeSpecFromExpr(pkg, file, exprType, exprType.Key, getPkg)
	if err != nil {
		return gast.TypeSpecResolution{}, fmt.Errorf("failed to resolve map key - %v", err)
	}

	resolvedValue, err := gast.ResolveTypeSpecFromExpr(pkg, file, exprType, exprType.Value, getPkg)
	if err != nil {
		return gast.TypeSpecResolution{}, fmt.Errorf("failed to resolve map key - %v", err)
	}

	return gast.TypeSpecResolution{
		TypeName:       "map",
		TypeNameIdent:  exprIdent,
		IsUniverse:     true,
		TypeParameters: []gast.TypeSpecResolution{resolvedKey, resolvedValue},
	}, nil
}
