package arbitrators

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"strconv"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/gast"
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
	typeVisitor TypeVisitor,
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
		fields, err := typeVisitor.VisitField(pkg, file, field)
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
	typeVisitor TypeVisitor,
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
		fields, err := typeVisitor.VisitField(pkg, file, field)
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
					SymbolKind:  common.SymKindParameter,
					PkgPath:     retValField.PkgPath,
					Annotations: funcAnnotations,
					FVersion:    retValField.FVersion,
				},
				Ordinal: index, // This accounts for params like "a, b string"
				Type:    retValField.Type,
			},
		)

	}

	return params, nil
}

func (arb *AstArbitrator) GetImportType(file *ast.File, expr ast.Expr) (common.ImportType, error) {
	switch e := expr.(type) {

	case *ast.Ident:
		// Check for universe type first
		if gast.IsUniverseType(e.Name) {
			return common.ImportTypeNone, nil
		}

		// Try to detect dot-import
		relevantPkg, err := arb.GetPackageFromDotImportedIdent(file, e)
		if err != nil {
			return common.ImportTypeNone, err
		}
		if relevantPkg != nil {
			return common.ImportTypeDot, nil
		}

		// If not dot, and not selector, assume it's local
		return common.ImportTypeNone, nil

	case *ast.SelectorExpr:
		// If it's a selector, assume it's an aliased import (either default or custom alias)
		return common.ImportTypeAlias, nil

	case *ast.StarExpr:
		// Dereference and recurse
		return arb.GetImportType(file, e.X)

	case *ast.ArrayType:
		return arb.GetImportType(file, e.Elt)

	case *ast.MapType:
		// IMPORTANT NOTE: Maps CAN have two separate import types! (Key & Value)
		// Currently, we're making a pretty ugly assumption that the key is a universe type and not an imported one!
		//
		// Return the import type of the value type
		return arb.GetImportType(file, e.Value)

	case *ast.ChanType:
		return arb.GetImportType(file, e.Value)

	default:
		return common.ImportTypeNone, fmt.Errorf("unsupported expression type: %T", expr)
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

// ExtractEnumAliasType attempts to determine the underlying type and possible value for the given TypeName,
// assuming it is an enumeration.
func (arb *AstArbitrator) ExtractEnumAliasType(pkg *packages.Package, typeName *types.TypeName) (*definitions.AliasMetadata, error) {

	basic, isBasicType := typeName.Type().Underlying().(*types.Basic)

	if !isBasicType {
		return nil, fmt.Errorf("type %s is not a basic type", typeName.Name())
	}

	aliasMetadata := definitions.AliasMetadata{
		Name:      typeName.Name(),
		AliasType: basic.String(),
		Values:    []string{},
	}

	// Iterate through all objects in package scope
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if obj == nil {
			continue
		}

		// Check if it's a constant
		constVal, isConst := obj.(*types.Const)
		if !isConst {
			continue
		}

		// Check if this constant has the enum type we're looking for
		if !types.Identical(constVal.Type(), typeName.Type()) {
			continue
		}

		// Extract the value based on the basic kind
		val := ""
		switch basic.Kind() {
		case types.String:
			val = constant.StringVal(constVal.Val())
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
			if intVal, ok := constant.Int64Val(constVal.Val()); ok {
				val = strconv.FormatInt(intVal, 10)
			}
		case types.Float32, types.Float64:
			if floatVal, ok := constant.Float64Val(constVal.Val()); ok {
				val = strconv.FormatFloat(floatVal, 'f', -1, 64)
			}
		case types.Bool:
			boolVal := constant.BoolVal(constVal.Val())
			val = strconv.FormatBool(boolVal)
		default:
			return nil, fmt.Errorf("unsupported alias to basic type %s", basic.String())
		}

		aliasMetadata.Values = append(aliasMetadata.Values, val)
	}

	return &aliasMetadata, nil
}
