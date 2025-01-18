package generator

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/haimkastner/gleece/cmd/arguments"
	"github.com/haimkastner/gleece/definitions"
	"github.com/haimkastner/gleece/extractor"
	"github.com/haimkastner/gleece/generator/swagen"
	Logger "github.com/haimkastner/gleece/infrastructure/logger"
	"github.com/haimkastner/gleece/infrastructure/validation"
)

func getConfig(configPath string) (*definitions.GleeceConfig, error) {

	// Read the JSON file
	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err.Error())
	}

	// Unmarshal the JSON content into the struct
	var config definitions.GleeceConfig
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	// Validate the struct
	err = validation.ValidatorInstance.Struct(config)
	if err != nil {
		// Validation failed
		for _, err := range err.(validator.ValidationErrors) {
			Logger.Error("Field '%s' failed validation: %s", err.StructField(), err.Tag())
		}
		return nil, fmt.Errorf("validation failed")
	}

	return &config, nil
}

func getConfigAndMetadata(args arguments.CliArguments) (*definitions.GleeceConfig, []definitions.ControllerMetadata, error) {
	config, err := getConfig(args.ConfigPath)
	if err != nil {
		return nil, []definitions.ControllerMetadata{}, err
	}

	Logger.Info("Generating spec. Configuration file: %s", args.ConfigPath)

	defs, err := extractor.GetMetadata()
	if err != nil {
		Logger.Fatal("Could not collect metadata - %v", err)
		return nil, []definitions.ControllerMetadata{}, err
	}

	return config, defs, nil
}

func GenerateSpec(args arguments.CliArguments) error {
	config, meta, err := getConfigAndMetadata(args)
	if err != nil {
		return err
	}

	// Generate the spec
	if err := swagen.GenerateAndOutputSpec(&config.OpenAPIGeneratorConfig, meta, []definitions.ModelMetadata{}); err != nil {
		Logger.Fatal("Failed to generate OpenAPI spec - %v", err)
		return err
	}

	return nil
}

func GenerateRoutes(args arguments.CliArguments) error {
	/*
		config, meta, err := getConfigAndMetadata(args)
		if err != nil {
			return err
		}
	*/
	return nil
}

func GenerateSpecAndRoutes(args arguments.CliArguments) error {
	config, meta, err := getConfigAndMetadata(args)
	if err != nil {
		return err
	}

	// Generate the spec
	if err := swagen.GenerateAndOutputSpec(&config.OpenAPIGeneratorConfig, meta, []definitions.ModelMetadata{}); err != nil {
		Logger.Fatal("Failed to generate OpenAPI spec - %v", err)
		return err
	}

	//////////////////////
	// Add gen routes here
	//////////////////////

	return nil
}
