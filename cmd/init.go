package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/gopher-fleece/gleece/v2/definitions"
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
	w := newCliWizard(bufio.NewReader(os.Stdin))

	fmt.Println("Welcome to the Gleece configuration wizard!")
	fmt.Println("Please answer the following questions to generate your gleece.config.json")
	fmt.Println()

	config := definitions.GleeceConfig{}

	askCommonConfig(w, &config)
	askRoutesConfig(w, &config)
	askAuthConfig(w, &config)
	askOpenApiConfig(w, &config)
	askSecSchemas(w, &config)
	askExperimentalConfigs(w, &config)
	saveConfig(w, config)
}

func askCommonConfig(w *cliWizard, config *definitions.GleeceConfig) {
	fmt.Println("--- Common Configuration ---")
	config.CommonConfig.ControllerGlobs = w.askVarArgs("Controller globs (e.g. ./**/*.go)", []string{"./**/controllers/**/*.go"})
	config.CommonConfig.AllowPackageLoadFailures = w.askBool("Allow package load failures?", false)
	fmt.Println()
}

func askRoutesConfig(w *cliWizard, config *definitions.GleeceConfig) {
	fmt.Println("--- Routes Configuration ---")
	config.RoutesConfig.Engine = definitions.RoutingEngineType(w.askSelection("Routing engine", []string{"gin", "echo", "mux", "fiber", "chi"}, "gin"))
	config.RoutesConfig.PackageName = w.askOpenEnded("Go package name for generated routes", "routes")
	config.RoutesConfig.OutputPath = w.askFilePath(
		"Output path for generated routes",
		fmt.Sprintf("./%s/gleece.go", config.RoutesConfig.PackageName),
		"go",
	)
	config.RoutesConfig.OutputFilePerms = w.askFilePerms("Output file permissions", "0644")
	config.RoutesConfig.ValidateResponsePayload = w.askBool("Validate response payload?", false)
	config.RoutesConfig.SkipGenerateDateComment = w.askBool("Skip generation date comment?", true)
}

func askAuthConfig(w *cliWizard, config *definitions.GleeceConfig) {
	fmt.Println("--- Authorization Configuration ---")
	config.RoutesConfig.AuthorizationConfig.AuthFileFullPackageName = w.askOpenEnded("Full package name for auth middleware file", "authentication")
	config.RoutesConfig.AuthorizationConfig.EnforceSecurityOnAllRoutes = w.askBool("Enforce security on all routes?", true)
	fmt.Println()
}

func askOpenApiConfig(w *cliWizard, config *definitions.GleeceConfig) {
	fmt.Println("--- OpenAPI Generator Configuration ---")
	config.OpenAPIGeneratorConfig.OpenAPI = w.askSelection("OpenAPI version", []string{"3.0.0", "3.1.0"}, "3.0.0")
	config.OpenAPIGeneratorConfig.Info.Title = w.askOpenEnded("API Title", "My API")
	config.OpenAPIGeneratorConfig.Info.Description = w.askOpenEnded("API Description", "")
	config.OpenAPIGeneratorConfig.Info.Version = w.askOpenEnded("API Version", "1.0.0")
	config.OpenAPIGeneratorConfig.Info.TermsOfService = w.askUrl("Terms of Service URL", "")

	if w.askBool("Include contact information?", false) {
		config.OpenAPIGeneratorConfig.Info.Contact = &definitions.OpenAPIContact{
			Name:  w.askOpenEnded("Contact Name", ""),
			URL:   w.askUrl("Contact URL", ""),
			Email: w.askEmail("Contact Email", ""),
		}
	}

	if w.askBool("Include license information?", false) {
		config.OpenAPIGeneratorConfig.Info.License = &definitions.OpenAPILicense{
			Name: w.askOpenEnded("License Name", ""),
			URL:  w.askUrl("License URL", ""),
		}
	}

	config.OpenAPIGeneratorConfig.BaseURL = w.askUrl("Base URL", "http://localhost:8080")
	config.OpenAPIGeneratorConfig.SpecGeneratorConfig.OutputPath = w.askFilePath(
		"OpenAPI spec output path",
		"./docs/swagger.json",
		"",
	)
	fmt.Println()
}

func askSecSchemas(w *cliWizard, config *definitions.GleeceConfig) {
	fmt.Println("--- Security Schemes ---")
	if w.askBool("Add a security scheme?", false) {
		config.OpenAPIGeneratorConfig.SecuritySchemes = w.askSecuritySchemes()
	}
}

func askExperimentalConfigs(w *cliWizard, config *definitions.GleeceConfig) {
	fmt.Println("--- Experimental Configuration ---")
	config.ExperimentalConfig.ValidateTopLevelOnlyEnum = w.askBool("Validate top-level only enums?", false)
	config.ExperimentalConfig.GenerateEnumValidator = w.askBool("Generate enum validator?", false)
	fmt.Println()
}

func removeNilValuesRecursive(v any) any {
	switch valueType := v.(type) {
	case map[string]any:
		for key, val := range valueType {
			if val == nil {
				delete(valueType, key)
				continue
			}
			valueType[key] = removeNilValuesRecursive(val)
		}
		return valueType

	case []any:
		cleaned := make([]any, 0, len(valueType))
		for _, item := range valueType {
			if item == nil {
				continue
			}
			cleaned = append(cleaned, removeNilValuesRecursive(item))
		}
		return cleaned

	default:
		return v
	}
}

func marshalConfig(config definitions.GleeceConfig) ([]byte, error) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config to JSON: %v", err)
	}

	asMap := make(map[string]any)
	err = json.Unmarshal(data, &asMap)
	if err != nil {
		return data, errors.New("failed to perform configuration cleanup. Null fields may be shown in the final configuration")
	}

	cleanedConfig := removeNilValuesRecursive(asMap)

	cleanedData, err := json.MarshalIndent(cleanedConfig, "", "  ")
	if err != nil {
		return data, errors.New("failed to marshal cleaned configuration file. Null fields may be shown in the final configuration")
	}

	return cleanedData, nil
}

func saveConfig(w *cliWizard, config definitions.GleeceConfig) {
	configBytes, err := marshalConfig(config)
	if err != nil {
		if configBytes == nil {
			fmt.Println(fmt.Sprintf("Encountered a fatal error whilst preparing the configuration file for writing - %v", err))
			return
		}
		fmt.Println(
			"Encountered an error whilst preparing the configuration file for writing." +
				" The resulting configuration file may have null fields or be otherwise faulty",
		)
	}

	filename := "gleece.config.json"

	if _, statErr := os.Stat(filename); statErr == nil {
		// Prompt before overwriting
		if !w.askBlockingConfirm("A gleece.config.json file already exists. Overwrite?") {
			fmt.Println("Aborting...")
			return
		}
	}

	err = os.WriteFile(filename, configBytes, 0644)
	if err == nil {
		fmt.Printf("Successfully created %s\n", filename)
	} else {
		fmt.Println(fmt.Sprintf("Failed to write the configuration file -  %v", err))
		fmt.Println(fmt.Sprintf("Dumping the configuration file contents to console:\n%s", string(configBytes)))
	}
}
