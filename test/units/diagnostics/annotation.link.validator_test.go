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

	It("Returns a diagnostic warning when an annotation cannot be applied to a receiver", func() {
		receiver.Annotations = utils.GetAnnotationHolderOrFail(
			[]string{
				"// @Route(/)",
				"// @Method(POST)",
			},
			annotations.CommentSourceRoute,
		)

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
