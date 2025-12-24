package arbitrators_test

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
	"testing"
	"time"

	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/core/arbitrators"
	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/core/visitors/providers"
	"github.com/gopher-fleece/gleece/v2/gast"
	"github.com/gopher-fleece/gleece/v2/graphs"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	"golang.org/x/tools/go/packages"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// testVisitor is a tiny test double that implements the exported arbitrators.TypeVisitor
// interface. Each test can inject behaviour by setting VisitFieldFunc.
type testVisitor struct {
	VisitFieldFunc      func(pkg *packages.Package, file *ast.File, field *ast.Field) ([]metadata.FieldMeta, error)
	VisitStructTypeFunc func(file *ast.File, nodeGenDecl *ast.GenDecl, node *ast.TypeSpec) (metadata.StructMeta, graphs.SymbolKey, error)
}

func (v *testVisitor) Visit(node ast.Node) ast.Visitor { return nil }

func (v *testVisitor) VisitStructType(file *ast.File, nodeGenDecl *ast.GenDecl, node *ast.TypeSpec) (metadata.StructMeta, graphs.SymbolKey, error) {
	if v.VisitStructTypeFunc != nil {
		return v.VisitStructTypeFunc(file, nodeGenDecl, node)
	}
	return metadata.StructMeta{}, graphs.SymbolKey{}, nil
}

func (v *testVisitor) VisitField(
	pkg *packages.Package,
	file *ast.File,
	field *ast.Field,
	kind common.SymKind,
) ([]metadata.FieldMeta, error) {
	if v.VisitFieldFunc != nil {
		return v.VisitFieldFunc(pkg, file, field)
	}
	// Default: pretend the field yields a single FieldMeta with the first name if present
	name := ""
	if len(field.Names) > 0 {
		name = field.Names[0].Name
	}
	return []metadata.FieldMeta{
		{
			SymNodeMeta: metadata.SymNodeMeta{
				Name:     name,
				Node:     field,
				PkgPath:  "pkg",
				FVersion: &gast.FileVersion{Path: "p", ModTime: time.Now(), Hash: "h"},
			},
			Type: metadata.TypeUsageMeta{},
		},
	}, nil
}

