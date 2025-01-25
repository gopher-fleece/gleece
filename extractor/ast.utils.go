package extractor

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	MapSet "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/definitions"
	"golang.org/x/tools/go/packages"
)

func IsFuncDeclReceiverForStruct(structName string, funcDecl *ast.FuncDecl) bool {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) <= 0 {
		return false
	}

	switch expr := funcDecl.Recv.List[0].Type.(type) {
	case *ast.Ident:
		return expr.Name == structName
	case *ast.StarExpr:
		return expr.X.(*ast.Ident).Name == structName
	default:
		return false
	}
}

// DoesStructEmbedStructByName Checks whether `structNode` embeds a struct who's name equals embeddedStructName.
// Note this function does NOT take into account
func DoesStructEmbedStructByName(structNode *ast.StructType, embeddedStructName string) bool {
	for _, field := range structNode.Fields.List {
		switch fieldType := field.Type.(type) {
		case *ast.Ident:
			if fieldType.Name == embeddedStructName {
				return true
			}
		case *ast.SelectorExpr:
			if fieldType.Sel.Name == embeddedStructName {
				return true
			}
		}
	}
	return false
}

func DoesStructEmbedStruct(
	sourceFile *ast.File,
	structFullPackageName string,
	structNode *ast.StructType,
	embeddedStructName string,
) bool {
	aliasedImports := GetImportAliases(sourceFile)

	// Iterate over the struct fields to check for embeds
	for _, field := range structNode.Fields.List {
		switch fieldType := field.Type.(type) {
		case *ast.Ident:
			// If the type is just an Ident (simple struct type), check the name
			if fieldType.Name == embeddedStructName {
				// If it's an Ident, check if it's a dot import or a direct match
				if isDotImported, _ := IsPackageDotImported(sourceFile, structFullPackageName); isDotImported {
					return true
				}
			}
		case *ast.SelectorExpr:
			// If the type is a SelectorExpr (meaning it's a struct from another package), check the package and name
			if ident, ok := fieldType.X.(*ast.Ident); ok {
				// Compare the package name and struct name
				sourcePackage := aliasedImports[ident.Name]
				isCorrectPackage := sourcePackage == structFullPackageName || IsAliasDefault(structFullPackageName, ident.Name)
				if isCorrectPackage && fieldType.Sel.Name == embeddedStructName {
					return true
				}
			}
		}
	}
	return false
}

func GetDefaultPackageAlias(file *ast.File) (string, error) {
	if file.Name != nil {
		return file.Name.Name, nil
	}
	return "", fmt.Errorf("source file does not have a name")
}

func GetFullPackageName(file *ast.File, fileSet *token.FileSet) (string, error) {
	// Get the file's full path using the fileSet
	position := fileSet.Position(file.Package)
	relativePath := position.Filename

	absFilePath, err := filepath.Abs(relativePath)
	if err != nil {
		return "", err
	}

	// Get the directory of the file
	dir := filepath.Dir(absFilePath)

	// Load package information for the file's directory
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedModule,
		Dir:  dir,
	}
	packages, err := packages.Load(cfg)
	if err != nil {
		return "", err
	}

	// Find the package matching the current file
	for _, pkg := range packages {
		for _, pkgFile := range pkg.GoFiles {
			if filepath.Clean(pkgFile) == filepath.Clean(absFilePath) {
				// Return the fully qualified package name
				return pkg.PkgPath, nil
			}
		}
	}

	return "", nil // Package not found
}

func IsPackageDotImported(file *ast.File, packageName string) (bool, string) {
	for _, imp := range file.Imports {
		// Check if it's a dot import (imp.Name == nil) and if the package path matches the expected package name
		if imp.Name != nil && imp.Path != nil {
			importedPackageName := strings.Trim(imp.Path.Value, `"`)
			if importedPackageName == packageName {
				// Ensure that the struct name is the same as the dot-imported struct
				// Since we know it's a dot import, any struct with this name should be from the expected package
				return true, importedPackageName
			}
		}
	}
	return false, ""
}

func IsAliasDefaultImport(file *ast.File, alias string) bool {
	for _, imp := range file.Imports {
		fullImport := strings.Trim(imp.Path.Value, `"`)
		if GetDefaultAlias(fullImport) == alias {
			return true
		}
	}
	return false
}

