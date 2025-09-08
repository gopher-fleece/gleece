package gast

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"math"
	"path/filepath"
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
		fileName := "'N\\A'"
		if file.Name != nil && file.Name.Name != "" {
			fileName = file.Name.Name
		}
		return "", fmt.Errorf("could not determine full path for file %v", fileName)
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

func GetIdentFromExpr(expr ast.Expr) *ast.Ident {
	switch t := expr.(type) {
	case *ast.Ident:
		return t

	case *ast.StarExpr:
		return GetIdentFromExpr(t.X)

	case *ast.SelectorExpr:
		// Returns the selector's "Baz" (e.g. `other.Baz`)
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

type TypeSpecResolution struct {
	IsUniverse       bool
	TypeName         string
	DeclaringPackage *packages.Package
	DeclaringAstFile *ast.File
	TypeSpec         *ast.TypeSpec
	TypeNameIdent    *ast.Ident
	GenDecl          *ast.GenDecl
}

func (t TypeSpecResolution) String() string {
	if t.IsUniverse {
		return fmt.Sprintf("Universe type %s", t.TypeName)
	}

	return fmt.Sprintf("Type %s.%s", t.DeclaringPackage.PkgPath, t.TypeName)
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
) (TypeSpecResolution, error) {
	return ResolveTypeSpecFromExpr(declaringPkg, declaringFile, field.Type, pkgResolver)
}

// ResolveTypeSpecFromExpr resolves a type expression to its defining TypeSpec,
// if it originates from a user-defined type, otherwise returns information about the universe type.
func ResolveTypeSpecFromExpr(
	pkg *packages.Package,
	file *ast.File,
	expr ast.Expr,
	getPkg func(pkgPath string) (*packages.Package, error), // typically your package resolver
) (TypeSpecResolution, error) {
	ident := GetIdentFromExpr(expr)
	if ident == nil {
		return TypeSpecResolution{}, fmt.Errorf(
			"cannot resolve type: expression has no base identifier or is an unsupported type such as an inline struct",
		)
	}

	obj := pkg.TypesInfo.Uses[ident]
	if obj == nil {
		// Fall back to type scope if ident is dot-imported or not recorded in Uses
		obj = pkg.Types.Scope().Lookup(ident.Name)
	}

	if obj == nil {
		return TypeSpecResolution{}, fmt.Errorf("cannot resolve identifier '%s' in file %s", ident.Name, file.Name.Name)
	}

	typeName, ok := obj.(*types.TypeName)
	if !ok {
		return TypeSpecResolution{}, fmt.Errorf("resolved object is not a type: %T", obj)
	}

	// Universe type fallback
	if obj.Pkg() == nil && types.Universe.Lookup(obj.Name()) == obj {
		return TypeSpecResolution{
			TypeName:   obj.Name(),
			IsUniverse: true,
		}, nil
	}

	declPkg, err := getPkg(obj.Pkg().Path())
	if err != nil || declPkg == nil {
		return TypeSpecResolution{}, fmt.Errorf("could not locate declaring package: %s", obj.Pkg().Path())
	}

	// Search for the TypeSpec by matching name in declPkg.Syntax
	for _, declFile := range declPkg.Syntax {
		for _, decl := range declFile.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if typeSpec.Name.Name == typeName.Name() {
					return TypeSpecResolution{
						DeclaringPackage: declPkg,
						DeclaringAstFile: declFile,
						TypeSpec:         typeSpec,
						GenDecl:          genDecl,
						TypeName:         typeName.Name(),
						TypeNameIdent:    ident,
						IsUniverse:       false,
					}, nil
				}
			}
		}
	}

	return TypeSpecResolution{}, fmt.Errorf(
		"could not find TypeSpec for type '%s' in package '%s'",
		typeName.Name(),
		obj.Pkg().Path(),
	)
}

func GetAstFileName(fSet *token.FileSet, file *ast.File) string {
	return fSet.Position(file.Package).Filename
}

