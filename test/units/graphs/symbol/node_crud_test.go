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

	Context("AddController", func() {
		It("Adds a controller node successfully", func() {
			structMeta := metadata.StructMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Node:     utils.MakeIdent("MyController"),
					FVersion: fVersion,
				},
			}
			request := symboldg.CreateControllerNode{
				Data: metadata.ControllerMeta{
					Struct:    structMeta,
					Receivers: []metadata.ReceiverMeta{},
				},
				Annotations: nil,
			}

			node, err := graph.AddController(request)
			Expect(err).ToNot(HaveOccurred())
			Expect(node).ToNot(BeNil())
			Expect(node.Kind).To(Equal(common.SymKindController))
			Expect(node.Data).To(Equal(structMeta))
		})

		It("Returns an error when createAndAddSymNode returns an error", func() {
			// Pass a request with nil node to cause idempotencyGuard error
			structMeta := metadata.StructMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Node:     nil,
					FVersion: fVersion,
				},
			}
			request := symboldg.CreateControllerNode{
				Data: metadata.ControllerMeta{
					Struct:    structMeta,
					Receivers: []metadata.ReceiverMeta{},
				},
				Annotations: nil,
			}

			node, err := graph.AddController(request)
			Expect(err).To(HaveOccurred())
			Expect(node).To(BeNil())
		})
	})

	Context("AddRoute", func() {
		It("Adds a route and links it to its parent controller", func() {
			// Prepare parent controller node
			ctrlMeta := symboldg.KeyableNodeMeta{
				Decl:     utils.MakeIdent("MyController"),
				FVersion: *fVersion,
			}

			// Prepare route metadata
			routeMeta := &metadata.ReceiverMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:       "MyRoute",
					Node:       utils.MakeIdent("MyRoute"),
					SymbolKind: common.SymKindReceiver,
					PkgPath:    "example/pkg",
					FVersion:   fVersion,
				},
				Params:  []metadata.FuncParam{},
				RetVals: []metadata.FuncReturnValue{},
			}

			// Add the route
			routeNode, err := graph.AddRoute(symboldg.CreateRouteNode{
				Data:             routeMeta,
				ParentController: ctrlMeta,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(routeNode).ToNot(BeNil())

			// The parent controller should have the route as a child (EdgeKindReceiver)
			children := graph.Children(
				&symboldg.SymbolNode{Id: ctrlMeta.SymbolKey()},
				&symboldg.TraversalBehavior{
					Filtering: symboldg.TraversalFilter{
						EdgeKinds: []symboldg.SymbolEdgeKind{symboldg.EdgeKindReceiver},
					},
				})
			Expect(children).To(HaveLen(1))
			Expect(children[0].Id).To(Equal(routeNode.Id))
		})

		It("Returns an error when Data.Node is nil", func() {
			ctrlMeta := symboldg.KeyableNodeMeta{
				Decl:     utils.MakeIdent("MyController"),
				FVersion: *fVersion,
			}

			// Invalid ReceiverMeta: Node is nil -> triggers idempotencyGuard error
			badRouteMeta := &metadata.ReceiverMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:       "BadRoute",
					Node:       nil, // <- invalid to hit the error branch
					SymbolKind: common.SymKindReceiver,
					PkgPath:    "example/pkg",
					FVersion:   fVersion,
				},
			}

			routeNode, err := graph.AddRoute(symboldg.CreateRouteNode{
				Data:             badRouteMeta,
				ParentController: ctrlMeta,
			})

			Expect(err).To(HaveOccurred())
			Expect(routeNode).To(BeNil())
		})
	})

	Context("AddRouteParam", func() {
		var controllerMeta *metadata.ControllerMeta
		var receiverMeta *metadata.ReceiverMeta
		var routeNode *symboldg.SymbolNode

		BeforeEach(func() {
			// create a controller (parent for route)
			controllerMeta = &metadata.ControllerMeta{
				Struct: metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:     "MyController",
						Node:     utils.MakeIdent("MyController"),
						FVersion: fVersion,
					},
				},
			}
			_, err := graph.AddController(symboldg.CreateControllerNode{
				Data: *controllerMeta,
			})
			Expect(err).ToNot(HaveOccurred())

			// create a route (parent for params)
			receiverMeta = &metadata.ReceiverMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:     "MyRoute",
					Node:     utils.MakeIdent("MyRoute"),
					FVersion: fVersion,
				},
			}

			var err2 error
			routeNode, err2 = graph.AddRoute(symboldg.CreateRouteNode{
				Data: receiverMeta,
				ParentController: symboldg.KeyableNodeMeta{
					Decl:     controllerMeta.Struct.Node,
					FVersion: *fVersion,
				},
			})
			Expect(err2).ToNot(HaveOccurred())
		})

		It("Adds a parameter node and links it to the route", func() {
			paramMeta := metadata.FuncParam{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:     "MyParam",
					Node:     utils.MakeIdent("MyParam"),
					FVersion: fVersion,
				},
				Ordinal: 1,
				Type: metadata.TypeUsageMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:     "int",
						Node:     nil,
						FVersion: fVersion,
					},
				},
			}

			// Build a KeyableNodeMeta that points to the route we created above.
			// routeNode.Data is stored as the original *metadata.ReceiverMeta (not a SymNodeMeta),
			// so assert that concrete type before pulling out the inner AST node.
			receiverData, ok := routeNode.Data.(*metadata.ReceiverMeta)
			Expect(ok).To(BeTrue(), "expected routeNode.Data to be *metadata.ReceiverMeta")

			parentRoute := symboldg.KeyableNodeMeta{
				Decl:     receiverData.SymNodeMeta.Node,
				FVersion: *routeNode.Version,
			}

			paramNode, err := graph.AddRouteParam(symboldg.CreateParameterNode{
				Data:        paramMeta,
				ParentRoute: parentRoute,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(paramNode).ToNot(BeNil())

			// Route should have the param as a child via EdgeKindParam
			children := graph.Children(routeNode, &symboldg.TraversalBehavior{
				Filtering: symboldg.TraversalFilter{
					EdgeKinds: []symboldg.SymbolEdgeKind{symboldg.EdgeKindParam},
				},
			})
			Expect(children).To(HaveLen(1))
			Expect(children[0].Id).To(Equal(paramNode.Id))
		})

		It("Returns an error when parameter's Node is nil (idempotencyGuard error path)", func() {
			badParamMeta := metadata.FuncParam{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:     "BadParam",
					Node:     nil, // <- triggers idempotencyGuard error
					FVersion: fVersion,
				},
				Ordinal: 0,
			}

			receiverData, ok := routeNode.Data.(*metadata.ReceiverMeta)
			Expect(ok).To(BeTrue())

			parentRoute := symboldg.KeyableNodeMeta{
				Decl:     receiverData.SymNodeMeta.Node,
				FVersion: *routeNode.Version,
			}

			paramNode, err := graph.AddRouteParam(symboldg.CreateParameterNode{
				Data:        badParamMeta,
				ParentRoute: parentRoute,
			})
			Expect(err).To(HaveOccurred())
			Expect(paramNode).To(BeNil())
		})
	})

	Context("AddRouteRetVal", func() {
		var controllerMeta *metadata.ControllerMeta
		var receiverMeta *metadata.ReceiverMeta
		var routeNode *symboldg.SymbolNode

		BeforeEach(func() {
			// Create a controller (parent for the route)
			controllerMeta = &metadata.ControllerMeta{
				Struct: metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:     "MyController",
						Node:     utils.MakeIdent("MyController"),
						FVersion: fVersion,
					},
				},
				Receivers: nil,
			}
			_, err := graph.AddController(symboldg.CreateControllerNode{
				Data: *controllerMeta,
			})
			Expect(err).ToNot(HaveOccurred())

			// Create a route under that controller
			receiverMeta = &metadata.ReceiverMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:     "MyRoute",
					Node:     utils.MakeIdent("MyRoute"),
					FVersion: fVersion,
				},
			}

			var err2 error
			routeNode, err2 = graph.AddRoute(symboldg.CreateRouteNode{
				Data: receiverMeta,
				ParentController: symboldg.KeyableNodeMeta{
					Decl:     controllerMeta.Struct.Node,
					FVersion: *fVersion,
				},
			})
			Expect(err2).ToNot(HaveOccurred())
		})

		It("Adds a return value node and links it to the route", func() {
			retValMeta := metadata.FuncReturnValue{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:     "MyRetVal",
					Node:     utils.MakeIdent("MyRetVal"),
					FVersion: fVersion,
				},
				Ordinal: 1,
				Type: metadata.TypeUsageMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:     "string",
						Node:     nil,
						FVersion: fVersion,
					},
				},
			}

			receiverData, ok := routeNode.Data.(*metadata.ReceiverMeta)
			Expect(ok).To(BeTrue(), "expected routeNode.Data to be *metadata.ReceiverMeta")

			parentRoute := symboldg.KeyableNodeMeta{
				Decl:     receiverData.SymNodeMeta.Node,
				FVersion: *routeNode.Version,
			}

			retValNode, err := graph.AddRouteRetVal(symboldg.CreateReturnValueNode{
				Data:        retValMeta,
				ParentRoute: parentRoute,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(retValNode).ToNot(BeNil())

			children := graph.Children(routeNode, &symboldg.TraversalBehavior{
				Filtering: symboldg.TraversalFilter{
					EdgeKinds: []symboldg.SymbolEdgeKind{symboldg.EdgeKindRetVal},
				},
			})
			Expect(children).To(HaveLen(1))
			Expect(children[0].Id).To(Equal(retValNode.Id))
		})

		It("Returns an error when return value's Node is nil (idempotencyGuard error path)", func() {
			badRetValMeta := metadata.FuncReturnValue{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:     "BadRetVal",
					Node:     nil, // triggers idempotencyGuard error
					FVersion: fVersion,
				},
				Ordinal: 0,
			}

			receiverData, ok := routeNode.Data.(*metadata.ReceiverMeta)
			Expect(ok).To(BeTrue())

			parentRoute := symboldg.KeyableNodeMeta{
				Decl:     receiverData.SymNodeMeta.Node,
				FVersion: *routeNode.Version,
			}

			retValNode, err := graph.AddRouteRetVal(symboldg.CreateReturnValueNode{
				Data:        badRetValMeta,
				ParentRoute: parentRoute,
			})
			Expect(err).To(HaveOccurred())
			Expect(retValNode).To(BeNil())
		})
	})

	Context("AddStruct", func() {
		It("Registers fields as edges when creating a struct", func() {
			fieldMeta := metadata.FieldMeta{
				SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("FieldA"), FVersion: fVersion},
				Type: metadata.TypeUsageMeta{
					SymNodeMeta: metadata.SymNodeMeta{Name: "int", FVersion: fVersion},
				},
			}

			structMeta := metadata.StructMeta{
				SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("MyStruct"), FVersion: fVersion},
				Fields:      []metadata.FieldMeta{fieldMeta},
			}

			structNode, err := graph.AddStruct(symboldg.CreateStructNode{Data: structMeta})
			Expect(err).ToNot(HaveOccurred())

			// Edge to field exists, but field node not added yet
			children := graph.Children(structNode, nil)
			Expect(children).To(BeEmpty())
		})

		It("Returns an error when struct Node is nil", func() {
			badMeta := metadata.StructMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:     "BadStruct",
					Node:     nil, // triggers idempotencyGuard error
					FVersion: fVersion,
				},
				Fields: nil,
			}

			structNode, err := graph.AddStruct(symboldg.CreateStructNode{
				Data:        badMeta,
				Annotations: nil,
			})
			Expect(err).To(HaveOccurred())
			Expect(structNode).To(BeNil())
		})
	})

	Context("AddEnum", func() {
		var enumMeta metadata.EnumMeta

		BeforeEach(func() {
			enumMeta = metadata.EnumMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:     "MyEnum",
					Node:     utils.MakeIdent("MyEnum"),
					FVersion: fVersion,
				},
				ValueKind: metadata.EnumValueKindString,
				Values: []metadata.EnumValueDefinition{
					{
						SymNodeMeta: metadata.SymNodeMeta{
							Name:     "ValueOne",
							Node:     utils.MakeIdent("ValueOne"),
							FVersion: fVersion,
						},
						Value: "First",
					},
					{
						SymNodeMeta: metadata.SymNodeMeta{
							Name:     "ValueTwo",
							Node:     utils.MakeIdent("ValueTwo"),
							FVersion: fVersion,
						},
						Value: "Second",
					},
				},
			}
		})

		It("Adds an enum node and links its values and underlying type", func() {
			enumNode, err := graph.AddEnum(symboldg.CreateEnumNode{
				Data:        enumMeta,
				Annotations: nil,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(enumNode).ToNot(BeNil())

			// Enum should have Value edges to its constants
			valueChildren := graph.Children(enumNode, &symboldg.TraversalBehavior{
				Filtering: symboldg.TraversalFilter{
					EdgeKinds: []symboldg.SymbolEdgeKind{symboldg.EdgeKindValue},
				},
			})
			Expect(valueChildren).To(HaveLen(len(enumMeta.Values)))

			// Each value should also have a Reference edge to the underlying type
			valueSymKey := graphs.NewUniverseSymbolKey(string(enumMeta.ValueKind))
			for _, valueChild := range valueChildren {
				referenceEdges := graph.Children(valueChild, &symboldg.TraversalBehavior{
					Filtering: symboldg.TraversalFilter{
						EdgeKinds: []symboldg.SymbolEdgeKind{symboldg.EdgeKindReference},
					},
				})
				Expect(referenceEdges).To(HaveLen(1))
				Expect(referenceEdges[0].Id).To(Equal(valueSymKey))
			}
		})

		It("Returns an error when enum Node is nil", func() {
			badMeta := metadata.EnumMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:     "BadEnum",
					Node:     nil, // triggers createAndAddSymNode error
					FVersion: fVersion,
				},
				ValueKind: metadata.EnumValueKindInt,
				Values:    nil,
			}

			enumNode, err := graph.AddEnum(symboldg.CreateEnumNode{
				Data:        badMeta,
				Annotations: nil,
			})
			Expect(err).To(HaveOccurred())
			Expect(enumNode).To(BeNil())
		})

		It("Returns an error when creating an enum value fails", func() {
			badValuesMeta := enumMeta
			badValuesMeta.Values = []metadata.EnumValueDefinition{
				enumMeta.Values[0],
				{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:     "BadValue",
						Node:     nil, // triggers createAndAddSymNode error for this value
						FVersion: fVersion,
					},
					Value: "Broken",
				},
			}

			enumNode, err := graph.AddEnum(symboldg.CreateEnumNode{
				Data:        badValuesMeta,
				Annotations: nil,
			})
			Expect(err).To(HaveOccurred())
			Expect(enumNode).To(BeNil())
		})

		It("Returns an error when an enum references a non-primitive type that does not exist in the graph", func() {
			badValueKind := enumMeta
			badValueKind.ValueKind = "NonPrimitiveType"

			enumNode, err := graph.AddEnum(symboldg.CreateEnumNode{
				Data:        badValueKind,
				Annotations: nil,
			})

			Expect(err).To(MatchError(
				ContainSubstring("value kind for enum 'MyEnum' is 'NonPrimitiveType' which is unexpected")),
			)
			Expect(enumNode).To(BeNil())
		})
	})

	Context("AddField", func() {
		var fieldMeta metadata.FieldMeta
		var baseTypeKey graphs.SymbolKey

		BeforeEach(func() {
			baseTypeKey = graphs.NewUniverseSymbolKey("float32")

			fieldMeta = metadata.FieldMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:     "MyField",
					Node:     utils.MakeIdent("MyField"),
					FVersion: fVersion,
				},
				Type: metadata.TypeUsageMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:     "float32",
						Node:     nil,
						FVersion: fVersion,
					},
					Import: common.ImportTypeNone,
				},
				IsEmbedded: false,
			}
		})

		It("Adds a field node and links it to its base type", func() {
			fieldNode, err := graph.AddField(symboldg.CreateFieldNode{
				Data:        fieldMeta,
				Annotations: nil,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(fieldNode).ToNot(BeNil())

			// Add the type node so links actually exist
			graph.AddPrimitive(symboldg.PrimitiveTypeFloat32)

			children := graph.Children(fieldNode, &symboldg.TraversalBehavior{
				Filtering: symboldg.TraversalFilter{
					EdgeKinds: []symboldg.SymbolEdgeKind{symboldg.EdgeKindType},
				},
			})
			Expect(children).To(HaveLen(1))
			Expect(children[0].Id).To(Equal(baseTypeKey))
		})

		It("Returns an error when the field Node is nil", func() {
			badMeta := fieldMeta
			badMeta.Node = nil // Should cause createAndAddSymNode to error

			fieldNode, err := graph.AddField(symboldg.CreateFieldNode{
				Data:        badMeta,
				Annotations: nil,
			})
			Expect(err).To(HaveOccurred())
			Expect(fieldNode).To(BeNil())
		})

		It("Returns an error when getTypeRef fails due to missing base type", func() {
			badMeta := fieldMeta
			// badMeta.Type.Layers = nil // no base layer → getTypeRef error

			fieldNode, err := graph.AddField(symboldg.CreateFieldNode{
				Data:        badMeta,
				Annotations: nil,
			})
			Expect(err).To(HaveOccurred())
			Expect(fieldNode).To(BeNil())
		})
	})

	Context("AddConst", func() {
		It("Adds a const without error", func() {
			node, err := graph.AddConst(symboldg.CreateConstNode{
				Data: metadata.ConstMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name:     "SomeConst",
						Node:     utils.MakeIdent("SomeConst"),
						FVersion: fVersion,
					},
					Value: "some value",
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{Name: "string", FVersion: fVersion},
					},
				},
			})

			Expect(err).To(BeNil())
			Expect(node).ToNot(BeNil())
			Expect(node.Id.Name).To(Equal("SomeConst"))
			Expect(node.Kind).To(Equal(common.SymKindConstant))
		})
	})

	Context("AddPrimitive", func() {
		It("Adds a primitive and returns a non-nil node", func() {
			graph := symboldg.NewSymbolGraph()
			node := graph.AddPrimitive(symboldg.PrimitiveTypeBool)
			Expect(node).ToNot(BeNil())
		})

		It("Returns the same node when adding a duplicate primitive", func() {
			graph := symboldg.NewSymbolGraph()
			n1 := graph.AddPrimitive(symboldg.PrimitiveTypeBool)
			n2 := graph.AddPrimitive(symboldg.PrimitiveTypeBool)
			Expect(n2).To(Equal(n1))
		})
	})

	Context("AddSpecial", func() {
		It("Adds a special type and returns a non-nil node", func() {
			graph := symboldg.NewSymbolGraph()
			node := graph.AddSpecial(symboldg.SpecialTypeError)
			Expect(node).ToNot(BeNil())
		})
	})

	Context("RemoveNode", func() {
		It("Is a no-op when the target node does not exist", func() {
			// create a struct so the graph isn't empty
			parent, err := graph.AddStruct(symboldg.CreateStructNode{
				Data: metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("ParentNoop"), FVersion: fVersion},
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(parent).ToNot(BeNil())

			// remove a non-existent key
			fakeKey := graphs.NewSymbolKey(utils.MakeIdent("does-not-exist"), fVersion)
			graph.RemoveNode(fakeKey) // should be a no-op

			// the original node should still be present
			found := false
			for _, s := range graph.FindByKind(common.SymKindStruct) {
				if s.Id.Equals(parent.Id) {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("Removes a node that has no dependents", func() {
			n, err := graph.AddStruct(symboldg.CreateStructNode{
				Data: metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("Lonely"), FVersion: fVersion},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// Ensure it exists
			Expect(graph.FindByKind(common.SymKindStruct)).To(ContainElement(n))

			// Remove it
			graph.RemoveNode(n.Id)

			// It should be gone
			results := graph.FindByKind(common.SymKindStruct)
			for _, r := range results {
				Expect(r.Id).ToNot(Equal(n.Id))
			}
		})

		It("Does not evict a dependent that still has other outgoing dependencies", func() {
			// Setup:
			// dependent -> target
			// dependent -> other (so dependent should NOT be orphaned when target removed)
			target, err := graph.AddField(symboldg.CreateFieldNode{
				Data: metadata.FieldMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("TargetField"), FVersion: fVersion},
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{FVersion: fVersion},
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			dep, err := graph.AddStruct(symboldg.CreateStructNode{
				Data: metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("DepStruct"), FVersion: fVersion},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			other := graph.AddPrimitive(symboldg.PrimitiveTypeBool) // another node for dep to point to
			Expect(other).ToNot(BeNil())

			// Add dependent->target and dependent->other
			graph.AddEdge(dep.Id, target.Id, symboldg.EdgeKindField, nil)    // dep -> target
			graph.AddEdge(dep.Id, other.Id, symboldg.EdgeKindReference, nil) // dep -> other

			// Remove the target
			graph.RemoveNode(target.Id)

			// Dependent should still exist (it still points to `other`)
			found := false
			for _, s := range graph.FindByKind(common.SymKindStruct) {
				if s.Id.Equals(dep.Id) {
					found = true
				}
			}
			Expect(found).To(BeTrue())

			// And its children should still include 'other'
			children := graph.Children(dep, nil)
			Expect(children).To(ContainElement(other))
		})

		It("Recursively evicts dependents that become orphaned", func() {
			// Build a chain: A <- B <- C (where "<-" means "dep -> target")
			// We'll remove A and expect B and C to be removed as well if they have no other deps.

			// A (target)
			a, err := graph.AddField(symboldg.CreateFieldNode{
				Data: metadata.FieldMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("A_field"), FVersion: fVersion},
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{FVersion: fVersion},
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// B depends on A
			b, err := graph.AddStruct(symboldg.CreateStructNode{
				Data: metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("B_struct"), FVersion: fVersion},
				},
			})
			Expect(err).ToNot(HaveOccurred())
			graph.AddEdge(b.Id, a.Id, symboldg.EdgeKindField, nil)

			// C depends on B
			c, err := graph.AddStruct(symboldg.CreateStructNode{
				Data: metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("C_struct"), FVersion: fVersion},
				},
			})
			Expect(err).ToNot(HaveOccurred())
			graph.AddEdge(c.Id, b.Id, symboldg.EdgeKindField, nil)

			// Sanity: B and C exist
			Expect(graph.FindByKind(common.SymKindStruct)).To(ContainElement(b))
			Expect(graph.FindByKind(common.SymKindStruct)).To(ContainElement(c))

			// Remove A -> should recursively evict B and then C
			graph.RemoveNode(a.Id)

			// Neither B nor C should be present now
			structs := graph.FindByKind(common.SymKindStruct)
			Expect(structs).ToNot(ContainElement(b))
			Expect(structs).ToNot(ContainElement(c))

			// Also children of C/B shouldn't exist; Descendants from a (if a still existed) would be empty
		})

		It("Handles cycles without infinite recursion (mutual deps are removed appropriately)", func() {
			// Create two nodes A and B with mutual edges A->B and B->A
			a, err := graph.AddStruct(symboldg.CreateStructNode{
				Data: metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("CycleA"), FVersion: fVersion},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			b, err := graph.AddStruct(symboldg.CreateStructNode{
				Data: metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("CycleB"), FVersion: fVersion},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// A depends on B and B depends on A
			graph.AddEdge(a.Id, b.Id, symboldg.EdgeKindReference, nil)
			graph.AddEdge(b.Id, a.Id, symboldg.EdgeKindReference, nil)

			// Removing A should not hang; it should remove both if they become orphaned.
			graph.RemoveNode(a.Id)

			// After removal, neither node should be present
			structs := graph.FindByKind(common.SymKindStruct)
			Expect(structs).ToNot(ContainElement(a))
			Expect(structs).ToNot(ContainElement(b))
		})

		It("Cleans up edges/indices so no stale references remain", func() {
			// Create structure: Parent -> Child, and also Parent -> SomePrimitive
			child, err := graph.AddField(symboldg.CreateFieldNode{
				Data: metadata.FieldMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("CleanupChild"), FVersion: fVersion},
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{FVersion: fVersion},
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			parent, err := graph.AddStruct(symboldg.CreateStructNode{
				Data: metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("CleanupParent"), FVersion: fVersion},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			p := graph.AddPrimitive(symboldg.PrimitiveTypeInt)
			Expect(p).ToNot(BeNil())

			// Parent -> Child and Parent -> p
			graph.AddEdge(parent.Id, child.Id, symboldg.EdgeKindField, nil)
			graph.AddEdge(parent.Id, p.Id, symboldg.EdgeKindReference, nil)

			// Remove the child node
			graph.RemoveNode(child.Id)

			// Parent should still exist and still have the primitive child
			Expect(graph.FindByKind(common.SymKindStruct)).To(ContainElement(parent))
			children := graph.Children(parent, nil)
			Expect(children).To(ContainElement(p))

			// And asking for parents of the removed node should return empty slice (no stale revDeps)
			// We can't call Parents on the removed node (node object is gone). Instead confirm that
			// no node claims the removed child's Id in its children
			foundAnyParent := false
			for _, n := range graph.FindByKind(common.SymKindStruct) {
				for _, ch := range graph.Children(n, nil) {
					if ch.Id.Equals(child.Id) {
						foundAnyParent = true
					}
				}
			}
			Expect(foundAnyParent).To(BeFalse())
		})
	})
})
