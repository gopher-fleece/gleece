package arbitrators

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strconv"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"golang.org/x/tools/go/packages"
)

// AstArbitrator serves as an abstraction and interconnect for Go's AST and packages components.
// The arbitrator is used to obtain high level metadata for given AST structures like FuncDecls
type AstArbitrator struct {
	pkgFacade *PackagesFacade
	fileSet   *token.FileSet
}

// Creates a new AST Arbitrator
func NewAstArbitrator(pkgFacade *PackagesFacade, fileSet *token.FileSet) AstArbitrator {
	return AstArbitrator{
		pkgFacade: pkgFacade,
		fileSet:   fileSet,
	}
}

// GetFuncParameterTypeList gets metadata for all parameters of the given function in the given AST file
func (arb *AstArbitrator) GetFuncParameterTypeList(file *ast.File, funcDecl *ast.FuncDecl) ([]definitions.ParamMeta, error) {
	paramTypes := []definitions.ParamMeta{}

	if funcDecl.Type.Params == nil || funcDecl.Type.Params.List == nil {
		return paramTypes, nil
	}

	for _, field := range funcDecl.Type.Params.List {
		meta, err := arb.GetFieldMetadata(file, field)
		if err != nil {
			return paramTypes, err
		}

		paramTypes = append(
			paramTypes,
			definitions.ParamMeta{
				Name: field.Names[0].Name, TypeMeta: meta,
				// A special case- Go Contexts should be explicitly marked so they can be injected via the template
				IsContext: meta.Name == "Context" && meta.FullyQualifiedPackage == "context",
			},
		)
	}

	return paramTypes, nil
}

// GetFuncReturnTypesMetadata gets type metadata for all return values of the given function in the given AST file
func (arb *AstArbitrator) GetFuncReturnTypeList(file *ast.File, funcDecl *ast.FuncDecl) ([]definitions.TypeMetadata, error) {
	returnTypes := []definitions.TypeMetadata{}

	if funcDecl.Type.Results == nil {
		return returnTypes, nil
	}

	for _, field := range funcDecl.Type.Results.List {
		meta, err := arb.GetFieldMetadata(file, field)
		if err != nil {
			return returnTypes, err
		}
		returnTypes = append(returnTypes, meta)
	}
	return returnTypes, nil
}

// GetFieldMetadata gets metadata for the given field in the given AST file
func (arb *AstArbitrator) GetFieldMetadata(file *ast.File, value *ast.Field) (definitions.TypeMetadata, error) {
	return arb.GetTypeMetaForExpr(file, value.Type)
}

// GetTypeMetaForExpr gets type metadata for the given expression in the given AST file
// Note that if the type of expression is not currently supported, an error will be returned.
func (arb *AstArbitrator) GetTypeMetaForExpr(file *ast.File, expr ast.Expr) (definitions.TypeMetadata, error) {
	switch fieldType := expr.(type) {
	case *ast.Ident:
		return arb.GetTypeMetaByIdent(file, fieldType)
	case *ast.SelectorExpr:
		return arb.GetTypeMetaBySelectorExpr(file, fieldType)
	case *ast.StarExpr:
		meta, err := arb.GetFieldMetadata(file, &ast.Field{Type: fieldType.X})
		meta.IsByAddress = true
		return meta, err
	case *ast.ArrayType:
		meta, err := arb.GetTypeMetaForExpr(file, fieldType.Elt)
		meta.Name = "[]" + meta.Name // Passing the array information as a string - easy to unwrap later on.
		return meta, err
	default:
		fieldTypeString := extractor.GetFieldTypeString(fieldType)
		return definitions.TypeMetadata{}, fmt.Errorf("field type '%s' is not currently supported", fieldTypeString)
	}
}

// GetTypeMetaByIdent gets type metadata for the given Ident in the given AST file
func (arb *AstArbitrator) GetTypeMetaByIdent(file *ast.File, ident *ast.Ident) (definitions.TypeMetadata, error) {
	comments := extractor.GetCommentsFromIdent(arb.fileSet, file, ident)
	holder, err := annotations.NewAnnotationHolder(comments, annotations.CommentSourceProperty)

	meta := definitions.TypeMetadata{Name: ident.Name}

	if err != nil {
		return meta, err
	}

	meta.Description = holder.GetDescription()

	if extractor.IsUniverseType(ident.Name) {
		// The identifier is a member of the universe, e.g. 'error'.
		// Nothing to do here. Leave the package empty so the downstream generator knows no import/alias is needed
		meta.IsUniverseType = true
		meta.Import = definitions.ImportTypeNone
		meta.SymbolKind = definitions.SymKindUnknown
		return meta, nil
	}

	relevantPkg := arb.GetPackageFromDotImportedIdent(file, ident)
	if relevantPkg != nil {
		// The identifier is a type from a dot imported package
		meta.Import = definitions.ImportTypeDot
		meta.FullyQualifiedPackage = relevantPkg.PkgPath
		meta.DefaultPackageAlias = relevantPkg.Name
		kind, err := extractor.GetSymbolKind(relevantPkg, ident.Name)
		if err != nil {
			return meta, err
		}
		meta.SymbolKind = kind
	} else {
		// If we've gotten here, the ident is a locally defined entity;
		//
		// Were it a an aliased import, it've been resolved by GetTypeMetaBySelectorExpr.
		// For dot-imports, we'd have flowed to the above 'if'.
		currentPackageName, err := extractor.GetFullPackageName(file, arb.fileSet)
		if err != nil {
			return meta, err
		}

		pkg, err := arb.pkgFacade.GetPackage(currentPackageName)
		if err != nil {
			return meta, err
		}

		if pkg == nil {
			return meta, fmt.Errorf("could not find package '%s' in the given list of packages", currentPackageName)
		}

		typeName, err := extractor.GetTypeNameOrError(pkg, ident.Name)
		if err != nil {
			return meta, err
		}

		// Verify the identifier does in fact exist in the current package.
		// Not strictly needed but helps with safety.
		symbolKind := extractor.GetSymbolKindFromObject(typeName)

		if symbolKind == definitions.SymKindUnknown {
			return meta, fmt.Errorf("could not determine entity kind for '%s.%s", currentPackageName, ident.Name)
		}

		meta.Import = definitions.ImportTypeNone
		meta.FullyQualifiedPackage = currentPackageName
		meta.DefaultPackageAlias = extractor.GetDefaultAlias(currentPackageName)
		meta.SymbolKind = symbolKind

		if symbolKind == definitions.SymKindAlias {
			aliasMetadata, err := arb.ExtractEnumAliasType(pkg, typeName)
			if err != nil {
				return meta, err
			}
			meta.AliasMetadata = aliasMetadata
		}
	}

	return meta, nil
}

