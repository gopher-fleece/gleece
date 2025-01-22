package main

import (
	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/cmd/arguments"
	Logger "github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/infrastructure/validation"
)

/*
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
*/

// CLI Main. Uncomment when ready
func main() {
	validation.InitValidator()
	err := cmd.GenerateSpecAndRoutes(arguments.CliArguments{ConfigPath: "./gleece.json"})
	if err != nil {
		Logger.Fatal("Failed to generate routes: %v", err)
	}
	// cmd.Execute()
}
