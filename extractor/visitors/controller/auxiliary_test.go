package controller

import (
	"testing"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auxiliary Tests", func() {
	var visitor ControllerVisitor

	BeforeEach(func() {
		visitor = ControllerVisitor{}
	})

	Context("when processing template context attributes", func() {
		It("should successfully create template context map", func() {
			attributes, _ := annotations.NewAnnotationHolder([]string{
				"// @TemplateContext(testcontext, {key: \"value\"}) Test description",
			})

			// Act
			result, err := visitor.getTemplateContextMetadata(&attributes)

			// Assert
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result["testcontext"]).To(Equal(definitions.TemplateContext{
				Options:     map[string]any{"key": "value"},
				Description: "Test description",
			}))
		})

		It("should return error on duplicate template contexts", func() {
			// Arrange

			attributes, _ := annotations.NewAnnotationHolder([]string{
				"// @TemplateContext(duplicatecontext, {key1: \"value1\"}) First description",
				"// @TemplateContext(duplicatecontext, {key2: \"value2\"}) Second description",
			})

			// Act
			result, err := visitor.getTemplateContextMetadata(&attributes)

			// Assert
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Duplicate template context attribute"))
			Expect(result).To(BeNil())
		})

		It("should handle empty attributes", func() {
			// Arrange
			attributes, _ := annotations.NewAnnotationHolder([]string{})

			// Act
			result, err := visitor.getTemplateContextMetadata(&attributes)

			// Assert
			Expect(err).To(BeNil())
			Expect(result).To(BeEmpty())
		})
	})
})

func TestAuxiliary(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auxiliary Suite")
}
