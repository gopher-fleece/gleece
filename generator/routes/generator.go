package routes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aymerick/raymond"
	"github.com/iancoleman/strcase"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/generator/compilation"
	"github.com/gopher-fleece/gleece/generator/templates/gin"
	Logger "github.com/gopher-fleece/gleece/infrastructure/logger"
)

// dumpContext Dumps the routes context to the log for debugging purposes
func dumpContext(ctx any) {
	// It's basically impossible to break the marshaller without some pretty crazy, and very intentional hacks
	contextJson, _ := json.MarshalIndent(ctx, "\t", "\t")
	Logger.Debug("Routes Context:\n%s", contextJson)
}

func registerHelpers() {
	raymond.RegisterHelper("ToLowerCamel", func(arg string) string {
		return strcase.ToLowerCamel(arg)
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
}

func getDefaultTemplate(engine definitions.RoutingEngineType) string {
	switch engine {
	case definitions.RoutingEngineGin:
		return gin.RoutesTemplate
	}
	// This should not happen. It indicates a breakage in the build itself.
	panic(fmt.Sprintf("Could not find an embedded template for routing engine %v", engine))
}

// getTemplateString Gets the contents of the HandleBars template to use.
// If `templateFilePath` is empty, returns the default, built-in template, otherwise, the content of the provided template
func getTemplateString(engine definitions.RoutingEngineType, templateFilePath string) (string, error) {
	if len(templateFilePath) == 0 {
		return getDefaultTemplate(engine), nil
	}

	data, err := os.ReadFile(templateFilePath)
	if err != nil {
		Logger.Fatal("Could not read given template override from '%s' - %v", templateFilePath, err)
		return "", err
	}

	return string(data), nil
}

func registerPartials(engine definitions.RoutingEngineType) {
	switch engine {
	case definitions.RoutingEngineGin:
		gin.RegisterPartials()
	default:
		panic(fmt.Sprintf("Unknown routing engine type '%v'", engine))
	}
}

func getOutputFileMod(requestedPermissions string) os.FileMode {
	defaultPermission := os.FileMode(0644)
	if len(requestedPermissions) <= 0 {
		return defaultPermission
	}
	parsed, err := definitions.PermissionStringToFileMod(requestedPermissions)
	if err != nil {
		Logger.Warn("Could not convert permission string '%s' to FileMod. Using 0644 instead", requestedPermissions)
		return defaultPermission
	}
	return parsed

}

func GenerateRoutes(config *definitions.GleeceConfig, controllerMeta []definitions.ControllerMetadata) error {
	args := (*config).RoutesConfig

	registerPartials(args.Engine)
	registerHelpers()

	ctx, err := GetTemplateContext(args, controllerMeta)

	if err != nil {
		Logger.Fatal("Could not create a context for the template rendering process")
		return err
	}

	dumpContext(ctx)

	template, err := getTemplateString(args.Engine, args.TemplateOverrides[definitions.TemplateRoutes])
	if err != nil {
		Logger.Fatal("Could not obtain the template file's contents")
		return err
	}

	result, err := raymond.Render(template, ctx)
	if err != nil {
		Logger.Fatal("Could not render routes template - %v", err)
		return err
	}

	Logger.Debug("Formatting %d bytes of output code", len(result))
	formattedOutput, err := compilation.OptimizeImportsAndFormat(result)
	if err != nil {
		Logger.Warn("Could not format output - %v", err)
		formattedOutput = result
	}

	err = os.MkdirAll(filepath.Dir(args.OutputPath), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(args.OutputPath, []byte(formattedOutput), getOutputFileMod(args.OutputFilePerms))
	if err != nil {
		Logger.Fatal("Could not write output file at '%s' with permissions '%v' - %v", args.OutputPath, args.OutputFilePerms, err)
		return err
	}

	return nil
}
