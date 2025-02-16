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
	PackageName    string
	Controllers    []definitions.ControllerMetadata
	GenerationDate string
	AuthConfig     definitions.AuthorizationConfig
}

func GetTemplateContext(
	config definitions.RoutesConfig,
	controllers []definitions.ControllerMetadata,
) (RoutesContext, error) {
	ctx := RoutesContext{Controllers: controllers, AuthConfig: config.AuthorizationConfig}
	if len(config.PackageName) > 0 {
		ctx.PackageName = config.PackageName
	} else {
		ctx.PackageName = "routes"
	}

	ctx.GenerationDate = time.Now().Format(time.DateOnly)

	return ctx, nil
}
