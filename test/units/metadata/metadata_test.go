package metadata_test

import (
	"fmt"
	"go/ast"
	"go/types"
	"testing"

	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/visitors"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	ctx visitors.VisitContext
)

var _ = BeforeSuite(func() {
	ctx = utils.GetVisitContextByRelativeConfigOrFail("gleece.test.config.json")
})

var _ = Describe("Unit Tests - Metadata", func() {
	var _ = Describe("TypeLayer", func() {

		Context("NewPointerLayer", func() {
			It("creates a pointer layer with correct kind", func() {
				layer := metadata.NewPointerLayer()
				Expect(layer.Kind).To(Equal(metadata.TypeLayerKindPointer))
				Expect(layer.KeyType).To(BeNil())
				Expect(layer.ValueType).To(BeNil())
				Expect(layer.BaseTypeRef).To(BeNil())
			})
		})

		Context("NewArrayLayer", func() {
			It("creates an array layer with correct kind", func() {
				layer := metadata.NewArrayLayer()
				Expect(layer.Kind).To(Equal(metadata.TypeLayerKindArray))
				Expect(layer.KeyType).To(BeNil())
				Expect(layer.ValueType).To(BeNil())
				Expect(layer.BaseTypeRef).To(BeNil())
			})
		})

		Context("NewMapLayer", func() {
			It("creates a map layer with correct key and value", func() {
				key := graphs.NewUniverseSymbolKey("string")
				value := graphs.NewUniverseSymbolKey("int")
				layer := metadata.NewMapLayer(&key, &value)

				Expect(layer.Kind).To(Equal(metadata.TypeLayerKindMap))
				Expect(layer.KeyType).To(Equal(&key))
				Expect(layer.ValueType).To(Equal(&value))
				Expect(layer.BaseTypeRef).To(BeNil())
			})
		})

		Context("NewBaseLayer", func() {
			It("creates a base layer with correct base reference", func() {
				base := graphs.NewUniverseSymbolKey("MyStruct")
				layer := metadata.NewBaseLayer(&base)

				Expect(layer.Kind).To(Equal(metadata.TypeLayerKindBase))
				Expect(layer.BaseTypeRef).To(Equal(&base))
				Expect(layer.KeyType).To(BeNil())
				Expect(layer.ValueType).To(BeNil())
			})
		})
	})

	var _ = Describe("EnumValueKind", func() {
		Describe("NewEnumValueKind", func() {
			DescribeTable("returns correct EnumValueKind",
				func(kind types.BasicKind, expected metadata.EnumValueKind) {
					result, err := metadata.NewEnumValueKind(kind)
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(expected))
				},
				Entry("string", types.String, metadata.EnumValueKindString),
				Entry("int", types.Int, metadata.EnumValueKindInt),
				Entry("int8", types.Int8, metadata.EnumValueKindInt8),
				Entry("int16", types.Int16, metadata.EnumValueKindInt16),
				Entry("int32", types.Int32, metadata.EnumValueKindInt32),
				Entry("int64", types.Int64, metadata.EnumValueKindInt64),
				Entry("uint", types.Uint, metadata.EnumValueKindUInt),
				Entry("uint8", types.Uint8, metadata.EnumValueKindUInt8),
				Entry("uint16", types.Uint16, metadata.EnumValueKindUInt16),
				Entry("uint32", types.Uint32, metadata.EnumValueKindUInt32),
				Entry("uint64", types.Uint64, metadata.EnumValueKindUInt64),
				Entry("float32", types.Float32, metadata.EnumValueKindFloat32),
				Entry("float64", types.Float64, metadata.EnumValueKindFloat64),
				Entry("bool", types.Bool, metadata.EnumValueKindBool),
			)

			It("returns error on unsupported kind", func() {
				// types.UnsafePointer is not supported
				unsupported := types.UnsafePointer
				_, err := metadata.NewEnumValueKind(unsupported)
				Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("unsupported basic kind: %v", unsupported))))
			})
		})
	})

	var _ = Describe("StructMeta", func() {
		holder, err := annotations.NewAnnotationHolder(
			[]string{
				"// @Description This is a test struct",
				"// @Deprecated Use something else",
			},
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
				holder := getAnnotationHolderOrFail(
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
				holder := getAnnotationHolderOrFail(
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

func getAnnotationHolderOrFail(comments []string, appliedOn annotations.CommentSource) *annotations.AnnotationHolder {
	holder, err := annotations.NewAnnotationHolder(comments, appliedOn)
	Expect(err).ToNot(HaveOccurred())
	return &holder
}
func TestUnitCommons(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Metadata")
}
