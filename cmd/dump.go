package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gopher-fleece/gleece/core/pipeline"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/spf13/cobra"
)

type DumpFormat string

const (
	DumpFormatDot   DumpFormat = "dot"
	DumpFormatPlain DumpFormat = "plain"
)

var (
	gleeceConfigPath string
	dumpOutput       string
	dumpFormat       string
)

// dumpCmd is a command used to dump information such as controllers, models, graph representations etc.
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dumps information about a Gleece project",
	Long: `The dump command returns information on the given Gleece project.
This information is obtained by running the project through the analysis facilities`,
}

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Dumps the primary symbolic graph to a textual representation",
	Long: `The 'graph' command can be used to dump a full representation of Gleece primary symbolic graph.
This graph depict all entities and their relations and is most useful as means to visualize entity relations and dependencies`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		switch dumpFormat {
		case string(DumpFormatDot), string(DumpFormatPlain), "":
		default:
			return fmt.Errorf("invalid --format: '%s' (want dot|plain)", dumpFormat)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := dumpGraph(cmd)
		if err != nil {
			logger.Fatal("Failed to dump graph - %v", err)
		}
		return err
	},
}

func initDumpCommandHierarchy() {
	graphCmd.Flags().StringVarP(
		&dumpFormat,
		"format",
		"f",
		"dot",
		"-f=dot",
	)

	dumpCmd.PersistentFlags().StringVarP(
		&dumpOutput,
		"output",
		"o",
		"",
		"-o \"/target-directory/dump.dot\"",
	)

	dumpCmd.PersistentFlags().StringVarP(
		&gleeceConfigPath,
		"config",
		"c",
		"./gleece.config.json",
		"-c \"/project-directory/gleece.config.json\"",
	)

	dumpCmd.AddCommand(graphCmd)
}

func resetDumpCommand() {
	gleeceConfigPath = ""
	dumpOutput = ""
	dumpFormat = ""
}

func loadGleeceConfig(cmd *cobra.Command) (*definitions.GleeceConfig, error) {
	gleeceConfigPath, err := cmd.Flags().GetString("config")
	if err != nil {
		logger.Error("Failed to read Gleece config location from CLI - %v", err)
		return nil, fmt.Errorf("failed to read Gleece config location from CLI - %v", err)
	}

	trimmedPath := strings.Trim(gleeceConfigPath, "\"")
	if trimmedPath == "" {
		trimmedPath = "./gleece.config.json"
		logger.Info("Gleece config path was no specified. Defaulting to '%s'", trimmedPath)
	}

	gleeceConfig, err := LoadGleeceConfig(trimmedPath)
	if err != nil {
		logger.Error("Failed to load Gleece config from '%s' - %v", trimmedPath, err)
		return nil, fmt.Errorf("failed to load Gleece config from '%s' - %v", trimmedPath, err)
	}

	return gleeceConfig, nil

}

func getPipeline(cmd *cobra.Command) (*pipeline.GleecePipeline, error) {
	gleeceConfig, err := loadGleeceConfig(cmd)
	if err != nil {
		return nil, err
	}

	pipe, err := pipeline.NewGleecePipeline(gleeceConfig)
	if err != nil {
		logger.Error("Failed to construct the analysis pipeline - %v", err)
		return nil, fmt.Errorf("failed to construct the analysis pipeline - %v", err)
	}

	return &pipe, nil
}

func dumpGraph(cmd *cobra.Command) error {
	logger.Info("Dumping graph to '%s' format", dumpFormat)
	pipe, err := getPipeline(cmd)
	if err != nil {
		return err
	}

	err = pipe.GenerateGraph()
	if err != nil {
		logger.Error("Failed to construct a symbol graph for the project - %v", err)
		return fmt.Errorf("failed to construct a symbol graph for the project - %v", err)
	}

	var text string
	switch dumpFormat {
	case string(DumpFormatDot):
		text = pipe.Graph().ToDot(nil)
	case string(DumpFormatPlain):
		text = pipe.Graph().String()
	default:
		logger.Debug("Dump format not specified - defaulting to DOT")
		text = pipe.Graph().ToDot(nil)
	}

	if dumpOutput == "" {
		logger.Debug("Output not specified - using stdout")
		cmd.Println(string(text))
		return nil
	}

	logger.Debug("Writing %d bytes to '%s'", len(text), dumpOutput)
	if err = os.WriteFile(dumpOutput, []byte(text), 0o644); err != nil {
		logger.Error("Failed to write to '%s' - %v", dumpOutput, err)
		return fmt.Errorf("failed to write to '%s' - %v", dumpOutput, err)
	}

	logger.Info("Graph successfully written to '%s'", dumpOutput)
	return nil
}
