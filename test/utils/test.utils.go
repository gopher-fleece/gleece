package utils

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"

	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/cmd/arguments"
	"github.com/gopher-fleece/gleece/definitions"
	. "github.com/onsi/ginkgo/v2"
	"golang.org/x/tools/go/packages"
)

func GetControllersAndModelsOrFail() ([]definitions.ControllerMetadata, []definitions.StructMetadata, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		Fail(fmt.Sprintf("Could not determine process working directory - %v", err))
	}

	configPath := filepath.Join(cwd, "gleece.test.config.json")
	_, controllers, flatModels, hasStdError, err := cmd.GetConfigAndMetadata(arguments.CliArguments{ConfigPath: configPath})
	if err != nil {
		Fail(fmt.Sprintf("Could not generate routes - %v", err))
	}
	return controllers, flatModels.Structs, hasStdError
}

func GetAbsPathByRelativeOrFail(relativePath string) string {
	cwd, err := os.Getwd()
	if err != nil {
		Fail(fmt.Sprintf("Could not determine process working directory - %v", err))
	}

	return filepath.Join(cwd, relativePath)
}

func LoadPackageOrFail(fullName string, loadMode packages.LoadMode) *packages.Package {
	cfg := &packages.Config{Mode: loadMode}
	matchingPackages, err := packages.Load(cfg, fullName)
	if err != nil || len(matchingPackages) <= 0 {
		FailWithTestCodeError(fmt.Sprintf("Could not load package '%s' for testing", fullName))
	}

	return matchingPackages[0]
}

func LoadGleecePackageOrFail(loadMode packages.LoadMode) *packages.Package {
	return LoadPackageOrFail("github.com/gopher-fleece/gleece", loadMode)
}

func GetFunctionFromPackageOrFail(pkg *packages.Package, name string) *ast.FuncDecl {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if ok && funcDecl.Name.Name == name {
				return funcDecl
			}
		}
	}

	FailWithTestCodeError(fmt.Sprintf("Could not find function '%s' in package '%s'", name, pkg.Name))
	return nil
}

// FailWithTestCodeError Fails the test with a "This is a test issue, not a code issue" message.
// Used to signify something went wrong with test setup or such
func FailWithTestCodeError(message string) {
	Fail(fmt.Sprintf("%s. This indicates a test issue, not a code issue", message))
}

func GetAstFieldByNameOrFail(pkg *packages.Package, structName string, fieldName string) *ast.Field {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != structName {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					FailWithTestCodeError(fmt.Sprintf("Type %q is not a struct", structName))
					return nil
				}

				for _, field := range structType.Fields.List {
					for _, name := range field.Names {
						if name.Name == fieldName {
							return field
						}
					}
				}

				FailWithTestCodeError(fmt.Sprintf("Field %q not found in struct %q", fieldName, structName))
				return nil
			}
		}
	}

	FailWithTestCodeError(fmt.Sprintf("Struct %q not found in package", structName))
	return nil
}
