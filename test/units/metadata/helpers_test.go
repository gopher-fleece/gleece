package metadata_test

import (
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Metadata", func() {
	Context("GetMethodHideOpts", func() {
		It("Returns a 'Hide-Never' result when given nil annotation holder", func() {
			Expect(metadata.GetMethodHideOpts(nil)).To(Equal(
				definitions.MethodHideOptions{Type: definitions.HideMethodNever},
			))
		})
		It("Returns a 'Hide-Never' result when '@Hidden' annotation does not exist", func() {
			holder := utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Route(/)",
					"// @Method(POST)",
				},
				annotations.CommentSourceRoute,
			)
			Expect(metadata.GetMethodHideOpts(holder)).To(Equal(
				definitions.MethodHideOptions{Type: definitions.HideMethodNever},
			))
		})

		It("Returns a 'Hide-Always' result when '@Hidden' annotation exists", func() {
			holder := utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Route(/)",
					"// @Method(POST)",
					"// @Hidden",
				},
				annotations.CommentSourceRoute,
			)

			Expect(metadata.GetMethodHideOpts(holder)).To(Equal(
				definitions.MethodHideOptions{Type: definitions.HideMethodAlways},
			))
		})
	})

	Context("GetSecurityFromContext", func() {
		It("Returns an empty RouteSecurity slice when @Security attribute does not exist", func() {
			holder := utils.GetAnnotationHolderOrFail(
				[]string{},
				annotations.CommentSourceRoute,
			)

			sec, err := metadata.GetSecurityFromContext(holder)
			Expect(err).To(BeNil())
			Expect(sec).To(Equal([]definitions.RouteSecurity{}))
		})

		It("Returns an error if annotations include a @Security with an empty value", func() {
			holder := annotations.NewAnnotationHolderFromData(
				[]annotations.Attribute{
					{
						Name:  "Security",
						Value: "",
					},
				},
				[]annotations.NonAttributeComment{},
			)

			sec, err := metadata.GetSecurityFromContext(&holder)
			Expect(err).To(MatchError(ContainSubstring("a security schema's name cannot be empty")))
			Expect(sec).To(Equal([]definitions.RouteSecurity{}))
		})

		It("Returns an error when given a @Security attribute that has a non-string-slice 'scopes' property", func() {
			holder := utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Security(schemaName, { scopes: [1, 2, 3] })",
				},
				annotations.CommentSourceRoute,
			)

			sec, err := metadata.GetSecurityFromContext(holder)
			Expect(err).To(MatchError(ContainSubstring("failed to cast attribute property 'scopes'")))
			Expect(sec).To(Equal([]definitions.RouteSecurity{}))
		})

	})

	Context("GetRouteSecurityWithInheritance", func() {
		It("Returns an error when unable to retrieve security from the given context", func() {
			holder := annotations.NewAnnotationHolderFromData(
				[]annotations.Attribute{
					{
						Name:  "Security",
						Value: "",
					},
				},
				[]annotations.NonAttributeComment{},
			)

			_, err := metadata.GetRouteSecurityWithInheritance(&holder, []definitions.RouteSecurity{})
			Expect(err).To(MatchError("a security schema's name cannot be empty"))
		})
	})

	Context("GetTemplateContextMetadata", func() {
		It("Should successfully create template context map", func() {
			attributes := utils.GetAnnotationHolderOrFail(
				[]string{
					"// @TemplateContext(testContext, {key: \"value\"}) Test description",
				},
				annotations.CommentSourceRoute,
			)
			result, err := metadata.GetTemplateContextMetadata(attributes)

			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result["testContext"]).To(Equal(definitions.TemplateContext{
				Options:     map[string]any{"key": "value"},
				Description: "Test description",
			}))
		})

		It("Should return error on duplicate template contexts", func() {
			attributes := utils.GetAnnotationHolderOrFail(
				[]string{
					"// @TemplateContext(duplicateContext, {key1: \"value1\"}) First description",
					"// @TemplateContext(duplicateContext, {key2: \"value2\"}) Second description",
				},
				annotations.CommentSourceRoute,
			)

			result, err := metadata.GetTemplateContextMetadata(attributes)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("duplicate template context attribute"))
			Expect(result).To(BeNil())
		})

		It("Should handle empty attributes", func() {
			attributes := utils.GetAnnotationHolderOrFail([]string{}, annotations.CommentSourceRoute)
			result, err := metadata.GetTemplateContextMetadata(attributes)

			Expect(err).To(BeNil())
			Expect(result).To(BeEmpty())
		})
	})

	Context("GetResponseStatusCodeAndDescription", func() {
		It("Returns an error when given a non-integer status code", func() {
			holder := utils.GetAnnotationHolderOrFail(
				[]string{
					"// @Response(NOT_A_NUMBER)",
				},
				annotations.CommentSourceRoute,
			)
			_, _, err := metadata.GetResponseStatusCodeAndDescription(holder, true)
			Expect(err).To(MatchError(ContainSubstring("failed to convert HTTP status code string ")))
		})
	})
})
