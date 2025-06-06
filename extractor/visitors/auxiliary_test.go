package visitors

import (
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auxiliary Tests", func() {
	Context("when processing template context attributes", func() {
		It("should successfully create template context map", func() {
			attributes, _ := annotations.NewAnnotationHolder([]string{
				"// @TemplateContext(testcontext, {key: \"value\"}) Test description",
			}, annotations.CommentSourceRoute)

			// Act
			result, err := getTemplateContextMetadata(&attributes)

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
			}, annotations.CommentSourceRoute)

			// Act
			result, err := getTemplateContextMetadata(&attributes)

			// Assert
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("duplicate template context attribute"))
			Expect(result).To(BeNil())
		})

		It("should handle empty attributes", func() {
			// Arrange
			attributes, _ := annotations.NewAnnotationHolder([]string{}, annotations.CommentSourceRoute)

			// Act
			result, err := getTemplateContextMetadata(&attributes)

			// Assert
			Expect(err).To(BeNil())
			Expect(result).To(BeEmpty())
		})
	})
})
