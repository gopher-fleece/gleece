package commandline_test

import (
	"os"

	"github.com/gopher-fleece/gleece/v2/cmd"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	"github.com/gopher-fleece/gleece/v2/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dump Command", func() {
	AfterEach(func() {
		utils.DeleteDistInCurrentFolderOrFail()
		logger.SetLogLevel(logger.LogLevelNone)
		cmd.Reset()
		os.Remove("./gleece.txt")
		os.Remove("./gleece.manual.txt")
	})

	It("Returns an error if an invalid dump format is given", func() {
		result := cmd.ExecuteWithArgs([]string{"dump", "graph", "-f=invalid"})
		Expect(result.Error).To(MatchError(ContainSubstring("invalid --format")))
		Expect(result.StdErr).To(ContainSubstring("Error: invalid --format"))
	})

	It("Returns an error if an given a non-existent Gleece config", func() {
		result := cmd.ExecuteWithArgs([]string{"dump", "graph", "-c=./does.not.exist.json"})
		Expect(result.Error).To(MatchError(ContainSubstring("failed to load Gleece config from './does.not.exist.json'")))
		Expect(result.StdErr).To(ContainSubstring("Error: failed to load Gleece config from './does.not.exist.json'"))
	})

	It("Correctly dumps graph using default params", func() {
		result := cmd.ExecuteWithArgs([]string{"dump", "graph", "--no-banner"})
		Expect(result.StdErr).To(BeEmpty())
		Expect(result.StdOut).To(ContainSubstring("digraph SymbolGraph {"))
	})

	It("Correctly dumps graph using dot format and default path", func() {
		result := cmd.ExecuteWithArgs([]string{"dump", "graph", "-f=dot", "-o=./gleece.txt"})
		Expect(result.Logs).To(ContainSubstring("Graph successfully written"))

		fileData, err := os.ReadFile("./gleece.txt")
		Expect(err).To(BeNil(), "Expected ./gleece.txt to exist after dump command execution")
		Expect(string(fileData)).To(ContainSubstring("digraph SymbolGraph {"))
	})

	It("Correctly dumps graph using specific params", func() {
		result := cmd.ExecuteWithArgs([]string{
			"dump",
			"graph",
			"-c=./gleece.test.config.json",
			"-f=plain",
			"-o=./gleece.manual.txt",
		})

		Expect(result.Logs).To(ContainSubstring("Dumping graph to 'plain' format"))
		fileData, err := os.ReadFile("./gleece.manual.txt")
		Expect(err).To(BeNil(), "Expected ./gleece.manual.txt to exist after dump command execution")
		Expect(string(fileData)).To(ContainSubstring("=== SymbolGraph Dump ==="))
	})
})
