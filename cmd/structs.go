package cmd

import (
	"github.com/haimkastner/gleece/cmd/arguments"
)

type KnownTemplate string

const (
	TemplateRoutes                    KnownTemplate = "routes"
	TemplateControllerResponsePartial KnownTemplate = "controller.response.partial"
)

type RoutingEngineType string

const (
	RoutingEngineGin RoutingEngineType = "gin"
)

type RoutesConfig struct {
	Engine            RoutingEngineType
	TemplateOverrides map[KnownTemplate]string
	OutputPath        string
	OutputFilePerms   arguments.FileModeArg
}
