package diagnostics

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

		linkValidator = createValidatorOrFail(&receiver)
	})

	It("Returns a diagnostic warning when a non-context receiver parameter is unreferenced", func() {
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

	It("Does not return a diagnostic when a context receiver parameter is unreferenced", func() {
		receiver.Params[0].Type.Name = "Context"
		receiver.Params[0].Type.PkgPath = "context"

		diags := linkValidator.Validate()
		Expect(diags).To(HaveLen(0))
	})

	It("Returns a diagnostic error where a @Path annotation has a non-string 'name' attribute", func() {
		receiverAnnotations := utils.GetAnnotationHolderOrFail(
			[]string{
				"// @Route(/)",
				"// @Method(POST)",
				"// @Path(param1, { name: 12 }) This should raise an 'invalid property' type error",
			},
			annotations.CommentSourceRoute,
		)
		receiver.Annotations = receiverAnnotations
		validator := createValidatorOrFail(&receiver)
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
})

func TestUnitAnnotationLinkValidator(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Annotation Link Validation")
}

func createValidatorOrFail(receiver *metadata.ReceiverMeta) validators.AnnotationLinkValidator {
	validator, err := validators.NewAnnotationLinkValidator(receiver)
	if err != nil {
		Fail(fmt.Sprintf("Could not construct an annotation link validator for testing - %v", err))
	}

	return validator
}
