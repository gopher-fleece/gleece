package extractor

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"path/filepath"
	"strconv"
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

func IsAliasDefault(fullPackageName string, alias string) bool {
	packageName := GetDefaultAlias(fullPackageName)
	return alias == packageName
}

func GetDefaultAlias(fullyQualifiedPackage string) string {
	segments := strings.Split(fullyQualifiedPackage, "/")
	last := segments[len(segments)-1]

	// Check if last segment is a version (v2, v3, etc.)
	if strings.HasPrefix(last, "v") && len(last) > 1 && last[1] >= '0' && last[1] <= '9' {
		// If so, take the second-last segment
		if len(segments) > 1 {
			return segments[len(segments)-2]
		}
	}
	return last
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

func GetCommentsFromIdent(files *token.FileSet, file *ast.File, ident *ast.Ident) []string {
	if ident.Obj == nil || ident.Obj.Decl == nil {
		return []string{}
	}

	var docs *ast.CommentGroup
	switch expr := ident.Obj.Decl.(type) {
	case *ast.TypeSpec:
		if expr.Doc != nil && expr.Doc.List != nil && len(expr.Doc.List) > 0 {
			docs = expr.Doc
		} else {
			// Look for the parent GenDecl and grab the comments of it. Go works in mysterious ways.
			decl := FindGenDeclByIdent(files, file, ident)
			if decl != nil && decl.Doc != nil && decl.Doc.List != nil && len(decl.Doc.List) > 0 {
				docs = decl.Doc
			}
		}
	case *ast.FuncDecl:
		docs = expr.Doc
	case *ast.Field:
		docs = expr.Doc
	}

	if docs != nil && docs.List != nil && len(docs.List) > 0 {
		return MapDocListToStrings(docs.List)
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
	comments := GetCommentsFromIdent(fileSet, file, ident)

	meta := definitions.TypeMetadata{
		Name:        ident.Name,
		Description: FindAndExtract(comments, "@Description"),
	}

	if IsUniverseType(ident.Name) {
		// The identifier is a member of the universe, e.g. 'error'.
		// Nothing to do here. Leave the package empty so the downstream generator knows no import/alias is needed
		meta.IsUniverseType = true
		meta.Import = definitions.ImportTypeNone
		meta.EntityKind = definitions.AstNodeKindUnknown
		return meta, nil
	}

	relevantPkg := isIdentFromDotImportedPackage(file, packages, ident)
	if relevantPkg != nil {
		// The identifier is a type from a dot imported package
		meta.Import = definitions.ImportTypeDot
		meta.FullyQualifiedPackage = relevantPkg.PkgPath
		meta.DefaultPackageAlias = relevantPkg.Name

		typeName, err := LookupTypeName(relevantPkg, ident.Name)
		if err != nil {
			return meta, err
		}

		if typeName == nil {
			return meta, fmt.Errorf("type '%s' was not found in package %s", ident.Name, relevantPkg.Name)
		}

		kind, err := TryGetStructOrInterfaceKind(relevantPkg, ident.Name)
		if err != nil {
			return meta, err
		}
		meta.EntityKind = kind
		if meta.EntityKind == definitions.AstNodeKindAlias {
			aliasMetadata, err := ExtractAliasType(relevantPkg, typeName, ident)
			if err != nil {
				return meta, err
			}
			meta.AliasMetadata = aliasMetadata
		}

	} else {
		// If we've gotten here, the ident is a locally defined entity;
		//
		// Were it a an aliased import, it've been resolved by GetTypeMetaBySelectorExpr.
		// For dot-imports, we'd have flowed to the above 'if'.
		currentPackageName, err := GetFullPackageName(file, fileSet)
		if err != nil {
			return meta, err
		}

		pkg := FilterPackageByFullName(packages, currentPackageName)
		if pkg == nil {
			return meta, fmt.Errorf("could not find package '%s' in the given list of packages", currentPackageName)
		}

		typeName, err := LookupTypeName(pkg, ident.Name)
		if err != nil {
			return meta, err
		}

		if (typeName == nil) || (typeName.Type() == nil) {
			return meta, fmt.Errorf("could not find type '%s' in package '%s', are you sure it's included in the 'commonConfig->controllerGlobs' search paths?", ident.Name, currentPackageName)
		}

		// Verify the identifier does in fact exist in the current package.
		// Not strictly needed but helps with safety.
		exists, entityKind, err := DoesTypeOrInterfaceExistInPackage(typeName, ident)
		if err != nil {
			return meta, err
		}

		if !exists {
			return meta, fmt.Errorf("identifier %s does not correlate to a type or interface in package %s", ident.Name, currentPackageName)
		}

		meta.Import = definitions.ImportTypeNone
		meta.FullyQualifiedPackage = currentPackageName
		meta.DefaultPackageAlias = GetDefaultAlias(currentPackageName)
		meta.EntityKind = entityKind

		if entityKind == definitions.AstNodeKindAlias {
			aliasMetadata, err := ExtractAliasType(pkg, typeName, ident)
			if err != nil {
				return meta, err
			}
			meta.AliasMetadata = aliasMetadata
		}
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

	comments := GetCommentsFromIdent(fileSet, file, selector.Sel)
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

func ExtractAliasType(pkg *packages.Package, typeName *types.TypeName, ident *ast.Ident) (*definitions.AliasMetadata, error) {
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

func DoesTypeOrInterfaceExistInPackage(
	typeName *types.TypeName,
	ident *ast.Ident,
) (bool, definitions.AstNodeKind, error) {
	if _, isStruct := typeName.Type().Underlying().(*types.Struct); isStruct {
		return true, definitions.AstNodeKindStruct, nil
	}

	// Get the underlying type and check if it's an interface.
	if _, isInterface := typeName.Type().Underlying().(*types.Interface); isInterface {
		return true, definitions.AstNodeKindInterface, nil
	}

	// Check if that is an alias of a basic type (string, int, bool, etc.)
	if _, isBasicType := typeName.Type().Underlying().(*types.Basic); isBasicType {
		return true, definitions.AstNodeKindAlias, nil
	}

	return true, definitions.AstNodeKindUnknown, nil
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

func FindGenDeclByIdent(fileSet *token.FileSet, file *ast.File, ident *ast.Ident) *ast.GenDecl {
	var decl *ast.GenDecl

	// Walk the AST to locate the struct declaration
	ast.Inspect(file, func(n ast.Node) bool {
		// Look for a general declaration (const, var, type, etc.)
		if gd, ok := n.(*ast.GenDecl); ok && gd.Tok == token.TYPE {
			for _, spec := range gd.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					if ts.Name.Name == ident.Name {
						decl = gd
						return false // Stop traversal once found
					}
				}
			}
		}
		return true
	})

	return decl
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

func FindTypesStructInPackage(pkg *packages.Package, structName string) (*types.Struct, error) {
	typeName, err := LookupTypeName(pkg, structName)
	if typeName == nil || err != nil {
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

func TryGetStructOrInterfaceKind(pkg *packages.Package, name string) (definitions.AstNodeKind, error) {
	typeName, err := LookupTypeName(pkg, name)
	if err != nil {
		return definitions.AstNodeKindNone, err
	}

	if typeName == nil {
		return definitions.AstNodeKindUnknown, fmt.Errorf("type '%s' was not found in package %s", name, pkg.Name)
	}

	if _, isStruct := typeName.Type().Underlying().(*types.Struct); isStruct {
		return definitions.AstNodeKindStruct, nil
	}

	// Get the underlying type and check if it's an interface.
	if _, isInterface := typeName.Type().Underlying().(*types.Interface); isInterface {
		return definitions.AstNodeKindInterface, nil
	}

	// Check if that is an alias of a basic type (string, int, bool, etc.)
	if typeName.IsAlias() {
		return definitions.AstNodeKindAlias, nil
	}

	if _, isBasicType := typeName.Type().Underlying().(*types.Basic); isBasicType {
		return definitions.AstNodeKindAlias, nil
	}

	return definitions.AstNodeKindUnknown, nil
}

func GetUnderlyingTypeName(t types.Type) string {
	switch underlying := t.(type) {
	case *types.Basic:
		return underlying.Name() // e.g., "int64", "string"
	case *types.Named:
		return underlying.Obj().Name() // e.g., another named type
	case *types.Pointer:
		return "*" + GetUnderlyingTypeName(underlying.Elem()) // Handle pointer types
	case *types.Slice:
		return "[]" + GetUnderlyingTypeName(underlying.Elem()) // Handle slices
	case *types.Array:
		return fmt.Sprintf("[%d]%s", underlying.Len(), GetUnderlyingTypeName(underlying.Elem())) // Handle arrays
	case *types.Map:
		return fmt.Sprintf("map[%s]%s", GetUnderlyingTypeName(underlying.Key()), GetUnderlyingTypeName(underlying.Elem())) // Handle maps
	case *types.Chan:
		return "chan " + GetUnderlyingTypeName(underlying.Elem()) // Handle channels
	case *types.Signature:
		return "func(...)"
	default:
		return t.String() // Fallback
	}
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

func GetNodeKind(fieldType ast.Expr) definitions.AstNodeKind {
	switch fieldType.(type) {
	case *ast.Ident:
		return definitions.AstNodeKindIdent
	case *ast.SelectorExpr:
		return definitions.AstNodeKindSelector
	case *ast.StarExpr:
		return definitions.AstNodeKindPointer
	case *ast.ArrayType:
		return definitions.AstNodeKindArray
	case *ast.MapType:
		return definitions.AstNodeKindMap
	case *ast.ChanType:
		return definitions.AstNodeKindChannel
	case *ast.FuncType:
		return definitions.AstNodeKindFunction
	case *ast.InterfaceType:
		return definitions.AstNodeKindInterface
	case *ast.StructType:
		return definitions.AstNodeKindStruct
	case *ast.Ellipsis:
		return definitions.AstNodeKindVariadic
	case *ast.ParenExpr:
		return definitions.AstNodeKindParenthesis
	default:
		return definitions.AstNodeKindUnknown
	}
}

func IsAliasType(named *types.Named) bool {
	// First, check if it was declared using the alias syntax.
	if named.Obj().IsAlias() {
		return true
	}
	// For our purposes, if the underlying type is not a struct or interface,
	// we might want to treat it as an alias-like type.
	switch named.Underlying().(type) {
	case *types.Struct, *types.Interface:
		return false
	default:
		return true
	}
}

func GetIterableElementType(iterable types.Type) string {
	// Handle array or slice types
	var elemType types.Type
	if arr, ok := iterable.(*types.Array); ok {
		elemType = arr.Elem()
	} else if slice, ok := iterable.(*types.Slice); ok {
		elemType = slice.Elem()
	}

	// Check if the element type is a named type (enum/alias)
	if named, ok := elemType.(*types.Named); ok {
		// For arrays of named types, format as []TypeName
		return "[]" + named.Obj().Name()
	}

	// For arrays of primitive types
	return iterable.String() // <<< CHECK THIS WORKS!

}
