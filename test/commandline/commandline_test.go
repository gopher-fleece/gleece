package imports_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Errors Controller", func() {
	It("Simple errors should be properly detected and resolved", func() {
		defer func() {
			if r := recover(); r != nil {
				// If a panic occurs, fail the test
				Fail(fmt.Sprintf("CLI test panicked - %v", r))
			}
		}()

		absPath, err := filepath.Abs("./gleece.test.config.json")
		if err != nil {
			Fail(fmt.Sprintf("Failed to resolve absolute path for config - %v", err))
		}

		result := cmd.ExecuteWithArgs([]string{"generate", "spec-and-routes", "--no-banner", "-c", absPath}, true)
		Expect(result.Error).To(BeNil())
		Expect(result.StdErr).To(BeEmpty())
		Expect(result.Logs).ToNot(BeEmpty())
		Expect(result.Logs).To(ContainSubstring("Generating spec and routes"))
		Expect(result.Logs).To(ContainSubstring("Security spec generated successfully"))
		Expect(result.Logs).To(ContainSubstring("Models spec generated successfully"))
		Expect(result.Logs).To(ContainSubstring("Controllers spec generated successfully"))
		Expect(result.Logs).To(ContainSubstring("OpenAPI specification validated successfully"))
		Expect(result.Logs).To(ContainSubstring("OpenAPI specification generation completed successfully"))
		Expect(result.Logs).To(ContainSubstring("OpenAPI specification written to"))
		Expect(result.Logs).To(ContainSubstring("Spec and routes successfully generated"))
	})
})

func TestCommandline(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commandline")
}
