package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/gopher-fleece/gleece/v2/cmd/embed"
	"github.com/gopher-fleece/gleece/v2/definitions"
	"github.com/gopher-fleece/gleece/v2/gast"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Gleece project or re-configure an existing one",
	Long: `The init command is a wizard that guides you through the process of initializing a Gleece project.
Currently, this includes creating the necessary configuration and some of the necessary boilerplate code.
`,
	Run: func(cmd *cobra.Command, args []string) {
		runWizard()
	},
}

// runWizard initializes and executes the CLI wizard to gather Gleece configuration.
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

	askAboutCodeGeneration(w, &config)
}

// askCommonConfig prompts the user for common configuration settings like package name, templates, and target language.
func askCommonConfig(w *cliWizard, config *definitions.GleeceConfig) {
	fmt.Println("--- Common Configuration ---")
	config.CommonConfig.ControllerGlobs = w.askVarArgs("Controller globs (e.g. ./**/*.go)", []string{"./**/controllers/**/*.go"})
	config.CommonConfig.AllowPackageLoadFailures = w.askBool("Allow package load failures?", false)
	fmt.Println()
}

// askRoutesConfig prompts the user for routing engine and configuration.
func askRoutesConfig(w *cliWizard, config *definitions.GleeceConfig) {
	fmt.Println("--- Routes Configuration ---")
	config.RoutesConfig.Engine = definitions.RoutingEngineType(w.askSelection(
		"Routing engine",
		[]string{"gin", "echo", "mux", "fiber", "chi"},
		"gin",
		true,
	))
	config.RoutesConfig.PackageName = w.askOpenEnded("Go package name for generated routes", "routes")
	config.RoutesConfig.OutputPath = w.askFilePath(
		"Output path for generated routes",
		fmt.Sprintf("./%s/gleece.go", config.RoutesConfig.PackageName),
		"go",
	)
	config.RoutesConfig.OutputFilePerms = w.askFilePerms("Output file permissions", "0644")
	config.RoutesConfig.ValidateResponsePayload = w.askBool("Validate response payload?", false)
	config.RoutesConfig.SkipGenerateDateComment = w.askBool("Skip generation date comment?", true)
	fmt.Println()
}

// askAuthConfig prompts the user for authentication-related settings.
func askAuthConfig(w *cliWizard, config *definitions.GleeceConfig) {
	fmt.Println("--- Authorization Configuration ---")
	config.RoutesConfig.AuthorizationConfig.AuthFileFullPackageName = w.askOpenEnded("Full package name for auth middleware file", "authentication")
	config.RoutesConfig.AuthorizationConfig.EnforceSecurityOnAllRoutes = w.askBool("Enforce security on all routes?", true)
	fmt.Println()
}

