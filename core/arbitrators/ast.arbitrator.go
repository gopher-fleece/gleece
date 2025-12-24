package arbitrators

import (
	"fmt"
	"go/ast"

	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/core/annotations"
	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/gast"
	"golang.org/x/tools/go/packages"
)

// AstArbitrator serves as an abstraction and interconnect for Go's AST and packages components.
// The arbitrator is used to obtain high level metadata for given AST structures like FuncDecls
type AstArbitrator struct {
	pkgFacade *PackagesFacade
}

// Creates a new AST Arbitrator
func NewAstArbitrator(pkgFacade *PackagesFacade) AstArbitrator {
	return AstArbitrator{
		pkgFacade: pkgFacade,
	}
}

func (arb *AstArbitrator) GetFuncParametersMeta(
	typeVisitor FieldVisitor,
	pkg *packages.Package,
	file *ast.File,
	funcDecl *ast.FuncDecl,
	funcAnnotations *annotations.AnnotationHolder,
) ([]metadata.FuncParam, error) {
	params := []metadata.FuncParam{}

	if funcDecl.Type.Params == nil || funcDecl.Type.Params.List == nil {
		return params, nil
	}

	for paramOrdinal, field := range funcDecl.Type.Params.List {
		fields, err := typeVisitor.VisitField(pkg, file, field, common.SymKindParameter)
		if err != nil {
			return nil, err
		}

		for index, f := range fields {
			params = append(
				params,
				metadata.FuncParam{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:        f.Name,
						Node:        f.Node,
						SymbolKind:  common.SymKindParameter,
						PkgPath:     f.PkgPath,
						Annotations: funcAnnotations,
						FVersion:    f.FVersion,
						Range:       arb.getRangeForNode(field),
					},
					Ordinal: paramOrdinal + index, // This accounts for params like "a, b string"
					Type:    f.Type,
				},
			)
		}
	}

	return params, nil
}

func (arb *AstArbitrator) GetFuncRetValMeta(
	fieldVisitor FieldVisitor,
	pkg *packages.Package,
	file *ast.File,
	funcDecl *ast.FuncDecl,
	funcAnnotations *annotations.AnnotationHolder,
) ([]metadata.FuncReturnValue, error) {
	params := []metadata.FuncReturnValue{}

	if funcDecl.Type.Results == nil || funcDecl.Type.Results.List == nil {
		return params, nil
	}

	for index, field := range funcDecl.Type.Results.List {
		fields, err := fieldVisitor.VisitField(pkg, file, field, common.SymKindReturnType)
		if err != nil {
			return nil, err
		}

		if len(fields) < 1 {
			return nil, fmt.Errorf(
				"return value %d on function %s did not yield any information and could not be processed",
				index,
				funcDecl.Name,
			)
		}

		if len(fields) > 1 {
			return nil, fmt.Errorf(
				"return value %d on function %s is a multi-variable declaration which is not currently supported",
				index,
				funcDecl.Name,
			)
		}

		retValField := fields[0]
		params = append(
			params,
			metadata.FuncReturnValue{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:        retValField.Name,
					Node:        retValField.Node,
					SymbolKind:  common.SymKindReturnType,
					PkgPath:     retValField.PkgPath,
					Annotations: funcAnnotations,
					FVersion:    retValField.FVersion,
					Range:       arb.getRangeForNode(field),
				},
				Ordinal: index, // This accounts for params like "a, b string"
				Type:    retValField.Type,
			},
		)

	}

	return params, nil
}

// GetImportType returns the import classification for an expression.
// Policy summary:
//   - For Base[T] and Base[T1,...] -> return import type of Base.
//   - For pointer/slice/array/paren/chan -> return import type of the element/base.
//   - For map[K]V -> ImportTypeNone (K and V may differ).
//   - For selector -> ImportTypeAlias.
//   - For ident -> dot-import detection or ImportTypeNone.
//   - For func / struct (and other multi-origin/no-single-base) -> ImportTypeNone.
func (arb *AstArbitrator) GetImportType(file *ast.File, expr ast.Expr) (common.ImportType, error) {
	switch e := expr.(type) {
	case *ast.Ident:
		if gast.IsUniverseType(e.Name) {
			return common.ImportTypeNone, nil
		}
		pkg, err := arb.GetPackageFromDotImportedIdent(file, e)
		if err != nil {
			return common.ImportTypeNone, err
		}
		if pkg != nil {
			return common.ImportTypeDot, nil
		}
		return common.ImportTypeNone, nil

	case *ast.SelectorExpr:
		return common.ImportTypeAlias, nil

	// single-operand composites: defer to operand's import type
	case *ast.StarExpr:
		return arb.GetImportType(file, e.X)
	case *ast.ParenExpr:
		return arb.GetImportType(file, e.X)
	case *ast.ArrayType:
		return arb.GetImportType(file, e.Elt)
	case *ast.IndexExpr:
		// Base[Index] -> base determines import
		return arb.GetImportType(file, e.X)
	case *ast.IndexListExpr:
		return arb.GetImportType(file, e.X)
	case *ast.ChanType:
		return arb.GetImportType(file, e.Value)

	// Maps, funcs and structs are always 'not imported' - map is a universe type and Func/StructTypes are local declarations
	case *ast.MapType, *ast.FuncType, *ast.StructType:
		return common.ImportTypeNone, nil

	default:
		return common.ImportTypeNone, fmt.Errorf("unsupported expression type for import detection: %T", expr)
	}
}

// GetPackageFromDotImportedIdent returns the package from which a dot-imported ident was imported or nil
// the ident is not a dot-imported one.
//
// This method can be used as a "Is this Ident from a dot-import?"
func (arb *AstArbitrator) GetPackageFromDotImportedIdent(file *ast.File, ident *ast.Ident) (*packages.Package, error) {
	dotImports := gast.GetDotImportedPackageNames(file)
	for _, dotImport := range dotImports {
		pkg, err := arb.pkgFacade.GetPackage(dotImport)
		if err != nil {
			return nil, err
		}

		if pkg != nil {
			if gast.IsIdentInPackage(pkg, ident) {
				return pkg, nil
			}
		}
	}

	return nil, nil
}

func (arb *AstArbitrator) getRangeForNode(n ast.Node) common.ResolvedRange {
	return common.ResolvedRange{
		StartLine: arb.pkgFacade.FSet().Position(n.Pos()).Line - 1,
		StartCol:  arb.pkgFacade.FSet().Position(n.Pos()).Column - 1,
		EndLine:   arb.pkgFacade.FSet().Position(n.End()).Line - 1,
		EndCol:    arb.pkgFacade.FSet().Position(n.End()).Column - 1,
	}
}
