package cmd

import (
	"os"

	"github.com/gopher-fleece/gleece/cmd/arguments"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/spf13/cobra"
)

var cliArgs = arguments.CliArguments{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gleece",
	Short: "Gleece - A Simplified Framework for Building REST APIs in Go",
	Long: `Gleece - A Simplified Framework for Building REST APIs in Go



                                             ░▒▒▒▒▒                                                  
                                              ▒▒▒▒▒▒▒▒░                                              
                              ▒▒▒▒▒▒▒           ▒▒▒▒▒▒▒░         ▒▒▒▒▒▒░                             
                            ▒▒▒▒▒▒▒▒▒▒    ░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒   ▒▒▒▒▒▒▒▒▒▒░                           
                           ▒▒▒█████   ░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██████▒▒▒░                          
                           ▒▒▓███   ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██████▒▒▒                          
                           ▒▓▓██  ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▓████▒▒▒                          
                           ▓▓▓▓  ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒███  ██▒▒▒▒▒▒▒▒█▓▓▓▒                           
                             ▓ ░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒███    ██▒▒▒▒▒▒▒▓▓▓                             
                               ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██████  █▒▒▒▒▒▒▒▓██ ░                           
                           ░  ▒▒▒▒▒             ░▒▒▒▒▒████████▒▒▒▒▒▒▒▓██  ░                          
                          ░░ ▒▒▒                    ▒▒▒▒████▒▒▒▒▒▒▒▒▓▓██  ░░░                        
                         ░░▒░▒░    █████████          ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▓▓██  ░░░                        
                        ░░░██░       █████              ▒▒▒▒▒▒▒▒▒▒▓▓▓██   ░░░░                       
                        ░░░██         ██                  ▒▒▒▒▒▒▓▓▓███   ░░░░░                       
                        ▓░░██         ██                   ▓▓▓▓▓▓███    ░░░░░██                      
                       ░░░░▓█▒░    ██████████             ░░▓█████    ░░░░░░░██                      
                      ░░░░░▒██░░░   ██ ██ ██           ░▓█████      ░░░░░░░░░███                     
                       ▓█░░░▒██░░░░░████████░░░░░░░░█████       ░░░░░░░░░░░░░███                     
                       ██░░░▓▒██░░░░░░░░░░░░░░░░████▒      ░░░░░░░░░░░░░░░░░░███                     
                      ░██░░░░▓▒▒██░░░░░░░░░░░████      ░░░░░░░░░░░░░░░░░░░░░░███                     
                      ░██░░░░░░▒▒▓███░░░░░░███     ░░░░░░░░░░░░░░░░░░░░░░░░░▒██░                     
                       ██▒░░░░░░░▓▒▒███████     ░░░░░░░░░░░░░░░░░░░░░░░░░░░▒▒██                      
                       ▒██▒▒░░░░░░░▒▒▒▒██    ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░▒▒███                      
                        ███▒▒▒░░░░░▒▒██    ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░▒▒▒███                       
                         █████▓░░▒▒▒██   ░░░░░░░░░░░░░░░░░░░░░░░░░░░░▒▒▒▒▒███                        
                          ▒██▒▒▒▒▒▒██   ░░░░░░░░░░░░░░░░░░░█░░░░▒▒▒▒▒▒▒▓████                         
                           ░██▓▒▒▒██   ░░░░░░░░░░▓░░░░░░░░▒▓▓▓▓▒▒▒▒▒▓█▓████                          
                             ███▒██   ░░░░░░░░░░░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▓▓███                            
                               ████  ▒▒░░░░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒███                             
                                 ██  ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒█████                               
                                     ▒▒▒▒▓▓▓▓▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▓████                                 
                                    ░░░░▒▓▓███▒▒▒▒▒▒▒▒▒▒▒▒▒▒▓███                                     
                                     ░░░░░▓▓█████████████████                                        
                                                   ▒▓▒                                               
                                                                                                     
                                                                                                     
                                                                                                     
                                     ░░░░░░░░░░░░░░░░░░░░░░░░░░░░                                    
                                                                                                     
                                                                                                    
Gleece is a developer-focused CLI and framework that simplifies the creation of REST APIs in Go.
Using powerful code generation, it automates boilerplate and structure, letting you focus on your application's core logic.
Define your API contracts using Go-native types, and let Gleece handle generating routes, handlers, and validators.
By enforcing consistency between your contracts and implementation, Gleece helps prevent common issues like API mismatches and unexpected requests or responses.

Whether you're building a simple service or a complex application, Gleece ensures consistency, scalability, and developer productivity`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Handle the verbosity flag here if you want it executed for every subcommand
		if cmd.Flag("verbosity") == nil {
			logger.SetLogLevel(logger.LogLevelInfo)
			return
		}

		verbosityInt, err := cmd.Flags().GetUint8("verbosity")
		if err != nil {
			logger.SetLogLevel(logger.LogLevelAll)
			logger.Warn("Could not obtain verbosity level from arguments. Fell back to 'all'. Error - %v", err)
			return
		}

		verbosity := logger.LogLevel(verbosityInt)
		logger.SetLogLevel(verbosity)
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info(`Gleece called with no parameters. Assuming 'generate spec-and-routes -c "./gleece.config.json"'`)
		err := GenerateSpecAndRoutes(arguments.CliArguments{ConfigPath: "./gleece.config.json"})
		if err != nil {
			logger.Fatal("Failed to generate spec and routes: %v", err)
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
	rootCmd.PersistentFlags().Uint8VarP(
		&cliArgs.Verbosity,
		"verbosity",
		"v",
		2,
		"Determines the verbosity of Gleece's traces. 0 = Output everything, 5 = Output fatal errors only",
	)

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(generateCmd)
}
