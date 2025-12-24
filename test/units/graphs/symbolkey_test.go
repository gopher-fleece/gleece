package graphs_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"time"

	"github.com/gopher-fleece/gleece/v2/gast"
	"github.com/gopher-fleece/gleece/v2/graphs"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fakeNode struct{}

func (f fakeNode) Pos() token.Pos { return token.Pos(42) }
func (f fakeNode) End() token.Pos { return token.Pos(99) }

var _ = Describe("Unit Tests - SymbolKey", func() {

	Context("Id", func() {

		It("Returns the correct ID for universe types", func() {
			universeKey := graphs.NewUniverseSymbolKey("string")
			Expect(universeKey.Id()).To(Equal("UniverseType:string"))
		})

		It("Returns the correct ID for named symbol", func() {
			namedKey := graphs.SymbolKey{
				Name:     "Foo",
				Position: 123,
				FileId:   "path/to/file|mod|abc123",
			}
			Expect(namedKey.Id()).To(Equal("Foo@123@path/to/file|mod|abc123"))
		})

		It("Returns the correct ID for unnamed symbol", func() {
			unnamedKey := graphs.SymbolKey{
				Position: 456,
				FileId:   "path/to/file|mod|xyz999",
			}
			Expect(unnamedKey.Id()).To(Equal("@456@path/to/file|mod|xyz999"))
		})
	})

	Context("ShortLabel", func() {

		It("Includes base file name and short hash", func() {
			key := graphs.SymbolKey{
				Name:   "Bar",
				FileId: "path/to/thing.go|mymod|deadbeef1234",
			}

			Expect(key.ShortLabel()).To(Equal("Bar @thing.go|deadbee"))
		})

		It("Handles missing name and uses file info only", func() {
			key := graphs.SymbolKey{
				FileId: "some/path.go|mod|abcdef1234567890",
			}

			Expect(key.ShortLabel()).To(Equal("path.go|abcdef1"))
		})

		It("Shows built-in (non-universe) names unchanged", func() {
			builtInKey := graphs.NewNonUniverseBuiltInSymbolKey("error")

			// no file info attached -> label should be exactly the name
			Expect(builtInKey.ShortLabel()).To(Equal("error"))
		})

		It("Formats type-params and attaches file info", func() {
			fileVersion := &gast.FileVersion{
				Path:    "pkg/param.go",
				ModTime: time.Unix(99, 0),
				Hash:    "paramhash",
			}

			paramKey := graphs.NewParamSymbolKey(fileVersion, "TA", 0)

			shortLabel := paramKey.ShortLabel()

			Expect(shortLabel).To(ContainSubstring("TA#0"))
			Expect(shortLabel).To(ContainSubstring("param.go"))
			Expect(shortLabel).To(ContainSubstring("paramha")) // short hash (7 chars)
		})

		It("Formats instantiated names with universe and regular args", func() {
			baseKey := graphs.SymbolKey{
				Name:     "MultiGenericStruct",
				Position: 1,
				FileId:   "base.go|100|bhashvalue",
				FilePath: "base.go",
			}

			universeKey := graphs.NewUniverseSymbolKey("string")

			regularArg := graphs.SymbolKey{
				Name:     "T",
				Position: 2,
				FileId:   "arg.go|200|ahash",
				FilePath: "arg.go",
			}

			instKey := graphs.NewInstSymbolKey(
				baseKey,
				[]graphs.SymbolKey{
					universeKey,
					regularArg,
				},
			)

			shortLabel := instKey.ShortLabel()

			Expect(shortLabel).To(ContainSubstring("MultiGenericStruct["))
			Expect(shortLabel).To(ContainSubstring("string"))
			Expect(shortLabel).To(ContainSubstring("T"))
			Expect(shortLabel).To(ContainSubstring("base.go"))
		})

		It("Formats instantiated name with no args as just the base", func() {
			baseKey := graphs.SymbolKey{
				Name:     "OnlyBase",
				Position: 5,
				FileId:   "only.go|5|ohash",
				FilePath: "only.go",
			}

			instKey := graphs.NewInstSymbolKey(baseKey, nil)

			shortLabel := instKey.ShortLabel()

			Expect(shortLabel).To(ContainSubstring("OnlyBase"))
			Expect(shortLabel).To(ContainSubstring("only.go"))
		})

		It("Pretty prints composite slice/ptr/array/map/func/unknown kinds", func() {
			fileVersion := &gast.FileVersion{
				Path:    "pkg/thing.go",
				ModTime: time.Unix(1000, 0),
				Hash:    "complhash",
			}

			opSimple := graphs.SymbolKey{
				Name:   "SimpleStruct",
				FileId: "op.go|1|ohash",
			}

			// slice
			sliceKey := graphs.NewCompositeTypeKey(
				graphs.CompositeKindSlice,
				fileVersion,
				[]graphs.SymbolKey{
					opSimple,
				},
			)

			Expect(sliceKey.Name).To(ContainSubstring("comp:slice"))
			Expect(sliceKey.ShortLabel()).To(ContainSubstring("[]SimpleStruct"))
			Expect(sliceKey.ShortLabel()).To(ContainSubstring("thing.go"))

			// ptr
			ptrKey := graphs.NewCompositeTypeKey(
				graphs.CompositeKindPtr,
				fileVersion,
				[]graphs.SymbolKey{
					{Name: "Foo", FileId: "foo.go|1|fhash"},
				},
			)

			Expect(ptrKey.ShortLabel()).To(ContainSubstring("*Foo"))

			// array: if operand name is "10" -> "[10]"
			arrayOperand := graphs.SymbolKey{
				Name:   "10",
				FileId: "alen.go|1|alen",
			}

			arrayKey := graphs.NewCompositeTypeKey(
				graphs.CompositeKindArray,
				fileVersion,
				[]graphs.SymbolKey{
					arrayOperand,
				},
			)

			Expect(arrayKey.ShortLabel()).To(ContainSubstring("[10]"))

			// map with two operands
			mapKey := graphs.NewCompositeTypeKey(
				graphs.CompositeKindMap,
				fileVersion,
				[]graphs.SymbolKey{
					{Name: "Key", FileId: "k.go|1|k"},
					{Name: "Val", FileId: "v.go|1|v"},
				},
			)

			Expect(mapKey.ShortLabel()).To(ContainSubstring("map[Key]Val"))

			// map with single operand -> fallback: map[inner]
			mapSingleKey := graphs.NewCompositeTypeKey(
				graphs.CompositeKindMap,
				fileVersion,
				[]graphs.SymbolKey{
					{Name: "Only", FileId: "o.go|1|o"},
				},
			)

			Expect(mapSingleKey.ShortLabel()).To(ContainSubstring("map[Only]"))

			// func with two operands
			funcKey := graphs.NewCompositeTypeKey(
				graphs.CompositeKindFunc,
				fileVersion,
				[]graphs.SymbolKey{
					{Name: "Args", FileId: "a.go|1|a"},
					{Name: "Rets", FileId: "r.go|1|r"},
				},
			)

			Expect(funcKey.ShortLabel()).To(ContainSubstring("func(Args)(Rets)"))

			// func with single operand -> fallback
			funcSingleKey := graphs.NewCompositeTypeKey(
				graphs.CompositeKindFunc,
				fileVersion,
				[]graphs.SymbolKey{
					{Name: "OnlyArgs", FileId: "oa.go|1|oa"},
				},
			)

			Expect(funcSingleKey.ShortLabel()).To(ContainSubstring("func(OnlyArgs)"))

			// unknown kind -> trim inner
			unknownKey := graphs.NewCompositeTypeKey(
				graphs.CompositeKind("weird"),
				fileVersion,
				[]graphs.SymbolKey{
					{Name: "WeirdInner", FileId: "w.go|1|w"},
				},
			)

			Expect(unknownKey.ShortLabel()).To(ContainSubstring("WeirdInner"))
		})

		It("Works with nil fileVersion (no file info attached)", func() {
			op := graphs.SymbolKey{
				Name:   "Elem",
				FileId: "e.go|1|ehash",
			}

			compositeNoFile := graphs.NewCompositeTypeKey(
				graphs.CompositeKindSlice,
				nil,
				[]graphs.SymbolKey{
					op,
				},
			)

			shortLabel := compositeNoFile.ShortLabel()

			Expect(shortLabel).To(ContainSubstring("[]Elem"))
			Expect(shortLabel).ToNot(ContainSubstring("@"))
		})

		It("Attaches file base without short hash when present", func() {
			key := graphs.SymbolKey{
				Name:   "Label",
				FileId: "path/to/file.go|mymod", // no hash part -> shortHash == ""
			}

			shortLabel := key.ShortLabel()

			// exact expected format: "Label @file.go"
			Expect(shortLabel).To(Equal("Label @file.go"))
		})

		It("Returns file base only when name is empty and no short hash", func() {
			key := graphs.SymbolKey{
				Name:   "",                     // empty name -> fall back to file info only
				FileId: "some/dir/file.go|mod", // no hash part
			}

			shortLabel := key.ShortLabel()

			// should return only the file base
			Expect(shortLabel).To(Equal("file.go"))
		})

		It("Returns '?' when both name and file info are missing", func() {
			key := graphs.SymbolKey{
				Name:   "",
				FileId: "",
			}

			shortLabel := key.ShortLabel()

			Expect(shortLabel).To(Equal("?"))
		})

		It("Handles composite names that lack brackets (parseCompositeName br<0 branch)", func() {
			compositeKey := graphs.SymbolKey{
				Name:     "comp:slice",                       // no '[...]' -> parseCompositeName will return kind="slice", args=""
				FileId:   "pkg/nomod/sym.go|mod|comp1234567", // include a hash so file info is attached
				FilePath: "sym.go",
			}

			shortLabel := compositeKey.ShortLabel()

			// composite kind "slice" with empty args should produce "[]" prefix
			Expect(shortLabel).To(ContainSubstring("[]"))

			// file base should be attached; short hash is truncated to 7 chars
			Expect(shortLabel).To(ContainSubstring("sym.go"))
		})

	})

	Context("PrettyPrint", func() {

		It("Prints cleanly for universe types", func() {
			key := graphs.NewUniverseSymbolKey("UniverseType:string")
			Expect(key.PrettyPrint()).To(Equal("string"))
		})

		It("Prints details for named symbol", func() {
			namedKey := graphs.SymbolKey{
				Name:     "MyFunc",
				Position: 789,
				FileId:   "src/main.go|mod|abcd1234",
			}

			result := namedKey.PrettyPrint()

			Expect(result).To(ContainSubstring("MyFunc"))
			Expect(result).To(ContainSubstring("• src/main.go"))
			Expect(result).To(ContainSubstring("• mod"))
			Expect(result).To(ContainSubstring("• abcd1234"))
		})

		It("Pretty prints a symbol with no name using position fallback", func() {
			fileSet := token.NewFileSet()
			src := `package main; var _ = 123`

			fileNode, err := parser.ParseFile(fileSet, "dummy.go", src, parser.AllErrors)
			Expect(err).To(BeNil())

			fileVersion := &gast.FileVersion{
				Path:    "dummy.go",
				ModTime: time.Unix(1234, 0),
				Hash:    "abcd1234",
			}

			// Use a node type not explicitly handled by NewSymbolKey to exercise fallback
			symbolKey := graphs.NewSymbolKey(fileNode, fileVersion)

			Expect(symbolKey.Name).To(Equal(""))
			Expect(symbolKey.Position).ToNot(Equal(token.NoPos))

			output := symbolKey.PrettyPrint()

			Expect(output).To(HavePrefix(fmt.Sprintf("@%d\n", symbolKey.Position)))
			Expect(output).To(ContainSubstring("• dummy.go"))
			Expect(output).To(ContainSubstring("• abcd1234"))
		})
	})

	Context("Equals", func() {

		It("Returns true for equal universe keys", func() {
			a := graphs.NewUniverseSymbolKey("bool")
			b := graphs.NewUniverseSymbolKey("bool")
			Expect(a.Equals(b)).To(BeTrue())
		})

		It("Returns false for unequal universe keys", func() {
			a := graphs.NewUniverseSymbolKey("string")
			b := graphs.NewUniverseSymbolKey("int")
			Expect(a.Equals(b)).To(BeFalse())
		})

		It("Returns true for equal non-universe keys", func() {
			a := graphs.SymbolKey{Name: "Foo", Position: 1, FileId: "id"}
			b := graphs.SymbolKey{Name: "Foo", Position: 1, FileId: "id"}
			Expect(a.Equals(b)).To(BeTrue())
		})

		It("Returns false if one field differs", func() {
			a := graphs.SymbolKey{Name: "Foo", Position: 1, FileId: "id"}
			b := graphs.SymbolKey{Name: "Foo", Position: 2, FileId: "id"}
			Expect(a.Equals(b)).To(BeFalse())
		})
	})

	Context("NewSymbolKey", func() {

		var fileVersion *gast.FileVersion

		BeforeEach(func() {
			fileVersion = &gast.FileVersion{
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
			fileSet := token.NewFileSet()
			src := `package main; func MyFunc() {}`

			parsedFile, err := parser.ParseFile(fileSet, "fake.go", src, parser.AllErrors)
			Expect(err).To(BeNil())

			var fn *ast.FuncDecl
			for _, decl := range parsedFile.Decls {
				if fd, ok := decl.(*ast.FuncDecl); ok {
					fn = fd
					break
				}
			}
			Expect(fn).ToNot(BeNil())

			funcFileVersion := &gast.FileVersion{
				Path:    "fake.go",
				ModTime: time.Unix(123, 0),
				Hash:    "abc123",
			}

			symKey := graphs.NewSymbolKey(fn, funcFileVersion)

			Expect(symKey.Name).To(Equal("MyFunc"))
			Expect(symKey.FileId).To(Equal("fake.go|123|abc123"))
			Expect(symKey.Position).ToNot(Equal(token.NoPos))
			Expect(symKey.Id()).To(ContainSubstring("MyFunc@"))
		})

		It("Creates from TypeSpec", func() {
			spec := &ast.TypeSpec{
				Name: &ast.Ident{Name: "MyType"},
			}
			typedKey := graphs.NewSymbolKey(spec, fileVersion)
			Expect(typedKey.Name).To(Equal("MyType"))
		})

		It("Creates from ValueSpec", func() {
			valueSpec := &ast.ValueSpec{
				Names: []*ast.Ident{{Name: "A"}, {Name: "B"}},
			}
			valueKey := graphs.NewSymbolKey(valueSpec, fileVersion)
			Expect(valueKey.Name).To(Equal("A,B"))
		})

		It("Creates from Field", func() {
			field := &ast.Field{
				Names: []*ast.Ident{{Name: "FieldName"}},
			}
			fieldKey := graphs.NewSymbolKey(field, fileVersion)
			Expect(fieldKey.Name).To(Equal("FieldName"))
		})

		It("Creates from Ident", func() {
			id := &ast.Ident{Name: "VarX"}
			idKey := graphs.NewSymbolKey(id, fileVersion)
			Expect(idKey.Name).To(Equal("VarX"))
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

	Context("NewUniverseSymbolKey", func() {

		It("Sets IsUniverse and IsBuiltIn", func() {
			key := graphs.NewUniverseSymbolKey("string")
			Expect(key.IsUniverse).To(BeTrue())
			Expect(key.IsBuiltIn).To(BeTrue())
		})
	})

	Context("NewNonUniverseBuiltInSymbolKey", func() {

		It("Sets IsBuiltIn but not IsUniverse", func() {
			key := graphs.NewNonUniverseBuiltInSymbolKey("error")
			Expect(key.IsBuiltIn).To(BeTrue())
			Expect(key.IsUniverse).To(BeFalse())
		})
	})
})

func TestUnitGraphs(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - SymbolKey")
}
