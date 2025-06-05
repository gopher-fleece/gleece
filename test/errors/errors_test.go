package imports_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var metadata []definitions.ControllerMetadata
var models []definitions.StructMetadata

var _ = BeforeSuite(func() {
	controllers, flatModels, _ := utils.GetControllersAndModelsOrFail()
	metadata = controllers
	models = flatModels
})

var _ = Describe("Errors Controller", func() {
	It("Simple errors should be properly detected and resolved", func() {
		route := metadata[0].Routes[0]

		Expect(route.Responses).To(HaveLen(1))
		Expect(route.Responses[0].TypeMetadata.Name).To(Equal("SimpleCustomError"))
		Expect(route.Responses[0].TypeMetadata.PkgPath).To(Equal("github.com/gopher-fleece/gleece/test/errors"))
		Expect(route.Responses[0].TypeMetadata.DefaultPackageAlias).To(Equal("errors"))
		Expect(route.Responses[0].TypeMetadata.Import).To(Equal(definitions.ImportTypeNone))
		Expect(route.Responses[0].TypeMetadata.IsUniverseType).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.IsByAddress).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.SymbolKind).To(Equal(definitions.SymKindStruct))
	})

	It("Complex errors should be properly detected and resolved", func() {
		route := metadata[0].Routes[1]

		Expect(route.Responses).To(HaveLen(1))
		Expect(route.Responses[0].TypeMetadata.Name).To(Equal("ComplexCustomError"))
		Expect(route.Responses[0].TypeMetadata.PkgPath).To(Equal("github.com/gopher-fleece/gleece/test/errors"))
		Expect(route.Responses[0].TypeMetadata.DefaultPackageAlias).To(Equal("errors"))
		Expect(route.Responses[0].TypeMetadata.Import).To(Equal(definitions.ImportTypeNone))
		Expect(route.Responses[0].TypeMetadata.IsUniverseType).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.IsByAddress).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.SymbolKind).To(Equal(definitions.SymKindStruct))
	})
})

func TestErrorsController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Errors Controller")
}
