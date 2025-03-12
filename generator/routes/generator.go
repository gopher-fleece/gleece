package routes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/iancoleman/strcase"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/generator/compilation"
	"github.com/gopher-fleece/gleece/generator/templates/chi"
	"github.com/gopher-fleece/gleece/generator/templates/echo"
	"github.com/gopher-fleece/gleece/generator/templates/fiber"
	"github.com/gopher-fleece/gleece/generator/templates/gin"
	"github.com/gopher-fleece/gleece/generator/templates/mux"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

var RoutesTemplateName = "Routes"

var helpersRegistered bool
var lastEngine *definitions.RoutingEngineType

// dumpContext Dumps the routes context to the log for debugging purposes
func dumpContext(ctx any) {
	// It's basically impossible to break the marshaller without some pretty crazy, and very intentional hacks
	contextJson, _ := json.MarshalIndent(ctx, "\t", "\t")
	logger.Debug("Routes Context:\n%s", contextJson)
}

func registerHelpers() {
	raymond.RegisterHelper("ToLowerCamel", func(arg string) string {
		return strcase.ToLowerCamel(arg)
	})

	raymond.RegisterHelper("ToUpperCamel", func(arg string) string {
		return strcase.ToCamel(arg)
	})

	raymond.RegisterHelper("LastTypeNameEquals", func(types []definitions.FuncReturnValue, value string, options *raymond.Options) string {
		if len(types) <= 0 {
			panic("LastTypeNameEquals received a 0-length array")
		}

		if types[len(types)-1].Name == value {
			return options.Fn()
		}

		return options.Inverse()
	})

	raymond.RegisterHelper("ifEqual", func(a interface{}, b interface{}, options *raymond.Options) string {
		if raymond.Str(a) == raymond.Str(b) {
			return options.Fn()
		}

		return options.Inverse()
	})

	raymond.RegisterHelper("ifAnyParamRequiresConversion", func(params []definitions.FuncParam, options *raymond.Options) string {
		for _, param := range params {
			if param.TypeMeta.Name != "string" && param.TypeMeta.FullyQualifiedPackage != "" && param.TypeMeta.EntityKind != definitions.AstNodeKindAlias {
				// Currently, only 'string' parameters don't undergo any validation
				return options.Fn()
			}
		}

		return options.Inverse()
	})

	raymond.RegisterHelper("LastTypeIsByAddress", func(types []definitions.FuncReturnValue, options *raymond.Options) string {
		if len(types) <= 0 {
			panic("LastTypeIsByAddress received a 0-length array")
		}

		if types[len(types)-1].IsByAddress {
			return options.Fn()
		}

		return options.Inverse()
	})

	raymond.RegisterHelper("GetLastTyeFullyQualified", func(types []definitions.FuncReturnValue) string {
		if len(types) <= 0 {
			panic("GetLastTyeFullyQualified received a 0-length array")
		}

		last := types[len(types)-1]
		return fmt.Sprintf("Response%d%s.%s", last.UniqueImportSerial, last.Name, last.Name)
	})

	raymond.RegisterHelper("OrEqual", func(val1, comp1, val2, comp2 interface{}) bool {
		isEqual1 := (val1 == comp1)
		isEqual2 := (val2 == comp2)
		return isEqual1 || isEqual2
	})

	helpersRegistered = true
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
	if lastEngine != nil {
		if *lastEngine == config.RoutesConfig.Engine {
			logger.Debug("Last engine is %v, partials are already registered", *lastEngine)
			return nil
		}

		logger.Debug("Last used engine was %v, removing all partials before re-registration'", *lastEngine)
		raymond.RemoveAllPartials()
	}

	logger.Debug("Registering partials for '%v'", config.RoutesConfig.Engine)
	raymond.RegisterPartials(partials)
	logger.Debug("Registering extensions for '%v'", config.RoutesConfig.Engine)
	raymond.RegisterPartials(extensions)

	lastEngine = &engine
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

func GenerateRoutes(config *definitions.GleeceConfig, controllerMeta []definitions.ControllerMetadata) error {
	args := (*config).RoutesConfig

	if err := registerPartials(config); err != nil {
		return err
	}

	if !helpersRegistered {
		registerHelpers()
	}

	ctx, err := GetTemplateContext(args, controllerMeta)

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
