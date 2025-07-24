package routes

import (
	"time"

	"github.com/gopher-fleece/gleece/core/pipeline"
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
	Imports                 map[string][]string
	Controllers             []definitions.ControllerMetadata
	GenerationDate          string
	AuthConfig              definitions.AuthorizationConfig
	ValidateResponsePayload bool
	ExperimentalConfig      definitions.ExperimentalConfig
	Models                  definitions.Models
}

func GetTemplateContext(
	config *definitions.GleeceConfig,
	fullMeta pipeline.GleeceFlattenedMetadata,
) (RoutesContext, error) {
	ctx := RoutesContext{
		Imports:                 fullMeta.Imports,
		Controllers:             fullMeta.Flat,
		AuthConfig:              config.RoutesConfig.AuthorizationConfig,
		ValidateResponsePayload: config.RoutesConfig.ValidateResponsePayload,
		ExperimentalConfig:      config.ExperimentalConfig,
		Models:                  fullMeta.Models,
	}
	if len(config.RoutesConfig.PackageName) > 0 {
		ctx.PackageName = config.RoutesConfig.PackageName
	} else {
		ctx.PackageName = "routes"
	}

	ctx.GenerationDate = time.Now().Format(time.DateOnly)

	return ctx, nil
}
