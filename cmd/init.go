package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gopher-fleece/gleece/v2/definitions"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Gleece configuration",
	Long:  `The init command starts a wizard that asks questions and creates a gleece.config.json file.`,
	Run: func(cmd *cobra.Command, args []string) {
		runWizard()
	},
}

func runWizard() {
	w := &wizard{
		reader: bufio.NewReader(os.Stdin),
	}

	fmt.Println("Welcome to the Gleece configuration wizard!")
	fmt.Println("Please answer the following questions to generate your gleece.config.json")
	fmt.Println()

	config := definitions.GleeceConfig{}

	// Common Config
	fmt.Println("--- Common Configuration ---")
	config.CommonConfig.ControllerGlobs = w.askVarArgs("Controller globs (e.g. ./**/*.go)", []string{"./**/controllers/**/*.go"})
	config.CommonConfig.AllowPackageLoadFailures = w.askBool("Allow package load failures?", false)
	fmt.Println()

	// Routes Config
	fmt.Println("--- Routes Configuration ---")
	config.RoutesConfig.Engine = definitions.RoutingEngineType(w.askSelection("Routing engine", []string{"gin", "echo", "mux", "fiber", "chi"}, "gin"))
	config.RoutesConfig.PackageName = w.askOpenEnded("Go package name for generated routes", "routes")
	config.RoutesConfig.OutputPath = w.askOpenEnded("Output path for generated routes", fmt.Sprintf("./%s/gleece.go", config.RoutesConfig.PackageName))
	config.RoutesConfig.OutputFilePerms = w.askOpenEnded("Output file permissions", "0644")
	config.RoutesConfig.ValidateResponsePayload = w.askBool("Validate response payload?", false)
	config.RoutesConfig.SkipGenerateDateComment = w.askBool("Skip generation date comment?", true)

	// Authorization Config
	config.RoutesConfig.AuthorizationConfig.AuthFileFullPackageName = w.askOpenEnded("Full package name for auth middleware file", "")
	config.RoutesConfig.AuthorizationConfig.EnforceSecurityOnAllRoutes = w.askBool("Enforce security on all routes?", true)
	fmt.Println()

	// OpenAPI Config
	fmt.Println("--- OpenAPI Generator Configuration ---")
	config.OpenAPIGeneratorConfig.OpenAPI = w.askSelection("OpenAPI version", []string{"3.0.0", "3.1.0"}, "3.0.0")
	config.OpenAPIGeneratorConfig.Info.Title = w.askOpenEnded("API Title", "My API")
	config.OpenAPIGeneratorConfig.Info.Description = w.askOpenEnded("API Description", "")
	config.OpenAPIGeneratorConfig.Info.Version = w.askOpenEnded("API Version", "1.0.0")
	config.OpenAPIGeneratorConfig.BaseURL = w.askOpenEnded("Base URL", "http://localhost:8080")

	config.OpenAPIGeneratorConfig.SpecGeneratorConfig.OutputPath = w.askOpenEnded("OpenAPI spec output path", "./docs/swagger.json")

	fmt.Println()
	fmt.Println("--- Security Schemes ---")
	if w.askBool("Add a security scheme?", false) {
		config.OpenAPIGeneratorConfig.SecuritySchemes = w.askSecuritySchemes()
	}

	// Experimental Config
	fmt.Println("--- Experimental Configuration ---")
	config.ExperimentalConfig.ValidateTopLevelOnlyEnum = w.askBool("Validate top-level only enums?", false)
	config.ExperimentalConfig.GenerateEnumValidator = w.askBool("Generate enum validator?", false)

	fmt.Println()
	saveConfig(config)
}

func saveConfig(config definitions.GleeceConfig) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		logger.Fatal("Failed to marshal config to JSON: %v", err)
	}

	filename := "gleece.config.json"

	// Prompt before overwriting
	if _, statErr := os.Stat(filename); statErr == nil {
		fmt.Printf("File %s already exists. Overwrite? (y/N): ", filename)

		var confirm string
		_, scanErr := fmt.Scanln(&confirm)
		if scanErr != nil {
			fmt.Printf("Failed to read confirmation: %v", scanErr)
		}

		if strings.ToLower(confirm) != "y" {
			fmt.Println("Aborted.")
			return
		}
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		logger.Fatal("Failed to write config file: %v", err)
	}

	fmt.Printf("Successfully created %s\n", filename)
}
