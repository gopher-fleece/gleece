package validators_test

import (
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/arbitrators"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/validators"
	"github.com/gopher-fleece/gleece/core/validators/diagnostics"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/gopher-fleece/gleece/test/utils/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Controller Validator", func() {

	var gleeceConfig *definitions.GleeceConfig
	var pkgFacade *arbitrators.PackagesFacade
	var validator validators.ControllerValidator
	var controller metadata.ControllerMeta

	BeforeEach(func() {
		ctx := utils.GetVisitContextByRelativeConfigOrFail("./gleece.test.config.json")
		gleeceConfig = ctx.GleeceConfig
		pkgFacade = ctx.ArbitrationProvider.Pkg()

		controller = metadata.ControllerMeta{
			Struct: metadata.StructMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:    "ExampleController",
					PkgPath: "example.com/test",
				},
			},
		}

		validator = validators.NewControllerValidator(gleeceConfig, pkgFacade, &controller)
	})

	It("Returns a DiagAnnotationUnknown error for unknown annotations", func() {
		controller.Struct.Annotations = utils.GetAnnotationHolderOrFail(
			[]string{
				"// @Tag(Test)",
				"// @UnknownAnnotation",
			},
			annotations.CommentSourceController,
		)

		validator = validators.NewControllerValidator(gleeceConfig, pkgFacade, &controller)
		diag, err := validator.Validate()
		Expect(err).To(BeNil())
		Expect(diag.EntityKind).To(BeEquivalentTo("Controller"))
		Expect(diag.EntityName).To(BeEquivalentTo("ExampleController"))
		Expect(diag.Children).To(BeEmpty())
		Expect(diag.Diagnostics).To(HaveLen(1))
		Expect(diag.Diagnostics[0]).To(BeDiagnosticErrorWithCodeAndMessage(
			diagnostics.DiagAnnotationUnknown,
			"Unknown annotation '@UnknownAnnotation'",
		))
	})

	It("Returns a DiagAnnotationValueMustExist when annotation lacks required value", func() {
		controller.Struct.Annotations = utils.GetAnnotationHolderOrFail(
			[]string{"// @Tag"},
			annotations.CommentSourceController,
		)

		validator = validators.NewControllerValidator(gleeceConfig, pkgFacade, &controller)
		diag, err := validator.Validate()
		Expect(err).To(BeNil())

		Expect(diag.Diagnostics).To(HaveLen(1))
		Expect(diag.Diagnostics[0]).To(BeDiagnosticErrorWithCodeAndMessage(
			diagnostics.DiagAnnotationValueMustExist,
			"Annotation '@Tag' requires a value",
		))
	})

	It("Returns a DiagAnnotationInvalidInContext when annotation cannot be used on a controller", func() {
		controller.Struct.Annotations = utils.GetAnnotationHolderOrFail(
			[]string{
				"// @Tag(Test)",
				"// @Method(POST)",
			},
			annotations.CommentSourceController,
		)

		validator = validators.NewControllerValidator(gleeceConfig, pkgFacade, &controller)
		diag, err := validator.Validate()
		Expect(err).To(BeNil())

		Expect(diag.Diagnostics).To(HaveLen(1))
		Expect(diag.Diagnostics[0]).To(BeDiagnosticWarningWithCodeAndMessage(
			diagnostics.DiagAnnotationInvalidInContext,
			"Annotation '@Method' is not valid in the context of a controller",
		))
	})
})