func IsAliasDefault(fullPackageName string, alias string) bool {
	packageName := GetDefaultAlias(fullPackageName)
	return alias == packageName
}

func GetDefaultAlias(fullyQualifiedPackage string) string {
	segments := strings.Split(fullyQualifiedPackage, "/")
	return segments[len(segments)-1]
}

func GetDotImportedPackageNames(file *ast.File) []string {
	dotImports := []string{}

	for _, imp := range file.Imports {
		packagePath := strings.Trim(imp.Path.Value, `"`)
		if imp.Name != nil && imp.Name.Name == "." {
			dotImports = append(dotImports, packagePath)
		}
	}

	return dotImports
}

func GetImportAliases(file *ast.File) map[string]string {
	aliases := make(map[string]string)
	for _, imp := range file.Imports {
		packagePath := strings.Trim(imp.Path.Value, `"`)
		alias := packagePath
		if imp.Name != nil {
			alias = imp.Name.Name
		}
		aliases[alias] = packagePath
	}
	return aliases
}

func isIdentFromDotImportedPackage(file *ast.File, packages []*packages.Package, ident *ast.Ident) *packages.Package {
	dotImports := GetDotImportedPackageNames(file)
	for _, dotImport := range dotImports {
		pkg := FilterPackageByFullName(packages, dotImport)
		if pkg != nil {
			if IsIdentInPackage(pkg, ident) {
				return pkg
			}
		}
	}
	return nil
}

func GetCommentsFromIdent(ident *ast.Ident) []string {
	if ident.Obj == nil || ident.Obj.Decl == nil {
		return []string{}
	}
	switch expr := ident.Obj.Decl.(type) {
	case *ast.TypeSpec:
	case *ast.FuncDecl:
	case *ast.Field:
		return MapDocListToStrings(expr.Doc.List)
	}

	// A bit hacky but we don't currently need parse everything
	return nil
}

func GetTypeMetaByIdent(
	file *ast.File,
	fileSet *token.FileSet,
	packages []*packages.Package,
	ident *ast.Ident,
) (definitions.TypeMetadata, error) {
	comments := GetCommentsFromIdent(ident)

	meta := definitions.TypeMetadata{
		Name:        ident.Name,
		Description: FindAndExtract(comments, "@Description"),
	}

	if IsUniverseType(ident.Name) {
		// The identifier is a member of the universe, e.g. 'error'.
		// Nothing to do here. Leave the package empty so the downstream generator knows no import/alias is needed
		meta.IsUniverseType = true
		meta.Import = definitions.ImportTypeNone
		meta.EntityKind = definitions.AstEntityUniverse
		return meta, nil
	}

	relevantPkg := isIdentFromDotImportedPackage(file, packages, ident)
	if relevantPkg != nil {
		// The identifier is a type from a dot imported package
		meta.Import = definitions.ImportTypeDot
		meta.FullyQualifiedPackage = relevantPkg.PkgPath
		meta.DefaultPackageAlias = relevantPkg.Name
		kind, err := TryGetStructOrInterfaceKind(relevantPkg, ident.Name)
		if err != nil {
			return meta, err
		}
		meta.EntityKind = kind
	} else {
		// The identifier is locally defined
		currentPackageName, err := GetFullPackageName(file, fileSet)
		if err != nil {
			return meta, err
		}

		// Verify the identifier does in fact exist in the current package.
		// Not strictly needed but helps with safety.
		exists, entityKind, err := DoesTypeOrInterfaceExistInPackage(packages, currentPackageName, ident)
		if err != nil {
			return meta, err
		}

		if !exists {
			return meta, fmt.Errorf("identifier %s does not correlate to a type or interface in package %s", ident.Name, currentPackageName)
		}

		meta.Import = definitions.ImportTypeAlias
		meta.FullyQualifiedPackage = currentPackageName
		meta.DefaultPackageAlias = GetDefaultAlias(currentPackageName)
		meta.EntityKind = entityKind
	}

	return meta, nil
}

