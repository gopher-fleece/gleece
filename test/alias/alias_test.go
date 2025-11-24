package generics_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/common/linq"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/pipeline"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const typedefAliasName = "TypedefAlias"
const assignedAliasName = "AssignedAlias"

var pipe pipeline.GleecePipeline

var _ = BeforeSuite(func() {
	pipe = utils.GetPipelineOrFail()
	err := pipe.GenerateGraph()
	Expect(err).To(BeNil())
})

var _ = Describe("Alias Controller", func() {

	Context("ReceivesTypedefAliasQuery", func() {
		It("Processes a TypeDef Alias parameter", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"AliasController",
				"ReceivesTypedefAliasQuery",
				[]string{"alias"},
			)

			Expect(info.Params).To(HaveLen(1))
			Expect(info.RetVals).To(HaveLen(1))

			aliasTypeParam := utils.GetSingularChildTypeNode(pipe.Graph(), info.Params[0].Node)

			Expect(aliasTypeParam).ToNot(BeNil())
			Expect(aliasTypeParam.Kind).To(Equal(common.SymKindAlias))

			aliasMeta := utils.MustAliasMeta(aliasTypeParam)
			Expect(aliasMeta.Name).To(Equal(typedefAliasName))

			relevantEdges := common.MapValues(pipe.Graph().GetEdges(
				aliasTypeParam.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType}),
			)

			outgoingAliasEdge := linq.First(relevantEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == aliasTypeParam && edge.Edge.To.Name == "string"
			})

			aliasedType := pipe.Graph().Get(outgoingAliasEdge.Edge.To)

			Expect(aliasedType).ToNot(BeNil())
			Expect(aliasedType.Kind).To(Equal(common.SymKindBuiltin))
			Expect(aliasedType.Id.Name).To(Equal("string"))
		})
	})

	Context("ReceivesATypedefAliasInBody", func() {
		It("Processes a Typedef Alias inside a body parameter", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"AliasController",
				"ReceivesATypedefAliasInBody",
				[]string{"body"},
			)

			Expect(info.Params).To(HaveLen(1))
			Expect(info.RetVals).To(HaveLen(1))

			bodyNode := utils.GetSingularChildTypeNode(pipe.Graph(), info.Params[0].Node)
			Expect(bodyNode).ToNot(BeNil())
			Expect(bodyNode.Kind).To(Equal(common.SymKindStruct))

			// find the "Ally" field edge from the struct
			fieldEdges := common.MapValues(pipe.Graph().GetEdges(
				bodyNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindField},
			))

			allyFieldEdge := linq.First(fieldEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return edge.Edge.To.Name == "Ally"
			})
			allyField := pipe.Graph().Get(allyFieldEdge.Edge.To)
			Expect(allyField).ToNot(BeNil())

			// get the type of the "Ally" field
			allyTypeEdges := common.MapValues(pipe.Graph().GetEdges(
				allyField.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			allyTypeEdge := linq.First(allyTypeEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == allyField
			})
			allyType := pipe.Graph().Get(allyTypeEdge.Edge.To)

			Expect(allyType).ToNot(BeNil())
			Expect(allyType.Kind).To(Equal(common.SymKindAlias))

			aliasMeta := utils.MustAliasMeta(allyType)
			Expect(aliasMeta.Name).To(Equal(typedefAliasName))
			Expect(aliasMeta.AliasType).To(Equal(metadata.AliasKindTypedef))

			// ensure alias points to builtin string
			relevantEdges := common.MapValues(pipe.Graph().GetEdges(
				allyType.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			outgoing := linq.First(relevantEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == allyType && edge.Edge.To.Name == "string"
			})
			aliasedType := pipe.Graph().Get(outgoing.Edge.To)
			Expect(aliasedType).ToNot(BeNil())
			Expect(aliasedType.Kind).To(Equal(common.SymKindBuiltin))
			Expect(aliasedType.Id.Name).To(Equal("string"))
		})
	})

	Context("ReturnsATypedefAlias", func() {
		It("Processes a Typedef Alias return value", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"AliasController",
				"ReturnsATypedefAlias",
				[]string{},
			)

			Expect(info.Params).To(HaveLen(0))
			Expect(info.RetVals).To(HaveLen(2))

			var retTypeNode *symboldg.SymbolNode
			for _, retVal := range info.RetVals {
				node := utils.GetSingularChildTypeNode(pipe.Graph(), retVal.Node)
				if node != nil && node.Id.Name == typedefAliasName {
					retTypeNode = node
					break
				}
			}

			Expect(retTypeNode).ToNot(BeNil())
			Expect(retTypeNode.Kind).To(Equal(common.SymKindAlias))

			aliasMeta := utils.MustAliasMeta(retTypeNode)
			Expect(aliasMeta.Name).To(Equal(typedefAliasName))
			Expect(aliasMeta.AliasType).To(Equal(metadata.AliasKindTypedef))

			relevantEdges := common.MapValues(pipe.Graph().GetEdges(
				retTypeNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))

			outgoing := linq.First(relevantEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == retTypeNode && edge.Edge.To.Name == "string"
			})

			aliasedType := pipe.Graph().Get(outgoing.Edge.To)
			Expect(aliasedType).ToNot(BeNil())
			Expect(aliasedType.Kind).To(Equal(common.SymKindBuiltin))
			Expect(aliasedType.Id.Name).To(Equal("string"))
		})
	})

	Context("ReceivesAssignedAliasQuery", func() {
		It("Processes an Assigned Alias parameter", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"AliasController",
				"ReceivesAssignedAliasQuery",
				[]string{"alias"},
			)

			Expect(info.Params).To(HaveLen(1))
			Expect(info.RetVals).To(HaveLen(1))

			assignedTypeParam := utils.GetSingularChildTypeNode(pipe.Graph(), info.Params[0].Node)

			Expect(assignedTypeParam).ToNot(BeNil())
			Expect(assignedTypeParam.Kind).To(Equal(common.SymKindAlias))

			assignedMeta := utils.MustAliasMeta(assignedTypeParam)
			Expect(assignedMeta.Name).To(Equal(assignedAliasName))
			Expect(assignedMeta.AliasType).To(Equal(metadata.AliasKindAssigned))

			relevantEdges := common.MapValues(pipe.Graph().GetEdges(
				assignedTypeParam.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))

			outgoingAssignedEdge := linq.First(relevantEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == assignedTypeParam && edge.Edge.To.Name == "string"
			})

			aliasedType := pipe.Graph().Get(outgoingAssignedEdge.Edge.To)

			Expect(aliasedType).ToNot(BeNil())
			Expect(aliasedType.Kind).To(Equal(common.SymKindBuiltin))
			Expect(aliasedType.Id.Name).To(Equal("string"))
		})
	})

	Context("ReceivesAnAssignedAliasInBody", func() {
		It("Processes an Assigned Alias inside a body parameter", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"AliasController",
				"ReceivesAnAssignedAliasInBody",
				[]string{"body"},
			)

			Expect(info.Params).To(HaveLen(1))
			Expect(info.RetVals).To(HaveLen(1))

			bodyNode := utils.GetSingularChildTypeNode(pipe.Graph(), info.Params[0].Node)
			Expect(bodyNode).ToNot(BeNil())
			Expect(bodyNode.Kind).To(Equal(common.SymKindStruct))

			fieldEdges := common.MapValues(pipe.Graph().GetEdges(
				bodyNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindField},
			))
			allyFieldEdge := linq.First(fieldEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return edge.Edge.To.Name == "Ally"
			})
			allyField := pipe.Graph().Get(allyFieldEdge.Edge.To)
			Expect(allyField).ToNot(BeNil())

			allyTypeEdges := common.MapValues(pipe.Graph().GetEdges(
				allyField.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			allyTypeEdge := linq.First(allyTypeEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == allyField
			})
			allyType := pipe.Graph().Get(allyTypeEdge.Edge.To)

			Expect(allyType).ToNot(BeNil())
			Expect(allyType.Kind).To(Equal(common.SymKindAlias))

			aliasMeta := utils.MustAliasMeta(allyType)
			Expect(aliasMeta.Name).To(Equal(assignedAliasName))
			Expect(aliasMeta.AliasType).To(Equal(metadata.AliasKindAssigned))

			relevantEdges := common.MapValues(pipe.Graph().GetEdges(
				allyType.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			outgoing := linq.First(relevantEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == allyType && edge.Edge.To.Name == "string"
			})
			aliasedType := pipe.Graph().Get(outgoing.Edge.To)
			Expect(aliasedType).ToNot(BeNil())
			Expect(aliasedType.Kind).To(Equal(common.SymKindBuiltin))
			Expect(aliasedType.Id.Name).To(Equal("string"))
		})
	})

	Context("ReturnsAnAssignedAlias", func() {
		It("Processes an Assigned Alias return value", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"AliasController",
				"ReturnsAnAssignedAlias",
				[]string{},
			)

			Expect(info.Params).To(HaveLen(0))
			Expect(info.RetVals).To(HaveLen(2))

			var retTypeNode *symboldg.SymbolNode
			for _, retVal := range info.RetVals {
				node := utils.GetSingularChildTypeNode(pipe.Graph(), retVal.Node)
				if node != nil && node.Id.Name == assignedAliasName {
					retTypeNode = node
					break
				}
			}

			Expect(retTypeNode).ToNot(BeNil())
			Expect(retTypeNode.Kind).To(Equal(common.SymKindAlias))

			aliasMeta := utils.MustAliasMeta(retTypeNode)
			Expect(aliasMeta.Name).To(Equal(assignedAliasName))
			Expect(aliasMeta.AliasType).To(Equal(metadata.AliasKindAssigned))

			relevantEdges := common.MapValues(pipe.Graph().GetEdges(
				retTypeNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			outgoing := linq.First(relevantEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == retTypeNode && edge.Edge.To.Name == "string"
			})
			aliasedType := pipe.Graph().Get(outgoing.Edge.To)
			Expect(aliasedType).ToNot(BeNil())
			Expect(aliasedType.Kind).To(Equal(common.SymKindBuiltin))
			Expect(aliasedType.Id.Name).To(Equal("string"))
		})
	})

	Context("ReceivesNestedTypedefAliasQuery", func() {
		It("Processes a Nested Typedef Alias parameter", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"AliasController",
				"ReceivesNestedTypedefAliasQuery",
				[]string{"alias"},
			)

			Expect(info.Params).To(HaveLen(1))
			Expect(info.RetVals).To(HaveLen(1))

			nestedTypeParam := utils.GetSingularChildTypeNode(pipe.Graph(), info.Params[0].Node)
			Expect(nestedTypeParam).ToNot(BeNil())
			Expect(nestedTypeParam.Kind).To(Equal(common.SymKindAlias))

			nestedMeta := utils.MustAliasMeta(nestedTypeParam)
			Expect(nestedMeta.Name).To(Equal("NestedTypedefAlias"))
			Expect(nestedMeta.AliasType).To(Equal(metadata.AliasKindTypedef))

			// nested alias should point to TypedefAlias
			relevantEdges := common.MapValues(pipe.Graph().GetEdges(
				nestedTypeParam.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			outgoingToTypedef := linq.First(relevantEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == nestedTypeParam && edge.Edge.To.Name == typedefAliasName
			})
			aliasedType := pipe.Graph().Get(outgoingToTypedef.Edge.To)
			Expect(aliasedType).ToNot(BeNil())
			Expect(aliasedType.Kind).To(Equal(common.SymKindAlias))

			// and that TypedefAlias resolves further to builtin string
			typedefEdges := common.MapValues(pipe.Graph().GetEdges(
				aliasedType.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			outgoingToBuiltin := linq.First(typedefEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == aliasedType && edge.Edge.To.Name == "string"
			})
			aliasedBuiltin := pipe.Graph().Get(outgoingToBuiltin.Edge.To)
			Expect(aliasedBuiltin).ToNot(BeNil())
			Expect(aliasedBuiltin.Kind).To(Equal(common.SymKindBuiltin))
			Expect(aliasedBuiltin.Id.Name).To(Equal("string"))
		})
	})

	Context("ReceivesAnNestedTypedefAliasInBody", func() {
		It("Processes a Nested Typedef Alias inside a body parameter", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"AliasController",
				"ReceivesAnNestedTypedefAliasInBody",
				[]string{"body"},
			)

			Expect(info.Params).To(HaveLen(1))
			Expect(info.RetVals).To(HaveLen(1))

			bodyNode := utils.GetSingularChildTypeNode(pipe.Graph(), info.Params[0].Node)
			Expect(bodyNode).ToNot(BeNil())
			Expect(bodyNode.Kind).To(Equal(common.SymKindStruct))

			fieldEdges := common.MapValues(pipe.Graph().GetEdges(
				bodyNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindField},
			))
			allyFieldEdge := linq.First(fieldEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return edge.Edge.To.Name == "Ally"
			})
			allyField := pipe.Graph().Get(allyFieldEdge.Edge.To)
			Expect(allyField).ToNot(BeNil())

			allyTypeEdges := common.MapValues(pipe.Graph().GetEdges(
				allyField.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			allyTypeEdge := linq.First(allyTypeEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == allyField
			})
			allyType := pipe.Graph().Get(allyTypeEdge.Edge.To)

			Expect(allyType).ToNot(BeNil())
			Expect(allyType.Kind).To(Equal(common.SymKindAlias))

			aliasMeta := utils.MustAliasMeta(allyType)
			Expect(aliasMeta.Name).To(Equal("NestedTypedefAlias"))
			Expect(aliasMeta.AliasType).To(Equal(metadata.AliasKindTypedef))

			// Nested alias -> TypedefAlias -> string
			relevantEdges := common.MapValues(pipe.Graph().GetEdges(
				allyType.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			toTypedef := linq.First(relevantEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == allyType && edge.Edge.To.Name == typedefAliasName
			})
			typedefNode := pipe.Graph().Get(toTypedef.Edge.To)
			Expect(typedefNode).ToNot(BeNil())

			typedefEdges := common.MapValues(pipe.Graph().GetEdges(
				typedefNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			toBuiltin := linq.First(typedefEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == typedefNode && edge.Edge.To.Name == "string"
			})
			builtin := pipe.Graph().Get(toBuiltin.Edge.To)
			Expect(builtin).ToNot(BeNil())
			Expect(builtin.Kind).To(Equal(common.SymKindBuiltin))
			Expect(builtin.Id.Name).To(Equal("string"))
		})
	})

	Context("ReturnsAnNestedTypedefAlias", func() {
		It("Processes a Nested Typedef Alias return value", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"AliasController",
				"ReturnsAnNestedTypedefAlias",
				[]string{},
			)

			Expect(info.Params).To(HaveLen(0))
			Expect(info.RetVals).To(HaveLen(2))

			var retTypeNode *symboldg.SymbolNode
			for _, retVal := range info.RetVals {
				node := utils.GetSingularChildTypeNode(pipe.Graph(), retVal.Node)
				if node != nil && node.Id.Name == "NestedTypedefAlias" {
					retTypeNode = node
					break
				}
			}

			Expect(retTypeNode).ToNot(BeNil())
			Expect(retTypeNode.Kind).To(Equal(common.SymKindAlias))

			aliasMeta := utils.MustAliasMeta(retTypeNode)
			Expect(aliasMeta.Name).To(Equal("NestedTypedefAlias"))
			Expect(aliasMeta.AliasType).To(Equal(metadata.AliasKindTypedef))

			// Nested -> TypedefAlias -> string
			nestedEdges := common.MapValues(pipe.Graph().GetEdges(
				retTypeNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			toTypedef := linq.First(nestedEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == retTypeNode && edge.Edge.To.Name == typedefAliasName
			})
			typedefNode := pipe.Graph().Get(toTypedef.Edge.To)
			Expect(typedefNode).ToNot(BeNil())

			typedefEdges := common.MapValues(pipe.Graph().GetEdges(
				typedefNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			toBuiltin := linq.First(typedefEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == typedefNode && edge.Edge.To.Name == "string"
			})
			builtin := pipe.Graph().Get(toBuiltin.Edge.To)
			Expect(builtin).ToNot(BeNil())
			Expect(builtin.Kind).To(Equal(common.SymKindBuiltin))
			Expect(builtin.Id.Name).To(Equal("string"))
		})
	})

	Context("ReceivesNestedAssignedAliasQuery", func() {
		It("Processes a Nested Assigned Alias parameter", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"AliasController",
				"ReceivesNestedAssignedAliasQuery",
				[]string{"alias"},
			)

			Expect(info.Params).To(HaveLen(1))
			Expect(info.RetVals).To(HaveLen(1))

			nestedTypeParam := utils.GetSingularChildTypeNode(pipe.Graph(), info.Params[0].Node)
			Expect(nestedTypeParam).ToNot(BeNil())
			Expect(nestedTypeParam.Kind).To(Equal(common.SymKindAlias))

			nestedMeta := utils.MustAliasMeta(nestedTypeParam)
			Expect(nestedMeta.Name).To(Equal("NestedAssignedAlias"))
			Expect(nestedMeta.AliasType).To(Equal(metadata.AliasKindAssigned))

			// nested assigned alias should point to TypedefAlias
			relevantEdges := common.MapValues(pipe.Graph().GetEdges(
				nestedTypeParam.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			toTypedef := linq.First(relevantEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == nestedTypeParam && edge.Edge.To.Name == typedefAliasName
			})
			typedefNode := pipe.Graph().Get(toTypedef.Edge.To)
			Expect(typedefNode).ToNot(BeNil())

			// TypedefAlias -> string
			typedefEdges := common.MapValues(pipe.Graph().GetEdges(
				typedefNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			toBuiltin := linq.First(typedefEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == typedefNode && edge.Edge.To.Name == "string"
			})
			builtin := pipe.Graph().Get(toBuiltin.Edge.To)
			Expect(builtin).ToNot(BeNil())
			Expect(builtin.Kind).To(Equal(common.SymKindBuiltin))
			Expect(builtin.Id.Name).To(Equal("string"))
		})
	})

	Context("ReceivesAnNestedAssignedAliasInBody", func() {
		It("Processes a Nested Assigned Alias inside a body parameter", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"AliasController",
				"ReceivesAnNestedAssignedAliasInBody",
				[]string{"body"},
			)

			Expect(info.Params).To(HaveLen(1))
			Expect(info.RetVals).To(HaveLen(1))

			bodyNode := utils.GetSingularChildTypeNode(pipe.Graph(), info.Params[0].Node)
			Expect(bodyNode).ToNot(BeNil())
			Expect(bodyNode.Kind).To(Equal(common.SymKindStruct))

			fieldEdges := common.MapValues(pipe.Graph().GetEdges(
				bodyNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindField},
			))
			allyFieldEdge := linq.First(fieldEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return edge.Edge.To.Name == "Ally"
			})
			allyField := pipe.Graph().Get(allyFieldEdge.Edge.To)
			Expect(allyField).ToNot(BeNil())

			allyTypeEdges := common.MapValues(pipe.Graph().GetEdges(
				allyField.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			allyTypeEdge := linq.First(allyTypeEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == allyField
			})
			allyType := pipe.Graph().Get(allyTypeEdge.Edge.To)

			Expect(allyType).ToNot(BeNil())
			Expect(allyType.Kind).To(Equal(common.SymKindAlias))

			aliasMeta := utils.MustAliasMeta(allyType)
			Expect(aliasMeta.Name).To(Equal("NestedAssignedAlias"))
			Expect(aliasMeta.AliasType).To(Equal(metadata.AliasKindAssigned))

			// NestedAssignedAlias -> TypedefAlias -> string
			relevantEdges := common.MapValues(pipe.Graph().GetEdges(
				allyType.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			toTypedef := linq.First(relevantEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == allyType && edge.Edge.To.Name == typedefAliasName
			})
			typedefNode := pipe.Graph().Get(toTypedef.Edge.To)
			Expect(typedefNode).ToNot(BeNil())

			typedefEdges := common.MapValues(pipe.Graph().GetEdges(
				typedefNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			toBuiltin := linq.First(typedefEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == typedefNode && edge.Edge.To.Name == "string"
			})
			builtin := pipe.Graph().Get(toBuiltin.Edge.To)
			Expect(builtin).ToNot(BeNil())
			Expect(builtin.Kind).To(Equal(common.SymKindBuiltin))
			Expect(builtin.Id.Name).To(Equal("string"))
		})
	})

	Context("ReturnsAnNestedAssignedAlias", func() {
		It("Processes a Nested Assigned Alias return value", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"AliasController",
				"ReturnsAnNestedAssignedAlias",
				[]string{},
			)

			Expect(info.Params).To(HaveLen(0))
			Expect(info.RetVals).To(HaveLen(2))

			var retTypeNode *symboldg.SymbolNode
			for _, retVal := range info.RetVals {
				node := utils.GetSingularChildTypeNode(pipe.Graph(), retVal.Node)
				if node != nil && node.Id.Name == "NestedAssignedAlias" {
					retTypeNode = node
					break
				}
			}

			Expect(retTypeNode).ToNot(BeNil())
			Expect(retTypeNode.Kind).To(Equal(common.SymKindAlias))

			aliasMeta := utils.MustAliasMeta(retTypeNode)
			Expect(aliasMeta.Name).To(Equal("NestedAssignedAlias"))
			Expect(aliasMeta.AliasType).To(Equal(metadata.AliasKindAssigned))

			// NestedAssignedAlias -> TypedefAlias -> string
			nestedEdges := common.MapValues(pipe.Graph().GetEdges(
				retTypeNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			toTypedef := linq.First(nestedEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == retTypeNode && edge.Edge.To.Name == typedefAliasName
			})
			typedefNode := pipe.Graph().Get(toTypedef.Edge.To)
			Expect(typedefNode).ToNot(BeNil())

			typedefEdges := common.MapValues(pipe.Graph().GetEdges(
				typedefNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))
			toBuiltin := linq.First(typedefEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == typedefNode && edge.Edge.To.Name == "string"
			})
			builtin := pipe.Graph().Get(toBuiltin.Edge.To)
			Expect(builtin).ToNot(BeNil())
			Expect(builtin.Kind).To(Equal(common.SymKindBuiltin))
			Expect(builtin.Id.Name).To(Equal("string"))
		})
	})

	Context("ReceivesATypedefSpecialAliasQuery", func() {
		It("Processes a Typedef-special alias parameter (TypedefSpecialAlias -> time.Time)", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"AliasController",
				"ReceivesATypedefSpecialAliasQuery",
				[]string{"alias"},
			)

			Expect(info.Params).To(HaveLen(1))
			Expect(info.RetVals).To(HaveLen(1))

			aliasParamNode := utils.GetSingularChildTypeNode(pipe.Graph(), info.Params[0].Node)
			Expect(aliasParamNode).ToNot(BeNil())
			Expect(aliasParamNode.Kind).To(Equal(common.SymKindAlias))

			aliasMeta := utils.MustAliasMeta(aliasParamNode)
			Expect(aliasMeta.Name).To(Equal("TypedefSpecialAlias"))
			Expect(aliasMeta.AliasType).To(Equal(metadata.AliasKindTypedef))

			// find outgoing type-edge from the alias that points to Time
			relevantEdges := common.MapValues(pipe.Graph().GetEdges(
				aliasParamNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))

			outgoingToTime := linq.First(relevantEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == aliasParamNode && edge.Edge.To.Name == "time.Time"
			})
			Expect(outgoingToTime).ToNot(BeNil())
			aliasedType := pipe.Graph().Get(outgoingToTime.Edge.To)

			Expect(aliasedType).ToNot(BeNil())
			Expect(aliasedType.Id.Name).To(Equal("time.Time"))
			Expect(aliasedType.Kind).To(Equal(common.SymKindSpecialBuiltin))
		})
	})

	Context("ReceivesAnAssignedSpecialAliasQuery", func() {
		It("Processes an Assigned-special alias parameter (AssignedSpecialAlias = time.Time)", func() {
			info := utils.GetApiEndpointHierarchy(
				pipe.Graph(),
				"AliasController",
				"ReceivesAnAssignedSpecialAliasQuery",
				[]string{"alias"},
			)

			Expect(info.Params).To(HaveLen(1))
			Expect(info.RetVals).To(HaveLen(1))

			aliasParamNode := utils.GetSingularChildTypeNode(pipe.Graph(), info.Params[0].Node)
			Expect(aliasParamNode).ToNot(BeNil())
			Expect(aliasParamNode.Kind).To(Equal(common.SymKindAlias))

			aliasMeta := utils.MustAliasMeta(aliasParamNode)
			Expect(aliasMeta.Name).To(Equal("AssignedSpecialAlias"))
			Expect(aliasMeta.AliasType).To(Equal(metadata.AliasKindAssigned))

			// find outgoing type-edge from the alias that points to Time
			relevantEdges := common.MapValues(pipe.Graph().GetEdges(
				aliasParamNode.Id,
				[]symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
			))

			outgoingToTime := linq.First(relevantEdges, func(edge symboldg.SymbolEdgeDescriptor) bool {
				return pipe.Graph().Get(edge.Edge.From) == aliasParamNode && edge.Edge.To.Name == "time.Time"
			})
			Expect(outgoingToTime).ToNot(BeNil())

			aliasedType := pipe.Graph().Get(outgoingToTime.Edge.To)

			Expect(aliasedType).ToNot(BeNil())
			Expect(aliasedType.Id.Name).To(Equal("time.Time"))
			Expect(aliasedType.Kind).To(Equal(common.SymKindSpecialBuiltin))
		})
	})

})

func TestAliasController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Alias Controller")
}
