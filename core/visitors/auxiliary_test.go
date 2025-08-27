package visitors_test

import (
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auxiliary Tests", func() {
	Context("When processing template context attributes", func() {
		It("Should successfully create template context map", func() {
			attributes := utils.GetAnnotationHolderOrFail([]string{
				"// @TemplateContext(testcontext, {key: \"value\"}) Test description",
			}, annotations.CommentSourceRoute)

			// Act
			result, err := metadata.GetTemplateContextMetadata(attributes)

			// Assert
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result["testcontext"]).To(Equal(definitions.TemplateContext{
				Options:     map[string]any{"key": "value"},
				Description: "Test description",
			}))
		})

		It("Should return error on duplicate template contexts", func() {
			// Arrange

			attributes := utils.GetAnnotationHolderOrFail([]string{
				"// @TemplateContext(duplicatecontext, {key1: \"value1\"}) First description",
				"// @TemplateContext(duplicatecontext, {key2: \"value2\"}) Second description",
			}, annotations.CommentSourceRoute)

			// Act
			result, err := metadata.GetTemplateContextMetadata(attributes)

			// Assert
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("duplicate template context attribute"))
			Expect(result).To(BeNil())
		})

		It("Should handle empty attributes", func() {
			// Arrange
			attributes := utils.GetAnnotationHolderOrFail([]string{}, annotations.CommentSourceRoute)

			// Act
			result, err := metadata.GetTemplateContextMetadata(attributes)

			// Assert
			Expect(err).To(BeNil())
			Expect(result).To(BeEmpty())
		})
	})
})
