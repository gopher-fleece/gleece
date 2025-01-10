package main

import (
	"github.com/haimkastner/gleece/cmd"
	"github.com/haimkastner/gleece/cmd/arguments"
	"github.com/haimkastner/gleece/extractor"
	"github.com/haimkastner/gleece/generator/routes"
)

func main() {
	defs, _ := extractor.ExtractMetadata()
	routes.GenerateRoutes(
		cmd.RoutesConfig{
			Engine:            cmd.RoutingEngineGin,
			TemplateOverrides: map[cmd.KnownTemplate]string{},
			OutputPath:        "/mnt/7e91759c-6dd7-4c99-8d38-e6422452a469/git/gleece/dist/routes.go",
			OutputFilePerms:   arguments.NewFileModeArg("0644"),
		},
		defs,
	)
}
