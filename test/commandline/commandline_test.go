package imports_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = AfterEach(func() {
	distPath := utils.GetAbsPathByRelative("dist")
	os.RemoveAll(distPath)
})

var _ = Describe("Commandline", func() {
	It("Generate spec should complete successfully", func() {
		defer func() {
			if r := recover(); r != nil {
				// If a panic occurs, fail the test
				Fail(fmt.Sprintf("CLI test panicked - %v", r))
			}
		}()

		absPath := utils.GetAbsPathByRelative("./gleece.test.config.json")

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
		defer func() {
			if r := recover(); r != nil {
				// If a panic occurs, fail the test
				Fail(fmt.Sprintf("CLI test panicked - %v", r))
			}
		}()

		absPath := utils.GetAbsPathByRelative("./gleece.test.config.json")

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

		absPath := utils.GetAbsPathByRelative("./gleece.test.config.json")

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
})

func TestCommandline(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commandline")
}
