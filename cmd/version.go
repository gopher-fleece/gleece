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
		out := cmd.OutOrStdout()
		fmt.Fprintln(out, "Gleece")
		fmt.Fprintf(out, "Version: %s\n", Version)
		fmt.Fprintf(out, "Build Date: %s\n", BuildDate)
		fmt.Fprintf(out, "Commit: %s\n", Commit)
		fmt.Fprintf(out, "Target architecture: %s\n", TargetOs)
		fmt.Fprintf(out, "Target platform: %s\n", TargetArch)
	},
}
