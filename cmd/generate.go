package cmd

import (
	"os"

	"github.com/gopher-fleece/gleece/cmd/arguments"
	Logger "github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/spf13/cobra"
)

var cliArgs = arguments.CliArguments{}

var generateCmd = &cobra.Command{
	Use:   "generate spec-and-routes --config \"/path/to/gleece.config.json\"",
	Short: "Generate OpenAPI schema and routing middlewares from a Go project",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var specCommand = &cobra.Command{
	Use:   "spec",
	Short: "Generates an OpenAPI schema from the codebase using the specified configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		err := generateSpec(cliArgs)
		if err != nil {
			Logger.Fatal("Failed to generate spec: %v", err)
			os.Exit(1)
		}
	},
}

var routesCommand = &cobra.Command{
	Use:   "routes",
	Short: "Generates a routing middleware file from the codebase using the specified configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		err := GenerateRoutes(cliArgs)
		if err != nil {
			Logger.Fatal("Failed to generate routes: %v", err)
			os.Exit(1)
		}
	},
}

var specAndRoutesCommand = &cobra.Command{
	Use:   "spec-and-routes",
	Short: "Generates an OpenAPI schema and a routing middleware file from the codebase using the specified configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		err := generateSpecAndRoutes(cliArgs)
		if err != nil {
			Logger.Fatal("Failed to generate spec and routes: %v", err)
			os.Exit(1)
		}
	},
}

func initGenerateCommandHierarchy() {
	generateCmd.PersistentFlags().StringVarP(&cliArgs.ConfigPath, "config", "c", "", "/project-directory/gleece.config.json")
	generateCmd.MarkFlagRequired("config")

	generateCmd.AddCommand(specCommand)
	generateCmd.AddCommand(routesCommand)
	generateCmd.AddCommand(specAndRoutesCommand)
}
