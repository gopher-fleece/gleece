package errorhandling_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/cmd/arguments"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error-handling", func() {
	It("Returns a clear error when configuration is not found", func() {
		_, _, _, _, err := cmd.GetConfigAndMetadata(arguments.CliArguments{ConfigPath: "/this/path/does/not/exist.json"})
		Expect(err).To(MatchError(ContainSubstring("could not read config file from")))
		// Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
	})

	It("Returns a clear error when configuration is syntactically broken", func() {
		configPath := utils.GetAbsPathByRelative("gleece.broken.json.config")
		_, _, _, _, err := cmd.GetConfigAndMetadata(arguments.CliArguments{ConfigPath: configPath})
		Expect(err).To(MatchError(ContainSubstring("could not unmarshal config file")))
		Expect(err).To(MatchError(ContainSubstring("invalid character")))
	})

	It("Returns a clear error when configuration fails validation", func() {
		configPath := utils.GetAbsPathByRelative("gleece.invalid.config.json")
		_, _, _, _, err := cmd.GetConfigAndMetadata(arguments.CliArguments{ConfigPath: configPath})
		Expect(err).To(MatchError(ContainSubstring("Field 'ControllerGlobs' failed validation with tag 'min'")))
	})

	It("Returns a clear error when configuration has a non-existent template override", func() {
		configPath := utils.GetAbsPathByRelative("gleece.missing.partial.config.json")
		err := cmd.GenerateRoutes(arguments.CliArguments{ConfigPath: configPath})

		Expect(err).To(MatchError(ContainSubstring("could not read given template Imports override at")))
		// Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
	})

	It("Returns a clear error when configuration references a non-existent partial", func() {
		configPath := utils.GetAbsPathByRelative("gleece.unknown.partial.config.json")
		err := cmd.GenerateRoutes(arguments.CliArguments{ConfigPath: configPath})

		Expect(err).To(MatchError(ContainSubstring("partial 'thisPartialDoesNotExist' is not a valid gin partial")))
	})

	It("Returns a clear error when configuration has a syntactically broken template override", func() {
		configPath := utils.GetAbsPathByRelative("gleece.broken.override.syntax.config.json")
		err := cmd.GenerateRoutes(arguments.CliArguments{ConfigPath: configPath})

		Expect(err).To(MatchError(ContainSubstring("Evaluation error")))
		Expect(err).To(MatchError(ContainSubstring("Lexer error")))
		Expect(err).To(MatchError(ContainSubstring("Unclosed expression")))
	})
})

func TestErrorHandling(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Error-handling")
}
