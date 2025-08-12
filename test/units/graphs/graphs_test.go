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
				children := graph.Children(&symboldg.SymbolNode{Id: ctrlMeta.SymbolKey()}, &symboldg.TraversalFilter{
					EdgeKind: common.Ptr(symboldg.EdgeKindReceiver),
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
				children := graph.Children(routeNode, &symboldg.TraversalFilter{
					EdgeKind: common.Ptr(symboldg.EdgeKindParam),
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

				children := graph.Children(routeNode, &symboldg.TraversalFilter{
					EdgeKind: common.Ptr(symboldg.EdgeKindRetVal),
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
				valueChildren := graph.Children(enumNode, &symboldg.TraversalFilter{
					EdgeKind: common.Ptr(symboldg.EdgeKindValue),
				})
				Expect(valueChildren).To(HaveLen(len(enumMeta.Values)))

				// Each value should also have a Reference edge to the underlying type
				valueSymKey := graphs.NewUniverseSymbolKey(string(enumMeta.ValueKind))
				for _, valueChild := range valueChildren {
					referenceEdges := graph.Children(valueChild, &symboldg.TraversalFilter{
						EdgeKind: common.Ptr(symboldg.EdgeKindReference),
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

				children := graph.Children(fieldNode, &symboldg.TraversalFilter{
					EdgeKind: common.Ptr(symboldg.EdgeKindType),
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

			It("Returns only children matching the filter", func() {
				filter := &symboldg.TraversalFilter{EdgeKind: common.Ptr(symboldg.EdgeKindField)}
				result := graph.Children(structNode, filter)
				Expect(result).To(HaveLen(1))
				Expect(result[0].Id).To(Equal(fieldNode.Id))
			})

			It("Skips edges that do not match a given filter", func() {
				filter := &symboldg.TraversalFilter{EdgeKind: common.Ptr(symboldg.EdgeKindCall)} // This should drop the edge
				result := graph.Children(structNode, filter)
				Expect(result).To(HaveLen(0))
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
				filter := &symboldg.TraversalFilter{
					EdgeKind: common.Ptr(symboldg.EdgeKindValue), // deliberately a different edge kind
				}

				result := graph.Parents(fieldNode, filter)
				Expect(result).To(HaveLen(0))
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
				filter := &symboldg.TraversalFilter{
					NodeKind: common.Ptr(common.SymKindSpecialBuiltin),
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

		Context("String and ToDot outputs", func() {
			It("Produces a non-empty String dump and ToDot output", func() {
				g := symboldg.NewSymbolGraph()
				// Populate with a couple nodes
				g.AddPrimitive(symboldg.PrimitiveTypeBool)
				fv := utils.MakeFileVersion("dump", "")
				_, err := g.AddStruct(symboldg.CreateStructNode{
					Data: metadata.StructMeta{
						SymNodeMeta: metadata.SymNodeMeta{
							Node:     utils.MakeIdent("D"),
							FVersion: fv,
						},
					},
				},
				)
				Expect(err).ToNot(HaveOccurred())

				s := g.String()
				Expect(s).To(ContainSubstring("=== SymbolGraph Dump ==="))
				Expect(len(s)).To(BeNumerically(">", 0))

				// ToDot should not panic and should return a string
				dot := g.ToDot(nil)
				Expect(len(dot)).To(BeNumerically(">", 0))
			})
		})
	})
})

func TestUnitGraphs(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Graphs")
}
