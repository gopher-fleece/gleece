package main

import (
	"os"

	"github.com/haimkastner/gleece/cmd"
	"github.com/haimkastner/gleece/cmd/arguments"
	"github.com/haimkastner/gleece/extractor"
	"github.com/haimkastner/gleece/generator/routes"
	Logger "github.com/haimkastner/gleece/infrastructure/logger"
)

func main() {
	defs, err := extractor.GetMetadata()
	if err != nil {
		Logger.Fatal("Could not collect metadata - %v", err)
		os.Exit(1)
	}

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
