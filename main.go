package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	go_validator "github.com/go-playground/validator/v10"
	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/cmd/arguments"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor"
	"github.com/gopher-fleece/gleece/generator/routes"
	"github.com/gopher-fleece/gleece/generator/swagen"
	Logger "github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/infrastructure/validator"
)

func obtainConfig(configPath string) (*definitions.GleeceConfig, error) {

	// Read the JSON file
	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error reading file: %v", err.Error()))
	}

	// Unmarshal the JSON content into the struct
	var config definitions.GleeceConfig
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error unmarshaling JSON: %v", err))
	}

	// Validate the struct
	err = validator.Validate.Struct(config)
	if err != nil {
		// Validation failed
		for _, err := range err.(go_validator.ValidationErrors) {
			Logger.Error("Field '%s' failed validation: %s", err.StructField(), err.Tag())
		}
		return nil, errors.New("Validation failed")
	}

	return &config, nil
}

func main() {
	validator.InitValidator()

	defaultPath := "gleece.json"
	var filePath string
	flag.StringVar(&filePath, "path", defaultPath, "Path to the JSON file")
	flag.Parse()
	gleeceConfig, err := obtainConfig(defaultPath)

	if err != nil {
		Logger.Fatal("Could not load configuration - %s", err.Error())
		os.Exit(1)
	}

	Logger.Info("Configuration loaded successfully")

	defs, err := extractor.GetMetadata()
	if err != nil {
		Logger.Fatal("Could not collect metadata - %v", err)
		os.Exit(1)
	}

	// Generate the spec
	if err := swagen.GenerateAndOutputSpec(&gleeceConfig.OpenAPIGeneratorConfig, defs, []definitions.ModelMetadata{}); err != nil {
		Logger.Fatal("Failed to generate OpenAPI spec - %v", err)
		os.Exit(1)
	}

	routes.GenerateRoutes(
		cmd.RoutesConfig{
			Engine:            cmd.RoutingEngineGin,
			TemplateOverrides: map[cmd.KnownTemplate]string{},
			OutputPath:        "/mnt/7e91759c-6dd7-4c99-8d38-e6422452a469/git/gleece/dist/routes.go",
			OutputFilePerms:   arguments.NewFileModeArg("0644"),
		},
		defs,
	)
}

/*
	CLI Main. Uncomment when ready
func main() {
	cmd.Execute()
}

*/
