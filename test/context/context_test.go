package sanity_test

import (
	"fmt"
	"testing"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/generator/routes"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = AfterEach(func() {
	utils.DeleteDistInCurrentFolderOrFail()
})

var _ = Describe("Context Controller", func() {
	It("Does not raise an error when a receiver has a Go Context as a parameter", func() {
		utils.GetControllersAndModelsOrFail()
	})

	It("Does not include Go Context parameters in the model list", func() {
		_, modelsList, _ := utils.GetControllersAndModelsOrFail()
		Expect(modelsList).ToNot(ContainElement(Satisfy(func(model definitions.StructMetadata) bool {
			return model.Name == "Context" && model.FullyQualifiedPackage == "context"
		})))
	})

	It("Correctly inject context in the routing file when the first parameter is a context.Context", func() {
		// Process the metadata once
		config, metadata, structs, _, err := utils.GetDefaultConfigAndMetadataOrFail()
		Expect(err).To(BeNil())

		// Each engine has a different way of accessing the raw HTTP context
		expectedSubstringPerEngine := map[string]string{
			string(definitions.RoutingEngineGin):   "opError := controller.MethodWithContext(ctx.Request.Context(), *idRawPtr)",
			string(definitions.RoutingEngineEcho):  "opError := controller.MethodWithContext(ctx.Request().Context(), *idRawPtr)",
			string(definitions.RoutingEngineMux):   "opError := controller.MethodWithContext(ctx.Context(), *idRawPtr)",
			string(definitions.RoutingEngineFiber): "opError := controller.MethodWithContext(ctx.UserContext(), *idRawPtr)",
			string(definitions.RoutingEngineChi):   "opError := controller.MethodWithContext(ctx.Context(), *idRawPtr)",
		}

		// For each engine, make sure the context is correctly injected
		for _, engine := range definitions.SupportedRoutingEngineStrings {
			config.RoutesConfig.Engine = definitions.RoutingEngineType(engine)

			err := routes.GenerateRoutes(config, metadata, &definitions.Models{Structs: structs})
			Expect(err).To(BeNil())

			// Read the generated routes file
			data := utils.ReadFileByRelativePathOrFail("dist/gleece.go")
			Expect(data).ToNot(BeNil(), fmt.Sprintf("Could not read resulting routing file for engine '%s'", engine))
			Expect(data).To(
				ContainSubstring(expectedSubstringPerEngine[engine]),
				fmt.Sprintf("Routing file for engine '%s' does not contain expected context injection", engine),
			)
		}
	})

	It("Correctly inject context in the routing file when the last parameter is a context.Context", func() {
		// Process the metadata once
		config, metadata, structs, _, err := utils.GetDefaultConfigAndMetadataOrFail()
		Expect(err).To(BeNil())

		// Each engine has a different way of accessing the raw HTTP context
		expectedSubstringPerEngine := map[string]string{
			string(definitions.RoutingEngineGin):   "opError := controller.MethodWithLastParamContext(*idRawPtr, ctx.Request.Context())",
			string(definitions.RoutingEngineEcho):  "opError := controller.MethodWithLastParamContext(*idRawPtr, ctx.Request().Context())",
			string(definitions.RoutingEngineMux):   "opError := controller.MethodWithLastParamContext(*idRawPtr, ctx.Context())",
			string(definitions.RoutingEngineFiber): "opError := controller.MethodWithLastParamContext(*idRawPtr, ctx.UserContext())",
			string(definitions.RoutingEngineChi):   "opError := controller.MethodWithLastParamContext(*idRawPtr, ctx.Context())",
		}

		// For each engine, make sure the context is correctly injected
		for _, engine := range definitions.SupportedRoutingEngineStrings {
			config.RoutesConfig.Engine = definitions.RoutingEngineType(engine)

			err := routes.GenerateRoutes(config, metadata, &definitions.Models{Structs: structs})
			Expect(err).To(BeNil())

			// Read the generated routes file
			data := utils.ReadFileByRelativePathOrFail("dist/gleece.go")
			Expect(data).ToNot(BeNil(), fmt.Sprintf("Could not read resulting routing file for engine '%s'", engine))
			Expect(data).To(
				ContainSubstring(expectedSubstringPerEngine[engine]),
				fmt.Sprintf("Routing file for engine '%s' does not contain expected context injection", engine),
			)
		}
	})
})

func TestContextController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Context Controller")
}
