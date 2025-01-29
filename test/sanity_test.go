package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/cmd/arguments"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/external"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var metadata []definitions.ControllerMetadata
var models []definitions.ModelMetadata
var schemaShouldHaveStdError bool

var _ = BeforeSuite(func() {
	cwd, err := os.Getwd()
	if err != nil {
		Fail(fmt.Sprintf("Could not determine process working directory - %v", err))
	}

	configPath := filepath.Join(cwd, "configs", "sanity.config.json")
	_, controllers, flatModels, hasStdError, err := cmd.GetConfigAndMetadata(arguments.CliArguments{ConfigPath: configPath})
	if err != nil {
		Fail(fmt.Sprintf("Could not generate routes - %v", err))
	}

	metadata = controllers
	models = flatModels
	schemaShouldHaveStdError = hasStdError
})

var _ = Describe("Sanity Controller", func() {
	It("Creates metadata and model lists have length of 1 and RFC7807 should be present", func() {
		Expect(metadata).ToNot(BeNil())
		Expect(metadata).To(HaveLen(1))
		Expect(models).ToNot(BeNil())
		Expect(models).To(HaveLen(1))
		Expect(schemaShouldHaveStdError).To(BeTrue())
	})

	It("Produces correct controller level metadata", func() {
		controllerMeta := metadata[0]
		Expect(controllerMeta.Name).To(Equal("SanityController"))
		Expect(controllerMeta.Package).To(Equal("fixtures"))
		Expect(controllerMeta.FullyQualifiedPackage).To(Equal("github.com/gopher-fleece/gleece/test/fixtures"))
		Expect(controllerMeta.Tag).To(Equal("Sanity Controller Tag"))
		Expect(controllerMeta.RestMetadata.Path).To(Equal("/test/sanity"))
		Expect(controllerMeta.Description).To(Equal("Sanity Controller"))
		Expect(controllerMeta.Routes).To(HaveLen(1))
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
		Expect(route.RestMetadata.Path).To(Equal("/{routeParam}"))
		Expect(route.FuncParams).To(HaveLen(3))
		Expect(route.Responses).To(HaveLen(2))
		Expect(route.HasReturnValue).To(BeTrue())
		Expect(route.ResponseDescription).To(Equal("Description for HTTP 200"))
		Expect(route.ResponseSuccessCode).To(Equal(external.StatusOK))
		Expect(route.ErrorResponses).To(HaveLen(2))
		Expect(route.ErrorResponses[0].HttpStatusCode).To(Equal(external.StatusInternalServerError))
		Expect(route.ErrorResponses[0].Description).To(Equal("Code 500"))
		Expect(route.ErrorResponses[1].HttpStatusCode).To(Equal(external.StatusBadGateway))
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
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.FullyQualifiedPackage).To(Equal(""))
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.DefaultPackageAlias).To(Equal(""))
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.Description).To(Equal(""))
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.Import).To(Equal(definitions.ImportTypeNone))
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.IsUniverseType).To(BeTrue())
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.IsByAddress).To(BeFalse())
		Expect(route.FuncParams[0].ParamMeta.TypeMeta.EntityKind).To(Equal(definitions.AstNodeKindUnknown))

		Expect(route.FuncParams[1].PassedIn).To(Equal(definitions.PassedInQuery))
		Expect(route.FuncParams[1].NameInSchema).To(Equal("queryParam"))
		Expect(route.FuncParams[1].Description).To(Equal(""))
		Expect(route.FuncParams[1].UniqueImportSerial).To(Equal(uint64(1)))
		Expect(route.FuncParams[1].Validator).To(Equal("required"))
		Expect(route.FuncParams[1].Deprecation).To(BeNil())
		Expect(route.FuncParams[1].ParamMeta.Name).To(Equal("queryParam"))
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.Name).To(Equal("int"))
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.FullyQualifiedPackage).To(Equal(""))
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.DefaultPackageAlias).To(Equal(""))
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.Description).To(Equal(""))
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.Import).To(Equal(definitions.ImportTypeNone))
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.IsUniverseType).To(BeTrue())
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.IsByAddress).To(BeFalse())
		Expect(route.FuncParams[1].ParamMeta.TypeMeta.EntityKind).To(Equal(definitions.AstNodeKindUnknown))

		Expect(route.FuncParams[2].PassedIn).To(Equal(definitions.PassedInHeader))
		Expect(route.FuncParams[2].NameInSchema).To(Equal("headerParam"))
		Expect(route.FuncParams[2].Description).To(Equal(""))
		Expect(route.FuncParams[2].UniqueImportSerial).To(Equal(uint64(2)))
		Expect(route.FuncParams[2].Validator).To(Equal("required"))
		Expect(route.FuncParams[2].Deprecation).To(BeNil())
		Expect(route.FuncParams[2].ParamMeta.Name).To(Equal("headerParam"))
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.Name).To(Equal("float32"))
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.FullyQualifiedPackage).To(Equal(""))
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.DefaultPackageAlias).To(Equal(""))
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.Description).To(Equal(""))
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.Import).To(Equal(definitions.ImportTypeNone))
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.IsUniverseType).To(BeTrue())
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.IsByAddress).To(BeFalse())
		Expect(route.FuncParams[2].ParamMeta.TypeMeta.EntityKind).To(Equal(definitions.AstNodeKindUnknown))

		Expect(route.Responses[0].UniqueImportSerial).To(Equal(uint64(3)))
		Expect(route.Responses[0].TypeMetadata.Name).To(Equal("SimpleResponseModel"))
		Expect(route.Responses[0].TypeMetadata.FullyQualifiedPackage).To(Equal("github.com/gopher-fleece/gleece/test/fixtures"))
		Expect(route.Responses[0].TypeMetadata.DefaultPackageAlias).To(Equal("fixtures"))
		Expect(route.Responses[0].TypeMetadata.Description).To(Equal("This should be the actual description"))
		Expect(route.Responses[0].TypeMetadata.Import).To(Equal(definitions.ImportTypeAlias))
		Expect(route.Responses[0].TypeMetadata.IsUniverseType).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.IsByAddress).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.EntityKind).To(Equal(definitions.AstNodeKindStruct))

		Expect(route.Responses[1].UniqueImportSerial).To(Equal(uint64(4)))
		Expect(route.Responses[1].TypeMetadata.Name).To(Equal("error"))
		Expect(route.Responses[1].TypeMetadata.FullyQualifiedPackage).To(Equal(""))
		Expect(route.Responses[1].TypeMetadata.DefaultPackageAlias).To(Equal(""))
		Expect(route.Responses[1].TypeMetadata.Description).To(Equal(""))
		Expect(route.Responses[1].TypeMetadata.Import).To(Equal(definitions.ImportTypeNone))
		Expect(route.Responses[1].TypeMetadata.IsUniverseType).To(BeTrue())
		Expect(route.Responses[1].TypeMetadata.IsByAddress).To(BeFalse())
		Expect(route.Responses[1].TypeMetadata.EntityKind).To(Equal(definitions.AstNodeKindUnknown))
	})

	It("Produces correct models list", func() {
		Expect(models[0].Name).To(Equal("SimpleResponseModel"))
		Expect(models[0].FullyQualifiedPackage).To(Equal("github.com/gopher-fleece/gleece/test/fixtures"))
		Expect(models[0].Description).To(Equal("This should be the actual description"))
		Expect(models[0].Fields).To(HaveLen(1))
		Expect(models[0].Fields[0].Name).To(Equal("SomeValue"))
		Expect(models[0].Fields[0].Type).To(Equal("int"))
		Expect(models[0].Fields[0].Description).To(Equal("A description for the value"))
		Expect(models[0].Fields[0].Tag).To(Equal("validate:\"required,min=0,max=10\""))
		Expect(models[0].Fields[0].Deprecation.Deprecated).To(BeFalse())
		Expect(models[0].Fields[0].Deprecation.Description).To(Equal(""))
		Expect(models[0].Deprecation.Deprecated).To(BeFalse())
		Expect(models[0].Deprecation.Description).To(Equal(""))
	})
})

func TestSanityController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sanity Controller")
}
