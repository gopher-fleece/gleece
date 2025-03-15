package routes

import (
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