var _ = Describe("Unit Tests - AST Arbitrator (external)", func() {
	var arbProvider *providers.ArbitrationProvider
	var astArb *arbitrators.AstArbitrator
	var pkgFacade any

	BeforeEach(func() {
		var err error
		arbProvider, err = providers.NewArbitrationProviderFromGleeceConfig(nil)
		Expect(err).To(BeNil())

		// Use exported provider API to get the arbitrator.
		astArb = arbProvider.Ast()

		// The provider exposes a package facade via Pkg(); keep a local reference if needed.
		pkgFacade = arbProvider.Pkg()
		Expect(astArb).ToNot(BeNil())
		Expect(pkgFacade).ToNot(BeNil())
	})

	Context("GetFuncParametersMeta", func() {
		It("Returns empty when function has no parameters", func() {
			fn := &ast.FuncDecl{
				Name: ast.NewIdent("NoParams"),
				Type: &ast.FuncType{
					Params: nil,
				},
			}
			tv := &testVisitor{}
			params, err := astArb.GetFuncParametersMeta(tv, nil, nil, fn, nil)
			Expect(err).To(BeNil())
			Expect(params).To(HaveLen(0))
		})

		It("Returns error when visitor returns an error for a parameter field", func() {
			fn := &ast.FuncDecl{
				Name: ast.NewIdent("ParamErr"),
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: []*ast.Field{
							{Type: ast.NewIdent("int")},
						},
					},
				},
			}
			tv := &testVisitor{
				VisitFieldFunc: func(pkg *packages.Package, file *ast.File, field *ast.Field) ([]metadata.FieldMeta, error) {
					return nil, fmt.Errorf("visit failed")
				},
			}
			_, err := astArb.GetFuncParametersMeta(tv, nil, nil, fn, nil)
			Expect(err).To(MatchError(ContainSubstring("visit failed")))
		})

		It("Correctly expands a multi-name field (a, b string) into two parameters with ordinals", func() {
			a := ast.NewIdent("a")
			b := ast.NewIdent("b")
			fn := &ast.FuncDecl{
				Name: ast.NewIdent("Multi"),
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{a, b},
								Type:  ast.NewIdent("string"),
							},
						},
					},
				},
			}

			fmA := metadata.FieldMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:       "a",
					Node:       a,
					PkgPath:    "pkg",
					FVersion:   &gast.FileVersion{Path: "p", ModTime: time.Now(), Hash: "h"},
					SymbolKind: common.SymKindParameter,
				},
				Type: metadata.TypeUsageMeta{},
			}
			fmB := fmA
			fmB.Name = "b"

			tv := &testVisitor{
				VisitFieldFunc: func(pkg *packages.Package, file *ast.File, field *ast.Field) ([]metadata.FieldMeta, error) {
					// Simulate that the field yields two entries (a and b)
					return []metadata.FieldMeta{fmA, fmB}, nil
				},
			}

			params, err := astArb.GetFuncParametersMeta(tv, nil, nil, fn, nil)
			Expect(err).To(BeNil())
			Expect(params).To(HaveLen(2))
			Expect(params[0].Name).To(Equal("a"))
			Expect(params[0].Ordinal).To(Equal(0))
			Expect(params[1].Name).To(Equal("b"))
			Expect(params[1].Ordinal).To(Equal(1))
		})

		It("Aggregates multiple field entries across multiple parameter fields", func() {
			a := ast.NewIdent("a")
			b := ast.NewIdent("b")
			c := ast.NewIdent("c")
			fn := &ast.FuncDecl{
				Name: ast.NewIdent("Combine"),
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{a},
								Type:  ast.NewIdent("int"),
							},
							{
								Names: []*ast.Ident{b, c},
								Type:  ast.NewIdent("string"),
							},
						},
					},
				},
			}

			fma := metadata.FieldMeta{SymNodeMeta: metadata.SymNodeMeta{Name: "a"}}
			fmb := metadata.FieldMeta{SymNodeMeta: metadata.SymNodeMeta{Name: "b"}}
			fmc := metadata.FieldMeta{SymNodeMeta: metadata.SymNodeMeta{Name: "c"}}

			tv := &testVisitor{
				VisitFieldFunc: func(pkg *packages.Package, file *ast.File, field *ast.Field) ([]metadata.FieldMeta, error) {
					if len(field.Names) == 1 && field.Names[0].Name == "a" {
						return []metadata.FieldMeta{fma}, nil
					}
					return []metadata.FieldMeta{fmb, fmc}, nil
				},
			}

			params, err := astArb.GetFuncParametersMeta(tv, nil, nil, fn, nil)
			Expect(err).To(BeNil())
			Expect(params).To(HaveLen(3))
			Expect(params[0].Name).To(Equal("a"))
			Expect(params[1].Name).To(Equal("b"))
			Expect(params[2].Name).To(Equal("c"))
			Expect(params[0].Ordinal).To(Equal(0))
			Expect(params[1].Ordinal).To(Equal(1))
			Expect(params[2].Ordinal).To(Equal(2))
		})
	})

	Context("GetFuncRetValMeta", func() {
		It("Returns empty when function has no results", func() {
			fn := &ast.FuncDecl{
				Name: ast.NewIdent("NoResults"),
				Type: &ast.FuncType{
					Results: nil,
				},
			}
			tv := &testVisitor{}
			out, err := astArb.GetFuncRetValMeta(tv, nil, nil, fn, nil)
			Expect(err).To(BeNil())
			Expect(out).To(HaveLen(0))
		})

		It("Returns error when visitor returns an error for a result field", func() {
			fn := &ast.FuncDecl{
				Name: ast.NewIdent("ResErr"),
				Type: &ast.FuncType{
					Results: &ast.FieldList{
						List: []*ast.Field{{Type: ast.NewIdent("int")}},
					},
				},
			}
			tv := &testVisitor{
				VisitFieldFunc: func(pkg *packages.Package, file *ast.File, field *ast.Field) ([]metadata.FieldMeta, error) {
					return nil, fmt.Errorf("visit results failed")
				},
			}
			_, err := astArb.GetFuncRetValMeta(tv, nil, nil, fn, nil)
			Expect(err).To(MatchError(ContainSubstring("visit results failed")))
		})

		It("Returns error when visitor returns zero fields for a result entry", func() {
			fn := &ast.FuncDecl{
				Name: ast.NewIdent("ZeroFields"),
				Type: &ast.FuncType{
					Results: &ast.FieldList{
						List: []*ast.Field{{Type: ast.NewIdent("T")}},
					},
				},
			}
			tv := &testVisitor{
				VisitFieldFunc: func(pkg *packages.Package, file *ast.File, field *ast.Field) ([]metadata.FieldMeta, error) {
					return []metadata.FieldMeta{}, nil
				},
			}
			_, err := astArb.GetFuncRetValMeta(tv, nil, nil, fn, nil)
			Expect(err).To(MatchError(ContainSubstring("did not yield any information")))
			Expect(err.Error()).To(ContainSubstring("ZeroFields"))
		})

		It("Returns error when visitor returns multiple fields for a single result (multi-variable declaration)", func() {
			fn := &ast.FuncDecl{
				Name: ast.NewIdent("MultiRet"),
				Type: &ast.FuncType{
					Results: &ast.FieldList{
						List: []*ast.Field{{Type: ast.NewIdent("T")}},
					},
				},
			}
			tv := &testVisitor{
				VisitFieldFunc: func(pkg *packages.Package, file *ast.File, field *ast.Field) ([]metadata.FieldMeta, error) {
					return []metadata.FieldMeta{
						{SymNodeMeta: metadata.SymNodeMeta{Name: "r1"}},
						{SymNodeMeta: metadata.SymNodeMeta{Name: "r2"}},
					}, nil
				},
			}
			_, err := astArb.GetFuncRetValMeta(tv, nil, nil, fn, nil)
			Expect(err).To(MatchError(ContainSubstring("is a multi-variable declaration")))
			Expect(err.Error()).To(ContainSubstring("MultiRet"))
		})

		It("Successfully returns single return value metadata", func() {
			fn := &ast.FuncDecl{
				Name: ast.NewIdent("OneRet"),
				Type: &ast.FuncType{
					Results: &ast.FieldList{
						List: []*ast.Field{{Type: ast.NewIdent("T")}},
					},
				},
			}
			fm := metadata.FieldMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:     "out",
					Node:     &ast.Ident{Name: "out"},
					PkgPath:  "p",
					FVersion: &gast.FileVersion{Path: "p", ModTime: time.Now(), Hash: "h"},
				},
				Type: metadata.TypeUsageMeta{},
			}
			tv := &testVisitor{
				VisitFieldFunc: func(pkg *packages.Package, file *ast.File, field *ast.Field) ([]metadata.FieldMeta, error) {
					return []metadata.FieldMeta{fm}, nil
				},
			}
			out, err := astArb.GetFuncRetValMeta(tv, nil, nil, fn, nil)
			Expect(err).To(BeNil())
			Expect(out).To(HaveLen(1))
			Expect(out[0].Name).To(Equal("out"))
			Expect(out[0].Ordinal).To(Equal(0))
		})
	})

	Context("GetImportType", func() {
		It("Returns None for Universe identifiers (builtins)", func() {
			ident := &ast.Ident{Name: "int"} // universe type
			tp, err := astArb.GetImportType(nil, ident)
			Expect(err).To(BeNil())
			Expect(tp).To(Equal(common.ImportTypeNone))
		})

		It("Returns Alias for selector expressions", func() {
			se := &ast.SelectorExpr{
				X:   ast.NewIdent("fmt"),
				Sel: ast.NewIdent("Println"),
			}
			tp, err := astArb.GetImportType(nil, se)
			Expect(err).To(BeNil())
			Expect(tp).To(Equal(common.ImportTypeAlias))
		})

		It("Dereferences pointers (StarExpr) and returns the underlying import type", func() {
			se := &ast.SelectorExpr{
				X:   ast.NewIdent("pkg"),
				Sel: ast.NewIdent("T"),
			}
			st := &ast.StarExpr{X: se}
			tp, err := astArb.GetImportType(nil, st)
			Expect(err).To(BeNil())
			Expect(tp).To(Equal(common.ImportTypeAlias))
		})

		It("Inspects array element types", func() {
			arr := &ast.ArrayType{Elt: ast.NewIdent("int")}
			tp, err := astArb.GetImportType(nil, arr)
			Expect(err).To(BeNil())
			Expect(tp).To(Equal(common.ImportTypeNone))
		})

		It("Inspects map value types and returns their import type", func() {
			mp := &ast.MapType{
				Key:   ast.NewIdent("string"),
				Value: &ast.SelectorExpr{X: ast.NewIdent("pkg"), Sel: ast.NewIdent("T")},
			}
			tp, err := astArb.GetImportType(nil, mp)
			Expect(err).To(BeNil())
			Expect(tp).To(Equal(common.ImportTypeNone))
		})

		It("Inspects channel element types", func() {
			ch := &ast.ChanType{Value: ast.NewIdent("int")}
			tp, err := astArb.GetImportType(nil, ch)
			Expect(err).To(BeNil())
			Expect(tp).To(Equal(common.ImportTypeNone))
		})

		It("Detects dot-imported idents via GetPackageFromDotImportedIdent and returns Dot", func() {
			// Create an AST file with a dot import of "fmt"
			file := &ast.File{
				Imports: []*ast.ImportSpec{
					{
						Name: ast.NewIdent("."),
						Path: &ast.BasicLit{Kind: token.STRING, Value: `"fmt"`},
					},
				},
			}
			ident := &ast.Ident{Name: "Println"} // fmt.Println should exist in stdlib types
			tp, err := astArb.GetImportType(file, ident)
			Expect(err).To(BeNil())
			Expect(tp).To(Equal(common.ImportTypeDot))
		})

		It("Propagates error when a dot-import path is excessively long", func() {
			// It's a tad difficult to get the underlying packages.Load call to return an error.
			// Here, we create a package pattern that exceeds the allowed maximum arg length so we can get coverage of the error branch
			tooLong := strings.Repeat("a", 200000) // > ARG_MAX

			file := &ast.File{
				Imports: []*ast.ImportSpec{
					{
						Name: ast.NewIdent("."),
						Path: &ast.BasicLit{
							Kind:  token.STRING,
							Value: fmt.Sprintf("%q", tooLong),
						},
					},
				},
			}
			ident := &ast.Ident{Name: "SomeType"}

			tp, err := astArb.GetImportType(file, ident)
			Expect(tp).To(Equal(common.ImportTypeNone))
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("pattern is too long"))
		})
	})

	Context("GetPackageFromDotImportedIdent", func() {
		It("Returns nil when there are no dot imports", func() {
			file := &ast.File{Imports: nil}
			ident := &ast.Ident{Name: "Whatever"}
			pkg, err := astArb.GetPackageFromDotImportedIdent(file, ident)
			Expect(err).To(BeNil())
			Expect(pkg).To(BeNil())
		})

		It("Returns the package when ident exists in a dot-imported package", func() {
			file := &ast.File{
				Imports: []*ast.ImportSpec{
					{
						Name: ast.NewIdent("."),
						Path: &ast.BasicLit{Kind: token.STRING, Value: `"fmt"`},
					},
				},
			}
			ident := &ast.Ident{Name: "Printf"} // fmt.Printf is present in fmt
			found, err := astArb.GetPackageFromDotImportedIdent(file, ident)
			Expect(err).To(BeNil())
			Expect(found).ToNot(BeNil())
			Expect(found.Types.Scope().Lookup("Printf")).ToNot(BeNil())
		})

		It("Returns a proper error when one or more package loads returns an internal error", func() {
			file := &ast.File{
				Imports: []*ast.ImportSpec{
					{Name: ast.NewIdent("."), Path: &ast.BasicLit{Kind: token.STRING, Value: `"nonexistent_pkg_xyz"`}},
					{Name: ast.NewIdent("."), Path: &ast.BasicLit{Kind: token.STRING, Value: `"fmt"`}},
				},
			}
			ident := &ast.Ident{Name: "Sprintf"}
			_, err := astArb.GetPackageFromDotImportedIdent(file, ident)
			Expect(err).To(MatchError(MatchRegexp(
				`encountered \d+ errors over \d+ package/s \(.+?\) during load -`,
			)))
		})

		It("Returns an error when PackagesFacade.GetPackage returns an error", func() {
			// It's a tad difficult to get the underlying packages.Load call to return an error.
			// Here, we create a package pattern that exceeds the allowed maximum arg length so we can get coverage of the error branch.
			// Same as the GetImportType test case
			tooLong := strings.Repeat("x", 200000) // definitely > ARG_MAX

			file := &ast.File{
				Imports: []*ast.ImportSpec{
					{
						Name: ast.NewIdent("."),
						Path: &ast.BasicLit{
							Kind:  token.STRING,
							Value: fmt.Sprintf("%q", tooLong),
						},
					},
				},
			}

			pkg, err := astArb.GetPackageFromDotImportedIdent(file, ast.NewIdent("SomeSymbol"))

			Expect(pkg).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("pattern is too long"))
		})

	})
})

func TestUnitAstArbitrator(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - AST Arbitrator (external)")
}
