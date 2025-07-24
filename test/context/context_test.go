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
		meta := utils.GetControllersAndModelsOrFail()
		Expect(meta.Models.Structs).ToNot(ContainElement(Satisfy(func(model definitions.StructMetadata) bool {
			return model.Name == "Context" && model.PkgPath == "context"
		})))
	})

	It("Correctly inject context in the routing file when the first parameter is a context.Context", func() {
		// Process the metadata once
		config, meta, err := utils.GetDefaultConfigAndMetadataOrFail()
		Expect(err).To(BeNil())

		// Each engine has a different way of accessing the raw HTTP context
		expectedSubstringPerEngine := map[string]string{
			string(definitions.RoutingEngineGin):   "opError := controller.MethodWithContext(getRequestContext(ginCtx), *idRawPtr)",
			string(definitions.RoutingEngineEcho):  "opError := controller.MethodWithContext(getRequestContext(echoCtx), *idRawPtr)",
			string(definitions.RoutingEngineMux):   "opError := controller.MethodWithContext(getRequestContext(req), *idRawPtr)",
			string(definitions.RoutingEngineFiber): "opError := controller.MethodWithContext(getRequestContext(fiberCtx), *idRawPtr)",
			string(definitions.RoutingEngineChi):   "opError := controller.MethodWithContext(getRequestContext(req), *idRawPtr)",
		}

		// For each engine, make sure the context is correctly injected
		for _, engine := range definitions.SupportedRoutingEngineStrings {
			config.RoutesConfig.Engine = definitions.RoutingEngineType(engine)

			err := routes.GenerateRoutes(config, meta)
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
		config, meta, err := utils.GetDefaultConfigAndMetadataOrFail()
		Expect(err).To(BeNil())

		// Each engine has a different way of accessing the raw HTTP context
		expectedSubstringPerEngine := map[string]string{
			string(definitions.RoutingEngineGin):   "opError := controller.MethodWithLastParamContext(*idRawPtr, getRequestContext(ginCtx))",
			string(definitions.RoutingEngineEcho):  "opError := controller.MethodWithLastParamContext(*idRawPtr, getRequestContext(echoCtx))",
			string(definitions.RoutingEngineMux):   "opError := controller.MethodWithLastParamContext(*idRawPtr, getRequestContext(req))",
			string(definitions.RoutingEngineFiber): "opError := controller.MethodWithLastParamContext(*idRawPtr, getRequestContext(fiberCtx))",
			string(definitions.RoutingEngineChi):   "opError := controller.MethodWithLastParamContext(*idRawPtr, getRequestContext(req))",
		}

		// For each engine, make sure the context is correctly injected
		for _, engine := range definitions.SupportedRoutingEngineStrings {
			config.RoutesConfig.Engine = definitions.RoutingEngineType(engine)

			err := routes.GenerateRoutes(config, meta)
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
