package gast

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gopher-fleece/gleece/common"
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
		return GetDefaultPkgAliasByName(file.Name.Name), nil
	}
	return "", fmt.Errorf("source file does not have a name")
}

func GetFileFullPath(file *ast.File, fileSet *token.FileSet) (string, error) {
	if file == nil || fileSet == nil {
		return "", fmt.Errorf("GetFileFullPath was provided nil file or fileSet")
	}

	// Get the file's full path using the fileSet
	position := fileSet.Position(file.Package)
	relativePath := position.Filename
	if relativePath == "" {
		return "", fmt.Errorf("could not determine full path for file %v", file.Name.Name)
	}

	return filepath.Abs(relativePath)
}

func GetFullPackageName(file *ast.File, fileSet *token.FileSet) (string, error) {
	absFilePath, err := GetFileFullPath(file, fileSet)
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
	packageName := GetDefaultPkgAliasByName(fullPackageName)
	return alias == packageName
}

func GetDefaultPkgAliasByName(pkgPath string) string {
	segments := strings.Split(pkgPath, "/")
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

// FindDeclNodeByIdent looks up ident.Name in pkg.Types.Scope(), and if found,
// finds the AST node (FuncDecl, TypeSpec, or ValueSpec) whose Pos() contains
// the object’s position. Returns nil if nothing matches.
func FindDeclNodeByIdent(pkg *packages.Package, ident *ast.Ident) ast.Node {
	if pkg.Types == nil || pkg.Types.Scope() == nil {
		return nil
	}
	obj := pkg.Types.Scope().Lookup(ident.Name)
	if obj == nil {
		return nil
	}
	pos := obj.Pos()
	// Find which AST file contains this position
	for _, f := range pkg.Syntax {
		if pos < f.Pos() || pos >= f.End() {
			continue
		}
		// Walk only that file
		return FindDeclAtPos(f, pos)
	}
	return nil
}

// FindDeclForPos returns the *ast.FuncDecl, *ast.TypeSpec, or *ast.ValueSpec
// whose source-range encloses pos, or nil if none.
func FindDeclAtPos(file *ast.File, pos token.Pos) ast.Node {
	var found ast.Node
	ast.Inspect(file, func(n ast.Node) bool {
		if found != nil || n == nil {
			return false
		}
		if pos < n.Pos() || pos >= n.End() {
			return false
		}
		switch n := n.(type) {
		case *ast.FuncDecl, *ast.TypeSpec, *ast.ValueSpec:
			found = n
			return false
		}
		return true
	})
	return found
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

// FindTypeSpecByIdent walks 'file' and returns the *ast.TypeSpec
// whose Name matches ident.Name (or nil if not found).
func FindTypeSpecByIdent(file *ast.File, ident *ast.Ident) *ast.TypeSpec {
	var found *ast.TypeSpec
	ast.Inspect(file, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		if ts.Name.Name == ident.Name {
			found = ts
			return false // stop walking
		}
		return true
	})
	return found
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

func LookupIdentForObject(pkg *packages.Package, obj types.Object) *ast.Ident {
	for ident, definedObj := range pkg.TypesInfo.Defs {
		if definedObj == obj {
			return ident
		}
	}
	return nil
}

func FindContainingFile(pkg *packages.Package, ident *ast.Ident) *ast.File {
	for _, file := range pkg.Syntax {
		if ident.Pos() >= file.Pos() && ident.Pos() <= file.End() {
			return file
		}
	}
	return nil
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

func GetSymbolKind(pkg *packages.Package, name string) (common.SymKind, error) {
	typeName, err := GetTypeNameOrError(pkg, name)
	if err != nil {
		return common.SymKindUnknown, err
	}

	return GetSymbolKindFromObject(typeName), nil
}

func IsBasic(t types.Type) bool {
	_, ok := t.(*types.Basic)
	return ok
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
		switch t.Dir {
		case ast.SEND:
			dir = "send-only"
		case ast.RECV:
			dir = "receive-only"
		default:
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

// FindNamedTypePackage returns the *types.Package where the first underlying *types.Named is defined.
// Returns nil if no such package is found (e.g. built-in types, unnamed structs, etc).
func GetPackageOwnerOfType(t types.Type) *types.Package {
	for {
		switch underlyingType := t.(type) {
		case *types.Slice:
			t = underlyingType.Elem()
		case *types.Array:
			t = underlyingType.Elem()
		case *types.Pointer:
			t = underlyingType.Elem()
		case *types.Named:
			obj := underlyingType.Obj()
			if obj != nil {
				return obj.Pkg()
			}
			return nil
		default:
			// No deeper types to drill into; give up
			return nil
		}
	}
}

// MapDocListToStrings converts a list of comment nodes (ast.Comment) to a string array
func MapDocListToStrings(docList []*ast.Comment) []string {
	var result []string
	for _, comment := range docList {
		result = append(result, comment.Text)
	}
	return result
}

func GetSymbolKindFromObject(obj types.Object) common.SymKind {
	switch o := obj.(type) {

	case *types.PkgName:
		return common.SymKindPackage

	case *types.Const:
		return common.SymKindConstant

	case *types.Var:
		if o.IsField() {
			return common.SymKindField
		}
		// Parameters are Vars too, but we must distinguish
		// Parameters live in function signatures, not top-level scope
		if isParameter(o) {
			return common.SymKindParameter
		}
		return common.SymKindVariable

	case *types.Func:
		if sig, ok := o.Type().(*types.Signature); ok && sig.Recv() != nil {
			return common.SymKindReceiver
		}
		return common.SymKindFunction

	case *types.TypeName:
		if o.IsAlias() {
			return common.SymKindAlias
		}
		switch o.Type().Underlying().(type) {
		case *types.Struct:
			return common.SymKindStruct
		case *types.Interface:
			return common.SymKindInterface
		default:
			return common.SymKindAlias
		}
	}

	return common.SymKindUnknown
}

func isParameter(v *types.Var) bool {
	// Parameters do not have a parent scope associated with a file/package block
	if v.Parent() != nil {
		switch v.Parent().Parent() {
		case nil:
			// Probably not a parameter
			return false
		default:
			// Heuristic: check if the parent is a function Signature
			// Unfortunately, Go's type checker doesn't expose a perfect way to trace this
			// So this may need to be called only in known contexts (like while walking a Signature)
			return true
		}
	}
	return false
}

func GetIdentFromExpr(expr ast.Expr) *ast.Ident {
	switch t := expr.(type) {
	case *ast.Ident:
		return t

	case *ast.StarExpr:
		return GetIdentFromExpr(t.X)

	case *ast.SelectorExpr:
		// Returns the selector's "Baz" (e.g. `otherpkg.Baz`)
		return t.Sel

	case *ast.ArrayType:
		return GetIdentFromExpr(t.Elt)

	case *ast.MapType:
		// optional: you could return key or value ident
		return GetIdentFromExpr(t.Value)

	case *ast.ChanType:
		return GetIdentFromExpr(t.Value)

	case *ast.FuncType:
		// function types have no identifier
		return nil
	}

	return nil
}

// FindTypeSpecInPackage finds the AST TypeSpec for the given type name
// in the provided packages.Package. It returns the *ast.TypeSpec and
// the *ast.File it was found in, or (nil, nil) if not found.
func FindTypeSpecInPackage(pkg *packages.Package, typeName string) (*ast.TypeSpec, *ast.File) {
	for _, f := range pkg.Syntax {
		// fast-path: if the file’s token.File doesn’t contain any
		// decls that start near “typeName”, we could skip—but
		// for simplicity we just scan all TYPE decls here.
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if ts.Name.Name == typeName {
					return ts, f
				}
			}
		}
	}
	return nil, nil
}

type FieldTypeSpecResolution struct {
	IsUniverse       bool
	TypeName         string
	DeclaringPackage *packages.Package
	DeclaringAstFile *ast.File
	TypeSpec         *ast.TypeSpec
}

// ResolveTypeSpecFromField Resolves type information from the given field.
// Returns the declaring *packages.Package, *ast.File and the associated *ast.TypeSpec
// Has 3 possible outcomes:
//
// * Returns all of the above (the field references a concrete type somewhere)
// * Returns 4 nils - the field's type is a 'Universe' one
// * Returns an error
func ResolveTypeSpecFromField(
	declaringPkg *packages.Package, // <<<< Used only for locally defined type references
	declaringFile *ast.File,
	field *ast.Field,
	pkgResolver func(pkgPath string) (*packages.Package, error),
) (FieldTypeSpecResolution, error) {
	expr := UnwrapFirstNamed(field.Type)
	if expr == nil {
		return FieldTypeSpecResolution{}, fmt.Errorf("could not resolve named type from field")
	}

	var ident *ast.Ident
	var pkgPath string
	var err error

	switch t := expr.(type) {
	case *ast.Ident:
		ident = t
		pkgPath, err = ResolveUnqualifiedIdentPackage(declaringFile, ident)
		if err != nil {
			return FieldTypeSpecResolution{}, err
		}

	case *ast.SelectorExpr:
		ident = t.Sel
		pkgPath, err = ResolveImportPathForSelector(declaringFile, t)
		if err != nil {
			return FieldTypeSpecResolution{}, err
		}

	default:
		return FieldTypeSpecResolution{}, fmt.Errorf("unsupported resolved expression: %T", expr)
	}

	if ident == nil {
		return FieldTypeSpecResolution{}, fmt.Errorf("could not extract identifier")
	}

	// If still blank, check whether the type is a universe one.
	if pkgPath == "" {
		if IsUniverseType(ident.Name) {
			// Universe types are a special case
			return FieldTypeSpecResolution{IsUniverse: true, TypeName: ident.Name}, nil
		}
		// Fallback to the field's declaring package
		pkgPath = declaringPkg.PkgPath
	}

	pkg, err := pkgResolver(pkgPath)
	if err != nil {
		return FieldTypeSpecResolution{}, fmt.Errorf("failed to resolve package %q: %w", pkgPath, err)
	}

	spec, file := FindTypeSpecInPackage(pkg, ident.Name)
	if spec == nil || file == nil {
		return FieldTypeSpecResolution{}, fmt.Errorf("type %q not found in package %q", ident.Name, pkgPath)
	}

	return FieldTypeSpecResolution{
		IsUniverse:       false,
		TypeName:         ident.Name,
		DeclaringPackage: pkg,
		DeclaringAstFile: file,
		TypeSpec:         spec,
	}, nil
}

// unwrapFirstNamed walks through container expressions (e.g. *T, []T, map[K]V)
// and returns the first inner expression that is either *ast.Ident or *ast.SelectorExpr.
func UnwrapFirstNamed(expr ast.Expr) ast.Expr {
	for {
		switch t := expr.(type) {
		case *ast.ArrayType:
			expr = t.Elt
		case *ast.MapType:
			expr = t.Value
		case *ast.ChanType:
			expr = t.Value
		case *ast.StarExpr:
			// Only unwrap star if inner type is NOT a SelectorExpr
			if _, isSelector := t.X.(*ast.SelectorExpr); isSelector {
				// stop here — we need the selector to resolve the package alias
				return expr
			}
			expr = t.X
		default:
			return expr
		}
	}
}

func ResolveImportPathForSelector(source *ast.File, sel *ast.SelectorExpr) (string, error) {
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return "", fmt.Errorf("selector base is not an ident: %T", sel.X)
	}

	alias := ident.Name

	for _, imp := range source.Imports {
		rawPath, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}

		// Handle explicit alias
		if imp.Name != nil && imp.Name.Name == alias {
			return rawPath, nil
		}

		// Handle default alias
		if imp.Name == nil {
			parts := strings.Split(rawPath, "/")
			if len(parts) > 0 && parts[len(parts)-1] == alias {
				return rawPath, nil
			}
		}
	}

	return "", fmt.Errorf("no matching import for alias %q", alias)
}

func ResolveUnqualifiedIdentPackage(source *ast.File, ident *ast.Ident) (string, error) {
	for _, imp := range source.Imports {
		if imp.Name != nil && imp.Name.Name == "." {
			rawPath, err := strconv.Unquote(imp.Path.Value)
			if err == nil {
				return rawPath, nil // dot-import wins
			}
		}
	}

	// If not dot-imported, assume it's local
	// We can't resolve the actual import path from here alone, so:
	return "", nil // Signal: try local package (caller must patch it later)
}

func GetAstFileName(fSet *token.FileSet, file *ast.File) string {
	return fSet.Position(file.Package).Filename
}

func ExtractConstValue(kind types.BasicKind, constVal *types.Const) any {
	switch kind {
	case types.String:
		return constant.StringVal(constVal.Val())
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
		if val, ok := constant.Int64Val(constVal.Val()); ok {
			return val
		}
	case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		if val, ok := constant.Uint64Val(constVal.Val()); ok {
			return val
		}
	case types.Float32, types.Float64:
		if val, ok := constant.Float64Val(constVal.Val()); ok {
			return val
		}
	case types.Bool:
		return constant.BoolVal(constVal.Val())
	}
	return nil
}

func FindConstSpecNode(pkg *packages.Package, constName string) ast.Node {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.CONST {
				continue
			}
			for _, spec := range genDecl.Specs {
				valSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range valSpec.Names {
					if name.Name == constName {
						return valSpec
					}
				}
			}
		}
	}
	return nil
}

