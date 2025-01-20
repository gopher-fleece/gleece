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

func getMetadata(codeFileGlobs ...string) ([]definitions.ControllerMetadata, error) {
	var globs []string
	if len(codeFileGlobs) > 0 {
		globs = codeFileGlobs
	} else {
		globs = []string{"./*.go", "./**/*.go"}
	}

	visitor := &extractor.ControllerVisitor{}
	visitor.Init(globs)
	for _, file := range visitor.GetFiles() {
		ast.Walk(visitor, file)
	}

	lastErr := visitor.GetLastError()
	if lastErr != nil {
		Logger.Error("Visitor encountered at-least one error. Last error:\n%v\n\t%s", *lastErr, visitor.GetFormattedDiagnosticStack())
		return []definitions.ControllerMetadata{}, *lastErr
	}
	ctx, err := visitor.DumpContext()
	if err != nil {
		Logger.Error("Failed to dump context: %v", err)
	} else {
		Logger.Info("%v", ctx)
	}
	return visitor.GetControllers(), nil
}

func getConfigAndMetadata(args arguments.CliArguments) (*definitions.GleeceConfig, []definitions.ControllerMetadata, error) {
	config, err := getConfig(args.ConfigPath)
	if err != nil {
		return nil, []definitions.ControllerMetadata{}, err
	}

	Logger.Info("Generating spec. Configuration file: %s", args.ConfigPath)

	defs, err := getMetadata()
	if err != nil {
		Logger.Fatal("Could not collect metadata - %v", err)
		return nil, []definitions.ControllerMetadata{}, err
	}

	return config, defs, nil
}

func generateSpec(args arguments.CliArguments) error {
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
	config, meta, err := getConfigAndMetadata(args)
	if err != nil {
		return err
	}
	routes.GenerateRoutes(config, meta)
	return nil
}

func generateSpecAndRoutes(args arguments.CliArguments) error {
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