// askOpenApiConfig prompts the user for OpenAPI specific information including info object details and output paths.
func askOpenApiConfig(w *cliWizard, config *definitions.GleeceConfig) {
	fmt.Println("--- OpenAPI Generator Configuration ---")
	config.OpenAPIGeneratorConfig.OpenAPI = w.askSelection(
		"OpenAPI version",
		[]string{"3.0.0", "3.1.0"},
		"3.0.0",
		true,
	)
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

// askSecSchemas prompts the user to add security schemes to the OpenAPI generator config.
func askSecSchemas(w *cliWizard, config *definitions.GleeceConfig) {
	fmt.Println("--- Security Schemes ---")
	if w.askBool("Add a security scheme?", false) {
		config.OpenAPIGeneratorConfig.SecuritySchemes = w.askSecuritySchemes()
	}
	fmt.Println()
}

// askExperimentalConfigs prompts the user for experimental feature toggles.
func askExperimentalConfigs(w *cliWizard, config *definitions.GleeceConfig) {
	fmt.Println("--- Experimental Configuration ---")
	config.ExperimentalConfig.ValidateTopLevelOnlyEnum = w.askBool("Validate top-level only enums?", false)
	config.ExperimentalConfig.GenerateEnumValidator = w.askBool("Generate enum validator?", false)
	fmt.Println()
}

// askAboutCodeGeneration prompts the user if they want to generate boilerplate code
// such as the authentication middleware
func askAboutCodeGeneration(w *cliWizard, config *definitions.GleeceConfig) {
	fmt.Println("--- Boilerplate Code Generation ---")
	if !w.askBool("Generate authentication middleware skeleton code?", true) {
		fmt.Println()
		return
	}

	splitPkgPath := strings.Split(config.RoutesConfig.AuthorizationConfig.AuthFileFullPackageName, "/")
	concatenatedPath := append(splitPkgPath[1:], "authentication.go")
	suggestedOutputPath := filepath.Join(concatenatedPath...)

	selectedOutputPath := w.askFilePath("Authentication middleware output path", suggestedOutputPath, "go")
	configStr, err := generateAuthMiddleware(config)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to generate authentication middleware code - %v", err))
		return
	}

	var finalOutputPath string
	if filepath.IsAbs(selectedOutputPath) {
		finalOutputPath = selectedOutputPath
	} else {
		abs, absErr := filepath.Abs(suggestedOutputPath)
		if absErr != nil {
			fmt.Println("Failed to determine absolute path for output file. Will attempt to use user-provided path instead")
			finalOutputPath = selectedOutputPath
		} else {
			finalOutputPath = abs
		}
	}

	saveFileWithOverwriteConfirmation(w, finalOutputPath, []byte(configStr))
	fmt.Println()
}

// generateAuthMiddleware creates an authentication middleware file for the configured routing engine
func generateAuthMiddleware(config *definitions.GleeceConfig) (string, error) {
	var template string

	switch config.RoutesConfig.Engine {
	case definitions.RoutingEngineChi:
		template = embed.ChiAuthMiddleware
		break
	case definitions.RoutingEngineEcho:
		template = embed.EchoAuthMiddleware
		break
	case definitions.RoutingEngineFiber:
		template = embed.FiberAuthMiddleware
		break
	case definitions.RoutingEngineGin:
		template = embed.GinAuthMiddleware
		break
	case definitions.RoutingEngineMux:
		template = embed.MuxAuthMiddleware
		break
	}

	ctx := map[string]any{
		"PkgAlias": gast.GetDefaultPkgAliasByName(config.RoutesConfig.AuthorizationConfig.AuthFileFullPackageName),
	}

	result, err := raymond.Render(template, ctx)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to render authentication middleware code - %v", err))
	}

	return result, err
}

// removeNilValuesRecursive recursively traverses map/slice structures and deletes nil fields.
// Note this method is not perfect - some nulls may remain under some cases, but it's good enough here.
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

// marshalConfig converts a GleeceConfig struct to JSON while cleaning out nil values.
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

// saveConfig marshals and writes the configuration to a file.
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

	saveFileWithOverwriteConfirmation(w, "gleece.config.json", configBytes)
	fmt.Println()
}

// saveFileWithOverwriteConfirmation attempts to save the given file, prompting the user
//
//	to confirm overwriting as necessary
func saveFileWithOverwriteConfirmation(w *cliWizard, fileName string, data []byte) bool {
	_, statErr := os.Stat(fileName)

	if statErr == nil {
		// Prompt before overwriting
		if !w.askBlockingConfirm(fmt.Sprintf("File '%s' already exists. Overwrite?", fileName)) {
			fmt.Println("Configuration overwrite aborted")
			return false
		}
	} else if errors.Is(statErr, os.ErrNotExist) {
		dirPath := filepath.Dir(fileName)
		mkdirErr := os.MkdirAll(dirPath, 0755)
		if mkdirErr != nil {
			fmt.Println(fmt.Sprintf("Could not mkdir %s - %v", dirPath, mkdirErr))
		}
	}

	err := os.WriteFile(fileName, data, 0644)
	if err == nil {
		fmt.Printf("Successfully created %s\n", fileName)
		return true
	}

	fmt.Println(fmt.Sprintf("Failed to write file '%s' -  %v", fileName, err))
	fmt.Println(fmt.Sprintf("Dumping file contents to console:\n%s", string(data)))
	return false
}
