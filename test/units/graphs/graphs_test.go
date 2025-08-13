package graphs_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"time"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fakeNode struct{}

func (f fakeNode) Pos() token.Pos { return token.Pos(42) }
func (f fakeNode) End() token.Pos { return token.Pos(99) }

var _ = Describe("Unit Tests - Graphs", func() {

	var _ = Describe("SymbolKey", func() {
		Describe("Id", func() {
			It("returns the correct ID for universe types", func() {
				key := graphs.NewUniverseSymbolKey("string")
				Expect(key.Id()).To(Equal("UniverseType:string"))
			})

			It("returns the correct ID for named symbol", func() {
				key := graphs.SymbolKey{
					Name:     "Foo",
					Position: 123,
					FileId:   "path/to/file|mod|abc123",
				}
				Expect(key.Id()).To(Equal("Foo@123@path/to/file|mod|abc123"))
			})

			It("returns the correct ID for unnamed symbol", func() {
				key := graphs.SymbolKey{
					Position: 456,
					FileId:   "path/to/file|mod|xyz999",
				}
				Expect(key.Id()).To(Equal("@456@path/to/file|mod|xyz999"))
			})
		})

		Describe("ShortLabel", func() {
			It("includes base file name and short hash", func() {
				key := graphs.SymbolKey{
					Name:   "Bar",
					FileId: "path/to/thing.go|mymod|deadbeef1234",
				}
				Expect(key.ShortLabel()).To(Equal("Bar@thing.go|deadbee"))
			})

			It("handles missing name", func() {
				key := graphs.SymbolKey{
					FileId: "some/path.go|mod|abcdef1234567890",
				}
				Expect(key.ShortLabel()).To(Equal("path.go|abcdef1"))
			})
		})

		Describe("PrettyPrint", func() {
			It("Prints cleanly for universe types", func() {
				key := graphs.NewUniverseSymbolKey("UniverseType:string")
				Expect(key.PrettyPrint()).To(Equal("string"))
			})

			It("Prints details for named symbol", func() {
				key := graphs.SymbolKey{
					Name:     "MyFunc",
					Position: 789,
					FileId:   "src/main.go|mod|abcd1234",
				}
				result := key.PrettyPrint()
				Expect(result).To(ContainSubstring("MyFunc"))
				Expect(result).To(ContainSubstring("• src/main.go"))
				Expect(result).To(ContainSubstring("• mod"))
				Expect(result).To(ContainSubstring("• abcd1234"))
			})

			It("Pretty prints a symbol with no name using position fallback", func() {
				fSet := token.NewFileSet()
				src := `package main; var _ = 123`

				file, err := parser.ParseFile(fSet, "dummy.go", src, parser.AllErrors)
				Expect(err).To(BeNil())

				// Use the whole file node — not one of the recognized node types
				version := &gast.FileVersion{
					Path:    "dummy.go",
					ModTime: time.Unix(1234, 0),
					Hash:    "abcd1234",
				}

				// Trigger default case in NewSymbolKey and missing name fallback
				sk := graphs.NewSymbolKey(file, version)

				Expect(sk.Name).To(Equal("")) // Should trigger fallback
				Expect(sk.Position).ToNot(Equal(token.NoPos))

				output := sk.PrettyPrint()
				Expect(output).To(HavePrefix(fmt.Sprintf("@%d\n", sk.Position)))
				Expect(output).To(ContainSubstring("• dummy.go"))
				Expect(output).To(ContainSubstring("• abcd1234"))
			})

		})

		Describe("Equals", func() {
			It("returns true for equal universe keys", func() {
				a := graphs.NewUniverseSymbolKey("bool")
				b := graphs.NewUniverseSymbolKey("bool")
				Expect(a.Equals(b)).To(BeTrue())
			})

			It("returns false for unequal universe keys", func() {
				a := graphs.NewUniverseSymbolKey("string")
				b := graphs.NewUniverseSymbolKey("int")
				Expect(a.Equals(b)).To(BeFalse())
			})

			It("returns true for equal non-universe keys", func() {
				a := graphs.SymbolKey{Name: "Foo", Position: 1, FileId: "id"}
				b := graphs.SymbolKey{Name: "Foo", Position: 1, FileId: "id"}
				Expect(a.Equals(b)).To(BeTrue())
			})

			It("returns false if one field differs", func() {
				a := graphs.SymbolKey{Name: "Foo", Position: 1, FileId: "id"}
				b := graphs.SymbolKey{Name: "Foo", Position: 2, FileId: "id"}
				Expect(a.Equals(b)).To(BeFalse())
			})
		})

		Describe("NewSymbolKey", func() {
			var version *gast.FileVersion

			BeforeEach(func() {
				version = &gast.FileVersion{
					Path:    "/abs/path/to/file.go",
					ModTime: time.Now(),
					Hash:    "cafebabe12345678",
				}
			})

			It("Returns empty key on nil inputs", func() {
				key := graphs.NewSymbolKey(nil, nil)
				Expect(key).To(Equal(graphs.SymbolKey{}))
			})

			It("Creates from a real FuncDecl AST with valid token.Pos", func() {
				fSet := token.NewFileSet()
				src := `package main; func MyFunc() {}`

				file, err := parser.ParseFile(fSet, "fake.go", src, parser.AllErrors)
				Expect(err).To(BeNil())

				var fn *ast.FuncDecl
				for _, decl := range file.Decls {
					if fd, ok := decl.(*ast.FuncDecl); ok {
						fn = fd
						break
					}
				}
				Expect(fn).ToNot(BeNil())

				version := &gast.FileVersion{
					Path:    "fake.go",
					ModTime: time.Unix(123, 0),
					Hash:    "abc123",
				}

				symKey := graphs.NewSymbolKey(fn, version)

				Expect(symKey.Name).To(Equal("MyFunc"))
				Expect(symKey.FileId).To(Equal("fake.go|123|abc123"))
				Expect(symKey.Position).ToNot(Equal(token.NoPos))
				Expect(symKey.Id()).To(ContainSubstring("MyFunc@"))
			})

			It("Creates from TypeSpec", func() {
				spec := &ast.TypeSpec{
					Name: &ast.Ident{Name: "MyType"},
				}
				key := graphs.NewSymbolKey(spec, version)
				Expect(key.Name).To(Equal("MyType"))
			})

			It("Creates from ValueSpec", func() {
				spec := &ast.ValueSpec{
					Names: []*ast.Ident{{Name: "A"}, {Name: "B"}},
				}
				key := graphs.NewSymbolKey(spec, version)
				Expect(key.Name).To(Equal("A,B"))
			})

			It("Creates from Field", func() {
				field := &ast.Field{
					Names: []*ast.Ident{{Name: "FieldName"}},
				}
				key := graphs.NewSymbolKey(field, version)
				Expect(key.Name).To(Equal("FieldName"))
			})

			It("Creates from Ident", func() {
				id := &ast.Ident{Name: "VarX"}
				key := graphs.NewSymbolKey(id, version)
				Expect(key.Name).To(Equal("VarX"))
			})

			It("Falls back to default in NewSymbolKey for unknown node types", func() {
				version := &gast.FileVersion{
					Path:    "somefile.go",
					ModTime: time.Unix(4567, 0),
					Hash:    "def456",
				}

				node := fakeNode{}
				sk := graphs.NewSymbolKey(node, version)

				Expect(sk.Name).To(BeEmpty())
				Expect(sk.Position).To(Equal(token.Pos(42)))
				Expect(sk.FileId).To(Equal("somefile.go|4567|def456"))
			})

		})

		Describe("NewUniverseSymbolKey", func() {
			It("sets IsUniverse and IsBuiltIn", func() {
				key := graphs.NewUniverseSymbolKey("string")
				Expect(key.IsUniverse).To(BeTrue())
				Expect(key.IsBuiltIn).To(BeTrue())
			})
		})

		Describe("NewNonUniverseBuiltInSymbolKey", func() {
			It("sets IsBuiltIn but not IsUniverse", func() {
				key := graphs.NewNonUniverseBuiltInSymbolKey("error")
				Expect(key.IsBuiltIn).To(BeTrue())
				Expect(key.IsUniverse).To(BeFalse())
			})
		})
	})

	var _ = Describe("SymbolGraph", func() {
		var graph symboldg.SymbolGraph
		var fVersion *gast.FileVersion

		BeforeEach(func() {
			graph = symboldg.NewSymbolGraph()
			fVersion = utils.MakeFileVersion("file", "")
		})

		Context("NewSymbolGraph", func() {
			It("Creates a usable SymbolGraph", func() {
				g := symboldg.NewSymbolGraph()
				// Graph should be usable: add a primitive and check presence
				p := symboldg.PrimitiveTypeString
				n := g.AddPrimitive(p)
				Expect(n).ToNot(BeNil())
				Expect(g.IsPrimitivePresent(p)).To(BeTrue())
			})
		})

		Context("AddController", func() {
			It("Adds a controller node successfully", func() {
				controllerMeta := metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Node:     utils.MakeIdent("MyController"),
						FVersion: fVersion,
					},
				}
				request := symboldg.CreateControllerNode{
					Data:        controllerMeta,
					Annotations: nil,
				}

				node, err := graph.AddController(request)
				Expect(err).ToNot(HaveOccurred())
				Expect(node).ToNot(BeNil())
				Expect(node.Kind).To(Equal(common.SymKindController))
				Expect(node.Data).To(Equal(controllerMeta))
			})

			It("Returns an error when createAndAddSymNode returns an error", func() {
				// Pass a request with nil node to cause idempotencyGuard error
				controllerMeta := metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Node:     nil,
						FVersion: fVersion,
					},
				}
				request := symboldg.CreateControllerNode{
					Data:        controllerMeta,
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
							EdgeKind: common.Ptr(symboldg.EdgeKindReceiver),
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
				_, err := graph.AddController(symboldg.CreateControllerNode{Data: controllerMeta.Struct})
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
						Layers: []metadata.TypeLayer{
							metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("int"))),
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
						EdgeKind: common.Ptr(symboldg.EdgeKindParam),
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
					Data: controllerMeta.Struct,
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
						Layers: []metadata.TypeLayer{
							metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("string"))),
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
						EdgeKind: common.Ptr(symboldg.EdgeKindRetVal),
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
						Layers: []metadata.TypeLayer{
							metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("int"))),
						},
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
						EdgeKind: common.Ptr(symboldg.EdgeKindValue),
					},
				})
				Expect(valueChildren).To(HaveLen(len(enumMeta.Values)))

				// Each value should also have a Reference edge to the underlying type
				valueSymKey := graphs.NewUniverseSymbolKey(string(enumMeta.ValueKind))
				for _, valueChild := range valueChildren {
					referenceEdges := graph.Children(valueChild, &symboldg.TraversalBehavior{
						Filtering: symboldg.TraversalFilter{
							EdgeKind: common.Ptr(symboldg.EdgeKindReference),
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
						Layers: []metadata.TypeLayer{
							metadata.NewBaseLayer(&baseTypeKey),
						},
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
						EdgeKind: common.Ptr(symboldg.EdgeKindType),
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
				badMeta.Type.Layers = nil // no base layer → getTypeRef error

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
							Layers: []metadata.TypeLayer{
								metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("string"))),
							},
						},
					},
				})

				Expect(err).To(BeNil())
				Expect(node).ToNot(BeNil())
				Expect(node.Id.Name).To(Equal("SomeConst"))
				Expect(node.Kind).To(Equal(common.SymKindConstant))
			})
		})

		Context("FindByKind", func() {
			It("Correctly finds elements by the symbol kind", func() {
				// Add a couple of relevant nodes
				anyNode := graph.AddSpecial(symboldg.SpecialTypeAny)
				errNode := graph.AddSpecial(symboldg.SpecialTypeError)

				// Add a couple unrelated nodes that should be ignored

				graph.AddPrimitive(symboldg.PrimitiveTypeBool)
				graph.AddPrimitive(symboldg.PrimitiveTypeString)

				results := graph.FindByKind(common.SymKindSpecialBuiltin)

				Expect(results).To(HaveLen(2))
				Expect(results).To(ContainElements(anyNode, errNode))
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
							Layers:      []metadata.TypeLayer{metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("int")))},
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
							Layers:      []metadata.TypeLayer{metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("int")))},
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
							Layers:      []metadata.TypeLayer{metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("int")))},
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

		Context("IsPrimitivePresent", func() {
			It("Recognizes that a previously added primitive is present", func() {
				graph := symboldg.NewSymbolGraph()
				graph.AddPrimitive(symboldg.PrimitiveTypeBool)
				Expect(graph.IsPrimitivePresent(symboldg.PrimitiveTypeBool)).To(BeTrue())
			})

			It("Returns false for primitives that have not been added", func() {
				graph := symboldg.NewSymbolGraph()
				Expect(graph.IsPrimitivePresent(symboldg.PrimitiveTypeInt)).To(BeFalse())
			})
		})

		Context("IsSpecialPresent", func() {
			It("Recognizes a previously added special type", func() {
				graph := symboldg.NewSymbolGraph()
				graph.AddSpecial(symboldg.SpecialTypeError)
				Expect(graph.IsSpecialPresent(symboldg.SpecialTypeError)).To(BeTrue())
			})

			It("Returns false for special types not added", func() {
				graph := symboldg.NewSymbolGraph()
				Expect(graph.IsSpecialPresent(symboldg.SpecialTypeTime)).To(BeFalse())
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
						Layers: []metadata.TypeLayer{
							metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("int"))),
						},
					},
				}
				fieldNode, _ = graph.AddField(symboldg.CreateFieldNode{Data: fieldMeta})

				graph.AddEdge(structNode.Id, fieldNode.Id, symboldg.EdgeKindField, nil)
			})

			When("Unsorted", func() {
				It("Returns only children matching the filter", func() {
					filter := &symboldg.TraversalBehavior{
						Filtering: symboldg.TraversalFilter{
							EdgeKind: common.Ptr(symboldg.EdgeKindField),
						},
					}
					result := graph.Children(structNode, filter)
					Expect(result).To(HaveLen(1))
					Expect(result[0].Id).To(Equal(fieldNode.Id))
				})

				It("Skips edges that do not match a given filter", func() {
					filter := &symboldg.TraversalBehavior{
						Filtering: symboldg.TraversalFilter{
							EdgeKind: common.Ptr(symboldg.EdgeKindCall), // This should drop the edge
						},
					}
					result := graph.Children(structNode, filter)
					Expect(result).To(HaveLen(0))
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
							EdgeKind: common.Ptr(symboldg.EdgeKindCall),
						},
					}
					result := graph.Children(structNode, filter)
					Expect(result).To(HaveLen(0))
				})

				It("Skips nodes that do not match the given filter", func() {
					filter := &symboldg.TraversalBehavior{
						Sorting: symboldg.TraversalSortingOrdinalDesc,
						Filtering: symboldg.TraversalFilter{
							EdgeKind: common.Ptr(symboldg.EdgeKindField),
							NodeKind: common.Ptr(common.SymKindSpecialBuiltin),
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
							Layers: []metadata.TypeLayer{
								metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("int"))),
							},
						},
					},
				})

				otherNode, _ = graph.AddField(symboldg.CreateFieldNode{
					Data: metadata.FieldMeta{
						SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("OtherChild"), FVersion: fVersion},
						Type: metadata.TypeUsageMeta{
							SymNodeMeta: metadata.SymNodeMeta{Name: "string", FVersion: fVersion},
							Layers: []metadata.TypeLayer{
								metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("string"))),
							},
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
							EdgeKind: common.Ptr(symboldg.EdgeKindValue), // deliberately a different edge kind
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
							EdgeKind: common.Ptr(symboldg.EdgeKindValue), // deliberately a different edge kind
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
							Layers: []metadata.TypeLayer{
								metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("int"))),
							},
						},
					},
				})
				grandChildNode, _ = graph.AddField(symboldg.CreateFieldNode{
					Data: metadata.FieldMeta{
						SymNodeMeta: metadata.SymNodeMeta{Node: utils.MakeIdent("GrandChild"), FVersion: fVersion},
						Type: metadata.TypeUsageMeta{
							SymNodeMeta: metadata.SymNodeMeta{Name: "int", FVersion: fVersion},
							Layers: []metadata.TypeLayer{
								metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("int"))),
							},
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
						NodeKind: common.Ptr(common.SymKindSpecialBuiltin),
					},
				}

				result := graph.Descendants(structNode, filter)
				// We expect only the 'error' node here as the only filter match
				Expect(result).To(HaveLen(1))
				Expect(result[0].Id).To(Equal(errNode.Id))
			})
		})

		Context("Auxiliary", func() {
			It("Returns an error from the idempotency guard when the given FileVersion is nil", func() {
				_, err := graph.AddConst(symboldg.CreateConstNode{
					Data: metadata.ConstMeta{
						SymNodeMeta: metadata.SymNodeMeta{
							Name:     "SomeConst",
							Node:     utils.MakeIdent("SomeConst"),
							FVersion: nil,
						},
						Value: "some value",
						Type: metadata.TypeUsageMeta{
							SymNodeMeta: metadata.SymNodeMeta{Name: "string", FVersion: nil},
							Layers: []metadata.TypeLayer{
								metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("string"))),
							},
						},
					},
				})
				Expect(err).To(MatchError(ContainSubstring("idempotencyGuard received a nil version parameter")))
			})

			It("Evicts old nodes when adding same decl with a different FileVersion", func() {
				// Create a "same" AST node (same ident) but two different versions
				node := utils.MakeIdent("S")
				fv1 := utils.MakeFileVersion("v1", "some-hash")
				fv2 := utils.MakeFileVersion("v1", "some-different-hash")

				structMetaV1 := metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Node:     node,
						FVersion: fv1,
					},
				}
				n1, err := graph.AddStruct(symboldg.CreateStructNode{Data: structMetaV1})
				Expect(err).ToNot(HaveOccurred())
				Expect(n1).ToNot(BeNil())

				// Add a dependent: add a field belonging to this struct (so revDeps get populated)
				fieldMeta := metadata.FieldMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Node:     utils.MakeIdent("F1"),
						FVersion: fv1,
					},
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{
							FVersion: fv1,
						},
						Layers: []metadata.TypeLayer{
							metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("int"))),
						},
					},
				}
				_, err = graph.AddField(symboldg.CreateFieldNode{Data: fieldMeta})
				Expect(err).ToNot(HaveOccurred())

				// Now add the "same" struct but with a different file version -> should evict the previous
				structMetaV2 := metadata.StructMeta{SymNodeMeta: metadata.SymNodeMeta{Node: node, FVersion: fv2}}
				n2, err := graph.AddStruct(symboldg.CreateStructNode{Data: structMetaV2})
				Expect(err).ToNot(HaveOccurred())
				Expect(n2).ToNot(BeNil())
				// Ensure the new node is present and has the new version
				found := false
				for _, s := range graph.FindByKind(common.SymKindStruct) {
					if s.Id.Equals(n2.Id) {
						found = true
					}
				}
				Expect(found).To(BeTrue())
			})
		})

		Context("String", func() {
			It("Outputs an empty graph with 'headers' when empty", func() {
				text := graph.String()
				Expect(text).To(Equal("=== SymbolGraph Dump ===\nTotal nodes: 0\n\n=== End SymbolGraph ===\n"))
			})

			It("Outputs a correct graph when nodes exist but have no dependencies", func() {
				_, err := graph.AddConst(symboldg.CreateConstNode{
					Data: metadata.ConstMeta{
						SymNodeMeta: metadata.SymNodeMeta{
							Name:     "SomeConst",
							Node:     utils.MakeIdent("SomeConst"),
							FVersion: fVersion,
						},
						Value: "Some Value",
						Type: metadata.TypeUsageMeta{
							SymNodeMeta: metadata.SymNodeMeta{Name: "string", FVersion: fVersion},
							Layers: []metadata.TypeLayer{
								metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("string"))),
							},
						},
					},
				})
				Expect(err).To(BeNil())

				text := graph.String()
				expectedPattern := `^=== SymbolGraph Dump ===\n` +
					`Total nodes: 1\n\n` +
					`\[Constant\] SomeConst\n` +
					`    • file\n` +
					`    • \d+\n` + // matches timestamp
					`    • hash-file-\n\n` +
					`  Dependencies: \(none\)\n` +
					`  Dependents: \(none\)\n\n` +
					`=== End SymbolGraph ===\n$`

				Expect(text).To(MatchRegexp(expectedPattern))
			})

			It("Outputs a correct graph when nodes exist and have a dependent node", func() {
				constNode, err := graph.AddConst(symboldg.CreateConstNode{
					Data: metadata.ConstMeta{
						SymNodeMeta: metadata.SymNodeMeta{
							Name:     "SomeConst",
							Node:     utils.MakeIdent("SomeConst"),
							FVersion: fVersion,
						},
						Value: "Some Value",
						Type: metadata.TypeUsageMeta{
							SymNodeMeta: metadata.SymNodeMeta{Name: "string", FVersion: fVersion},
							Layers: []metadata.TypeLayer{
								metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("string"))),
							},
						},
					},
				})
				Expect(err).To(BeNil())

				strNode := graph.AddPrimitive(symboldg.PrimitiveTypeString)
				graph.AddEdge(constNode.Id, strNode.Id, symboldg.EdgeKindType, nil)

				text := graph.String()

				// Graph string output does not guarantee ordering so we have to test separately here
				const nodeBlockConstant = `(?s)\[Constant\] SomeConst\n` +
					`    • file\n` +
					`    • \d+\n` +
					`    • hash-file-\n\n` +
					`  Dependencies:\n` +
					`    • \[Builtin\] string \(ty\)\n` +
					`  Dependents: \(none\)`

				const nodeBlockBuiltin = `(?s)\[Builtin\] string\n` +
					`  Dependencies: \(none\)\n` +
					`  Dependents:\n` +
					`    • \[Constant\] SomeConst\n` +
					`    • file\n` +
					`    • \d+\n` +
					`    • hash-file-`

				// Assert that both node blocks exist somewhere in the dump
				Expect(text).To(MatchRegexp(nodeBlockConstant))
				Expect(text).To(MatchRegexp(nodeBlockBuiltin))

			})

			It("Outputs a correct graph when nodes exist and have a dependent edge without node", func() {
				constNode, err := graph.AddConst(symboldg.CreateConstNode{
					Data: metadata.ConstMeta{
						SymNodeMeta: metadata.SymNodeMeta{
							Name:     "SomeConst",
							Node:     utils.MakeIdent("SomeConst"),
							FVersion: fVersion,
						},
						Value: "Some Value",
						Type: metadata.TypeUsageMeta{
							SymNodeMeta: metadata.SymNodeMeta{Name: "string", FVersion: fVersion},
							Layers: []metadata.TypeLayer{
								metadata.NewBaseLayer(common.Ptr(graphs.NewUniverseSymbolKey("string"))),
							},
						},
					},
				})
				Expect(err).To(BeNil())

				// Add an edge without actually adding a node
				graph.AddEdge(constNode.Id, graphs.NewUniverseSymbolKey("string"), symboldg.EdgeKindType, nil)

				text := graph.String()

				const expectedRx = `(?m)^=== SymbolGraph Dump ===\n` +
					`Total nodes: 1\n\n` +
					`\[Constant\] SomeConst\n` +
					`    • file\n` +
					`    • \d+\n` +
					`    • hash-file-\n\n` +
					`  Dependencies:\n` +
					`    • \[string\] \(ty\)\n` +
					`  Dependents: \(none\)\n\n` +
					`=== End SymbolGraph ===$`

				Expect(text).To(MatchRegexp(expectedRx))
			})
		})

		Context("ToDot", func() {
			It("Outputs correct empty graph with default style when empty", func() {
				text := graph.ToDot(nil)

				const expectedDotGraph = "digraph SymbolGraph {\n" +
					"  rankdir=TB;\n" +
					"  subgraph cluster_legend {\n" +
					"    label = \"Legend\";\n" +
					"    style = dashed;\n" +
					"    L0 [label=\"Alias\", style=filled, shape=note, fillcolor=\"palegreen\"];\n" +
					"    L1 [label=\"Builtin\", style=filled, shape=box, fillcolor=\"gray80\"];\n" +
					"    L2 [label=\"Constant\", style=filled, shape=egg, fillcolor=\"plum\"];\n" +
					"    L3 [label=\"Controller\", style=filled, shape=octagon, fillcolor=\"lightcyan\"];\n" +
					"    L4 [label=\"Enum\", style=filled, shape=folder, fillcolor=\"mediumpurple\"];\n" +
					"    L5 [label=\"EnumValue\", style=filled, shape=note, fillcolor=\"plum\"];\n" +
					"    L6 [label=\"Field\", style=filled, shape=ellipse, fillcolor=\"gold\"];\n" +
					"    L7 [label=\"Function\", style=filled, shape=oval, fillcolor=\"darkseagreen\"];\n" +
					"    L8 [label=\"Interface\", style=filled, shape=component, fillcolor=\"lightskyblue\"];\n" +
					"    L9 [label=\"Package\", style=filled, shape=folder, fillcolor=\"lightyellow\"];\n" +
					"    L10 [label=\"Parameter\", style=filled, shape=parallelogram, fillcolor=\"khaki\"];\n" +
					"    L11 [label=\"Receiver\", style=filled, shape=hexagon, fillcolor=\"orange\"];\n" +
					"    L12 [label=\"RetType\", style=filled, shape=diamond, fillcolor=\"lightgrey\"];\n" +
					"    L13 [label=\"Struct\", style=filled, shape=box, fillcolor=\"lightblue\"];\n" +
					"    L14 [label=\"Unknown\", style=filled, shape=triangle, fillcolor=\"lightcoral\"];\n" +
					"    L15 [label=\"Variable\", style=filled, shape=circle, fillcolor=\"lightsteelblue\"];\n" +
					"  }\n" +
					"}\n"

				Expect(text).To(Equal(expectedDotGraph))
			})

			It("Outputs nodes and their edges", func() {
				anyNode := graph.AddSpecial(symboldg.SpecialTypeAny)
				strNode := graph.AddPrimitive(symboldg.PrimitiveTypeString)
				graph.AddEdge(anyNode.Id, strNode.Id, symboldg.EdgeKindType, nil)

				text := graph.ToDot(nil)

				Expect(text).To(ContainSubstring("N0 [label=\"any@.| (Special)\""))
				Expect(text).To(ContainSubstring("N1 [label=\"string@.| (Builtin)\""))
				Expect(text).To(ContainSubstring("N0 -> N1 [label=\"Type\""))
			})
		})
	})
})

func TestUnitGraphs(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Graphs")
}
