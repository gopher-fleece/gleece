package cmd

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gopher-fleece/gleece/cmd/arguments"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor"
	"github.com/gopher-fleece/gleece/generator/routes"
	"github.com/gopher-fleece/gleece/generator/swagen"
	Logger "github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/infrastructure/validation"
	"github.com/titanous/json5"
)

func getConfig(configPath string) (*definitions.GleeceConfig, error) {

	// Read the JSON file
	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err.Error())
	}

	// Unmarshal the JSON content into the struct
	var config definitions.GleeceConfig
	err = json5.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	// Validate the struct
	err = validation.ValidateStruct(config)
	if err != nil {
		// Validation failed
		for _, err := range err.(validator.ValidationErrors) {
			Logger.Error("Field '%s' failed validation: %s", err.StructField(), err.Tag())
		}
		return nil, fmt.Errorf("validation failed")
	}

	return &config, nil
}

func getMetadata(config *definitions.GleeceConfig) ([]definitions.ControllerMetadata, []definitions.ModelMetadata, bool, error) {
	visitor, err := extractor.NewControllerVisitor(config)
	if err != nil {
		return []definitions.ControllerMetadata{}, []definitions.ModelMetadata{}, false, err
	}

	for _, file := range visitor.GetFiles() {
		ast.Walk(visitor, file)
	}

	lastErr := visitor.GetLastError()
	if lastErr != nil {
		Logger.Error("Visitor encountered at-least one error. Last error:\n%v\n\t%s", *lastErr, visitor.GetFormattedDiagnosticStack())
		return nil, nil, false, *lastErr
	}

	flatModels, hasAnyErrorTypes, err := visitor.GetModelsFlat()
	if err != nil {
		Logger.Error("Failed to get models metadata: %v", err)
		return nil, nil, false, err
	}

	data, _ := json.MarshalIndent(flatModels, "", "\t")
	Logger.Info("%s", string(data))

	controllers := visitor.GetControllers()
	return controllers, flatModels, hasAnyErrorTypes, nil
}

func getConfigAndMetadata(args arguments.CliArguments) (
	*definitions.GleeceConfig,
	[]definitions.ControllerMetadata,
	[]definitions.ModelMetadata,
	bool,
	error,
) {
	config, err := getConfig(args.ConfigPath)
	if err != nil {
		return nil, nil, nil, false, err
	}

	Logger.Info("Generating spec. Configuration file: %s", args.ConfigPath)

	defs, models, hasAnyErrorTypes, err := getMetadata(config)
	if err != nil {
		Logger.Fatal("Could not collect metadata - %v", err)
		return nil, nil, nil, false, err
	}

	return config, defs, models, hasAnyErrorTypes, nil
}

func GenerateSpec(args arguments.CliArguments) error {
	config, meta, models, hasAnyErrorTypes, err := getConfigAndMetadata(args)
	if err != nil {
		return err
	}

	// Generate the spec
	if err := swagen.GenerateAndOutputSpec(&config.OpenAPIGeneratorConfig, meta, models, hasAnyErrorTypes); err != nil {
		Logger.Fatal("Failed to generate OpenAPI spec - %v", err)
		return err
	}

	return nil
}

func GenerateRoutes(args arguments.CliArguments) error {
	config, meta, _, _, err := getConfigAndMetadata(args)
	if err != nil {
		return err
	}

	if err := routes.GenerateRoutes(config, meta); err != nil {
		Logger.Fatal("Failed to generate routing file - %v", err)
		return err
	}

	return nil
}

func GenerateSpecAndRoutes(args arguments.CliArguments) error {
	config, meta, models, hasAnyErrorTypes, err := getConfigAndMetadata(args)
	if err != nil {
		return err
	}

	// Generate the routes first
	if err := routes.GenerateRoutes(config, meta); err != nil {
		Logger.Fatal("Failed to generate routes - %v", err)
		return err
	}

	// Generate the spec
	if err := swagen.GenerateAndOutputSpec(&config.OpenAPIGeneratorConfig, meta, models, hasAnyErrorTypes); err != nil {
		Logger.Fatal("Failed to generate OpenAPI spec - %v", err)
		return err
	}

	return nil
}
