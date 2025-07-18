package utils

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"

	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/cmd/arguments"
	"github.com/gopher-fleece/gleece/core/pipeline"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	. "github.com/onsi/ginkgo/v2"
	"golang.org/x/tools/go/packages"
)

func GetMetadataByRelativeConfig(relativeConfigPath string) ([]definitions.ControllerMetadata, []definitions.StructMetadata, bool, error) {
	_, controllers, modelsList, hasStdError, err := GetConfigAndMetadataOrFail(relativeConfigPath)
	return controllers, modelsList, hasStdError, err
}

func GetConfigAndMetadataOrFail(relativeConfigPath string) (
	*definitions.GleeceConfig,
	[]definitions.ControllerMetadata,
	[]definitions.StructMetadata,
	bool,
	error,
) {
	config, controllers, flatModels, hasStdError, err := cmd.GetConfigAndMetadata(
		arguments.CliArguments{
			ConfigPath: constructFullPathOrFail(relativeConfigPath, true),
		},
	)

	modelsList := []definitions.StructMetadata{}
	if flatModels != nil && len(flatModels.Structs) > 0 {
		modelsList = flatModels.Structs
	}

	return config, controllers, modelsList, hasStdError, err
}

func GetDefaultConfigAndMetadataOrFail() (
	*definitions.GleeceConfig,
	[]definitions.ControllerMetadata,
	[]definitions.StructMetadata,
	bool,
	error,
) {
	return GetConfigAndMetadataOrFail("gleece.test.config.json")
}

func GetMetadataByRelativeConfigOrFail(relativeConfigPath string) ([]definitions.ControllerMetadata, []definitions.StructMetadata, bool) {

	controllers, modelsList, hasStdError, generationErr := GetMetadataByRelativeConfig(relativeConfigPath)

	if generationErr != nil {
		Fail(fmt.Sprintf("Could not generate routes - %v", generationErr))
	}
	return controllers, modelsList, hasStdError
}

func GetControllersAndModelsOrFail() ([]definitions.ControllerMetadata, []definitions.StructMetadata, bool) {
	return GetMetadataByRelativeConfigOrFail("gleece.test.config.json")
}

func constructFullPathOrFail(relativePath string, failIfNotExists bool) string {
	cwd, err := os.Getwd()
	if err != nil {
		Fail(fmt.Sprintf("Could not determine process working directory - %v", err))
	}

	fullPath := filepath.Join(cwd, relativePath)

	if failIfNotExists && !FileOrFolderExists(fullPath) {
		Fail(fmt.Sprintf("Path %s does not exist", fullPath))
	}

	return fullPath
}

func FileOrFolderExists(fullPath string) bool {
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func ReadFileByRelativePathOrFail(relativePath string) string {
	filePath := constructFullPathOrFail(relativePath, true)
	data, err := os.ReadFile(filePath)
	if err != nil {
		Fail(fmt.Sprintf("Could not read file from '%s' - %v", filePath, err))
	}

	return string(data)
}

func WriteFileByRelativePathOrFail(relativePath string, data []byte) {
	filePath := constructFullPathOrFail(relativePath, false)
	_, err := os.Stat(filePath)
	if err != nil {
		dirPath := filepath.Dir(filePath)
		err = os.MkdirAll(dirPath, 0644)
		if err != nil {
			Fail(fmt.Sprintf("Could not mkdir %s - %v", dirPath, err))
		}

	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		Fail(fmt.Sprintf("Could not write to '%s' - %v", filePath, err))
	}
}

func GetAbsPathByRelativeOrFail(relativePath string) string {
	cwd, err := os.Getwd()
	if err != nil {
		Fail(fmt.Sprintf("Could not determine process working directory - %v", err))
	}

	return filepath.Join(cwd, relativePath)
}

func DeleteDistInCurrentFolderOrFail() {
	distPath := GetAbsPathByRelativeOrFail("dist")
	err := os.RemoveAll(distPath)
	if err != nil {
		Fail(fmt.Sprintf("Could not delete dist folder at '%s' - %v", distPath, err))
	}
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

func GetGraphByGleeceConfigOrFail() symboldg.SymbolGraphBuilder {
	configPath := constructFullPathOrFail("gleece.test.config.json", true)
	config, err := cmd.LoadGleeceConfig(configPath)
	if err != nil {
		Fail(fmt.Sprintf("could not load Gleece Config - %v", err))
	}

	pipe, err := pipeline.NewGleecePipeline(config)
	if err != nil {
		Fail(fmt.Sprintf("could not create a pipeline - %v", err))
	}

	pipe.GenerateGraph()
	return pipe.Graph()
}
