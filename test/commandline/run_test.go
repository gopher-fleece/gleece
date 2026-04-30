package commandline_test

import (
	"fmt"

	"github.com/gopher-fleece/gleece/v2/cmd"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Run", func() {
	defer func() {
		if r := recover(); r != nil {
			// If a panic occurs, fail the test
			Fail(fmt.Sprintf("CLI test panicked - %v", r))
		}
	}()

	Context("PersistentPreRun", func() {
		It("Prints banner if no-banner is not specified", func() {
			logger.SetLogLevel(logger.LogLevelAll)
			result := cmd.ExecuteWithArgs([]string{"version"})
			Expect(result.StdOut).To(ContainSubstring("▒▒▒██████  █▒▒▒▒"))
		})
	})

	Context("Run", func() {
		It("Correctly detects, runs and returns when called with no parameters", func() {
			result := cmd.ExecuteWithArgs([]string{})
			// We expect a failure since there's not gleece.config.file here - doesn't matter, we just need to verify the no-params case.
			Expect(result.Logs).To(ContainSubstring("Gleece called with no parameters"))
		})
	})
})
