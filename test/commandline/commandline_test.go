package imports_test

import (
	"fmt"
	"testing"

	"github.com/gopher-fleece/gleece/v2/cmd"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	"github.com/gopher-fleece/gleece/v2/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = AfterEach(func() {
	utils.DeleteDistInCurrentFolderOrFail()
	logger.SetLogLevel(logger.LogLevelNone)
})

var _ = Describe("Commandline", func() {
	defer func() {
		if r := recover(); r != nil {
			// If a panic occurs, fail the test
			Fail(fmt.Sprintf("CLI test panicked - %v", r))
		}
	}()

	Context("PersistentPreRun", func() {
		It("Prints banner if no-banner is not specified", func() {
			logger.SetLogLevel(logger.LogLevelAll)
			result := cmd.ExecuteWithArgs([]string{"version"}, true)
			Expect(result.Logs).To(ContainSubstring("▒▒▒██████  █▒▒▒▒"))
		})
	})

	Context("Run", func() {
		It("Correctly detects, runs and returns when called with no parameters", func() {
			result := cmd.ExecuteWithArgs([]string{}, true)
			// We expect a failure since there's not gleece.config.file here - doesn't matter, we just need to verify the no-params case.
			Expect(result.Logs).To(ContainSubstring("Gleece called with no parameters"))
		})
	})

	It("Generate spec should complete successfully", func() {
		absPath := utils.GetAbsPathByRelativeOrFail("./gleece.test.config.json")

		result := cmd.ExecuteWithArgs([]string{"generate", "spec", "--no-banner", "-c", absPath}, true)
		Expect(result.Error).To(BeNil())
		Expect(result.StdErr).To(BeEmpty())
		Expect(result.Logs).ToNot(BeEmpty())
		Expect(result.Logs).To(ContainSubstring("[INFO]   Generating spec"))
		Expect(result.Logs).To(ContainSubstring("Security spec generated successfully"))
		Expect(result.Logs).To(ContainSubstring("Models spec generated successfully"))
		Expect(result.Logs).To(ContainSubstring("Controllers spec generated successfully"))
		Expect(result.Logs).To(ContainSubstring("OpenAPI specification validated successfully"))
		Expect(result.Logs).To(ContainSubstring("OpenAPI specification generation completed successfully"))
		Expect(result.Logs).To(ContainSubstring("OpenAPI specification written to"))
		Expect(result.Logs).To(ContainSubstring("[INFO]   Spec successfully generated"))
	})

	It("Generate routes should complete successfully", func() {

		absPath := utils.GetAbsPathByRelativeOrFail("./gleece.test.config.json")

		result := cmd.ExecuteWithArgs([]string{"generate", "routes", "--no-banner", "-c", absPath}, true)
		Expect(result.Error).To(BeNil())
		Expect(result.StdErr).To(BeEmpty())
		Expect(result.Logs).ToNot(BeEmpty())
		Expect(result.Logs).To(ContainSubstring("[INFO]   Generating routes"))
		Expect(result.Logs).To(ContainSubstring("[INFO]   Routes successfully generated"))
	})

	It("Generate spec-and-routes should complete successfully", func() {
		defer func() {
			if r := recover(); r != nil {
				// If a panic occurs, fail the test
				Fail(fmt.Sprintf("CLI test panicked - %v", r))
			}
		}()

		absPath := utils.GetAbsPathByRelativeOrFail("./gleece.test.config.json")

		result := cmd.ExecuteWithArgs([]string{"generate", "spec-and-routes", "--no-banner", "-c", absPath}, true)
		Expect(result.Error).To(BeNil())
		Expect(result.StdErr).To(BeEmpty())
		Expect(result.Logs).ToNot(BeEmpty())
		Expect(result.Logs).To(ContainSubstring("[INFO]   Generating spec and routes"))
		Expect(result.Logs).To(ContainSubstring("Security spec generated successfully"))
		Expect(result.Logs).To(ContainSubstring("Models spec generated successfully"))
		Expect(result.Logs).To(ContainSubstring("Controllers spec generated successfully"))
		Expect(result.Logs).To(ContainSubstring("OpenAPI specification validated successfully"))
		Expect(result.Logs).To(ContainSubstring("OpenAPI specification generation completed successfully"))
		Expect(result.Logs).To(ContainSubstring("OpenAPI specification written to"))
		Expect(result.Logs).To(ContainSubstring("[INFO]   Spec and routes successfully generated"))
	})

	Context("Version Command", func() {
		It("Prints expected version information", func() {
			// Note that Version, Build Date, Commit, etc. should be populated during build so they're expected to be empty
			result := cmd.ExecuteWithArgs([]string{"version"}, true)
			Expect(result.StdOut).To(Equal("Gleece\nVersion: \nBuild Date: \nCommit: \nTarget architecture: \nTarget platform: \n"))

		})
	})
})

func TestCommandline(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commandline")
}