// GetCommentsFromNode extracts comments attached directly to the AST node (like FuncDecl.Doc).
// Returns a slice of comment texts or empty if none.
func GetCommentsFromNode(node ast.Node) []string {
	if node == nil {
		return nil
	}

	// Many node types have a Doc field of type *ast.CommentGroup.
	// We'll use a type switch to get the Doc comment group.
	switch n := node.(type) {
	case *ast.FuncDecl:
	case *ast.GenDecl:
	case *ast.TypeSpec:
	case *ast.ValueSpec:
	case *ast.Field:
		if n.Doc != nil {
			return MapDocListToStrings(n.Doc.List)
		}
	}

	return nil
}

// GetCommentsFromTypeSpec gets comments for a given TypeSpec.
// Similar to GetCommentsFromNode and GetCommentsFromIdent but also considers the given GenDecl whilst prioritizing comments
// on the type itself
func GetCommentsFromTypeSpec(typeSpec *ast.TypeSpec, owningGenDecl *ast.GenDecl) []string {
	// Comments are usually located on the nearest GenDecl but may also be inlined on the struct itself
	var commentSource *ast.CommentGroup
	if typeSpec.Doc != nil {
		commentSource = typeSpec.Doc
	} else {
		commentSource = owningGenDecl.Doc
	}

	// Do we want to fail if there are no attributes on the controller?
	if commentSource != nil {
		return MapDocListToStrings(commentSource.List)
	}

	return []string{}
}