func GetTypeMetaBySelectorExpr(
	file *ast.File,
	fileSet *token.FileSet,
	packages []*packages.Package,
	selector *ast.SelectorExpr,
) (definitions.TypeMetadata, error) {
	aliasedImports := GetImportAliases(file)

	typeOrInterfaceName := selector.Sel.Name

	comments := GetCommentsFromIdent(selector.Sel)
	meta := definitions.TypeMetadata{
		Name:        typeOrInterfaceName,
		Description: FindAndExtract(comments, "@Description"),
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
			if maybeFullPackageName == fullPackageName && IsAliasDefault(maybeFullPackageName, importAlias.Name) {
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
	meta.DefaultPackageAlias = GetDefaultAlias(realFullPackageName)

	pkg := FilterPackageByFullName(packages, realFullPackageName)
	if pkg == nil {
		return meta, fmt.Errorf("could not get *packages.Package for '%s' whilst processing '%s'", realFullPackageName, typeOrInterfaceName)
	}

	kind, err := TryGetStructOrInterfaceKind(pkg, typeOrInterfaceName)
	if err != nil {
		return meta, fmt.Errorf("could not determine entity type whilst processing '%s'", typeOrInterfaceName)
	}

	meta.EntityKind = kind
	return meta, nil
}

func GetFieldMetadata(
	file *ast.File,
	fileSet *token.FileSet,
	packages []*packages.Package,
	value *ast.Field,
) (definitions.TypeMetadata, error) {
	switch fieldType := value.Type.(type) {
	case *ast.Ident:
		return GetTypeMetaByIdent(file, fileSet, packages, fieldType)
	case *ast.SelectorExpr:
		return GetTypeMetaBySelectorExpr(file, fileSet, packages, fieldType)
	case *ast.StarExpr:
		meta, err := GetFieldMetadata(file, fileSet, packages, &ast.Field{Type: fieldType.X})
		if err == nil {
			meta.IsByAddress = true
		}
		return meta, err
	default:
		fieldTypeString := GetFieldTypeString(fieldType)
		return definitions.TypeMetadata{}, fmt.Errorf("field type '%s' is not currently supported", fieldTypeString)
	}
}

func GetFuncParameterTypeList(
	file *ast.File,
	fileSet *token.FileSet,
	packages []*packages.Package,
	funcDecl *ast.FuncDecl,
) ([]definitions.ParamMeta, error) {
	paramTypes := []definitions.ParamMeta{}

	if funcDecl.Type.Params == nil || funcDecl.Type.Params.List == nil {
		return paramTypes, nil
	}

	for _, field := range funcDecl.Type.Params.List {
		meta, err := GetFieldMetadata(file, fileSet, packages, field)
		if err != nil {
			return paramTypes, err
		}
		paramTypes = append(paramTypes, definitions.ParamMeta{Name: field.Names[0].Name, TypeMeta: meta})
	}

	return paramTypes, nil
}

func GetFuncReturnTypeList(
	file *ast.File,
	fileSet *token.FileSet,
	packages []*packages.Package,
	funcDecl *ast.FuncDecl,
) ([]definitions.TypeMetadata, error) {
	returnTypes := []definitions.TypeMetadata{}

	if funcDecl.Type.Results == nil {
		return returnTypes, nil
	}

	for _, field := range funcDecl.Type.Results.List {
		meta, err := GetFieldMetadata(file, fileSet, packages, field)
		if err != nil {
			return returnTypes, err
		}
		returnTypes = append(returnTypes, meta)
	}
	return returnTypes, nil
}

// IsIdentFromDotImport resolves whether an `*ast.Ident` refers to a type from a dot-imported package.
func IsIdentFromDotImport(file *ast.File, ident *ast.Ident, typeInfo *types.Info) (bool, error) {
	// Get the object corresponding to the identifier
	obj, ok := typeInfo.Uses[ident]
	if !ok {
		return false, nil // Identifier is not resolved
	}

	// Get the package where the object is defined
	pkg := obj.Pkg()
	if pkg == nil {
		return false, nil // Not a package-level object
	}

	// Check if the package was dot-imported
	for _, imp := range file.Imports {
		if imp.Name != nil && imp.Name.Name == "." {
			// Trim the quotes around the import path
			importPath := strings.Trim(imp.Path.Value, "\"")
			if pkg.Path() == importPath {
				return true, nil
			}
		}
	}

	return false, nil
}

func GetAllPackageFiles(codeFiles []*ast.File, fileSet *token.FileSet, fullPackageNameToFind string) ([]*ast.File, error) {
	packageFiles := []*ast.File{}
	for _, file := range codeFiles {
		packageName, err := GetFullPackageName(file, fileSet)
		if err != nil {
			return packageFiles, err
		}
		if packageName == fullPackageNameToFind {
			packageFiles = append(packageFiles, file)
		}
	}
	return packageFiles, nil
}

func GetFileByImportNode(codeFiles []*ast.File, fileSet *token.FileSet, importNode *ast.ImportSpec) (*ast.File, error) {
	cleanedPackageName := strings.Trim(importNode.Path.Value, "\"")
	for _, file := range codeFiles {
		filePackageName, err := GetFullPackageName(file, fileSet)
		if err != nil {
			return nil, err
		}
		if filePackageName == cleanedPackageName {
			return file, nil
		}
	}
	return nil, nil
}

func GetPackageAndDependencies(
	codeFiles []*ast.File,
	fileSet *token.FileSet,
	fullPackageNameToFind string,
	relevantFilesOutput *MapSet.Set[*ast.File],
) error {
	files, err := GetAllPackageFiles(codeFiles, fileSet, fullPackageNameToFind)
	if err != nil {
		return err
	}

	for _, file := range files {
		(*relevantFilesOutput).Add(file)
		for _, importNode := range file.Imports {
			relevantFile, err := GetFileByImportNode(codeFiles, fileSet, importNode)
			if relevantFile == nil || err != nil {
				return err
			}

			packageName, err := GetFullPackageName(relevantFile, fileSet)
			if err != nil {
				return err
			}

			err = GetPackageAndDependencies(codeFiles, fileSet, packageName, relevantFilesOutput)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func GetPackagesFromExpressions(packageExpressions []string) ([]*packages.Package, error) {
	cfg := &packages.Config{Mode: packages.LoadAllSyntax}
	return packages.Load(cfg, packageExpressions...)
}

func IsIdentInPackage(pkg *packages.Package, ident *ast.Ident) bool {
	return pkg.Types.Scope().Lookup(ident.Name) != nil
}

func DoesTypeOrInterfaceExistInPackage(
	packages []*packages.Package,
	packageFullName string,
	ident *ast.Ident,
) (bool, definitions.AstEntityKind, error) {
	pkg := FilterPackageByFullName(packages, packageFullName)
	if pkg == nil {
		return false, definitions.AstEntityNone, fmt.Errorf("could not find package '%s' in the given list of packages", packageFullName)
	}

	typeName, err := LookupTypeName(pkg, ident.Name)
	if err != nil {
		return false, definitions.AstEntityNone, err
	}

	if _, isStruct := typeName.Type().Underlying().(*types.Struct); isStruct {
		return true, definitions.AstEntityStruct, nil
	}

	// Get the underlying type and check if it's an interface.
	if _, isInterface := typeName.Type().Underlying().(*types.Interface); isInterface {
		return true, definitions.AstEntityInterface, nil
	}

	return true, definitions.AstEntityUnknown, nil
}

func IsUniverseType(typeName string) bool {
	obj := types.Universe.Lookup(typeName)
	if obj == nil {
		return false
	}

	_, isType := obj.(*types.TypeName)
	return isType
}

func FilterPackageByFullName(packages []*packages.Package, fullName string) *packages.Package {
	for _, pkg := range packages {
		if pkg.PkgPath == fullName {
			return pkg
		}
	}
	return nil
}

func FindGenDeclByName(pkg *packages.Package, typeSpecName string) *ast.GenDecl {
	// Traverse all the syntax trees (ASTs) in the loaded package
	for _, file := range pkg.Syntax {
		// Traverse all declarations in the AST file
		for _, decl := range file.Decls {
			// Look for type declarations (ast.GenDecl)
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				// Iterate over all the specs in the general declaration
				for _, spec := range genDecl.Specs {
					// Look for type specs (ast.TypeSpec)
					typeSpec, ok := spec.(*ast.TypeSpec)
					if ok && typeSpec.Name.Name == typeSpecName {
						return genDecl
					}
				}
			}
		}
	}
	return nil // Struct not found
}

func GetStructFromGenDecl(decl *ast.GenDecl) *ast.StructType {
	// Iterate over all the specs in the general declaration
	for _, spec := range decl.Specs {
		// Look for type specs (ast.TypeSpec)
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			// Check if the type is a struct type
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				return structType
			}
		}
	}

	return nil
}

func FindStructAst(pkg *packages.Package, structName string) *ast.StructType {
	// Traverse all the syntax trees (ASTs) in the loaded package
	for _, file := range pkg.Syntax {
		// Traverse all declarations in the AST file
		for _, decl := range file.Decls {
			// Look for type declarations (ast.GenDecl)
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				// Iterate over all the specs in the general declaration
				for _, spec := range genDecl.Specs {
					// Look for type specs (ast.TypeSpec)
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						// Check if the type is a struct type
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							// Match the struct's name with the input name
							if typeSpec.Name.Name == structName {
								return structType
							}
						}
					}
				}
			}
		}
	}
	return nil // Struct not found
}

