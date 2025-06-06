package sanity_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	"github.com/gopher-fleece/runtime"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var metadata []definitions.ControllerMetadata
var models []definitions.StructMetadata
var schemaShouldHaveStdErrorSanity bool

var _ = BeforeSuite(func() {
	controllers, flatModels, hasStdError := utils.GetControllersAndModelsOrFail()
	metadata = controllers
	models = flatModels
	schemaShouldHaveStdErrorSanity = hasStdError
})

var _ = Describe("Sanity Controller", func() {
	It("Created metadata and model lists have length of 1 and RFC7807 should be present", func() {
		Expect(metadata).ToNot(BeNil())
		Expect(metadata).To(HaveLen(1))
		Expect(models).ToNot(BeNil())
		Expect(models).To(HaveLen(3))
		Expect(schemaShouldHaveStdErrorSanity).To(BeTrue())
	})

	It("Produces correct controller level metadata", func() {
		controllerMeta := metadata[0]
		Expect(controllerMeta.Name).To(Equal("SanityController"))
		Expect(controllerMeta.Package).To(Equal("sanity_test"))
		Expect(controllerMeta.PkgPath).To(Equal("github.com/gopher-fleece/gleece/test/sanity"))
		Expect(controllerMeta.Tag).To(Equal("Sanity Controller Tag"))
		Expect(controllerMeta.RestMetadata.Path).To(Equal("/test/sanity"))
		Expect(controllerMeta.Description).To(Equal("Sanity Controller"))
		Expect(controllerMeta.Routes).To(HaveLen(2))
		Expect(controllerMeta.Security).To(HaveLen(1))
		Expect(controllerMeta.Security[0].SecurityAnnotation).To(HaveLen(1))
		Expect(controllerMeta.Security[0].SecurityAnnotation[0].SchemaName).To(Equal("sanitySchema"))
		Expect(controllerMeta.Security[0].SecurityAnnotation[0].Scopes).To(HaveLen(2))
		Expect(controllerMeta.Security[0].SecurityAnnotation[0].Scopes[0]).To(Equal("read"))
		Expect(controllerMeta.Security[0].SecurityAnnotation[0].Scopes[1]).To(Equal("write"))
	})

	It("Produces correct route level metadata", func() {
		route := metadata[0].Routes[0]

		Expect(route.OperationId).To(Equal("ValidMethodWithSimpleRouteQueryAndHeaderParameters"))
		Expect(route.HttpVerb).To(Equal(definitions.HttpPost))
		Expect(route.Hiding.Type).To(Equal(definitions.HideMethodNever))
		Expect(route.Hiding.Condition).To(BeEmpty())
		Expect(route.Deprecation.Deprecated).To(BeFalse())
		Expect(route.Deprecation.Description).To(BeEmpty())
		Expect(route.Description).To(Equal("A sanity test controller method"))
		Expect(route.RestMetadata.Path).To(Equal("/{routeParamAlias}"))
		Expect(route.FuncParams).To(HaveLen(3))
		Expect(route.Responses).To(HaveLen(2))
		Expect(route.HasReturnValue).To(BeTrue())
		Expect(route.ResponseDescription).To(Equal("Description for HTTP 200"))
		Expect(route.ResponseSuccessCode).To(Equal(runtime.StatusOK))
		Expect(route.ErrorResponses).To(HaveLen(2))
		Expect(route.ErrorResponses[0].HttpStatusCode).To(Equal(runtime.StatusInternalServerError))
		Expect(route.ErrorResponses[0].Description).To(Equal("Code 500"))
		Expect(route.ErrorResponses[1].HttpStatusCode).To(Equal(runtime.StatusBadGateway))
		Expect(route.ErrorResponses[1].Description).To(Equal("Code 502"))
		Expect(route.RequestContentType).To(Equal(definitions.ContentTypeJSON))
		Expect(route.ResponseContentType).To(Equal(definitions.ContentTypeJSON))
		Expect(route.Security).To(HaveLen(1))
		Expect(route.Security[0].SecurityAnnotation).To(HaveLen(1))
		Expect(route.Security[0].SecurityAnnotation[0].SchemaName).To(Equal("sanitySchema"))
		Expect(route.Security[0].SecurityAnnotation[0].Scopes).To(HaveLen(2))
		Expect(route.Security[0].SecurityAnnotation[0].Scopes[0]).To(Equal("read"))
		Expect(route.Security[0].SecurityAnnotation[0].Scopes[1]).To(Equal("write"))
	})

	It("Produces correct method level metadata", func() {
		route := metadata[0].Routes[0]
		Expect(route.FuncParams).To(HaveLen(3))
		Expect(route.Responses).To(HaveLen(2))

		Expect(route.FuncParams[0].PassedIn).To(Equal(definitions.PassedInPath))
		Expect(route.FuncParams[0].NameInSchema).To(Equal("routeParamAlias"))
		Expect(route.FuncParams[0].Description).To(Equal(""))
		Expect(route.FuncParams[0].UniqueImportSerial).To(Equal(uint64(0)))
		Expect(route.FuncParams[0].Validator).To(Equal("required"))
		Expect(route.FuncParams[0].Deprecation).To(BeNil())
		Expect(route.FuncParams[0].ParamMeta.Name).To(Equal("routeParam"))
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.Name).To(Equal("string"))
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.PkgPath).To(Equal(""))
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.DefaultPackageAlias).To(Equal(""))
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.Description).To(Equal(""))
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.Import).To(Equal(definitions.ImportTypeNone))
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.IsUniverseType).To(BeTrue())
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.IsByAddress).To(BeFalse())
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.SymbolKind).To(Equal(common.SymKindUnknown))

		Expect(route.FuncParams[1].PassedIn).To(Equal(definitions.PassedInQuery))
		Expect(route.FuncParams[1].NameInSchema).To(Equal("queryParam"))
		Expect(route.FuncParams[1].Description).To(Equal(""))
		Expect(route.FuncParams[1].UniqueImportSerial).To(Equal(uint64(1)))
		Expect(route.FuncParams[1].Validator).To(Equal("required"))
		Expect(route.FuncParams[1].Deprecation).To(BeNil())
		Expect(route.FuncParams[1].ParamMeta.Name).To(Equal("queryParam"))
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.Name).To(Equal("int"))
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.PkgPath).To(Equal(""))
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.DefaultPackageAlias).To(Equal(""))
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.Description).To(Equal(""))
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.Import).To(Equal(definitions.ImportTypeNone))
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.IsUniverseType).To(BeTrue())
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.IsByAddress).To(BeFalse())
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.SymbolKind).To(Equal(common.SymKindUnknown))

		Expect(route.FuncParams[2].PassedIn).To(Equal(definitions.PassedInHeader))
		Expect(route.FuncParams[2].NameInSchema).To(Equal("headerParam"))
		Expect(route.FuncParams[2].Description).To(Equal(""))
		Expect(route.FuncParams[2].UniqueImportSerial).To(Equal(uint64(2)))
		Expect(route.FuncParams[2].Validator).To(Equal("required"))
		Expect(route.FuncParams[2].Deprecation).To(BeNil())
		Expect(route.FuncParams[2].ParamMeta.Name).To(Equal("headerParam"))
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.Name).To(Equal("float32"))
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.PkgPath).To(Equal(""))
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.DefaultPackageAlias).To(Equal(""))
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.Description).To(Equal(""))
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.Import).To(Equal(definitions.ImportTypeNone))
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.IsUniverseType).To(BeTrue())
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.IsByAddress).To(BeFalse())
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.SymbolKind).To(Equal(common.SymKindUnknown))

		Expect(route.Responses[0].UniqueImportSerial).To(Equal(uint64(3)))
		Expect(route.Responses[0].TypeMetadata.Name).To(Equal("SimpleResponseModel"))
		Expect(route.Responses[0].TypeMetadata.PkgPath).To(Equal("github.com/gopher-fleece/gleece/test/sanity"))
		Expect(route.Responses[0].TypeMetadata.DefaultPackageAlias).To(Equal("sanity"))
		Expect(route.Responses[0].TypeMetadata.Description).To(Equal("This should be the actual description"))
		Expect(route.Responses[0].TypeMetadata.Import).To(Equal(definitions.ImportTypeNone))
		Expect(route.Responses[0].TypeMetadata.IsUniverseType).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.IsByAddress).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.SymbolKind).To(Equal(common.SymKindStruct))

		Expect(route.Responses[1].UniqueImportSerial).To(Equal(uint64(4)))
		Expect(route.Responses[1].TypeMetadata.Name).To(Equal("error"))
		Expect(route.Responses[1].TypeMetadata.PkgPath).To(Equal(""))
		Expect(route.Responses[1].TypeMetadata.DefaultPackageAlias).To(Equal(""))
		Expect(route.Responses[1].TypeMetadata.Description).To(Equal(""))
		Expect(route.Responses[1].TypeMetadata.Import).To(Equal(definitions.ImportTypeNone))
		Expect(route.Responses[1].TypeMetadata.IsUniverseType).To(BeTrue())
		Expect(route.Responses[1].TypeMetadata.IsByAddress).To(BeFalse())
		Expect(route.Responses[1].TypeMetadata.SymbolKind).To(Equal(common.SymKindUnknown))
	})

	It("Produces correct models list", func() {
		// Model 1
		model1 := models[0]
		Expect(model1.Name).To(Equal("SimpleResponseModel"))
		Expect(model1.PkgPath).To(Equal("github.com/gopher-fleece/gleece/test/sanity"))
		Expect(model1.Description).To(Equal("This should be the actual description"))
		Expect(model1.Fields).To(HaveLen(1))
		Expect(model1.Fields[0].Name).To(Equal("SomeValue"))
		Expect(model1.Fields[0].Type).To(Equal("int"))
		Expect(model1.Fields[0].Description).To(Equal("A description for the value"))
		Expect(model1.Fields[0].Tag).To(Equal("validate:\"required,min=0,max=10\""))
		Expect(model1.Fields[0].Deprecation.Deprecated).To(BeFalse())
		Expect(model1.Fields[0].Deprecation.Description).To(Equal(""))
		Expect(model1.Deprecation.Deprecated).To(BeFalse())
		Expect(model1.Deprecation.Description).To(Equal(""))

		// Model 2
		model2 := models[1]
		Expect(model2.Name).To(Equal("SimpleStruct"))
		Expect(model2.PkgPath).To(Equal("github.com/gopher-fleece/gleece/test/units"))
		Expect(model2.Description).To(BeEmpty())

		Expect(model2.Fields).To(HaveLen(2))
		Expect(model2.Fields[0].Name).To(Equal("FieldA"))
		Expect(model2.Fields[0].Type).To(Equal("string"))
		Expect(model2.Fields[0].Description).To(BeEmpty())
		Expect(model2.Fields[0].Tag).To(BeEmpty())
		Expect(model2.Fields[0].Deprecation.Deprecated).To(BeFalse())
		Expect(model2.Fields[0].Deprecation.Description).To(Equal(""))

		Expect(model2.Fields[1].Name).To(Equal("FieldB"))
		Expect(model2.Fields[1].Type).To(Equal("int"))
		Expect(model2.Fields[1].Description).To(BeEmpty())
		Expect(model2.Fields[1].Tag).To(BeEmpty())
		Expect(model2.Fields[1].Deprecation.Deprecated).To(BeFalse())
		Expect(model2.Fields[1].Deprecation.Description).To(Equal(""))

		Expect(model2.Deprecation.Deprecated).To(BeFalse())
		Expect(model2.Deprecation.Description).To(Equal(""))

		// Model 3
		model3 := models[2]
		Expect(model3.Name).To(Equal("StructWithStructSlice"))
		Expect(model3.PkgPath).To(Equal("github.com/gopher-fleece/gleece/test/units"))
		Expect(model3.Description).To(BeEmpty())

		Expect(model3.Fields).To(HaveLen(1))
		Expect(model3.Fields[0].Name).To(Equal("SimpleStructSlice"))
		Expect(model3.Fields[0].Type).To(Equal("[]SimpleStruct"))
		Expect(model3.Fields[0].Description).To(BeEmpty())
		Expect(model3.Fields[0].Tag).To(BeEmpty())
		Expect(model3.Fields[0].Deprecation.Deprecated).To(BeFalse())
		Expect(model3.Fields[0].Deprecation.Description).To(Equal(""))

		Expect(model3.Deprecation.Deprecated).To(BeFalse())
		Expect(model3.Deprecation.Description).To(Equal(""))
	})
})

func TestSanityController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sanity Controller")
}
