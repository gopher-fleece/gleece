package validators_test

import (
	"fmt"
	"testing"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/validators"
	"github.com/gopher-fleece/gleece/core/validators/diagnostics"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/gopher-fleece/gleece/test/utils/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Annotation Link Validation", func() {
	var receiver metadata.ReceiverMeta
	var linkValidator validators.AnnotationLinkValidator
	BeforeEach(func() {

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
			Params: []metadata.FuncParam{
				{
					SymNodeMeta: metadata.SymNodeMeta{
						Name: "param1",
						Range: common.ResolvedRange{
							StartLine: 4,
							EndLine:   4,
							StartCol:  6,
							EndCol:    12,
						},
					},
					Ordinal: 0,
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{
							Name:    "int",
							PkgPath: "",
						},
					},
				},
			},
		}

		linkValidator = createLinkValidatorOrFail(&receiver)
	})

	// Covers one instance of DiagLinkerUnreferencedParameter
	It("Returns a DiagLinkerUnreferencedParameter warning when a non-context receiver parameter is unreferenced", func() {
		diags := linkValidator.Validate()
		Expect(diags).To(HaveLen(1))
		Expect(diags[0].Message).To(Equal("Function parameter 'param1' is not referenced by a parameter annotation"))
		Expect(diags[0].Severity).To(Equal(diagnostics.DiagnosticError))
		Expect(diags[0].FilePath).To(ContainSubstring("annotation.link.validator_test.go"))
		Expect(diags[0].Range).To(Equal(common.ResolvedRange{
			StartLine: 4,
			EndLine:   4,
			StartCol:  6,
			EndCol:    12,
		}))
		Expect(diags[0].Code).To(BeEquivalentTo(diagnostics.DiagLinkerUnreferencedParameter))
		Expect(diags[0].Source).To(Equal("gleece"))
	})

	// Covers the negate branch of DiagLinkerUnreferencedParameter
	It("Does not return a diagnostic when a context receiver parameter is unreferenced", func() {
		receiver.Params[0].Type.Name = "Context"
		receiver.Params[0].Type.PkgPath = "context"

		diags := linkValidator.Validate()
		Expect(diags).To(HaveLen(0))
	})

	Context("validateRoute", func() {

		// Covers one instance of DiagAnnotationPropertiesInvalidValueForKey
		It("Returns a DiagAnnotationPropertiesInvalidValueForKey error where a @Path annotation has am invalid 'name' attribute", func() {
			receiverAnnotations := utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Route(/)",
					"// @Method(POST)",
					"// @Path(param1, { name: 12 }) This should raise an 'invalid property' type error",
				},
				annotations.CommentSourceRoute,
			)
			receiver.Annotations = receiverAnnotations
			validator := createLinkValidatorOrFail(&receiver)
			diags := validator.Validate()

			Expect(diags).To(HaveLen(1))
			Expect(diags[0].Message).To(Equal("Invalid value for property 'name' in attribute Path ('12')"))
			Expect(diags[0].Severity).To(Equal(diagnostics.DiagnosticError))
			Expect(diags[0].FilePath).To(ContainSubstring("annotation.link.validator_test.go"))
			Expect(diags[0].Range).To(Equal(common.ResolvedRange{
				StartLine: 0,
				EndLine:   0,
				StartCol:  17,
				EndCol:    29,
			}))
			Expect(diags[0].Code).To(BeEquivalentTo(diagnostics.DiagAnnotationPropertiesInvalidValueForKey))
			Expect(diags[0].Source).To(Equal("gleece"))
		})

		// Covers DiagLinkerDuplicateUrlParam
		It("Returns a DiagLinkerDuplicateUrlParam error when encountering a duplicate URL parameter", func() {
			receiverAnnotations := utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Route(/{url_param}/abc/{url_param})",
					"// @Method(POST)",
					"// @Path(param1, {name: 'url_param'})",
				},
				annotations.CommentSourceRoute,
			)
			receiver.Annotations = receiverAnnotations
			validator := createLinkValidatorOrFail(&receiver)
			diags := validator.Validate()

			Expect(diags).To(HaveLen(1))
			Expect(diags[0].Message).To(Equal("Duplicate URL parameter 'url_param'"))
			Expect(diags[0].Severity).To(Equal(diagnostics.DiagnosticError))
			Expect(diags[0].FilePath).To(ContainSubstring("annotation.link.validator_test.go"))
			Expect(diags[0].Range).To(Equal(common.ResolvedRange{
				StartLine: 0,
				EndLine:   0,
				StartCol:  11,
				EndCol:    22,
			}))
			Expect(diags[0].Code).To(BeEquivalentTo(diagnostics.DiagLinkerDuplicateUrlParam))
			Expect(diags[0].Source).To(Equal("gleece"))
		})

		// Covers DiagLinkerRouteMissingPath
		It("Returns a DiagLinkerRouteMissingPath error when a URL parameter has no matching @Path", func() {
			receiverAnnotations := utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Route(/{id}) This is unreferenced",
					"// @Method(POST)",
					"// @Path(param1, {name: 'not-id'})",
				},
				annotations.CommentSourceRoute,
			)
			receiver.Annotations = receiverAnnotations
			validator := createLinkValidatorOrFail(&receiver)
			diags := validator.Validate()

			Expect(diags).To(HaveLen(2))

			Expect(diags[0].Message).To(Equal("URL parameter 'id' does not have a corresponding @Path annotation"))
			Expect(diags[0].Severity).To(Equal(diagnostics.DiagnosticError))
			Expect(diags[0].FilePath).To(ContainSubstring("annotation.link.validator_test.go"))
			Expect(diags[0].Range).To(Equal(common.ResolvedRange{
				StartLine: 0,
				EndLine:   0,
				StartCol:  11,
				EndCol:    15,
			}))
			Expect(diags[0].Code).To(BeEquivalentTo(diagnostics.DiagLinkerRouteMissingPath))
			Expect(diags[0].Source).To(Equal("gleece"))

			Expect(diags[1].Message).To(Equal("Unknown @Path parameter alias 'not-id'. Did you mean 'id'?"))
			Expect(diags[1].Severity).To(Equal(diagnostics.DiagnosticError))
			Expect(diags[1].FilePath).To(ContainSubstring("annotation.link.validator_test.go"))
			// Note that this range is reliant on Comment positioning which is a bit of
			// a problem to reasonably emulate for tests and currently not done
			Expect(diags[1].Range).To(Equal(common.ResolvedRange{
				StartLine: 0,
				EndLine:   0,
				StartCol:  0,
				EndCol:    0,
			}))
			Expect(diags[1].Code).To(BeEquivalentTo(diagnostics.DiagLinkerPathInvalidRef))
			Expect(diags[1].Source).To(Equal("gleece"))
		})
	})

	Context("validatePathAnnotations", func() {

		// Covers DiagLinkerPathInvalidRef
		It("Returns a DiagLinkerPathInvalidRef if a @Path value references a non-existent parameter", func() {
			receiverAnnotations := utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Route(/)",
					"// @Method(POST)",
					"// @Path(param2) Param2 does not exist",
				},
				annotations.CommentSourceRoute,
			)
			receiver.Annotations = receiverAnnotations
			validator := createLinkValidatorOrFail(&receiver)
			diags := validator.Validate()

			Expect(diags).To(HaveLen(2))

			Expect(diags[0].Message).To(Equal("@Path 'param2' is not a parameter of TestReceiver. Did you mean 'param1'?"))
			Expect(diags[0].Severity).To(Equal(diagnostics.DiagnosticError))
			Expect(diags[0].FilePath).To(ContainSubstring("annotation.link.validator_test.go"))
			Expect(diags[0].Range).To(Equal(common.ResolvedRange{
				StartLine: 0,
				EndLine:   0,
				StartCol:  9,
				EndCol:    15,
			}))
			Expect(diags[0].Code).To(BeEquivalentTo(diagnostics.DiagLinkerPathInvalidRef))
			Expect(diags[0].Source).To(Equal("gleece"))

			Expect(diags[1]).To(BeDiagnosticErrorWithCodeAndMessage(
				diagnostics.DiagLinkerUnreferencedParameter,
				"Function parameter 'param1' is not referenced by a parameter annotation",
			))
		})

		// Covers DiagLinkerMultipleParameterRefs and DiagLinkerDuplicatePathParam
		It("Returns a DiagLinkerMultipleParameterRefs if a function param is referenced by multiple @Path annotations", func() {
			receiverAnnotations := utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Route(/{url_param1}/{url_param2})",
					"// @Method(POST)",
					"// @Path(param1, {name: 'url_param1'}) Ref1",
					"// @Path(param1, {name: 'url_param2'}) Ref1",
				},
				annotations.CommentSourceRoute,
			)
			receiver.Annotations = receiverAnnotations
			validator := createLinkValidatorOrFail(&receiver)
			diags := validator.Validate()

			Expect(diags).To(HaveLen(2))

			Expect(diags[0].Message).To(Equal("Function parameter 'param1' is referenced by multiple @Path attributes"))
			Expect(diags[0].Severity).To(Equal(diagnostics.DiagnosticError))
			Expect(diags[0].FilePath).To(ContainSubstring("annotation.link.validator_test.go"))
			Expect(diags[0].Range).To(Equal(common.ResolvedRange{
				StartLine: 0,
				EndLine:   0,
				StartCol:  9,
				EndCol:    15,
			}))
			Expect(diags[0].Code).To(BeEquivalentTo(diagnostics.DiagLinkerMultipleParameterRefs))
			Expect(diags[0].Source).To(Equal("gleece"))

			Expect(diags[1]).To(BeDiagnosticErrorWithCodeAndMessage(
				diagnostics.DiagLinkerDuplicatePathParam,
				"Duplicate @Path parameter reference 'param1'",
			))
		})
	})
})

func TestUnitAnnotationLinkValidator(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Annotation Link Validation")
}

func createLinkValidatorOrFail(receiver *metadata.ReceiverMeta) validators.AnnotationLinkValidator {
	validator, err := validators.NewAnnotationLinkValidator(receiver)
	if err != nil {
		Fail(fmt.Sprintf("Could not construct an annotation link validator for testing - %v", err))
	}

	return validator
}
