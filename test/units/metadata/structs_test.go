package metadata_test

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Metadata", func() {
	var _ = Describe("StructMeta", func() {
		holder, err := annotations.NewAnnotationHolder(
			utils.CommentsToCommentBlock([]string{
				"// @Description This is a test struct",
				"// @Deprecated Use something else",
			}, 1),
			annotations.CommentSourceProperty,
		)

		if err != nil {
			Fail("Could not create an annotation holder")
		}

		It("reduces StructMeta into StructMetadata with correct values", func() {
			structMeta := metadata.StructMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:        "TestStruct",
					PkgPath:     "example.com/mypkg",
					Node:        &ast.TypeSpec{}, // not actually used in Reduce
					Annotations: &holder,
				},
				Fields: []metadata.FieldMeta{
					{
						SymNodeMeta: metadata.SymNodeMeta{
							Name:        "Field1",
							Node:        &ast.Field{},
							Annotations: &holder,
						},
						Type: metadata.TypeUsageMeta{
							SymNodeMeta: metadata.SymNodeMeta{
								Name:        "string",
								PkgPath:     "",
								Annotations: &holder,
							},
							Layers: []metadata.TypeLayer{},
						},
					},
					{
						SymNodeMeta: metadata.SymNodeMeta{
							Name:        "Field2",
							Node:        &ast.Field{},
							Annotations: &holder,
						},
						Type: metadata.TypeUsageMeta{
							SymNodeMeta: metadata.SymNodeMeta{
								Name:        "int",
								PkgPath:     "",
								Annotations: &holder,
							},
							Layers: []metadata.TypeLayer{},
						},
						IsEmbedded: true,
					},
				},
			}

			reduced := structMeta.Reduce()

			Expect(reduced.Name).To(Equal("TestStruct"))
			Expect(reduced.PkgPath).To(Equal("example.com/mypkg"))
			Expect(reduced.Description).To(Equal("This is a test struct"))
			Expect(reduced.Deprecation.Description).To(Equal("Use something else"))

			Expect(reduced.Fields).To(HaveLen(2))

			Expect(reduced.Fields[0].Name).To(Equal("Field1"))
			Expect(reduced.Fields[0].Type).To(Equal("string"))
			Expect(reduced.Fields[0].IsEmbedded).To(BeFalse())

			Expect(reduced.Fields[1].Name).To(Equal("Field2"))
			Expect(reduced.Fields[1].Type).To(Equal("int"))
			Expect(reduced.Fields[1].IsEmbedded).To(BeTrue())
		})
	})

	var _ = Describe("ControllerMeta", func() {
		var controller metadata.ControllerMeta

		BeforeEach(func() {
			controller = metadata.ControllerMeta{
				Struct: metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:    "ExampleController",
						PkgPath: "example.com/test",
					},
				},
			}
		})

		Context("Reduce", func() {

			It("reduces successfully with explicit security", func() {
				holder := utils.GetAnnotationHolderOrFail(
					[]string{
						"// @Description Example controller",
						"// @Route(/example)",
						"// @Tag(Sample)",
						"// @Security(ApiKeyAuth, { scopes: [] })",
					},
					annotations.CommentSourceController,
				)

				controller.Struct.Annotations = holder

				result, err := controller.Reduce(ctx.GleeceConfig, ctx.MetadataCache, ctx.SyncedProvider)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Name).To(Equal("ExampleController"))
				Expect(result.Description).To(Equal("Example controller"))
				Expect(result.Tag).To(Equal("Sample"))
				Expect(result.RestMetadata.Path).To(Equal("/example"))
				Expect(result.Security).To(HaveLen(1))
				Expect(result.Security[0].SecurityAnnotation).To(HaveLen(1))
				Expect(result.Security[0].SecurityAnnotation[0].SchemaName).To(Equal("ApiKeyAuth"))
				Expect(result.Security[0].SecurityAnnotation[0].Scopes).To(BeEmpty())
			})

			It("falls back to default security if none are defined", func() {
				holder := utils.GetAnnotationHolderOrFail(
					[]string{
						"// @Description Example controller",
						"// @Route(/example)",
						"// @Tag(Sample)",
					},
					annotations.CommentSourceController,
				)
				controller.Struct.Annotations = holder

				result, err := controller.Reduce(ctx.GleeceConfig, ctx.MetadataCache, ctx.SyncedProvider)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Security).To(Equal(metadata.GetDefaultSecurity(ctx.GleeceConfig)))
			})

			It("Produces a correct error if given an empty security schema name", func() {
				// This error is basically impossible to trigger using normal flows
				holder := annotations.NewAnnotationHolderFromData(
					[]annotations.Attribute{
						{
							Name:  "Security",
							Value: "", // Empty on purpose, to trigger an error downstream
							Properties: map[string]any{
								"scopes": []string{},
							},
							Description: "",
						},
						{
							Name:        "Route",
							Value:       "/example", // Empty on purpose, to trigger an error downstream
							Properties:  nil,
							Description: "",
						},
					},
					[]annotations.NonAttributeComment{},
				)

				controller.Struct.Annotations = &holder

				_, err := controller.Reduce(ctx.GleeceConfig, ctx.MetadataCache, ctx.SyncedProvider)
				Expect(err).To(MatchError(ContainSubstring("a security schema's name cannot be empty")))
			})
		})
	})
})
