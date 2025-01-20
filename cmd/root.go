package cmd

import (
	"os"

	"github.com/gopher-fleece/gleece/cmd/arguments"
	Logger "github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gleece",
	Short: "Gleece - A Simplified Framework for Building REST APIs in Go",
	Long: `Gleece - A Simplified Framework for Building REST APIs in Go

                          ⢀⣀⣤⣄⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
			⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⡖⠋⠀⠀⠀⣉⣷⠤⣶⡒⠦⣤⡴⠒⠒⠲⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
			⠀⠀⠀⠀⠀⢠⠴⠚⠉⠉⠓⢦⡀⢠⠞⠁⠀⠀⠀⠉⢢⡀⠀⠀⠀⠀⠘⠒⠒⠲⣄⠀⠀⠀⠀⠀⠀⠀
			⠀⠀⠀⠀⠀⠏⠀⠀⠀⢀⣀⡀⢹⡏⢀⣀⡀⠀⠀⠀⠀⢷⠀⠀⠀⠀⠀⠀⠀⠀⠸⠦⢤⡀⠀⠀⠀⠀
			⠀⠀⠀⢀⡴⢧⠀⠀⠀⠻⣿⠏⡿⣇⠻⠿⠇⠀⠀⠀⠀⡞⢶⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⣆⠀⠀⠀
			⠀⠀⣰⠋⠀⠈⣿⠶⠤⠤⠔⠋⠁⠘⢦⡀⠀⠀⠀⣠⣾⡁⠀⠙⢦⠀⠀⠀⠀⠀⠀⠀⠀⠠⣿⣀⠀⠀
			⠀⣼⠁⠀⠀⢰⣿⠀⠀⠀⠀⠀⠀⠀⠀⠈⠉⠛⠉⠁⣿⡇⠀⠀⠈⢧⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⢳⠀
			⢰⠇⠀⠀⠀⣼⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣴⣀⠀⣿⡇⠀⠀⠀⢸⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⣸⠃
			⢸⡄⠀⠀⢀⡏⢸⡀⠀⠀⠀⠀⠀⠀⠀⠀⠘⠀⢻⣿⡏⢷⠀⠀⠀⣰⠇⠀⠀⠀⠀⠀⠀⠀⠀⢾⡁⠀
			⠀⠙⠢⢶⣞⠁⠀⢳⡀⠀⠦⣄⣀⢀⣀⣠⠀⠀⢀⡿⠀⠈⠓⠦⠶⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢹⠀
			⠀⠀⠀⠀⢨⡗⠀⠀⠑⢤⡀⠀⢹⠉⠁⠀⢀⣤⠞⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡾⠀
			⠀⠀⠀⠀⠸⡄⠀⠀⠀⠀⠉⠙⠚⠓⠒⠊⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠺⡏⠀⠀
			⠀⠀⠀⠀⠀⠙⠦⠤⠤⡖⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡇⠀⠀
			⠀⠀⠀⠀⠀⠀⠀⠀⠀⢷⡀⠀⠀⢀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⠦⠤⠴⠋⠀⠀⠀
			⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠛⡾⠉⣽⡀⠀⠀⠀⣰⡄⠀⠀⠀⢀⣶⡀⠀⠀⢀⣼⡄⠀⠀⠀⠀⠀⠀
			⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⠇⣰⢫⠟⣲⠖⠚⠁⠙⠲⠤⠔⠋⠀⣟⠓⡞⣏⠈⡇⠀⠀⠀⠀⠀⠀
			⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⡾⢠⣇⣼⢠⣯⣀⣤⣤⣤⣤⣤⣤⣄⣀⣼⡀⣧⣸⡀⣇⣀⠀⠀⠀⠀⠀
			⠀⠀⠀⠀⠀⠾⣿⣿⣿⣿⣿⣷⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣿⣿⣿⣿⣷⠶⠀
			⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠉⠉⠉⠉⠉⠉⠙⠛⠛⠛⠛⠛⠛⠛⠛⠋⠉⠉⠉⠉⠉⠉⠁⠀⠀⠀⠀⠀

Gleece is a developer-focused CLI and framework that simplifies the creation of REST APIs in Go.
Using powerful code generation, it automates boilerplate and structure, letting you focus on your application's core logic.
Define your API contracts using Go-native types, and let Gleece handle generating routes, handlers, and validators.
By enforcing consistency between your contracts and implementation, Gleece helps prevent common issues like API mismatches and unexpected requests or responses.

Whether you're building a simple service or a complex application, Gleece ensures consistency, scalability, and developer productivity`,
	Run: func(cmd *cobra.Command, args []string) {
		err := generateSpecAndRoutes(arguments.CliArguments{ConfigPath: "./gleece.config.json"})
		if err != nil {
			Logger.Fatal("Failed to generate spec and routes: %v", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	initGenerateCommandHierarchy()
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(generateCmd)
}
