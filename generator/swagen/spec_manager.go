package swagen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/generator/swagen/swagen30"
	"github.com/gopher-fleece/gleece/generator/swagen/swagen31"
	"github.com/gopher-fleece/gleece/generator/swagen/swagtool"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

// GenerateSpec generates the OpenAPI specification
func GenerateSpec(config *definitions.OpenAPIGeneratorConfig, defs []definitions.ControllerMetadata, models []definitions.ModelMetadata, hasAnyErrorTypes bool) ([]byte, error) {
	// In case of a default error in use, add the RFC-7807, otherwise skip and assume the user define it using structs by themselves
	swagtool.AppendErrorSchema(&models, hasAnyErrorTypes)

	// Since the tools and validation are WAY better for 3.0.0,
	// And our logic his focusing in 3.0 and not feature will be added if not can be support in it too,
	// We will use it any way for validation and alignment
	specV300, err := swagen30.GenerateSpec(config, defs, models)

	// In case of error, abort
	if err != nil {
		return nil, err
	}

	switch config.OpenAPI {
	case "3.0.0":
		return specV300, nil // In case of 3.0.0, we already have the spec
	case "3.1.0":
		return swagen31.GenerateSpec(config, defs, models)
	default:
		return nil, fmt.Errorf("Unsupported OpenAPI version: %s", config.OpenAPI)
	}
}

func GenerateAndOutputSpec(config *definitions.OpenAPIGeneratorConfig, defs []definitions.ControllerMetadata, models []definitions.ModelMetadata, hasAnyErrorTypes bool) error {
	jsonBytes, err := GenerateSpec(config, defs, models, hasAnyErrorTypes)

	if err != nil {
		return err
	}

	// Extract path from file path
	// Extract the directory path
	dirPath := filepath.Dir(config.SpecGeneratorConfig.OutputPath)
	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		logger.Error("Failed to create directory - %v", err)
		return err
	}

	// Write the JSON to the file
	if err := os.WriteFile(config.SpecGeneratorConfig.OutputPath, jsonBytes, 0644); err != nil {
		logger.Error("Failed to write file - %v", err)
		return err
	}

	// Print the path to the generated JSON file
	logger.Info("OpenAPI specification written to '%s'", config.SpecGeneratorConfig.OutputPath)
	return nil

}