func FindTypesStructInPackage(pkg *packages.Package, structName string) (*types.Struct, error) {
	typeName, err := LookupTypeName(pkg, structName)
	if err != nil {
		return nil, err
	}

	// Ensure the named type is a struct.
	structType, ok := typeName.Type().Underlying().(*types.Struct)
	if !ok {
		return nil, fmt.Errorf("%q is not a struct type", structName)
	}

	return structType, nil
}

func LookupTypeName(pkg *packages.Package, name string) (*types.TypeName, error) {
	if pkg.Types == nil || pkg.Types.Scope() == nil {
		return nil, fmt.Errorf("package %s does not have types or types scope", pkg.Name)
	}

	// Lookup the identifier in the package's scope.
	obj := pkg.Types.Scope().Lookup(name)
	if obj == nil {
		return nil, nil
	}

	// Check if the object is a type name.
	typeName, ok := obj.(*types.TypeName)
	if !ok {
		return nil, nil
	}

	return typeName, nil
}

func TryGetStructOrInterfaceKind(pkg *packages.Package, name string) (definitions.AstEntityKind, error) {
	typeName, err := LookupTypeName(pkg, name)
	if err != nil {
		return definitions.AstEntityNone, err
	}

	if _, isStruct := typeName.Type().Underlying().(*types.Struct); isStruct {
		return definitions.AstEntityStruct, nil
	}

	// Get the underlying type and check if it's an interface.
	if _, isInterface := typeName.Type().Underlying().(*types.Interface); isInterface {
		return definitions.AstEntityInterface, nil
	}

	return definitions.AstEntityUnknown, nil
}

