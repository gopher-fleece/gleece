package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version    string
	BuildDate  string
	Commit     string
	TargetOs   string
	TargetArch string
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display Gleece's version and build information",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Gleece")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Date: %s\n", BuildDate)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Target architecture: %s\n", TargetOs)
		fmt.Printf("Target platform: %s\n", TargetArch)
	},
}
