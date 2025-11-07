package symbol_test

import (
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - SymbolGraph", func() {
	var graph symboldg.SymbolGraph
	var fVersion *gast.FileVersion

	BeforeEach(func() {
		graph = symboldg.NewSymbolGraph()
		fVersion = utils.MakeFileVersion("file", "")
	})

	Context("AddEdge", func() {
		var (
			structNode *symboldg.SymbolNode
			typeNode   *symboldg.SymbolNode
			err        error
		)

		BeforeEach(func() {
			// Create a struct node
			fv := utils.MakeFileVersion("struct1", "")
			structMeta := metadata.StructMeta{
				SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("MyStruct"), FVersion: fv},
			}
			structNode, err = graph.AddStruct(symboldg.CreateStructNode{
				Data: structMeta,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(structNode).ToNot(BeNil())

			// Create a builtin/type node
			typeNode = graph.AddPrimitive(symboldg.PrimitiveTypeInt)
			Expect(typeNode).ToNot(BeNil())
		})

		It("Adds an edge from struct to typeNode", func() {
			graph.AddEdge(structNode.Id, typeNode.Id, symboldg.EdgeKindType, nil)

			children := graph.Children(structNode, nil)
			Expect(children).To(HaveLen(1))
			Expect(children[0].Id).To(Equal(typeNode.Id))
		})

		It("Does not add duplicate edges", func() {
			graph.AddEdge(structNode.Id, typeNode.Id, symboldg.EdgeKindType, nil)
			graph.AddEdge(structNode.Id, typeNode.Id, symboldg.EdgeKindType, nil) // duplicate

			children := graph.Children(structNode, nil)
			Expect(children).To(HaveLen(1))
		})
	})

	Context("Children", func() {
		var structNode *symboldg.SymbolNode
		var fieldNode *symboldg.SymbolNode

		BeforeEach(func() {
			// Build: Struct -> Field
			structMeta := metadata.StructMeta{
				SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("Parent"), FVersion: fVersion},
			}
			structNode, _ = graph.AddStruct(symboldg.CreateStructNode{Data: structMeta})

			fieldMeta := metadata.FieldMeta{
				SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("Child"), FVersion: fVersion},
				Type: metadata.TypeUsageMeta{
					SymNodeMeta: metadata.SymNodeMeta{Name: "int", FVersion: fVersion},
				},
			}
			fieldNode, _ = graph.AddField(symboldg.CreateFieldNode{Data: fieldMeta})

			graph.AddEdge(structNode.Id, fieldNode.Id, symboldg.EdgeKindField, nil)
		})

		When("Unsorted", func() {
			It("Returns only children matching the filter", func() {
				filter := &symboldg.TraversalBehavior{
					Filtering: symboldg.TraversalFilter{
						EdgeKinds: []symboldg.SymbolEdgeKind{symboldg.EdgeKindField},
					},
				}
				result := graph.Children(structNode, filter)
				Expect(result).To(HaveLen(1))
				Expect(result[0].Id).To(Equal(fieldNode.Id))
			})

			It("Skips edges that do not match a given filter", func() {
				filter := &symboldg.TraversalBehavior{
					Filtering: symboldg.TraversalFilter{
						EdgeKinds: []symboldg.SymbolEdgeKind{symboldg.EdgeKindCall}, // This should drop the edge
					},
				}
				result := graph.Children(structNode, filter)
				Expect(result).To(HaveLen(0))
			})

			It("Returns only matching nodes when a filter function is given", func() {
				filter := &symboldg.TraversalBehavior{
					Filtering: symboldg.TraversalFilter{
						FilterFunc: func(node *symboldg.SymbolNode) bool {
							return node.Id.Name == "Child"
						},
					},
				}

				// Add another field under the struct to verify filter works as expected
				fieldMeta := metadata.FieldMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("AnotherChild"), FVersion: fVersion},
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{Name: "int", FVersion: fVersion},
					},
				}
				fieldNode, _ = graph.AddField(symboldg.CreateFieldNode{Data: fieldMeta})
				graph.AddEdge(structNode.Id, fieldNode.Id, symboldg.EdgeKindField, nil)

				result := graph.Children(structNode, filter)
				Expect(result).To(HaveLen(1))
				Expect(result[0].Kind).To(Equal(common.SymKindField))
				Expect(result[0].Id.Name).To(Equal("Child"))
			})
		})

		When("Sorted", func() {
			var timeNode *symboldg.SymbolNode
			var anyNode *symboldg.SymbolNode

			BeforeEach(func() {
				// The graph itself doesn't verify semantics so we can use this un-real linkage
				// to test filtering logic
				primNode := graph.AddPrimitive(symboldg.PrimitiveTypeString)
				graph.AddEdge(structNode.Id, primNode.Id, symboldg.EdgeKindType, nil)

				timeNode = graph.AddSpecial(symboldg.SpecialTypeTime)
				graph.AddEdge(structNode.Id, timeNode.Id, symboldg.EdgeKindField, nil)

				anyNode = graph.AddSpecial(symboldg.SpecialTypeAny)
				graph.AddEdge(structNode.Id, anyNode.Id, symboldg.EdgeKindField, nil)

			})

			It("Skips edges that do not match the given filter", func() {
				filter := &symboldg.TraversalBehavior{
					Sorting: symboldg.TraversalSortingOrdinalDesc,
					Filtering: symboldg.TraversalFilter{
						EdgeKinds: []symboldg.SymbolEdgeKind{symboldg.EdgeKindCall},
					},
				}
				result := graph.Children(structNode, filter)
				Expect(result).To(HaveLen(0))
			})

			It("Skips nodes that do not match the given filter", func() {
				filter := &symboldg.TraversalBehavior{
					Sorting: symboldg.TraversalSortingOrdinalDesc,
					Filtering: symboldg.TraversalFilter{
						EdgeKinds: []symboldg.SymbolEdgeKind{symboldg.EdgeKindField},
						NodeKinds: []common.SymKind{common.SymKindSpecialBuiltin},
					},
				}
				result := graph.Children(structNode, filter)
				Expect(result).To(HaveLen(2))
				// Order should be DESC here - time node was inserted first and the any node second
				// so we expect a reverse order - any -> time
				Expect(result[0].Id).To(Equal(anyNode.Id))
				Expect(result[1].Id).To(Equal(timeNode.Id))
			})
		})
	})

	Context("Parents", func() {
		var structNode *symboldg.SymbolNode
		var fieldNode *symboldg.SymbolNode
		var otherNode *symboldg.SymbolNode

		BeforeEach(func() {
			structNode, _ = graph.AddStruct(symboldg.CreateStructNode{
				Data: metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("Parent"), FVersion: fVersion},
				},
			})
			fieldNode, _ = graph.AddField(symboldg.CreateFieldNode{
				Data: metadata.FieldMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("Child"), FVersion: fVersion},
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{Name: "int", FVersion: fVersion},
					},
				},
			})

			otherNode, _ = graph.AddField(symboldg.CreateFieldNode{
				Data: metadata.FieldMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("OtherChild"), FVersion: fVersion},
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{Name: "string", FVersion: fVersion},
					},
				},
			})

			// Add a valid edge from structNode to fieldNode
			graph.AddEdge(structNode.Id, fieldNode.Id, symboldg.EdgeKindField, nil)

			// Add an edge from structNode to otherNode (for later tests)
			graph.AddEdge(structNode.Id, otherNode.Id, symboldg.EdgeKindField, nil)
		})

		When("Unsorted", func() {
			It("Returns parent nodes correctly", func() {
				result := graph.Parents(fieldNode, nil)
				Expect(result).To(HaveLen(1))
				Expect(result[0].Id).To(Equal(structNode.Id))
			})

			It("Skips edges pointing to different nodes", func() {
				// Here we pass otherNode to Parents but edges from structNode point to fieldNode and otherNode
				// The edge to fieldNode should be skipped because edge.To != otherNode.Id
				result := graph.Parents(fieldNode, nil)
				Expect(result).To(HaveLen(1))
				Expect(result[0].Id).To(Equal(structNode.Id))

				// Now call Parents on otherNode, should return parent structNode because edge exists
				resultOther := graph.Parents(otherNode, nil)
				Expect(resultOther).To(HaveLen(1))
				Expect(resultOther[0].Id).To(Equal(structNode.Id))

				// To trigger continue edge.To != node.Id, create a fake node with no incoming edges
				fakeNode := &symboldg.SymbolNode{
					Id: graphs.NewUniverseSymbolKey("fakeNode"),
				}
				resultFake := graph.Parents(fakeNode, nil)
				Expect(resultFake).To(HaveLen(0)) // No parents, so effectively edges skipped
			})

			It("Skips edges excluded by EdgeKind filter", func() {
				// The existing edge kind is EdgeKindField
				// We set filter EdgeKind to a different kind, so shouldIncludeEdge returns false
				filter := &symboldg.TraversalBehavior{
					Filtering: symboldg.TraversalFilter{
						EdgeKinds: []symboldg.SymbolEdgeKind{symboldg.EdgeKindValue}, // deliberately a different edge kind
					},
				}

				result := graph.Parents(fieldNode, filter)
				Expect(result).To(HaveLen(0))
			})
		})

		When("Sorted", func() {
			var strNode *symboldg.SymbolNode
			BeforeEach(func() {
				strNode = graph.AddPrimitive(symboldg.PrimitiveTypeString)
				// As before this is *not* a valid tree but as the graph does not validate
				// semantics we can use this for testing
				graph.AddEdge(strNode.Id, fieldNode.Id, symboldg.EdgeKindType, nil)
			})

			It("Returns parent nodes correctly", func() {
				result := graph.Parents(
					fieldNode,
					&symboldg.TraversalBehavior{
						Sorting: symboldg.TraversalSortingOrdinalAsc,
					},
				)
				Expect(result).To(HaveLen(2))
				// Note we've passed a sort ASC here so we expect this specific order
				Expect(result[0].Id).To(Equal(structNode.Id))
				Expect(result[1].Id).To(Equal(strNode.Id))
			})

			It("Skips edges pointing to different nodes", func() {
				behavior := &symboldg.TraversalBehavior{
					Sorting: symboldg.TraversalSortingOrdinalDesc,
				}

				// Here we pass otherNode to Parents but edges from structNode point to fieldNode and otherNode
				// The edge to fieldNode should be skipped because edge.To != otherNode.Id
				result := graph.Parents(fieldNode, behavior)
				Expect(result).To(HaveLen(2))
				// Since we've passed a DESC order, we expect the SECOND node to be the 'struct' node
				Expect(result[1].Id).To(Equal(structNode.Id))

				// Now call Parents on otherNode, should return parent structNode because edge exists
				resultOther := graph.Parents(otherNode, behavior)
				Expect(resultOther).To(HaveLen(1))
				Expect(resultOther[0].Id).To(Equal(structNode.Id))

				// To trigger continue edge.To != node.Id, create a fake node with no incoming edges
				fakeNode := &symboldg.SymbolNode{
					Id: graphs.NewUniverseSymbolKey("fakeNode"),
				}
				resultFake := graph.Parents(fakeNode, nil)
				Expect(resultFake).To(HaveLen(0)) // No parents, so effectively edges skipped
			})

			It("Skips edges excluded by EdgeKind filter", func() {
				// The existing edge kind is EdgeKindField
				// We set filter EdgeKind to a different kind, so shouldIncludeEdge returns false
				filter := &symboldg.TraversalBehavior{
					Sorting: symboldg.TraversalSortingOrdinalAsc,
					Filtering: symboldg.TraversalFilter{
						EdgeKinds: []symboldg.SymbolEdgeKind{symboldg.EdgeKindValue}, // deliberately a different edge kind
					},
				}

				result := graph.Parents(fieldNode, filter)
				Expect(result).To(HaveLen(0))
			})

		})

	})

	Context("Descendants", func() {
		var structNode, childNode, grandChildNode *symboldg.SymbolNode

		BeforeEach(func() {
			structNode, _ = graph.AddStruct(symboldg.CreateStructNode{
				Data: metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("GrandParent"), FVersion: fVersion},
				},
			})
			childNode, _ = graph.AddField(symboldg.CreateFieldNode{
				Data: metadata.FieldMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("Child"), FVersion: fVersion},
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{Name: "int", FVersion: fVersion},
					},
				},
			})
			grandChildNode, _ = graph.AddField(symboldg.CreateFieldNode{
				Data: metadata.FieldMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("GrandChild"), FVersion: fVersion},
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{Name: "int", FVersion: fVersion},
					},
				},
			})

			graph.AddEdge(structNode.Id, childNode.Id, symboldg.EdgeKindField, nil)
			graph.AddEdge(childNode.Id, grandChildNode.Id, symboldg.EdgeKindField, nil)
		})

		It("Recursively traverses all descendants", func() {
			result := graph.Descendants(structNode, nil)
			Expect(result).To(HaveLen(2))
			Expect(result).To(ContainElements(childNode, grandChildNode))
		})

		It("Does not revisit already visited nodes", func() {
			// Add a cyclic edge to test visited map
			graph.AddEdge(grandChildNode.Id, childNode.Id, symboldg.EdgeKindField, nil)

			result := graph.Descendants(structNode, nil)
			// Should still only include each node once
			Expect(result).To(HaveLen(2))
			Expect(result).To(ContainElements(childNode, grandChildNode))
		})

		It("Respects the filter to exclude some children", func() {
			// Add a node we can refer to
			errNode := graph.AddSpecial("error")
			graph.AddEdge(structNode.Id, errNode.Id, symboldg.EdgeKindReference, nil)

			// Create a filter that excludes the grandChildNode by NodeKind mismatch
			filter := &symboldg.TraversalBehavior{
				Filtering: symboldg.TraversalFilter{
					NodeKinds: []common.SymKind{common.SymKindSpecialBuiltin},
				},
			}

			result := graph.Descendants(structNode, filter)
			// We expect only the 'error' node here as the only filter match
			Expect(result).To(HaveLen(1))
			Expect(result[0].Id).To(Equal(errNode.Id))
		})
	})
})
