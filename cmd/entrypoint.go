package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gopher-fleece/gleece/cmd/arguments"
	"github.com/gopher-fleece/gleece/core/pipeline"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/generator/routes"
	"github.com/gopher-fleece/gleece/generator/swagen"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/infrastructure/validation"
	"github.com/titanous/json5"
)

// LoadGleeceConfig gets the currently relevant Gleece Config file
func LoadGleeceConfig(configPath string) (*definitions.GleeceConfig, error) {

	// Read the JSON file
	var configPathToUse string

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		logger.Warn("Could not determine absolute path for config path '%s'. Will attempt to use as-is", configPath)
		configPathToUse = configPath
	} else {
		configPathToUse = absPath
	}

	fileContent, err := os.ReadFile(configPathToUse)
	if err != nil {
		return nil, fmt.Errorf(`could not read config file from "%s" - "%v"`, configPathToUse, err.Error())
	}

	// Unmarshal the JSON content into the struct
	var config definitions.GleeceConfig
	err = json5.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, fmt.Errorf(`could not unmarshal config file "%s" to JSON5 - "%v"`, configPathToUse, err)
	}

	// Validate the struct
	err = validation.ValidateStruct(config)
	if err != nil {
		return nil, fmt.Errorf(
			`configuration file "%s" is invalid - "%s"`,
			configPathToUse,
			validation.ExtractValidationErrorMessage(err, nil),
		)
	}

	return &config, nil
}

func getFullMetadata(config *definitions.GleeceConfig) (pipeline.GleeceFlattenedMetadata, error) {
	pipe, err := pipeline.NewGleecePipeline(config)
	if err != nil {
		return pipeline.GleeceFlattenedMetadata{}, err
	}

	return pipe.Run()
}

func GetConfigAndMetadata(args arguments.CliArguments) (
	*definitions.GleeceConfig,
	pipeline.GleeceFlattenedMetadata,
	error,
) {
	config, err := LoadGleeceConfig(args.ConfigPath)
	if err != nil {
		return config, pipeline.GleeceFlattenedMetadata{}, err
	}

	logger.Info("Generating spec. Configuration file: %s", args.ConfigPath)

	meta, err := getFullMetadata(config)
	return config, meta, err
}

func GenerateSpec(args arguments.CliArguments) error {
	logger.Info("Generating spec")
	config, meta, err := GetConfigAndMetadata(args)
	if err != nil {
		return err
	}

	// Generate the spec
	if err := swagen.GenerateAndOutputSpec(
		&config.OpenAPIGeneratorConfig,
		meta.Flat,
		&meta.Models,
		meta.PlainErrorPresent,
	); err != nil {
		logger.Fatal("Failed to generate OpenAPI spec - %v", err)
		return err
	}

	logger.Info("Spec successfully generated")
	return nil
}

func GenerateRoutes(args arguments.CliArguments) error {
	logger.Info("Generating routes")
	config, meta, err := GetConfigAndMetadata(args)
	if err != nil {
		return err
	}

	if err := routes.GenerateRoutes(config, meta); err != nil {
		logger.Fatal("Failed to generate routing file - %v", err)
		return err
	}

	logger.Info("Routes successfully generated")
	return nil
}

func GenerateSpecAndRoutes(args arguments.CliArguments) error {
	logger.Info("Generating spec and routes")
	config, meta, err := GetConfigAndMetadata(args)
	if err != nil {
		return err
	}

	// Generate the routes first
	if err := routes.GenerateRoutes(config, meta); err != nil {
		logger.Fatal("Failed to generate routes - %v", err)
		return err
	}

	// Generate the spec
	if err := swagen.GenerateAndOutputSpec(
		&config.OpenAPIGeneratorConfig,
		meta.Flat,
		&meta.Models,
		meta.PlainErrorPresent,
	); err != nil {
		logger.Fatal("Failed to generate OpenAPI spec - %v", err)
		return err
	}

	logger.Info("Spec and routes successfully generated")
	return nil
}
