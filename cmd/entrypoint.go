package cmd

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"os"

	"github.com/gopher-fleece/gleece/cmd/arguments"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/visitors/controller"
	"github.com/gopher-fleece/gleece/generator/routes"
	"github.com/gopher-fleece/gleece/generator/swagen"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/infrastructure/validation"
	"github.com/gopher-fleece/runtime"
	"github.com/titanous/json5"
)

func getConfig(configPath string) (*definitions.GleeceConfig, error) {

	// Read the JSON file
	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf(`could not read config file from "%s" - "%v"`, configPath, err.Error())
	}

	// Unmarshal the JSON content into the struct
	var config definitions.GleeceConfig
	err = json5.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, fmt.Errorf(`could not unmarshal config file "%s" to JSON5 - "%v"`, configPath, err)
	}

	// Validate the struct
	err = validation.ValidateStruct(config)
	if err != nil {
		return nil, fmt.Errorf(`configuration file "%s" is invalid - "%s"`, configPath, runtime.ExtractValidationErrorMessage(err, nil))
	}

	return &config, nil
}

func getMetadata(config *definitions.GleeceConfig) ([]definitions.ControllerMetadata, []definitions.ModelMetadata, bool, error) {
	visitor, err := controller.NewControllerVisitor(config)
	if err != nil {
		return []definitions.ControllerMetadata{}, []definitions.ModelMetadata{}, false, err
	}

	for _, file := range visitor.GetFiles() {
		ast.Walk(visitor, file)
	}

	lastErr := visitor.GetLastError()
	if lastErr != nil {
		logger.Error("Visitor encountered at-least one error. Last error:\n%v\n\t%s", *lastErr, visitor.GetFormattedDiagnosticStack())
		return nil, nil, false, *lastErr
	}

	flatModels, hasAnyErrorTypes, err := visitor.GetModelsFlat()
	if err != nil {
		logger.Error("Failed to get models metadata: %v", err)
		return nil, nil, false, err
	}

	data, _ := json.MarshalIndent(flatModels, "", "\t")
	logger.Debug("Flat models list:\n%s", string(data))

	controllers := visitor.GetControllers()
	return controllers, flatModels, hasAnyErrorTypes, nil
}

func GetConfigAndMetadata(args arguments.CliArguments) (
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

	logger.Info("Generating spec. Configuration file: %s", args.ConfigPath)

	defs, models, hasAnyErrorTypes, err := getMetadata(config)
	if err != nil {
		logger.Fatal("Could not collect metadata - %v", err)
		return nil, nil, nil, false, err
	}

	return config, defs, models, hasAnyErrorTypes, nil
}

func GenerateSpec(args arguments.CliArguments) error {
	logger.Info("Generating spec")
	config, meta, models, hasAnyErrorTypes, err := GetConfigAndMetadata(args)
	if err != nil {
		return err
	}

	// Generate the spec
	if err := swagen.GenerateAndOutputSpec(&config.OpenAPIGeneratorConfig, meta, models, hasAnyErrorTypes); err != nil {
		logger.Fatal("Failed to generate OpenAPI spec - %v", err)
		return err
	}

	logger.Info("Spec successfully generated")
	return nil
}

func GenerateRoutes(args arguments.CliArguments) error {
	logger.Info("Generating routes")
	config, meta, _, _, err := GetConfigAndMetadata(args)
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
	config, meta, models, hasAnyErrorTypes, err := GetConfigAndMetadata(args)
	if err != nil {
		return err
	}

	// Generate the routes first
	if err := routes.GenerateRoutes(config, meta); err != nil {
		logger.Fatal("Failed to generate routes - %v", err)
		return err
	}

	// Generate the spec
	if err := swagen.GenerateAndOutputSpec(&config.OpenAPIGeneratorConfig, meta, models, hasAnyErrorTypes); err != nil {
		logger.Fatal("Failed to generate OpenAPI spec - %v", err)
		return err
	}

	logger.Info("Spec and routes successfully generated")
	return nil
}
