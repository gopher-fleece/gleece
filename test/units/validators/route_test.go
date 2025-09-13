package validators_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/arbitrators"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/validators"
	"github.com/gopher-fleece/gleece/core/validators/diagnostics"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/gopher-fleece/gleece/test/utils/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Route Validator", func() {

	var gleeceConfig *definitions.GleeceConfig
	var pkgFacade *arbitrators.PackagesFacade
	var validator validators.ReceiverValidator
	var controller metadata.ControllerMeta
	var receiver metadata.ReceiverMeta

	BeforeEach(func() {
		ctx := utils.GetVisitContextByRelativeConfigOrFail("./gleece.test.config.json")
		gleeceConfig = ctx.GleeceConfig
		pkgFacade = ctx.ArbitrationProvider.Pkg()

		receiverAnnotations := utils.GetAnnotationHolderOrFail(
			[]string{
				"// @Route(/)",
				"// @Method(POST)",
			},
			annotations.CommentSourceRoute,
		)

		receiver = metadata.ReceiverMeta{
			SymNodeMeta: metadata.SymNodeMeta{
				Name:        "TestReceiver",
				Annotations: receiverAnnotations,
			},
		}

		controller = metadata.ControllerMeta{
			Struct: metadata.StructMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:        "ExampleController",
					PkgPath:     "example.com/test",
					Annotations: utils.GetAnnotationHolderOrFail([]string{}, annotations.CommentSourceController),
				},
			},
			Receivers: []metadata.ReceiverMeta{receiver},
		}

		validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)
	})

	It("Returns a DiagAnnotationInvalidInContext warning when an annotation is applied in an invalid context", func() {
		// Some annotations are only valid for certain types of entities.
		// In this example, @Tag is only valid for controllers.
		receiver.Annotations = utils.GetAnnotationHolderOrFail(
			[]string{
				"// @Tag(Test)",
				"// @Route(/)",
				"// @Method(POST)",
			},
			annotations.CommentSourceRoute,
		)
		// Simulate a no-params function for simplicity
		receiver.Params = []metadata.FuncParam{}
		controller.Receivers[0].Params = receiver.Params

		validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

		diag, err := validator.Validate()
		Expect(err).To(BeNil())
		Expect(diag).To(BeChildlessDiagOfReceiver(&receiver))

		Expect(diag.Diagnostics).To(HaveLen(2))

		Expect(diag.Diagnostics[0]).To(BeDiagnosticWarningWithCodeAndMessage(
			diagnostics.DiagAnnotationInvalidInContext,
			"Annotation '@Tag' is not valid in the context of a route",
		))

		Expect(diag.Diagnostics[1]).To(BeAReturnSigDiagnostic())
	})

	It("Returns a DiagAnnotationValueInvalid error when an annotation's value is invalid as per internal schema", func() {
		receiver.Annotations = utils.GetAnnotationHolderOrFail(
			[]string{
				"// @Route(/)",
				"// @Method(INVALID_METHOD)",
				"// @Response(INVALID)",
			},
			annotations.CommentSourceRoute,
		)

		validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

		diag, err := validator.Validate()
		Expect(err).To(BeNil())
		Expect(diag).To(BeChildlessDiagOfReceiver(&receiver))

		Expect(diag.Diagnostics).To(HaveLen(3))

		Expect(diag.Diagnostics[0]).To(BeDiagnosticErrorWithCodeAndMessage(
			diagnostics.DiagAnnotationValueInvalid,
			"Invalid HTTP verb 'INVALID_METHOD'. Supported verbs are: DELETE, GET, PATCH, POST, PUT",
		))

		Expect(diag.Diagnostics[1]).To(BeDiagnosticErrorWithCodeAndMessage(
			diagnostics.DiagAnnotationValueInvalid,
			"Non-numeric HTTP status code 'INVALID'",
		))

		Expect(diag.Diagnostics[1]).To(BeAReturnSigDiagnostic())
	})

	It("Returns a DiagAnnotationPropertiesShouldNotExist when a property-less annotations has properties", func() {
		receiver.Annotations = utils.GetAnnotationHolderOrFail(
			[]string{
				"// @Route(/)",
				"// @Method(POST, { invalid: \"properties\" })",
			},
			annotations.CommentSourceRoute,
		)

		validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

		diag, err := validator.Validate()
		Expect(err).To(BeNil())
		Expect(diag).To(BeChildlessDiagOfReceiver(&receiver))

		Expect(diag.Diagnostics).To(HaveLen(2))

		Expect(diag.Diagnostics[0]).To(BeDiagnosticWarningWithCodeAndMessage(
			diagnostics.DiagAnnotationPropertiesShouldNotExist,
			"Annotation '@Method' does not support properties",
		))

		Expect(diag.Diagnostics[1]).To(BeAReturnSigDiagnostic())
	})

	It("Returns a DiagAnnotationPropertiesUnknownKey warning when an annotation has an unknown property key", func() {
		receiver.Annotations = utils.GetAnnotationHolderOrFail(
			[]string{
				"// @Route(/)",
				"// @Method(POST)",
				"// @Query(value, { invalid: \"properties\" })",
			},
			annotations.CommentSourceRoute,
		)

		validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

		diag, err := validator.Validate()
		Expect(err).To(BeNil())
		Expect(diag).To(BeChildlessDiagOfReceiver(&receiver))

		Expect(diag.Diagnostics).To(HaveLen(3))

		Expect(diag.Diagnostics[0]).To(BeDiagnosticWarningWithCodeAndMessage(
			diagnostics.DiagAnnotationPropertyShouldNotExist,
			"Property 'invalid' is not allowed for annotation '@Query'",
		))

		Expect(diag.Diagnostics[1]).To(BeAReturnSigDiagnostic())

		Expect(diag.Diagnostics[2]).To(BeDiagnosticErrorWithCodeAndMessage(
			diagnostics.DiagLinkerPathInvalidRef,
			"@Query 'value' does not match any parameter of TestReceiver",
		))
	})

	It("Returns a DiagAnnotationPropertiesInvalidValueForKey error when property type is invalid", func() {
		receiver.Annotations = utils.GetAnnotationHolderOrFail(
			[]string{
				"// @Route(/)",
				"// @Method(POST)",
				"// @Query(value, { name: 123 })",
			},
			annotations.CommentSourceRoute,
		)

		validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

		diag, err := validator.Validate()
		Expect(err).To(BeNil())
		Expect(diag).To(BeChildlessDiagOfReceiver(&receiver))

		Expect(diag.Diagnostics).To(HaveLen(3))

		Expect(diag.Diagnostics[0]).To(BeDiagnosticWarningWithCodeAndMessage(
			diagnostics.DiagAnnotationPropertiesInvalidValueForKey,
			"Property 'name' expected to be a string",
		))

		Expect(diag.Diagnostics[1]).To(BeAReturnSigDiagnostic())

		Expect(diag.Diagnostics[2]).To(BeDiagnosticErrorWithCodeAndMessage(
			diagnostics.DiagLinkerPathInvalidRef,
			"@Query 'value' does not match any parameter of TestReceiver",
		))
	})

	When("Annotation combination", func() {

		It("Wrong Annotation combination - duplicate not allowed annotation type", func() {
			comments := []string{"// @Body(value1)", "// @Body(value2)"}
			nodes := utils.CommentsToCommentBlock(comments, 1)
			_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			Expect(err).To(MatchError(ContainSubstring("multiple instances of annotation @Body are not allowed")))
		})

		It("Wrong Annotation combination - two from different not allowed type", func() {
			comments := []string{"// @Body(value1)", "// @FormField(value2)"}
			nodes := utils.CommentsToCommentBlock(comments, 1)
			_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			Expect(err).To(MatchError(ContainSubstring("annotations @FormField and @Body cannot be used together")))
		})
	})

	When("Annotation values combination", func() {

		It("Wrong Annotation values combination - in the same annotation type", func() {
			comments := []string{"// @Query(the_value)", "// @Query(the_value)"}
			nodes := utils.CommentsToCommentBlock(comments, 1)
			_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			Expect(err).To(MatchError(ContainSubstring("duplicate value 'the_value' used in @Query and @Query annotations")))
		})

		It("Wrong Annotation combination - two from different not allowed type", func() {
			comments := []string{"// @Query(the_value)", "// @Header(the_value)"}
			nodes := utils.CommentsToCommentBlock(comments, 1)
			_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			Expect(err).To(MatchError(ContainSubstring("duplicate value 'the_value' used in @Query and @Header annotations")))
		})

		It("Valid Annotation combination - two and one not allowed", func() {
			comments := []string{"// @Query(the_value)", "// @Security(the_value)"}
			nodes := utils.CommentsToCommentBlock(comments, 1)
			_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			Expect(err).To(BeNil())
		})

		It("Valid Annotation combination - two and both allowed", func() {
			comments := []string{"// @Security(the_value)", "// @Security(the_value)"}
			nodes := utils.CommentsToCommentBlock(comments, 1)
			_, err := annotations.NewAnnotationHolder(nodes, annotations.CommentSourceRoute)
			Expect(err).To(BeNil())
		})
	})

})

func TestUnitRouteValidator(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Route Validator")
}
