package metadata_test

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/core/annotations"
	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/test/utils"
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
							Root: utils.MakeUniverseRoot("string"),
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
							Root: utils.MakeUniverseRoot("int"),
						},
						IsEmbedded: true,
					},
				},
			}

			// An empty context here is OK, for now - this is done to maintain a uniform reducer signature and is
			// is not currently being used
			reduced, err := structMeta.Reduce(metadata.ReductionContext{})
			Expect(err).To(BeNil())

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

				result, err := controller.Reduce(metadata.ReductionContext{
					GleeceConfig:   ctx.GleeceConfig,
					MetaCache:      ctx.MetadataCache,
					SyncedProvider: ctx.SyncedProvider,
				})

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

				result, err := controller.Reduce(metadata.ReductionContext{
					GleeceConfig:   ctx.GleeceConfig,
					MetaCache:      ctx.MetadataCache,
					SyncedProvider: ctx.SyncedProvider,
				})

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

				_, err := controller.Reduce(metadata.ReductionContext{
					GleeceConfig:   ctx.GleeceConfig,
					MetaCache:      ctx.MetadataCache,
					SyncedProvider: ctx.SyncedProvider,
				})

				Expect(err).To(MatchError(ContainSubstring("a security schema's name cannot be empty")))
			})
		})
	})

	var _ = Describe("ReceiverMeta", func() {
		var receiver metadata.ReceiverMeta

		BeforeEach(func() {
			receiver = metadata.ReceiverMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:    "ExampleReceiver",
					PkgPath: "example.com/test",
				},
			}
		})

		Context("RetValsRange", func() {
			It("Returns correct RetVal ranges for functions with more than a single return value", func() {
				// Get 3 ret-vals to trigger default case for RetValsRange
				receiver.RetVals = utils.GetMockRetVals(3)
				receiver.RetVals[0].Range = common.ResolvedRange{
					StartLine: 11,
					EndLine:   11,
					StartCol:  4,
					EndCol:    9,
				}
				receiver.RetVals[1].Range = common.ResolvedRange{
					StartLine: 12,
					EndLine:   12,
					StartCol:  4,
					EndCol:    9,
				}
				receiver.RetVals[2].Range = common.ResolvedRange{
					StartLine: 13,
					EndLine:   13,
					StartCol:  5,
					EndCol:    10,
				}

				Expect(receiver.RetValsRange()).To(Equal(common.ResolvedRange{
					StartLine: 11,
					EndLine:   13,
					StartCol:  4,
					EndCol:    10,
				}))
			})
		})
	})
})
