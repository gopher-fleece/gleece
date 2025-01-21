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

func getMetadata(codeFileGlobs ...string) ([]definitions.ControllerMetadata, []definitions.ModelMetadata, error) {
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
		return nil, nil, *lastErr
	}

	flatModels, err := visitor.GetModelsFlat()
	if err != nil {
		Logger.Error("Failed to get models metadata: %v", err)
		return nil, nil, err
	}

	data, _ := json.MarshalIndent(flatModels, "", "\t")
	Logger.Info("%s", string(data))

	controllers := visitor.GetControllers()
	return controllers, flatModels, nil
}

func getConfigAndMetadata(args arguments.CliArguments) (
	*definitions.GleeceConfig,
	[]definitions.ControllerMetadata,
	[]definitions.ModelMetadata, error,
) {
	config, err := getConfig(args.ConfigPath)
	if err != nil {
		return nil, nil, nil, err
	}

	Logger.Info("Generating spec. Configuration file: %s", args.ConfigPath)

	defs, models, err := getMetadata()
	if err != nil {
		Logger.Fatal("Could not collect metadata - %v", err)
		return nil, nil, nil, err
	}

	return config, defs, models, nil
}

func GenerateSpec(args arguments.CliArguments) error {
	config, meta, models, err := getConfigAndMetadata(args)
	if err != nil {
		return err
	}

	models = append(models, definitions.ModelMetadata{
		Name:        "Rfc7807Error",
		Description: "TBD ERROR DESCRIPTION",
		Fields: []definitions.FieldMetadata{
			{
				Name:        "code",
				Type:        "int",
				Description: "TBD ERROR CODE DESCRIPTION",
			},
		},
	})

	// Generate the spec
	if err := swagen.GenerateAndOutputSpec(&config.OpenAPIGeneratorConfig, meta, models); err != nil {
		Logger.Fatal("Failed to generate OpenAPI spec - %v", err)
		return err
	}

	return nil
}

func GenerateRoutes(args arguments.CliArguments) error {
	config, meta, _, err := getConfigAndMetadata(args)
	if err != nil {
		return err
	}
	routes.GenerateRoutes(config, meta)
	return nil
}

func GenerateSpecAndRoutes(args arguments.CliArguments) error {
	config, meta, _, err := getConfigAndMetadata(args)
	if err != nil {
		return err
	}

	// Generate the routes first
	if err := routes.GenerateRoutes(config, meta); err != nil {
		Logger.Fatal("Failed to generate routes - %v", err)
		return err
	}

	// Generate the spec
	if err := swagen.GenerateAndOutputSpec(&config.OpenAPIGeneratorConfig, meta, []definitions.ModelMetadata{}); err != nil {
		Logger.Fatal("Failed to generate OpenAPI spec - %v", err)
		return err
	}

	return nil
}
