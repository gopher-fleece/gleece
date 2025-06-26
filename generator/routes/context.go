package routes

import (
	"sort"
	"time"

	"github.com/gopher-fleece/gleece/definitions"
)

type Argument struct {
	Type      definitions.ParamPassedIn
	Name      string
	ValueType any
}

type RouteCtx struct {
	definitions.RouteMetadata
}

type ControllerMeta struct {
	Routes []RouteCtx
}

type PackageImport struct {
	FullPath string
	Alias    string
}

type RoutesContext struct {
	PackageName             string
	Controllers             []definitions.ControllerMetadata
	GenerationDate          string
	AuthConfig              definitions.AuthorizationConfig
	ValidateResponsePayload bool
	ExperimentalConfig      definitions.ExperimentalConfig
	Models                  definitions.Models
}

func GetTemplateContext(
	models *definitions.Models,
	config *definitions.GleeceConfig,
	controllers []definitions.ControllerMetadata,
) (RoutesContext, error) {

	if models == nil {
		models = &definitions.Models{
			Structs: make([]definitions.StructMetadata, 0),
			Enums:   make([]definitions.EnumMetadata, 0),
		}
	}

	// Sort template data arrays, so each generate will produce the same code.
	// It useful when the generated code is managed by source-control (e.g. git) and needs to avoid diffs without logic change.

	// Sort controllers by Name property
	sort.Slice(controllers, func(i, j int) bool {
		return controllers[i].Name < controllers[j].Name
	})

	// Sort Routes inside each controller by OperationId
	for i := range controllers {
		sort.Slice(controllers[i].Routes, func(a, b int) bool {
			return controllers[i].Routes[a].OperationId < controllers[i].Routes[b].OperationId
		})
	}

	ctx := RoutesContext{
		Controllers:             controllers,
		AuthConfig:              config.RoutesConfig.AuthorizationConfig,
		ValidateResponsePayload: config.RoutesConfig.ValidateResponsePayload,
		ExperimentalConfig:      config.ExperimentalConfig,
		Models:                  *models,
	}
	if len(config.RoutesConfig.PackageName) > 0 {
		ctx.PackageName = config.RoutesConfig.PackageName
	} else {
		ctx.PackageName = "routes"
	}

	ctx.GenerationDate = time.Now().Format(time.DateOnly)

	return ctx, nil
}
