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
		kind, err := extractor.GetEntityKind(relevantPkg, ident.Name)
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
		entityKind, err := extractor.GetEntityKindFromTypeName(typeName)
		if err != nil {
			return meta, err
		}

		if entityKind == definitions.AstNodeKindNone || entityKind == definitions.AstNodeKindUnknown {
			return meta, fmt.Errorf("could not determine entity kind for '%s.%s", currentPackageName, ident.Name)
		}

		meta.Import = definitions.ImportTypeNone
		meta.FullyQualifiedPackage = currentPackageName
		meta.DefaultPackageAlias = extractor.GetDefaultAlias(currentPackageName)
		meta.EntityKind = entityKind

		if entityKind == definitions.AstNodeKindAlias {
			aliasMetadata, err := arb.ExtractAliasType(pkg, typeName)
			if err != nil {
				return meta, err
			}
			meta.AliasMetadata = aliasMetadata
		}
	}

	return meta, nil
}

func (arb *AstArbitrator) GetTypeMetaBySelectorExpr(file *ast.File, selector *ast.SelectorExpr) (definitions.TypeMetadata, error) {
	aliasedImports := extractor.GetImportAliases(file)

	entityName := selector.Sel.Name

	comments := extractor.GetCommentsFromIdent(arb.fileSet, file, selector.Sel)
	meta := definitions.TypeMetadata{
		Name:        entityName,
		Description: extractor.FindAndExtract(comments, "@Description"),
		Import:      definitions.ImportTypeAlias,
	}

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

	kind, err := extractor.GetEntityKind(pkg, entityName)
	if err != nil {
		return meta, err
	}

	meta.EntityKind = kind

	if kind == definitions.AstNodeKindAlias {
		typeName, err := extractor.LookupTypeName(pkg, entityName)
		if err != nil {
			return meta, err
		}

		if typeName == nil {
			return meta, fmt.Errorf("type '%s' was not found in package %s", entityName, pkg.Name)
		}
		aliasMetadata, err := arb.ExtractAliasType(pkg, typeName)
		if err != nil {
			return meta, err
		}
		meta.AliasMetadata = aliasMetadata
	}
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

func (arb *AstArbitrator) ExtractAliasType(pkg *packages.Package, typeName *types.TypeName) (*definitions.AliasMetadata, error) {

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

func (arb *AstArbitrator) GetAstFileForNamed(named *types.Named) (*ast.File, error) {
	pkg := named.Obj().Pkg()
	if pkg == nil {
		return nil, nil // Built-in types, unnamed types, etc.
	}

	// Find the package where the struct is defined
	pkgPath := pkg.Path()
	targetPkg, err := arb.pkgFacade.GetPackage(pkgPath)
	if err != nil {
		return nil, err
	}

	if targetPkg == nil {
		return nil, nil
	}

	// Iterate over all AST files to find the struct declaration
	for _, file := range targetPkg.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				// If this is the struct we're looking for, return the file
				if typeSpec.Name.Name == named.Obj().Name() {
					return file, nil
				}
			}
		}
	}

	return nil, nil // Not found
}
