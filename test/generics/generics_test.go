package generics_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/common/linq"
	"github.com/gopher-fleece/gleece/v2/core/pipeline"
	"github.com/gopher-fleece/gleece/v2/definitions"
	"github.com/gopher-fleece/gleece/v2/graphs"
	"github.com/gopher-fleece/gleece/v2/graphs/symboldg"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	"github.com/gopher-fleece/gleece/v2/test/utils"
	"github.com/gopher-fleece/gleece/v2/test/utils/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var pipe pipeline.GleecePipeline

var _ = Describe("Generics Controller", func() {
	BeforeEach(func() {
		pipe = utils.GetPipelineOrFail()
		err := pipe.GenerateGraph()
		Expect(err).To(BeNil())

		pipe.Validate()
	})

	Context("RecvWithPrimitiveMapInBody", func() {
		It("Generates correct graph for primitive maps in parameter structs", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"GenericsController",
				"RecvWithPrimitiveMapInBody",
				[]string{"body"},
			)
			// Parameter tree structure checks
			Expect(info.Params).To(HaveLen(1))

			bodyParamNode := utils.GetSingularChildNode(pipe.Graph(), info.Params[0].Node, symboldg.EdgeKindParam)
			bodyStructNode := utils.GetSingularChildTypeNode(pipe.Graph(), bodyParamNode)
			Expect(bodyStructNode.Id.Name).To(Equal("BodyWithPrimitiveMap"))
			structMeta := utils.MustStructMeta(bodyStructNode)
			Expect(structMeta.Fields).To(HaveLen(1))
			utils.AssertFieldIsMap(structMeta, "Dict", "string", "int")

			dictFieldNode, _ := utils.AssertGetField(pipe.Graph(), bodyStructNode, "Dict")
			tParamChildren := utils.FollowThroughCompositeToTypeParams(pipe.Graph(), dictFieldNode)

			Expect(tParamChildren).To(matchers.MatchNodeIdNames([]string{"string", "int"}))

			// RetVals tree structure checks
			Expect(info.RetVals).To(HaveLen(1))
			retTypeNode := utils.GetSingularChildTypeNode(pipe.Graph(), info.RetVals[0].Node)
			Expect(retTypeNode.Id).To(Equal(graphs.NewUniverseSymbolKey("error")))
		})

		It("Creates correct models", func() {
			intermediate, err := pipe.GenerateIntermediate()
			Expect(err).To(BeNil())
			Expect(intermediate.Models.Structs).To(ContainElements(
				[]definitions.StructMetadata{
					{

						Name:    "MonoGenericStructString",
						PkgPath: "github.com/gopher-fleece/gleece/v2/test/generics",
						Fields: []definitions.FieldMetadata{{
							Name:        "Value",
							Type:        "string",
							Deprecation: common.Ptr(definitions.DeprecationOptions{Deprecated: false, Description: ""}),
						}},
					},
					{
						Name:    "MultiGenericStructBoolInt",
						PkgPath: "github.com/gopher-fleece/gleece/v2/test/generics",
						Fields: []definitions.FieldMetadata{
							{
								Name:        "ValueA",
								Type:        "bool",
								Deprecation: common.Ptr(definitions.DeprecationOptions{Deprecated: false, Description: ""}),
							},
							{
								Name:        "ValueB",
								Type:        "int",
								Deprecation: common.Ptr(definitions.DeprecationOptions{Deprecated: false, Description: ""}),
							},
						},
					},
					{
						Name:    "MultiGenericStructStringInt",
						PkgPath: "github.com/gopher-fleece/gleece/v2/test/generics",
						Fields: []definitions.FieldMetadata{
							{
								Name:        "ValueA",
								Type:        "string",
								Deprecation: common.Ptr(definitions.DeprecationOptions{Deprecated: false, Description: ""}),
							},
							{
								Name:        "ValueB",
								Type:        "int",
								Deprecation: common.Ptr(definitions.DeprecationOptions{Deprecated: false, Description: ""}),
							},
						},
					},
				}),
			)
		})
	})

	Context("RecvReturningAPrimitiveMap", func() {
		It("Generates correct graph for primitive maps in parameter structs", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"GenericsController",
				"RecvReturningAPrimitiveMap",
				nil,
			)

			Expect(info.Params).To(HaveLen(0))
			Expect(info.RetVals).To(HaveLen(2))

			mapRetValInfo := linq.First(
				info.RetVals,
				func(retVal utils.FuncRetValInfo) bool {
					retTypeNode := utils.GetSingularChildTypeNode(pipe.Graph(), retVal.Node)
					return retTypeNode.Id.Name != "error"
				},
			)

			Expect(mapRetValInfo).ToNot(BeNil())
			tParamChildren := utils.FollowThroughCompositeToTypeParams(pipe.Graph(), mapRetValInfo.Node)
			Expect(tParamChildren).To(matchers.MatchNodeIdNames([]string{"string", "int"}))
		})
	})

	Context("RecvWithNonPrimitiveMapBody", func() {
		It("Generates correct graph for primitive maps in parameter structs", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"GenericsController",
				"RecvWithNonPrimitiveMapBody",
				[]string{"body"},
			)

			Expect(info.Params).To(HaveLen(1))
			Expect(info.RetVals).To(HaveLen(1))

			mapTypeParam := utils.GetSingularChildTypeNode(pipe.Graph(), info.Params[0].Node)

			Expect(mapTypeParam).ToNot(BeNil())
			tParamChildren := utils.FollowThroughCompositeToTypeParams(pipe.Graph(), info.Params[0].Node)

			expectedStructName := "HoldsVeryNestedStructs"

			Expect(tParamChildren).To(matchers.MatchNodeIdNames([]string{"string", expectedStructName}))

			structNode := linq.First(tParamChildren, func(node *symboldg.SymbolNode) bool {
				return node.Id.Name == expectedStructName
			})

			Expect(structNode).ToNot(BeNil())
			Expect(*structNode).ToNot(BeNil())

			structMeta := utils.MustStructMeta(*structNode)
			Expect(structMeta.Name).To(Equal(expectedStructName))
			Expect(structMeta).To(matchers.HaveStructFields([]matchers.FieldDesc{
				{Name: "FieldA", TypeName: "float32"},
				{Name: "FieldB", TypeName: "uint"},
				{Name: "FieldC", TypeName: "SomeNestedStruct"},
			}))
		})
	})
})

func TestGenericsController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generics Controller")
}
