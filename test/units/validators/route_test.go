package validators_test

import (
	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/core/annotations"
	"github.com/gopher-fleece/gleece/v2/core/arbitrators"
	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/core/validators"
	"github.com/gopher-fleece/gleece/v2/core/validators/diagnostics"
	"github.com/gopher-fleece/gleece/v2/definitions"
	"github.com/gopher-fleece/gleece/v2/test/utils"
	. "github.com/gopher-fleece/gleece/v2/test/utils/matchers"
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

	Context("Individual annotations", func() {

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

			Expect(diag.Diagnostics[2]).To(BeAReturnSigDiagnostic())
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

		It("Returns a DiagFeatureUnsupported when a @Method annotation has a valid but unsupported HTTP verb", func() {
			receiver.Annotations = utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Route(/)",
					"// @Method(OPTIONS) Valid but unsupported",
				},
				annotations.CommentSourceRoute,
			)

			validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

			diag, err := validator.Validate()
			Expect(err).To(BeNil())
			Expect(diag).To(BeChildlessDiagOfReceiver(&receiver))

			Expect(diag.Diagnostics).To(HaveLen(2))

			Expect(diag.Diagnostics[0]).To(BeDiagnosticErrorWithCodeAndMessage(
				diagnostics.DiagFeatureUnsupported,
				"HTTP verb 'OPTIONS' is currently unsupported for @Method annotations. Supported verbs are: DELETE, GET, PATCH, POST, PUT",
			))

			Expect(diag.Diagnostics[1]).To(BeAReturnSigDiagnostic())
		})
	})

	Context("Annotation combinations", func() {

		It("Returns a DiagAnnotationDuplicate error when single-use annotations are duplicated", func() {
			receiver.Annotations = utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Route(/)",
					"// @Method(POST)",
					"// @Body(value1)",
					"// @Body(value2)",
				},
				annotations.CommentSourceRoute,
			)

			validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

			diag, err := validator.Validate()
			Expect(err).To(BeNil())
			Expect(diag).To(BeChildlessDiagOfReceiver(&receiver))

			Expect(diag.Diagnostics).To(HaveLen(4))

			Expect(diag.Diagnostics[0]).To(BeDiagnosticWarningWithCodeAndMessage(
				diagnostics.DiagAnnotationDuplicate,
				"Multiple instances of '@Body' annotations are not allowed",
			))

			Expect(diag.Diagnostics[1]).To(BeAReturnSigDiagnostic())

			Expect(diag.Diagnostics[2]).To(BeDiagnosticErrorWithCodeAndMessage(
				diagnostics.DiagLinkerPathInvalidRef,
				"@Body 'value1' does not match any parameter of TestReceiver",
			))

			Expect(diag.Diagnostics[3]).To(BeDiagnosticErrorWithCodeAndMessage(
				diagnostics.DiagLinkerPathInvalidRef,
				"@Body 'value2' does not match any parameter of TestReceiver",
			))
		})

		It("Wrong Annotation combination - two from different not allowed type", func() {
			receiver.Annotations = utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Route(/)",
					"// @Method(POST)",
					"// @Body(value1)",
					"// @FormField(value2)",
				},
				annotations.CommentSourceRoute,
			)

			validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

			diag, err := validator.Validate()
			Expect(err).To(BeNil())
			Expect(diag).To(BeChildlessDiagOfReceiver(&receiver))

			Expect(diag.Diagnostics).To(HaveLen(4))

			Expect(diag.Diagnostics[0]).To(BeDiagnosticErrorWithCodeAndMessage(
				diagnostics.DiagAnnotationMutuallyExclusive,
				"Annotations '@FormField' and '@Body' are mutually exclusive",
			))

			Expect(diag.Diagnostics[1]).To(BeAReturnSigDiagnostic())

			Expect(diag.Diagnostics[2]).To(BeDiagnosticErrorWithCodeAndMessage(
				diagnostics.DiagLinkerPathInvalidRef,
				"@Body 'value1' does not match any parameter of TestReceiver",
			))

			Expect(diag.Diagnostics[3]).To(BeDiagnosticErrorWithCodeAndMessage(
				diagnostics.DiagLinkerPathInvalidRef,
				"@FormField 'value2' does not match any parameter of TestReceiver",
			))
		})

		When("Separate annotations reference same value", func() {
			When("Multi-reference is not allowed", func() {
				It("Returns a DiagAnnotationDuplicateValue when used by different instances of the same annotation", func() {
					receiver.Annotations = utils.GetAnnotationHolderOrFail(
						[]string{
							"// @Route(/)",
							"// @Method(POST)",
							"// @Query(the_value)",
							"// @Query(the_value)",
						},
						annotations.CommentSourceRoute,
					)

					validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

					diag, err := validator.Validate()
					Expect(err).To(BeNil())
					Expect(diag).To(BeChildlessDiagOfReceiver(&receiver))

					Expect(diag.Diagnostics).To(HaveLen(4))

					Expect(diag.Diagnostics[0]).To(BeDiagnosticErrorWithCodeAndMessage(
						diagnostics.DiagAnnotationDuplicateValue,
						"Duplicate value 'the_value' referenced by multiple annotations",
					))

					Expect(diag.Diagnostics[1]).To(BeAReturnSigDiagnostic())

					// Note there are two DiagLinkerPathInvalidRef expected here - one for each instance of the annotation
					Expect(diag.Diagnostics[2]).To(BeDiagnosticErrorWithCodeAndMessage(
						diagnostics.DiagLinkerPathInvalidRef,
						"@Query 'the_value' does not match any parameter of TestReceiver",
					))
					Expect(diag.Diagnostics[2].Range).To(Equal(common.ResolvedRange{
						StartLine: 47,
						EndLine:   47,
						StartCol:  10,
						EndCol:    19,
					}))

					Expect(diag.Diagnostics[3]).To(BeDiagnosticErrorWithCodeAndMessage(
						diagnostics.DiagLinkerPathInvalidRef,
						"@Query 'the_value' does not match any parameter of TestReceiver",
					))
					Expect(diag.Diagnostics[3].Range).To(Equal(common.ResolvedRange{
						StartLine: 48,
						EndLine:   48,
						StartCol:  10,
						EndCol:    19,
					}))
				})

				It("Returns a DiagAnnotationDuplicateValue error when used by different types of related annotations", func() {
					receiver.Annotations = utils.GetAnnotationHolderOrFail(
						[]string{
							"// @Route(/)",
							"// @Method(POST)",
							"// @Query(the_value)",
							"// @Header(the_value)",
						},
						annotations.CommentSourceRoute,
					)

					validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

					diag, err := validator.Validate()
					Expect(err).To(BeNil())
					Expect(diag).To(BeChildlessDiagOfReceiver(&receiver))

					Expect(diag.Diagnostics).To(HaveLen(4))

					Expect(diag.Diagnostics[0]).To(BeDiagnosticErrorWithCodeAndMessage(
						diagnostics.DiagAnnotationDuplicateValue,
						"Duplicate value 'the_value' referenced by multiple annotations",
					))

					Expect(diag.Diagnostics[1]).To(BeAReturnSigDiagnostic())

					Expect(diag.Diagnostics[2]).To(BeDiagnosticErrorWithCodeAndMessage(
						diagnostics.DiagLinkerPathInvalidRef,
						"@Header 'the_value' does not match any parameter of TestReceiver",
					))
					Expect(diag.Diagnostics[2].Range).To(Equal(common.ResolvedRange{
						StartLine: 48,
						EndLine:   48,
						StartCol:  11,
						EndCol:    20,
					}))

					Expect(diag.Diagnostics[3]).To(BeDiagnosticErrorWithCodeAndMessage(
						diagnostics.DiagLinkerPathInvalidRef,
						"@Query 'the_value' does not match any parameter of TestReceiver",
					))
					Expect(diag.Diagnostics[3].Range).To(Equal(common.ResolvedRange{
						StartLine: 47,
						EndLine:   47,
						StartCol:  10,
						EndCol:    19,
					}))
				})

				When("Multi-reference is allowed", func() {
					It("Does not return an error when used by different, unrelated annotations", func() {
						receiver.Annotations = utils.GetAnnotationHolderOrFail(
							[]string{
								"// @Route(/)",
								"// @Method(POST)",
								"// @Query(the_value)",
								"// @Security(the_value)",
							},
							annotations.CommentSourceRoute,
						)

						validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

						diag, err := validator.Validate()
						Expect(err).To(BeNil())
						Expect(diag).To(BeChildlessDiagOfReceiver(&receiver))

						Expect(diag.Diagnostics).To(HaveLen(2))

						Expect(diag.Diagnostics[0]).To(BeAReturnSigDiagnostic())

						Expect(diag.Diagnostics[1]).To(BeDiagnosticErrorWithCodeAndMessage(
							diagnostics.DiagLinkerPathInvalidRef,
							"@Query 'the_value' does not match any parameter of TestReceiver",
						))
						Expect(diag.Diagnostics[1].Range).To(Equal(common.ResolvedRange{
							StartLine: 47,
							EndLine:   47,
							StartCol:  10,
							EndCol:    19,
						}))
					})

					It("Does not return an error when used by separate instances of the same annotation", func() {
						receiver.Annotations = utils.GetAnnotationHolderOrFail(
							[]string{
								"// @Route(/)",
								"// @Method(POST)",
								"// @Security(the_value)",
								"// @Security(the_value)",
							},
							annotations.CommentSourceRoute,
						)

						validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

						diag, err := validator.Validate()
						Expect(err).To(BeNil())
						Expect(diag).To(BeChildlessDiagOfReceiver(&receiver))

						Expect(diag.Diagnostics).To(HaveLen(1))

						Expect(diag.Diagnostics[0]).To(BeAReturnSigDiagnostic())
					})
				})
			})
		})
	})

	Context("Error handling", func() {
		It("Returns a 'could not validate parameter' error when an error prevents validation of parameter", func() {
			receiver.Annotations = utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Route(/)",
					"// @Method(POST)",
				},
				annotations.CommentSourceRoute,
			)
			receiver.Params = utils.GetMockParams(1)

			validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

			_, err := validator.Validate()
			Expect(err).To(MatchError(ContainSubstring("could not validate parameter")))
		})

		It("Returns a 'could not validate return types' error when an error prevents validation of return value", func() {
			receiver.Annotations = utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Route(/)",
					"// @Method(POST)",
				},
				annotations.CommentSourceRoute,
			)
			receiver.RetVals = utils.GetMockRetVals(1)
			receiver.RetVals[0].PkgPath = "non-existent/package"

			validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

			_, err := validator.Validate()
			Expect(err).To(MatchError(ContainSubstring("could not validate return types")))
		})

		It("Returns a 'could not validate security' error when an error prevents validation of return value", func() {
			receiver.Annotations = utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Route(/)",
					"// @Method(POST)",
				},
				annotations.CommentSourceRoute,
			)

			controller.Struct.Annotations = common.Ptr(annotations.NewAnnotationHolderFromData(
				[]annotations.Attribute{
					{Name: "Security"}, // This should trigger a security validation failure in when resolving parent security
				},
				[]annotations.NonAttributeComment{},
			))

			validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

			_, err := validator.Validate()
			Expect(err).To(MatchError(ContainSubstring("could not validate security")))
		})

		It("Returns a 'could not validate security' error when an error prevents validation of return value", func() {
			receiver.Annotations = utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Method(POST)",
				},
				annotations.CommentSourceRoute,
			)

			validator = validators.NewReceiverValidator(gleeceConfig, pkgFacade, &controller, &receiver)

			_, err := validator.Validate()
			Expect(err).To(MatchError(ContainSubstring("failed to construct an annotation link validator")))
		})
	})
})