func DoesStructEmbedType(pkg *packages.Package, structName string, embeddedStructFullPackage string, embeddedStructName string) (bool, error) {
	// Look for the struct definition in the package
	obj := pkg.Types.Scope().Lookup(structName)
	if obj == nil {
		return false, fmt.Errorf("struct '%s' not found in package '%s'", structName, pkg.PkgPath)
	}

	// Ensure it's a named type (i.e., a struct)
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return false, fmt.Errorf("type '%s' is not a named type", structName)
	}

	// Get the underlying struct type
	structType, ok := named.Underlying().(*types.Struct)
	if !ok {
		return false, fmt.Errorf("type '%s' is not a struct", structName)
	}

	// Check if any field in the struct embeds the target type
	var expectedFullyQualifiedName string
	if len(embeddedStructFullPackage) > 0 {
		expectedFullyQualifiedName = embeddedStructFullPackage + "." + embeddedStructName
	} else {
		expectedFullyQualifiedName = embeddedStructName
	}

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		if field.Embedded() {
			// Check if the field type matches the embedded struct we are looking for
			fieldType := field.Type().String()
			// Check if the type matches the embedded struct (full package path + type name)
			if fieldType == expectedFullyQualifiedName {
				return true, nil
			}
		}
	}

	return false, nil
}
