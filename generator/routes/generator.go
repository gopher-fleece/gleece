package routes

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/aymerick/raymond"

	"github.com/gopher-fleece/gleece/v2/core/pipeline"
	"github.com/gopher-fleece/gleece/v2/definitions"
	"github.com/gopher-fleece/gleece/v2/generator/compilation"
	"github.com/gopher-fleece/gleece/v2/generator/templates/chi"
	"github.com/gopher-fleece/gleece/v2/generator/templates/echo"
	"github.com/gopher-fleece/gleece/v2/generator/templates/fiber"
	"github.com/gopher-fleece/gleece/v2/generator/templates/gin"
	"github.com/gopher-fleece/gleece/v2/generator/templates/mux"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
)

var RoutesTemplateName = "Routes"

var helpersRegistered bool
var partialsRegistered bool

// dumpContext Dumps the routes context to the log for debugging purposes
func dumpContext(ctx any) {
	// It's basically impossible to break the marshaller without some pretty crazy, and very intentional hacks
	contextJson, _ := json.MarshalIndent(ctx, "\t", "\t")
	logger.Debug("Routes Context:\n%s", contextJson)
}

func getDefaultTemplate(engine definitions.RoutingEngineType) string {
	switch engine {
	case definitions.RoutingEngineGin:
		return gin.RoutesTemplate
	case definitions.RoutingEngineEcho:
		return echo.RoutesTemplate
	case definitions.RoutingEngineMux:
		return mux.RoutesTemplate
	case definitions.RoutingEngineFiber:
		return fiber.RoutesTemplate
	case definitions.RoutingEngineChi:
		return chi.RoutesTemplate
	}
	// This should not happen. It indicates a breakage in the build itself.
	panic(fmt.Sprintf("Could not find an embedded template for routing engine %v", engine))
}

// getRoutesTemplateString Gets the contents of the HandleBars template to use.
// If `templateFilePath` is empty, returns the default, built-in template, otherwise, the content of the provided template
func getRoutesTemplateString(engine definitions.RoutingEngineType, templateFilePath string) (string, error) {
	if len(templateFilePath) == 0 {
		return getDefaultTemplate(engine), nil
	}

	return getTemplateData(RoutesTemplateName, templateFilePath)
}

func getTemplateData(name string, path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("could not read given template %s override at '%s' - %s", name, path, err.Error())
	}
	return string(data), nil
}

func overrideTemplates(config *definitions.GleeceConfig, templatePartials map[string]string) error {
	for name, path := range config.RoutesConfig.TemplateOverrides {
		if name == RoutesTemplateName {
			// Do not handle the routes template here, as it is handled separately see getRoutesTemplateString
			continue
		}

		_, exists := templatePartials[name]
		if !exists {
			var knownPartials []string
			for partialName := range templatePartials {
				knownPartials = append(knownPartials, partialName)
			}
			return fmt.Errorf(
				"partial '%s' is not a valid %s partial. Known partials: %s",
				name,
				config.RoutesConfig.Engine,
				strings.Join(knownPartials, ", "),
			)
		}

		data, err := getTemplateData(name, path)
		if err != nil {
			return err
		}
		templatePartials[name] = data
	}
	return nil
}

func loadTemplatesExtensions(config *definitions.GleeceConfig, templateExtensions map[string]string) error {
	for name, path := range config.RoutesConfig.TemplateExtensions {
		_, exists := templateExtensions[name]
		if !exists {
			var knownExtensions []string
			for partialName := range templateExtensions {
				knownExtensions = append(knownExtensions, partialName)
			}
			return fmt.Errorf(
				"The extension '%s' is not a valid %s extension. Known extensions: %s",
				name,
				config.RoutesConfig.Engine,
				strings.Join(knownExtensions, ", "),
			)
		}

		data, err := getTemplateData(name, path)
		if err != nil {
			return err
		}
		templateExtensions[name] = data
	}
	return nil
}

func registerPartials(config *definitions.GleeceConfig) error {
	var partials map[string]string
	var extensions map[string]string
	engine := config.RoutesConfig.Engine
	switch engine {
	case definitions.RoutingEngineGin:
		partials = gin.Partials
		extensions = gin.TemplateExtensions
	case definitions.RoutingEngineEcho:
		partials = echo.Partials
		extensions = echo.TemplateExtensions
	case definitions.RoutingEngineMux:
		partials = mux.Partials
		extensions = mux.TemplateExtensions
	case definitions.RoutingEngineFiber:
		partials = fiber.Partials
		extensions = fiber.TemplateExtensions
	case definitions.RoutingEngineChi:
		partials = chi.Partials
		extensions = chi.TemplateExtensions
	default:
		panic(fmt.Sprintf("Unknown routing engine type '%v'", engine))
	}

	// Make sure to work on a clone, to not apply changes on the code's map
	extensions = maps.Clone(extensions)
	partials = maps.Clone(partials)

	if err := overrideTemplates(config, partials); err != nil {
		return err
	}

	if err := loadTemplatesExtensions(config, extensions); err != nil {
		return err
	}

	// Partials are currently registered on the package level.
	// Will likely need to find a way to register on the template level.
	//
	// Note that registering the same partials twice causes a panic.
	if partialsRegistered {
		logger.Debug("Removing all partials before re-registering")
		raymond.RemoveAllPartials()
	}

	logger.Debug("Registering partials for '%v'", config.RoutesConfig.Engine)
	raymond.RegisterPartials(partials)
	logger.Debug("Registering extensions for '%v'", config.RoutesConfig.Engine)
	raymond.RegisterPartials(extensions)

	partialsRegistered = true
	return nil
}

func getOutputFileMod(requestedPermissions string) os.FileMode {
	defaultPermission := os.FileMode(0644)
	if len(requestedPermissions) <= 0 {
		return defaultPermission
	}
	parsed, err := definitions.PermissionStringToFileMod(requestedPermissions)
	if err != nil {
		logger.Warn("Could not convert permission string '%s' to FileMod. Using 0644 instead", requestedPermissions)
		return defaultPermission
	}
	return parsed

}

func GenerateRoutes(
	config *definitions.GleeceConfig,
	fullMeta pipeline.GleeceFlattenedMetadata, // Possibly not ideal, if we want route regeneration using partial information
) error {

	args := (*config).RoutesConfig

	if err := registerPartials(config); err != nil {
		return err
	}

	if !helpersRegistered {
		registerHandlebarsHelpers()
		helpersRegistered = true
	}

	ctx, err := GetTemplateContext(config, fullMeta)

	if err != nil {
		logger.Fatal("Could not create a context for the template rendering process")
		return err
	}

	dumpContext(ctx)

	template, err := getRoutesTemplateString(args.Engine, args.TemplateOverrides[RoutesTemplateName])
	if err != nil {
		logger.Fatal("Could not obtain the template file's contents")
		return err
	}

	result, err := raymond.Render(template, ctx)
	if err != nil {
		logger.Fatal("Could not render routes template - %v", err)
		return err
	}

	logger.Debug("Formatting %d bytes of output code", len(result))
	formattedOutput, err := compilation.OptimizeImportsAndFormat(result)
	if err != nil {
		logger.Warn("Could not format output - %v", err)
		formattedOutput = result
	}

	err = os.MkdirAll(filepath.Dir(args.OutputPath), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(args.OutputPath, []byte(formattedOutput), getOutputFileMod(args.OutputFilePerms))
	if err != nil {
		logger.Fatal("Could not write output file at '%s' with permissions '%v' - %v", args.OutputPath, args.OutputFilePerms, err)
		return err
	}

	return nil
}
