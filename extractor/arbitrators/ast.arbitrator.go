package arbitrators

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor"
	"golang.org/x/tools/go/packages"
)

type AstArbitrator struct {
	pkgFacade *PackagesFacade
	fileSet   *token.FileSet
}

func NewAstArbitrator(pkgFacade *PackagesFacade, fileSet *token.FileSet) AstArbitrator {
	return AstArbitrator{
		pkgFacade: pkgFacade,
		fileSet:   fileSet,
	}
}

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
		paramTypes = append(paramTypes, definitions.ParamMeta{Name: field.Names[0].Name, TypeMeta: meta})
	}

	return paramTypes, nil
}

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

func (arb *AstArbitrator) GetFieldMetadata(file *ast.File, value *ast.Field) (definitions.TypeMetadata, error) {
	return arb.GetTypeMetaForExpr(file, value.Type)
}

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

func (arb *AstArbitrator) GetTypeMetaByIdent(file *ast.File, ident *ast.Ident) (definitions.TypeMetadata, error) {
	comments := extractor.GetCommentsFromIdent(arb.fileSet, file, ident)

	meta := definitions.TypeMetadata{
		Name:        ident.Name,
		Description: extractor.FindAndExtract(comments, "@Description"),
	}

	if extractor.IsUniverseType(ident.Name) {
		// The identifier is a member of the universe, e.g. 'error'.
		// Nothing to do here. Leave the package empty so the downstream generator knows no import/alias is needed
		meta.IsUniverseType = true
		meta.Import = definitions.ImportTypeNone
		meta.EntityKind = definitions.AstNodeKindUnknown
		return meta, nil
	}

	relevantPkg := arb.IsIdentFromDotImportedPackage(file, ident)
	if relevantPkg != nil {
		// The identifier is a type from a dot imported package
		meta.Import = definitions.ImportTypeDot
		meta.FullyQualifiedPackage = relevantPkg.PkgPath
		meta.DefaultPackageAlias = relevantPkg.Name
		kind, err := extractor.TryGetStructOrInterfaceKind(relevantPkg, ident.Name)
		if err != nil {
			return meta, err
		}
		meta.EntityKind = kind
	} else {
		// If we've gotten here, the ident is a locally defined entity;
		//
		// Were it a an aliased import, it've been resolved by GetTypeMetaBySelectorExpr.
		// For dot-imports, we'd have flowed to the above 'if'.
		currentPackageName, err := extractor.GetFullPackageName(file, arb.fileSet)
		if err != nil {
			return meta, err
		}

		// Verify the identifier does in fact exist in the current package.
		// Not strictly needed but helps with safety.
		exists, entityKind, err := arb.TypeOrInterfaceExistsIn(currentPackageName, ident)
		if err != nil {
			return meta, err
		}

		if !exists {
			return meta, fmt.Errorf("identifier %s does not correlate to a type or interface in package %s", ident.Name, currentPackageName)
		}

		meta.Import = definitions.ImportTypeNone
		meta.FullyQualifiedPackage = currentPackageName
		meta.DefaultPackageAlias = extractor.GetDefaultAlias(currentPackageName)
		meta.EntityKind = entityKind
	}

	return meta, nil
}

func (arb *AstArbitrator) GetTypeMetaBySelectorExpr(file *ast.File, selector *ast.SelectorExpr) (definitions.TypeMetadata, error) {
	aliasedImports := extractor.GetImportAliases(file)

	typeOrInterfaceName := selector.Sel.Name

	comments := extractor.GetCommentsFromIdent(arb.fileSet, file, selector.Sel)
	meta := definitions.TypeMetadata{
		Name:        typeOrInterfaceName,
		Description: extractor.FindAndExtract(comments, "@Description"),
		Import:      definitions.ImportTypeAlias,
	}

	// Resolve the importAlias part to a full package
	importAlias, ok := selector.X.(*ast.Ident)
	if !ok {
		return meta, fmt.Errorf("could not convert a selector expression's 'X' to an identifier. Sel name: %s", typeOrInterfaceName)
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
		return meta, fmt.Errorf("failed to retrieve package '%s' whilst processing '%s'", realFullPackageName, typeOrInterfaceName)
	}

	if pkg == nil {
		return meta, fmt.Errorf("could not find package '%s' whilst processing '%s'", realFullPackageName, typeOrInterfaceName)
	}

	kind, err := extractor.TryGetStructOrInterfaceKind(pkg, typeOrInterfaceName)
	if err != nil {
		return meta, fmt.Errorf("could not determine entity type whilst processing '%s'", typeOrInterfaceName)
	}

	meta.EntityKind = kind
	return meta, nil
}

func (arb *AstArbitrator) IsIdentFromDotImportedPackage(file *ast.File, ident *ast.Ident) *packages.Package {
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

func (arb *AstArbitrator) TypeOrInterfaceExistsIn(
	packageFullName string,
	ident *ast.Ident,
) (bool, definitions.AstNodeKind, error) {
	pkg := extractor.FilterPackageByFullName(arb.pkgFacade.GetAllPackages(), packageFullName)
	if pkg == nil {
		return false, definitions.AstNodeKindNone, fmt.Errorf("could not find package '%s' in the given list of packages", packageFullName)
	}

	typeName, err := extractor.LookupTypeName(pkg, ident.Name)
	if err != nil {
		return false, definitions.AstNodeKindNone, err
	}

	if (typeName == nil) || (typeName.Type() == nil) {
		return false, definitions.AstNodeKindNone, fmt.Errorf("could not find type '%s' in package '%s', are you sure it's included in the 'commonConfig->controllerGlobs' search paths?", ident.Name, packageFullName)
	}
	if _, isStruct := typeName.Type().Underlying().(*types.Struct); isStruct {
		return true, definitions.AstNodeKindStruct, nil
	}

	// Get the underlying type and check if it's an interface.
	if _, isInterface := typeName.Type().Underlying().(*types.Interface); isInterface {
		return true, definitions.AstNodeKindInterface, nil
	}

	// Check if that is an alias of a basic type (string, int, bool, etc.)
	if typeName.IsAlias() {
		return true, definitions.AstNodeKindAlias, nil
	}

	if _, isBasicType := typeName.Type().Underlying().(*types.Basic); isBasicType {
		return true, definitions.AstNodeKindAlias, nil
	}

	return true, definitions.AstNodeKindUnknown, nil
}
