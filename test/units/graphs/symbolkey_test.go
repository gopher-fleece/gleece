package graphs_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"time"

	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fakeNode struct{}

func (f fakeNode) Pos() token.Pos { return token.Pos(42) }
func (f fakeNode) End() token.Pos { return token.Pos(99) }

var _ = Describe("Unit Tests - SymbolKey", func() {

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

func TestUnitGraphs(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - SymbolKey")
}
