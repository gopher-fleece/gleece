package security_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/cmd/arguments"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var metadata []definitions.ControllerMetadata

var _ = BeforeSuite(func() {
	cwd, err := os.Getwd()
	if err != nil {
		Fail(fmt.Sprintf("Could not determine process working directory - %v", err))
	}

	configPath := filepath.Join(cwd, "security.config.json")
	_, controllers, _, _, err := cmd.GetConfigAndMetadata(arguments.CliArguments{ConfigPath: configPath})
	if err != nil {
		Fail(fmt.Sprintf("Could not generate routes - %v", err))
	}

	metadata = controllers
})

var _ = Describe("Security Controller", func() {
	It("Created metadata has length", func() {
		Expect(metadata).ToNot(BeNil())
		Expect(metadata).To(HaveLen(1))
	})

	It("Produces correct controller level security", func() {
		controllerMeta := metadata[0]
		Expect(controllerMeta.Security).To(HaveLen(0))
	})

	It("Produces correct route level security", func() {
		route := metadata[0].Routes[0]
		Expect(route.Security).To(HaveLen(2))
		Expect(route.Security[0].SecurityAnnotation).To(HaveLen(1))
		Expect(route.Security[0].SecurityAnnotation[0].SchemaName).To(Equal("secSchema1"))
		Expect(route.Security[0].SecurityAnnotation[0].Scopes).To(HaveLen(1))
		Expect(route.Security[0].SecurityAnnotation[0].Scopes[0]).To(Equal("scope1"))

		Expect(route.Security[1].SecurityAnnotation).To(HaveLen(1))
		Expect(route.Security[1].SecurityAnnotation[0].SchemaName).To(Equal("secSchema2"))
		Expect(route.Security[1].SecurityAnnotation[0].Scopes).To(HaveLen(2))
		Expect(route.Security[1].SecurityAnnotation[0].Scopes).To(HaveExactElements("2", "3"))
	})
})

func TestSecurityController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Security Controller")
}