// GetTypeMetaBySelectorExpr gets type metadata for the given Selector Expression in the given AST file
func (arb *AstArbitrator) GetTypeMetaBySelectorExpr(file *ast.File, selector *ast.SelectorExpr) (definitions.TypeMetadata, error) {
	aliasedImports := extractor.GetImportAliases(file)

	entityName := selector.Sel.Name

	meta := definitions.TypeMetadata{
		Name:   entityName,
		Import: definitions.ImportTypeAlias,
	}

	comments := extractor.GetCommentsFromIdent(arb.fileSet, file, selector.Sel)
	holder, err := annotations.NewAnnotationHolder(comments, annotations.CommentSourceProperty)
	if err != nil {
		return meta, err
	}

	meta.Description = holder.GetDescription()

	// Resolve the importAlias part to a full package
	importAlias, ok := selector.X.(*ast.Ident)
	if !ok {
		return meta, fmt.Errorf("could not convert a selector expression's 'X' to an identifier. Sel name: %s", entityName)
	}

	var realFullPackageName string

	aliasedFullName := aliasedImports[importAlias.Name]
	if len(aliasedFullName) == 0 { // If there's no alias, the string will be empty
		for maybeFullPackageName, fullPackageName := range aliasedImports {
			if maybeFullPackageName == fullPackageName && extractor.IsAliasDefault(maybeFullPackageName, importAlias.Name) {
				// A reverse check - if the import uses a default alias, we look in the map in reverse;
				// Since the SelectorExpr's X is the default alias, we can check each import to see if its default alias matches the X.
				// If it does, it's a match.
				// The secondary 'maybeFullPackageName == fullPackageName' check is mostly just-in-case - for default aliases
				// we expect the the mapped key to equal the mapped value.
				realFullPackageName = fullPackageName
				break
			}
		}
	} else {
		// Imported with a custom alias
		realFullPackageName = aliasedFullName
	}

	meta.FullyQualifiedPackage = realFullPackageName
	meta.DefaultPackageAlias = extractor.GetDefaultAlias(realFullPackageName)

	pkg, err := arb.pkgFacade.GetPackage(realFullPackageName)
	if err != nil {
		return meta, fmt.Errorf("failed to retrieve package '%s' whilst processing '%s'", realFullPackageName, entityName)
	}

	if pkg == nil {
		return meta, fmt.Errorf("could not find package '%s' whilst processing '%s'", realFullPackageName, entityName)
	}

	kind, err := extractor.GetSymbolKind(pkg, entityName)
	if err != nil {
		return meta, err
	}

	meta.SymbolKind = kind

	if kind == definitions.SymKindAlias {
		typeName, err := extractor.LookupTypeName(pkg, entityName)
		if err != nil {
			return meta, err
		}

		if typeName == nil {
			return meta, fmt.Errorf("type '%s' was not found in package %s", entityName, pkg.Name)
		}
		aliasMetadata, err := arb.ExtractEnumAliasType(pkg, typeName)
		if err != nil {
			return meta, err
		}
		meta.AliasMetadata = aliasMetadata
	}
	return meta, nil
}

// GetPackageFromDotImportedIdent returns the package from which a dot-imported ident was imported or nil
// the ident is not a dot-imported one.
//
// This method can be used as a "Is this Ident from a dot-import?"
func (arb *AstArbitrator) GetPackageFromDotImportedIdent(file *ast.File, ident *ast.Ident) *packages.Package {
	dotImports := extractor.GetDotImportedPackageNames(file)
	for _, dotImport := range dotImports {
		pkg := extractor.FilterPackageByFullName(arb.pkgFacade.GetAllPackages(), dotImport)
		if pkg != nil {
			if extractor.IsIdentInPackage(pkg, ident) {
				return pkg
			}
		}
	}
	return nil
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
