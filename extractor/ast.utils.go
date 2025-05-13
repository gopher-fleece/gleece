package extractor

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/gopher-fleece/gleece/definitions"
	"golang.org/x/tools/go/packages"
)

// IsFuncDeclReceiverForStruct determines if the given FuncDecl is a receiver for a struct with the given name
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

// DoesStructEmbedStruct tests the given structToCheck in the *ast.File `sourceFile`
// to determine whether it embeds struct `embeddedStructName` from package `embeddedStructFullPackageName`
func DoesStructEmbedStruct(
	sourceFile *ast.File,
	structToCheck *ast.StructType,
	embeddedStructFullPackageName string,
	embeddedStructName string,
) bool {
	aliasedImports := GetImportAliases(sourceFile)

	// Iterate over the struct fields to check for embeds
	for _, field := range structToCheck.Fields.List {
		switch fieldType := field.Type.(type) {
		case *ast.Ident:
			// If the type is just an Ident (simple struct type), check the name.
			if fieldType.Name == embeddedStructName {
				// IMPORTANT NOTE:
				// Believe there's an edge case here where detection may fail if both structs in from the same package.
				// This is not relevant to the standard use-case but we'll need to verify
				//
				//
				// If it's an Ident, check if it's a dot import or a direct match
				if isDotImported, _ := IsPackageDotImported(sourceFile, embeddedStructFullPackageName); isDotImported {
					return true
				}
			}
		case *ast.SelectorExpr:
			// If the type is a SelectorExpr (meaning it's a struct from another package), check the package and name
			if ident, ok := fieldType.X.(*ast.Ident); ok {
				// Compare the package name and struct name
				sourcePackage := aliasedImports[ident.Name]
				isCorrectPackage := sourcePackage == embeddedStructFullPackageName || IsAliasDefault(embeddedStructFullPackageName, ident.Name)
				if isCorrectPackage && fieldType.Sel.Name == embeddedStructName {
					return true
				}
			}
		}
	}
	return false
}

// GetDefaultPackageAlias returns the default import alias for the given file
func GetDefaultPackageAlias(file *ast.File) (string, error) {
	if file.Name != nil {
		return GetDefaultAlias(file.Name.Name), nil
	}
	return "", fmt.Errorf("source file does not have a name")
}

func GetFullPackageName(file *ast.File, fileSet *token.FileSet) (string, error) {
	// Get the file's full path using the fileSet
	position := fileSet.Position(file.Package)
	relativePath := position.Filename

	absFilePath, err := filepath.Abs(relativePath)
	if err != nil {
		// This is nearly impossible to break - filepath.Abs is extremely lenient.
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

func IsIdentInPackage(pkg *packages.Package, ident *ast.Ident) bool {
	return pkg.Types.Scope().Lookup(ident.Name) != nil
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

func GetTypeNameOrError(pkg *packages.Package, name string) (*types.TypeName, error) {
	typeName, err := LookupTypeName(pkg, name)
	if err != nil {
		return nil, err
	}

	if typeName == nil {
		return nil, fmt.Errorf("type '%s' was not found in package '%s'", name, pkg.Name)
	}

	if typeName.Type() == nil {
		// This failure is pretty much impossible to trigger without some black magic.
		// Even NeedTypes populates the type information and, if something was missing, it would've been caught by LookupTypeName
		return nil, fmt.Errorf("type '%s.%s' does not have Type() information", pkg.Name, name)
	}

	return typeName, nil
}

func GetEntityKind(pkg *packages.Package, name string) (definitions.AstNodeKind, error) {
	typeName, err := GetTypeNameOrError(pkg, name)
	if err != nil {
		return definitions.AstNodeKindNone, err
	}

	return GetEntityKindFromTypeName(typeName)
}

func GetEntityKindFromTypeName(typeName *types.TypeName) (definitions.AstNodeKind, error) {
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
		// This takes care of *slices* like '[]int'
		if t.Len == nil {
			return fmt.Sprintf("[]%s", GetFieldTypeString(t.Elt))
		}

		// This handles fixed-size arrays like '[3]int'
		if lit, ok := t.Len.(*ast.BasicLit); ok {
			return fmt.Sprintf("[%s]%s", lit.Value, GetFieldTypeString(t.Elt))
		}

		// And a fallback for weird edge cases
		return fmt.Sprintf("[?]%s", GetFieldTypeString(t.Elt))

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