func ExtractConstValue(kind types.BasicKind, constVal *types.Const) any {
	switch kind {
	case types.String:
		return constant.StringVal(constVal.Val())
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
		if val, exact := constant.Int64Val(constVal.Val()); exact {
			return val
		}
	case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		if val, exact := constant.Uint64Val(constVal.Val()); exact {
			return val
		}
	case types.Float32, types.Float64:
		// For floats, we don't *really* case about 'exact' values - things get rounded to sane
		// precision and that's generally fine.
		// We DO want to reject on INF/NaN cause those do indicate something unexpected.
		// A bit of an ugly one.
		//
		// For INT types, we do want to throw non-exact values as they may indicate something major like an overflow
		val, _ := constant.Float64Val(constVal.Val())
		if math.IsInf(val, 0) || math.IsNaN(val) {
			return nil
		}
		return val
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

func IsEnumLike(pkg *packages.Package, spec *ast.TypeSpec) bool {
	// Must be an alias like: `type MyEnum string`
	_, ok := spec.Type.(*ast.Ident)
	if !ok {
		return false
	}

	// Look up the declared type
	obj := pkg.Types.Scope().Lookup(spec.Name.Name)
	typeName, ok := obj.(*types.TypeName)
	if !ok {
		return false
	}

	// Underlying type must be basic (string/int/etc)
	_, isBasic := typeName.Type().Underlying().(*types.Basic)
	if !isBasic {
		return false
	}

	// Look for constants in same package with that alias type
	for _, name := range pkg.Types.Scope().Names() {
		obj := pkg.Types.Scope().Lookup(name)
		constVal, ok := obj.(*types.Const)
		if !ok {
			continue
		}
		if types.Identical(constVal.Type(), typeName.Type()) {
			return true
		}
	}

	return false
}

func GetAstFileNameOrFallback(file *ast.File, fallback *string) string {
	getFallback := func(typePrefix string) string {
		if fallback != nil {
			return *fallback
		}
		return fmt.Sprintf("%s_FILE", strings.ToUpper(typePrefix))
	}

	if file == nil {
		return getFallback("NIL")
	}

	if file.Name == nil {
		return getFallback("UNNAMED")
	}

	if file.Name.Name == "" {
		return getFallback("UNNAMED")
	}

	return file.Name.Name
}

// MapDocListToCommentBlock converts a list of *ast.Comment into a CommentBlock.
// Assumes docList is the ordered list coming from a single ast.CommentGroup
// (i.e. comments attached to the same node/file).
//
// Coordinates:
//   - Lines and columns are 0-based.
//   - EndLine/EndCol are exclusive (half-open interval).
func MapDocListToCommentBlock(docList []*ast.Comment, fileSet *token.FileSet) CommentBlock {
	if len(docList) == 0 {
		return CommentBlock{
			Comments: []CommentNode{},
			FileName: "",
			Range: common.ResolvedRange{
				StartLine: 0,
				StartCol:  0,
				EndLine:   0,
				EndCol:    0,
			},
		}
	}

	comments := make([]CommentNode, 0, len(docList))
	var fileName string

	// Build per-comment nodes
	for i, c := range docList {
		var startPos, endPos token.Position
		if fileSet != nil {
			startPos = fileSet.Position(c.Pos())
			endPos = fileSet.Position(c.End())
		}
		// normalize to 0-based; if fileSet==nil these will be zero
		startLine := max(0, startPos.Line-1)
		startCol := max(0, startPos.Column-1)
		endLine := max(0, endPos.Line-1)
		endCol := max(0, endPos.Column-1)

		if fileName == "" && startPos.Filename != "" {
			fileName = startPos.Filename
		}

		comments = append(comments, CommentNode{
			Text: c.Text,
			Position: CommentPosition{
				StartLine: startLine,
				StartCol:  startCol,
				EndLine:   endLine,
				EndCol:    endCol,
			},
			Index: i,
		})
	}

	// derive block-level range from first and last comment
	first := comments[0].Position
	last := comments[len(comments)-1].Position

	return CommentBlock{
		Comments: comments,
		FileName: fileName,
		Range: common.ResolvedRange{
			StartLine: first.StartLine,
			StartCol:  first.StartCol,
			EndLine:   last.EndLine,
			EndCol:    last.EndCol,
		},
	}
}

// GetCommentsFromNode extracts comments attached directly to node (like FuncDecl.Doc)
// and returns a CommentBlock with resolved positions (via fileSet).
// If fileSet is nil the Position fields will be zero-valued (normalized to 0).
func GetCommentsFromNode(node ast.Node, fileSet *token.FileSet) CommentBlock {
	if node == nil {
		return CommentBlock{
			Comments: []CommentNode{},
		}
	}

	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Doc != nil {
			return MapDocListToCommentBlock(n.Doc.List, fileSet)
		}
	case *ast.GenDecl:
		if n.Doc != nil {
			return MapDocListToCommentBlock(n.Doc.List, fileSet)
		}
	case *ast.TypeSpec:
		if n.Doc != nil {
			return MapDocListToCommentBlock(n.Doc.List, fileSet)
		}
	case *ast.ValueSpec:
		if n.Doc != nil {
			return MapDocListToCommentBlock(n.Doc.List, fileSet)
		}
	case *ast.Field:
		if n.Doc != nil {
			return MapDocListToCommentBlock(n.Doc.List, fileSet)
		}
	}

	return CommentBlock{
		Comments: []CommentNode{},
	}
}

// GetCommentsFromTypeSpec behaves like your previous GetCommentsFromTypeSpec but
// returns a CommentBlock and resolves positions. It prioritizes comments on the
// TypeSpec over the owning GenDecl (same semantics as your original helper).
func GetCommentsFromTypeSpec(
	typeSpec *ast.TypeSpec,
	owningGenDecl *ast.GenDecl,
	fileSet *token.FileSet,
) CommentBlock {
	var commentSource *ast.CommentGroup
	if typeSpec != nil && typeSpec.Doc != nil {
		commentSource = typeSpec.Doc
	} else if owningGenDecl != nil && owningGenDecl.Doc != nil {
		commentSource = owningGenDecl.Doc
	}

	if commentSource != nil {
		return MapDocListToCommentBlock(commentSource.List, fileSet)
	}

	return CommentBlock{
		Comments: []CommentNode{},
	}
}
