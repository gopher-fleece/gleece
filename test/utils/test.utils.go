package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/cmd/arguments"
	"github.com/gopher-fleece/gleece/definitions"
	. "github.com/onsi/ginkgo/v2"
	"golang.org/x/tools/go/packages"
)

func GetControllersAndModels() ([]definitions.ControllerMetadata, []definitions.StructMetadata, bool) {
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

func GetAbsPathByRelative(relativePath string) string {
	cwd, err := os.Getwd()
	if err != nil {
		Fail(fmt.Sprintf("Could not determine process working directory - %v", err))
	}

	return filepath.Join(cwd, relativePath)
}

func LoadPackage(fullName string, loadMode packages.LoadMode) *packages.Package {
	cfg := &packages.Config{Mode: loadMode}
	matchingPackages, err := packages.Load(cfg, fullName)
	if err != nil || len(matchingPackages) <= 0 {
		Fail(fmt.Sprintf("Could not load package '%s' for testing. This indicates a test issue, not a code issue", fullName))
	}

	return matchingPackages[0]
}

func LoadGleecePackage(loadMode packages.LoadMode) *packages.Package {
	return LoadPackage("github.com/gopher-fleece/gleece", loadMode)
}
