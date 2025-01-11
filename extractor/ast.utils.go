package extractor

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

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
				if IsPackageDotImported(sourceFile, structFullPackageName) {
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

func IsPackageDotImported(file *ast.File, packageName string) bool {
	for _, imp := range file.Imports {
		// Check if it's a dot import (imp.Name == nil) and if the package path matches the expected package name
		if imp.Name != nil && imp.Path != nil && strings.Trim(imp.Path.Value, `"`) == packageName {
			// Ensure that the struct name is the same as the dot-imported struct
			// Since we know it's a dot import, any struct with this name should be from the expected package
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
