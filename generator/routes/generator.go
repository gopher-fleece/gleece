package routes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aymerick/raymond"
	"github.com/iancoleman/strcase"

	"github.com/gopher-fleece/gleece/cmd"
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
}

func getDefaultTemplate(engine cmd.RoutingEngineType) string {
	switch engine {
	case cmd.RoutingEngineGin:
		return gin.RoutesTemplate
	}
	// This should not happen. It indicates a breakage in the build itself.
	panic(fmt.Sprintf("Could not find an embedded template for routing engine %v", engine))
}

// getTemplateString Gets the contents of the HandleBars template to use.
// If `templateFilePath` is empty, returns the default, built-in template, otherwise, the content of the provided template
func getTemplateString(engine cmd.RoutingEngineType, templateFilePath string) (string, error) {
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

func registerPartials(engine cmd.RoutingEngineType) {
	switch engine {
	case cmd.RoutingEngineGin:
		gin.RegisterPartials()
	default:
		panic(fmt.Sprintf("Unknown routing engine type '%v'", engine))
	}
}

func GenerateRoutes(args cmd.RoutesConfig, controllerMeta []definitions.ControllerMetadata) error {
	registerPartials(args.Engine)
	registerHelpers()

	ctx, err := GetTemplateContext(args, controllerMeta)

	if err != nil {
		Logger.Fatal("Could not create a context for the template rendering process")
		return err
	}

	dumpContext(ctx)

	template, err := getTemplateString(args.Engine, args.TemplateOverrides[cmd.TemplateRoutes])
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
	formattedOutput, _ := compilation.OptimizeImportsAndFormat(result)
	if err := os.MkdirAll(filepath.Dir(args.OutputPath), 0755); err != nil {
		return err
	}
	err = os.WriteFile(args.OutputPath, []byte(formattedOutput), args.OutputFilePerms.FileMode())
	if err != nil {
		Logger.Fatal("Could not write output file at '%s' with permissions '%v' - %v", args.OutputPath, args.OutputFilePerms, err)
		return err
	}

	return nil
}