func GetFieldTypeString(fieldType ast.Expr) string {
	switch t := fieldType.(type) {
	case *ast.Ident:
		return t.Name

	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", GetFieldTypeString(t.X), t.Sel.Name)

	case *ast.StarExpr:
		return fmt.Sprintf("*%s", GetFieldTypeString(t.X))

	case *ast.ArrayType:
		return fmt.Sprintf("[]%s", GetFieldTypeString(t.Elt))

	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", GetFieldTypeString(t.Key), GetFieldTypeString(t.Value))

	case *ast.ChanType:
		dir := ""
		if t.Dir == ast.SEND {
			dir = "send-only"
		} else if t.Dir == ast.RECV {
			dir = "receive-only"
		} else {
			dir = "bidirectional"
		}
		return fmt.Sprintf("Channel (%s, type: %s)", dir, GetFieldTypeString(t.Value))

	case *ast.FuncType:
		return "Function"

	case *ast.InterfaceType:
		return "Interface"

	case *ast.StructType:
		return "Struct"

	case *ast.Ellipsis:
		return fmt.Sprintf("Variadic (...%s)", GetFieldTypeString(t.Elt))

	case *ast.ParenExpr:
		return fmt.Sprintf("Parenthesized (%s)", GetFieldTypeString(t.X))

	default:
		return fmt.Sprintf("Unknown type (%T)", fieldType)
	}
}
