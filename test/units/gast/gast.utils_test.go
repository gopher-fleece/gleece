package gast_test

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/go/packages"
)

const auxTypesPackageName = "github.com/gopher-fleece/gleece/test/units"

var typesPkgLoadOnly *packages.Package
var typesPkgFullSyntax *packages.Package

var _ = BeforeSuite(func() {
	typesPkgLoadOnly = utils.LoadPackageOrFail(auxTypesPackageName, packages.LoadFiles)
	typesPkgFullSyntax = utils.LoadPackageOrFail(auxTypesPackageName, packages.LoadAllSyntax)
})

var _ = Describe("Unit Tests - AST", func() {
	Context("IsFuncDeclReceiverForStruct", func() {
		It("Returns false if the given FuncDecl is not a receiver", func() {
			funcDecl := utils.GetFunctionFromPackageOrFail(typesPkgFullSyntax, "NotAReceiver")
			Expect(gast.IsFuncDeclReceiverForStruct("StructWithReceivers", funcDecl)).To(BeFalse())
		})

		It("Returns true if the given FuncDecl is a value receiver", func() {
			funcDecl := utils.GetFunctionFromPackageOrFail(typesPkgFullSyntax, "ValueReceiverForStructWithReceivers")
			Expect(gast.IsFuncDeclReceiverForStruct("StructWithReceivers", funcDecl)).To(BeTrue())
		})

		It("Returns true if the given FuncDecl is a pointer receiver", func() {
			funcDecl := utils.GetFunctionFromPackageOrFail(typesPkgFullSyntax, "PointerReceiverForStructWithReceivers")
			Expect(gast.IsFuncDeclReceiverForStruct("StructWithReceivers", funcDecl)).To(BeTrue())
		})

		It("Returns false if the receiver type is not Ident or SelectorExpr", func() {
			// This is a synthetic test and not expected to be encountered in the wild
			receiverType := &ast.SelectorExpr{
				X:   ast.NewIdent("pkg"),
				Sel: ast.NewIdent("Type"),
			}

			funcDecl := &ast.FuncDecl{
				Recv: &ast.FieldList{
					List: []*ast.Field{
						{
							Type: receiverType,
						},
					},
				},
			}

			Expect(gast.IsFuncDeclReceiverForStruct("StructWithReceivers", funcDecl)).To(BeFalse())
		})
	})

	Context("GetDefaultPackageAlias", func() {
		It("Returns alias when file has name", func() {
			file := &ast.File{Name: &ast.Ident{Name: "mypkg"}}
			alias, err := gast.GetDefaultPackageAlias(file)
			Expect(err).ToNot(HaveOccurred())
			Expect(alias).To(Equal(gast.GetDefaultPkgAliasByName("mypkg")))
		})

		It("Returns an error if the given ast.File has no name", func() {
			file := &ast.File{
				Name: nil,
			}
			alias, err := gast.GetDefaultPackageAlias(file)
			Expect(err).To(MatchError(ContainSubstring("source file does not have a name")))
			Expect(alias).To(BeEmpty())
		})
	})

	Context("GetFileFullPath", func() {
		It("Returns an error when given a nil file or fileSet", func() {
			_, err := gast.GetFileFullPath(nil, nil)
			Expect(err).To(MatchError(ContainSubstring("GetFileFullPath was provided nil file or fileSet")))
		})

		It("Returns an error when the position filename is empty", func() {
			fs := token.NewFileSet()
			// Create a token.File with empty filename
			f := fs.AddFile("", -1, 0)
			astFile := &ast.File{
				Name:    ast.NewIdent("mypkg"),
				Package: token.Pos(f.Base()),
			}

			_, err := gast.GetFileFullPath(astFile, fs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not determine full path"))
		})

		It("Returns an absolute path when the file has a filename in the file set", func() {
			tmpDir := GinkgoT().TempDir()
			goPath := filepath.Join(tmpDir, "afile.go")
			Expect(os.WriteFile(goPath, []byte("package testpkg\n"), 0644)).To(Succeed())

			fs := token.NewFileSet()
			// register the absolute path in the file set
			f := fs.AddFile(goPath, -1, len("package testpkg\n"))
			astFile := &ast.File{
				Name:    ast.NewIdent("testpkg"),
				Package: token.Pos(f.Base()),
			}

			got, err := gast.GetFileFullPath(astFile, fs)
			Expect(err).ToNot(HaveOccurred())
			absExpect, _ := filepath.Abs(goPath)
			Expect(got).To(Equal(absExpect))
		})
	})

	Context("GetFullPackageName", func() {
		It("Returns a correct error when given nil inputs", func() {
			// Calling with nils makes GetFileFullPath return a clear error path.
			_, err := gast.GetFullPackageName(nil, nil)
			Expect(err).To(MatchError(ContainSubstring("GetFileFullPath was provided nil file or fileSet")))
		})

		It("Returns a correct error if packages.Load fails", func() {
			// This will be a fake file path that doesn't exist
			fakeFilename := "/this/definitely/does/not/exist.go"

			// Create a token.FileSet and register the fake file with it
			fileSet := token.NewFileSet()
			f := fileSet.AddFile(fakeFilename, -1, 100)
			f.SetLines([]int{0}) // dummy line info

			// Assign a position within the file
			pos := f.Pos(0) // get a token.Pos that points inside this file

			// Create a fake *ast.File with the fake Pos
			astFile := &ast.File{
				Name:    ast.NewIdent("main"),
				Package: pos,
			}

			// Now call your function
			fullPkg, err := gast.GetFullPackageName(astFile, fileSet)
			Expect(err).To(
				Or(
					MatchError(ContainSubstring("no such file or directory")),
					MatchError(ContainSubstring("cannot find the path")),
				),
			)
			Expect(fullPkg).To(BeEmpty())
		})

		It("Returns empty string without error if no matching package file is found", func() {
			tmpDir := GinkgoT().TempDir()

			// Write a fake go.mod file to make packages.Load happy
			Expect(os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module testpkg\n"), 0644)).To(Succeed())

			// Write a Go file that will be parsed
			goFilePath := filepath.Join(tmpDir, "somefile.go")
			Expect(os.WriteFile(goFilePath, []byte(`package testpkg`), 0644)).To(Succeed())

			// Parse it just to get a valid *ast.File and token.FileSet
			fileSet := token.NewFileSet()
			parsedFile, err := parser.ParseFile(fileSet, goFilePath, nil, parser.PackageClauseOnly)
			Expect(err).To(BeNil())

			// Now lie about the file's position so it won't match anything in pkg.GoFiles
			f := fileSet.AddFile(filepath.Join(tmpDir, "otherfile.go"), -1, 100)
			f.SetLines([]int{0})
			parsedFile.Package = f.Pos(0)

			pkgName, err := gast.GetFullPackageName(parsedFile, fileSet)
			Expect(err).To(BeNil())       // packages.Load succeeded
			Expect(pkgName).To(BeEmpty()) // but file wasn't found in the loaded package
		})

		It("Returns the package path when packages.Load contains the file", func() {
			tmpDir := GinkgoT().TempDir()
			modulePath := "example.com/testmod"

			// create go.mod so packages.Load treats tmpDir as a module root
			Expect(os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module "+modulePath+"\n"), 0644)).To(Succeed())

			// create a Go file in the module root
			goFilePath := filepath.Join(tmpDir, "somefile.go")
			Expect(os.WriteFile(goFilePath, []byte("package testpkg\n"), 0644)).To(Succeed())

			// parse it to obtain *ast.File and *token.FileSet where positions reference the actual filename
			fileSet := token.NewFileSet()
			parsedFile, err := parser.ParseFile(fileSet, goFilePath, nil, parser.PackageClauseOnly)
			Expect(err).ToNot(HaveOccurred())
			Expect(parsedFile).ToNot(BeNil())

			// run the function under test
			pkgPath, err := gast.GetFullPackageName(parsedFile, fileSet)
			Expect(err).ToNot(HaveOccurred())

			// packages.Load should return the module root path as the package path for a file at the module root
			Expect(pkgPath).To(Equal(modulePath))
		})

	})

	Context("IsPackageDotImported", func() {
		It("Returns true when an import spec has a non-nil Name and matching path", func() {
			// Note: the implementation checks imp.Name != nil (so treat this as the 'dot-like' case per the code).
			file := &ast.File{
				Imports: []*ast.ImportSpec{
					{
						// Name is non-nil per the implementation's check
						Name: ast.NewIdent("."),
						Path: &ast.BasicLit{Kind: token.STRING, Value: `"github.com/example/pkg"`},
					},
				},
			}

			found, pkg := gast.IsPackageDotImported(file, "github.com/example/pkg")
			Expect(found).To(BeTrue())
			Expect(pkg).To(Equal("github.com/example/pkg"))
		})

		It("Returns false when import names/paths don't match or Name is nil", func() {
			// Case: Name is nil -> code path requires Name != nil, so should return false
			file1 := &ast.File{
				Imports: []*ast.ImportSpec{
					{
						Name: nil,
						Path: &ast.BasicLit{Kind: token.STRING, Value: `"github.com/example/pkg"`},
					},
				},
			}
			found1, pkg1 := gast.IsPackageDotImported(file1, "github.com/example/pkg")
			Expect(found1).To(BeFalse())
			Expect(pkg1).To(Equal(""))

			// Case: Name non-nil but path doesn't match
			file2 := &ast.File{
				Imports: []*ast.ImportSpec{
					{
						Name: ast.NewIdent("."),
						Path: &ast.BasicLit{Kind: token.STRING, Value: `"github.com/other/pkg"`},
					},
				},
			}
			found2, pkg2 := gast.IsPackageDotImported(file2, "github.com/example/pkg")
			Expect(found2).To(BeFalse())
			Expect(pkg2).To(Equal(""))
		})
	})

	Context("GetDefaultPkgAliasByName", func() {
		It("Returns the last segment for a typical import path", func() {
			Expect(gast.GetDefaultPkgAliasByName("github.com/user/pkg")).To(Equal("pkg"))
		})

		It("Returns the second-last segment when the last segment is a version (v2/v10 etc.)", func() {
			Expect(gast.GetDefaultPkgAliasByName("github.com/user/pkg/v2")).To(Equal("pkg"))
			Expect(gast.GetDefaultPkgAliasByName("github.com/user/pkg/v10")).To(Equal("pkg"))
		})

		It("Handles single-segment edge cases (returns the segment itself)", func() {
			// single segment "v2" should return "v2" because len(segments) == 1
			Expect(gast.GetDefaultPkgAliasByName("v2")).To(Equal("v2"))
		})
	})

	Context("FindTypesStructInPackage", func() {
		It("Returns correct error when package does not have type information", func() {
			typeName, err := gast.FindTypesStructInPackage(typesPkgLoadOnly, "StructA")
			Expect(err).To(MatchError(ContainSubstring("does not have types or types scope")))
			Expect(typeName).To(BeNil())
		})

		It("Returns no value and no error when given name does not exist in given package", func() {
			typeName, err := gast.FindTypesStructInPackage(typesPkgFullSyntax, "ThisNameDoesNotExistInThisPackage")
			Expect(err).To(BeNil())
			Expect(typeName).To(BeNil())
		})

		It("Returns correct error when given name exist in given package but is not a struct", func() {
			typeName, err := gast.FindTypesStructInPackage(typesPkgFullSyntax, "InterfaceA")
			Expect(err).To(MatchError(ContainSubstring("is not a struct type")))
			Expect(typeName).To(BeNil())
		})
	})

	Context("LookupTypeName", func() {
		It("Returns correct error when given package has no type information", func() {
			typeName, err := gast.LookupTypeName(typesPkgLoadOnly, "NA")
			Expect(typeName).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("does not have types or types scope")))
		})

		It("Returns no value and no error when given name does not exist in given package", func() {
			typeName, err := gast.LookupTypeName(typesPkgFullSyntax, "ThisNameDoesNotExistInThisPackage")
			Expect(typeName).To(BeNil())
			Expect(err).To(BeNil())
		})

		It("Returns no value and no error when given name exist in given package but is not a TypeName", func() {
			typeName, err := gast.LookupTypeName(typesPkgFullSyntax, "ConstA")
			Expect(typeName).To(BeNil())
			Expect(err).To(BeNil())
		})

		It("Returns correct value when given name exist in given package and is a TypeName", func() {
			typeName, err := gast.LookupTypeName(typesPkgFullSyntax, "StructA")
			Expect(err).To(BeNil())
			Expect(typeName.Name()).ToNot(BeNil())

			typeName, err = gast.LookupTypeName(typesPkgFullSyntax, "InterfaceA")
			Expect(err).To(BeNil())
			Expect(typeName.Name()).ToNot(BeNil())

			typeName, err = gast.LookupTypeName(typesPkgFullSyntax, "EnumTypeA")
			Expect(err).To(BeNil())
			Expect(typeName.Name()).ToNot(BeNil())
		})
	})

	Context("GetTypeNameOrError", func() {
		It("Returns correct error when package does not have type information", func() {
			typeName, err := gast.GetTypeNameOrError(typesPkgLoadOnly, "StructA")
			Expect(err).To(MatchError(ContainSubstring("does not have types or types scope")))
			Expect(typeName).To(BeNil())
		})

		It("Returns correct error given name does not exist in given package", func() {
			typeName, err := gast.GetTypeNameOrError(typesPkgFullSyntax, "ThisNameDoesNotExistInThisPackage")
			Expect(err).To(MatchError(ContainSubstring("was not found in package")))
			Expect(typeName).To(BeNil())
		})

		It("Returns a correct error when the found TypeName has nil Type()", func() {
			// Create a types.Package and insert a TypeName whose underlying type is nil.
			// This produces a *types.TypeName where Type() == nil (the branch we want).
			typesPkg := types.NewPackage("example.com/fake", "fakepkg")
			tn := types.NewTypeName(token.NoPos, typesPkg, "BadTypeWithNoType", nil)
			typesPkg.Scope().Insert(tn)

			// Create a fake *packages.Package that points to our types.Package.
			fakePkg := &packages.Package{
				Name:  "fakepkg",
				Types: typesPkg,
			}

			got, err := gast.GetTypeNameOrError(fakePkg, "BadTypeWithNoType")
			Expect(got).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("does not have Type() information")))
		})

	})

	Context("GetUnderlyingTypeName", func() {
		It("Correctly resolves basic and named types", func() {
			testStruct, err := gast.FindTypesStructInPackage(typesPkgFullSyntax, "StructForGetUnderlyingTypeName")
			Expect(err).To(BeNil())
			Expect(testStruct).ToNot(BeNil())

			tests := map[string]string{
				"FieldIntAlias":     "IntAlias",       // named type
				"FieldStringPtr":    "*string",        // pointer
				"FieldIntSlice":     "[]int",          // slice
				"FieldStringIntMap": "map[string]int", // map
				"FieldChannelInt":   "chan int",       // channel
				"FieldFunc":         "func(...)",      // function signature
				"FieldIntArray":     "[3]int",         // array
				"FieldInterface":    "any",            // fallback
				"FieldStruct":       "struct{}",
				"FieldInt":          "int",
				"FieldComment":      "Comment",
			}

			for i := range testStruct.NumFields() {
				field := testStruct.Field(i)
				name := field.Name()
				typ := field.Type()

				expected := tests[name]
				actual := gast.GetUnderlyingTypeName(typ)
				Expect(actual).To(Equal(expected), fmt.Sprintf("field %s", name))
			}
		})
	})

	Context("GetFieldTypeString", func() {
		It("Correctly parses field AST expressions into string descriptions", func() {
			typeStruct, err := gast.FindTypesStructInPackage(typesPkgFullSyntax, "StructForGetUnderlyingTypeName")
			Expect(err).To(BeNil())
			Expect(typeStruct).ToNot(BeNil())

			expected := map[string]string{
				"FieldIntAlias":     "IntAlias",
				"FieldStringPtr":    "*string",
				"FieldIntSlice":     "[]int",
				"FieldIntArray":     "[3]int",
				"FieldStringIntMap": "map[string]int",
				"FieldChannelInt":   "Channel (bidirectional, type: int)",
				"FieldFunc":         "Function",
				"FieldInterface":    "any",
				"FieldStruct":       "Struct",
				"FieldInt":          "Parenthesized (int)",
				"FieldComment":      "ast.Comment",
			}

			for i := range typeStruct.NumFields() {
				field := typeStruct.Field(i)
				fieldName := field.Name()

				expectedStr, ok := expected[fieldName]
				if !ok {
					continue // skip fields we're not explicitly testing
				}

				astField := utils.GetAstFieldByNameOrFail(typesPkgFullSyntax, "StructForGetUnderlyingTypeName", fieldName)
				actualStr := gast.GetFieldTypeString(astField.Type)

				Expect(actualStr).To(Equal(expectedStr), fmt.Sprintf("Mismatch on field %q", fieldName))
			}
		})

		It("Correctly parses variadic function parameters", func() {
			funcDecl := utils.GetFunctionFromPackageOrFail(typesPkgFullSyntax, "SimpleVariadicFunc")
			params := funcDecl.Type.Params.List
			if params == nil || len(params) != 1 {
				utils.FailWithTestCodeError("Expected test variadic function 'SimpleVariadicFunc' to have exactly one argument")
			}

			ellipsisExpr := params[0].Type
			if ellipsisExpr == nil {
				utils.FailWithTestCodeError("Function argument's type is not an ellipsis expression")
			}

			Expect(gast.GetFieldTypeString(ellipsisExpr)).To(Equal("Variadic (...int)"))
		})

		It("Correctly falls back for unsupported array length expressions", func() {
			arrayExpr := &ast.ArrayType{
				Len: &ast.BinaryExpr{
					X:  &ast.BasicLit{Kind: token.INT, Value: "1"},
					Op: token.ADD,
					Y:  &ast.BasicLit{Kind: token.INT, Value: "2"},
				},
				Elt: &ast.Ident{Name: "int"},
			}

			result := gast.GetFieldTypeString(arrayExpr)
			Expect(result).To(Equal("[?]int"))
		})

		It("Returns send-only channel type string", func() {
			channel := &ast.ChanType{
				Dir:   ast.SEND,
				Value: &ast.Ident{Name: "int"},
			}
			Expect(gast.GetFieldTypeString(channel)).To(Equal("Channel (send-only, type: int)"))
		})

		It("Returns receive-only channel type string", func() {
			channel := &ast.ChanType{
				Dir:   ast.RECV,
				Value: &ast.Ident{Name: "string"},
			}
			Expect(gast.GetFieldTypeString(channel)).To(Equal("Channel (receive-only, type: string)"))
		})

		It("Returns 'Interface' for *ast.InterfaceType", func() {
			iFace := &ast.InterfaceType{}
			Expect(gast.GetFieldTypeString(iFace)).To(Equal("Interface"))
		})

		It("Returns Unknown type for unhandled AST type", func() {
			// ast.BadExpr is never specifically handled, so it will hit the default branch.
			bad := &ast.BadExpr{}
			got := gast.GetFieldTypeString(bad)
			Expect(got).To(MatchRegexp(`Unknown type \(\*ast\.BadExpr\)`))
		})
	})

	Context("GetIdentFromExpr", func() {
		It("Returns the ident when given an *ast.Ident", func() {
			id := &ast.Ident{Name: "x"}
			Expect(gast.GetIdentFromExpr(id)).To(Equal(id))
		})

		It("Unwraps a star expression", func() {
			id := &ast.Ident{Name: "x"}
			Expect(gast.GetIdentFromExpr(&ast.StarExpr{X: id})).To(Equal(id))
		})

		It("Returns selector's Sel for SelectorExpr", func() {
			sel := &ast.Ident{Name: "Baz"}
			se := &ast.SelectorExpr{X: ast.NewIdent("pkg"), Sel: sel}
			Expect(gast.GetIdentFromExpr(se)).To(Equal(sel))
		})

		It("Unwraps array element type", func() {
			id := &ast.Ident{Name: "x"}
			arr := &ast.ArrayType{Elt: id}
			Expect(gast.GetIdentFromExpr(arr)).To(Equal(id))
		})

		It("Returns value ident for map types", func() {
			id := &ast.Ident{Name: "val"}
			m := &ast.MapType{Key: &ast.Ident{Name: "k"}, Value: id}
			Expect(gast.GetIdentFromExpr(m)).To(Equal(id))
		})

		It("Returns value ident for channel types", func() {
			id := &ast.Ident{Name: "c"}
			ch := &ast.ChanType{Value: id}
			Expect(gast.GetIdentFromExpr(ch)).To(Equal(id))
		})

		It("Returns nil for function types", func() {
			Expect(gast.GetIdentFromExpr(&ast.FuncType{})).To(BeNil())
		})

		It("Returns nil for unknown/unsupported expressions", func() {
			Expect(gast.GetIdentFromExpr(&ast.BadExpr{})).To(BeNil())
		})
	})

	Context("ResolveTypeSpecFromExpr", func() {
		It("Returns error when ident is nil", func() {
			// expr that has no base ident (GetIdentFromExpr will return nil)
			expr := &ast.StructType{}

			pkg := &packages.Package{
				TypesInfo: &types.Info{
					Uses: map[*ast.Ident]types.Object{},
				},
			}

			res, err := gast.ResolveTypeSpecFromExpr(pkg, &ast.File{Name: ast.NewIdent("f")}, expr, func(string) (*packages.Package, error) {
				return nil, nil
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("expression has no base identifier"))
			Expect(res).To(Equal(gast.TypeSpecResolution{}))
		})

		It("Returns error when identifier not found in Uses or Scope", func() {
			ident := ast.NewIdent("MissingType")

			pkg := &packages.Package{
				TypesInfo: &types.Info{
					Uses: map[*ast.Ident]types.Object{}, // empty -> not found in Uses
				},
				Types: types.NewPackage("mypkg/path", "mypkg"),
			}

			res, err := gast.ResolveTypeSpecFromExpr(pkg, &ast.File{Name: ast.NewIdent("f")}, ident, func(string) (*packages.Package, error) {
				return nil, nil
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cannot resolve identifier"))
			Expect(res).To(Equal(gast.TypeSpecResolution{}))
		})

		It("Returns error when resolved object is not a TypeName", func() {
			declPkg := types.NewPackage("pkg/path", "pkg")
			nonTypeObj := types.NewVar(token.NoPos, declPkg, "NotAType", types.Typ[types.Int])

			ident := ast.NewIdent("NotAType")
			pkg := &packages.Package{
				TypesInfo: &types.Info{
					Uses: map[*ast.Ident]types.Object{
						ident: nonTypeObj,
					},
				},
			}

			res, err := gast.ResolveTypeSpecFromExpr(pkg, &ast.File{Name: ast.NewIdent("f")}, ident, func(string) (*packages.Package, error) {
				return nil, nil
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resolved object is not a type"))
			Expect(res).To(Equal(gast.TypeSpecResolution{}))
		})

		It("Returns universe type when object comes from types.Universe", func() {
			// Use the universe 'int' TypeName
			obj := types.Universe.Lookup("int").(*types.TypeName)
			ident := ast.NewIdent("int")

			pkg := &packages.Package{
				TypesInfo: &types.Info{
					Uses: map[*ast.Ident]types.Object{
						ident: obj,
					},
				},
			}

			res, err := gast.ResolveTypeSpecFromExpr(pkg, &ast.File{Name: ast.NewIdent("f")}, ident, func(string) (*packages.Package, error) {
				return nil, nil
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(res.IsUniverse).To(BeTrue())
			Expect(res.TypeName).To(Equal("int"))
		})

		It("Returns error when getPkg fails to locate declaring package", func() {
			// Create a TypeName that points to a package path; getPkg will fail
			declPkgObj := types.NewPackage("decl/pg", "decl")
			obj := types.NewTypeName(token.NoPos, declPkgObj, "MyType", types.Typ[types.Int])
			ident := ast.NewIdent("MyType")

			pkg := &packages.Package{
				TypesInfo: &types.Info{
					Uses: map[*ast.Ident]types.Object{
						ident: obj,
					},
				},
			}

			res, err := gast.ResolveTypeSpecFromExpr(pkg, &ast.File{Name: ast.NewIdent("f")}, ident, func(path string) (*packages.Package, error) {
				return nil, fmt.Errorf("fail-resolve")
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not locate declaring package"))
			Expect(res).To(Equal(gast.TypeSpecResolution{}))
		})

		It("Returns TypeSpecResolution on happy path when TypeSpec found in declaring package", func() {
			// Declare a TypeSpec in a synthetic AST file
			typeSpec := &ast.TypeSpec{
				Name: ast.NewIdent("MyType"),
				Type: &ast.StructType{},
			}
			genDecl := &ast.GenDecl{
				Tok:   token.TYPE,
				Specs: []ast.Spec{typeSpec},
			}
			declFile := &ast.File{
				Name:  ast.NewIdent("declfile"),
				Decls: []ast.Decl{genDecl},
			}

			// Create a TypeName which claims to be declared in 'decl/pg'
			declPkgObj := types.NewPackage("decl/pg", "decl")
			obj := types.NewTypeName(token.NoPos, declPkgObj, "MyType", types.Typ[types.Int])
			ident := ast.NewIdent("MyType")

			mainPkg := &packages.Package{
				TypesInfo: &types.Info{
					Uses: map[*ast.Ident]types.Object{
						ident: obj,
					},
				},
			}

			declPkg := &packages.Package{
				Syntax: []*ast.File{declFile},
			}

			res, err := gast.ResolveTypeSpecFromExpr(mainPkg, &ast.File{Name: ast.NewIdent("f")}, ident, func(path string) (*packages.Package, error) {
				// Only accept the path used above
				if path == "decl/pg" {
					return declPkg, nil
				}
				return nil, fmt.Errorf("not found")
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(res.IsUniverse).To(BeFalse())
			Expect(res.TypeName).To(Equal("MyType"))
			Expect(res.TypeSpec).To(Equal(typeSpec))
			Expect(res.GenDecl).To(Equal(genDecl))
			Expect(res.DeclaringPackage).To(Equal(declPkg))
			Expect(res.DeclaringAstFile).To(Equal(declFile))
		})

		It("Returns error when TypeSpec not found in declaring package", func() {
			// TypeName pointing to decl/pkg but the declaring package has no matching TypeSpec
			declPkgObj := types.NewPackage("decl2/pg", "decl2")
			obj := types.NewTypeName(token.NoPos, declPkgObj, "MissingType", types.Typ[types.Int])
			ident := ast.NewIdent("MissingType")

			mainPkg := &packages.Package{
				TypesInfo: &types.Info{
					Uses: map[*ast.Ident]types.Object{
						ident: obj,
					},
				},
			}

			declPkg := &packages.Package{
				Syntax: []*ast.File{ // file with no type decls
					{Name: ast.NewIdent("empty")},
				},
			}

			res, err := gast.ResolveTypeSpecFromExpr(mainPkg, &ast.File{Name: ast.NewIdent("f")}, ident, func(path string) (*packages.Package, error) {
				if path == "decl2/pg" {
					return declPkg, nil
				}
				return nil, fmt.Errorf("not found")
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not find TypeSpec for type"))
			Expect(res).To(Equal(gast.TypeSpecResolution{}))
		})

		It("Continues when a GenDecl.Spec is not a *ast.TypeSpec (covers inner continue)", func() {
			// Setup a TypeName that claims to be declared in "decl3/pg"
			declPkgObj := types.NewPackage("decl3/pg", "decl3")
			obj := types.NewTypeName(token.NoPos, declPkgObj, "MissingTypeViaValueSpec", types.Typ[types.Int])
			ident := ast.NewIdent("MissingTypeViaValueSpec")

			// mainPkg where Uses maps ident -> obj (so obj != nil and we reach the search in declPkg.Syntax)
			mainPkg := &packages.Package{
				TypesInfo: &types.Info{
					Uses: map[*ast.Ident]types.Object{
						ident: obj,
					},
				},
			}

			// Create a GenDecl with Tok == token.TYPE but its Specs contain a non-TypeSpec
			// (ValueSpec implements ast.Spec, so the cast to *ast.TypeSpec will fail and trigger the continue).
			valueSpec := &ast.ValueSpec{
				Names:  []*ast.Ident{ast.NewIdent("someConst")},
				Type:   &ast.Ident{Name: "int"},
				Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}},
			}
			genDecl := &ast.GenDecl{
				Tok:   token.TYPE, // must be TYPE so the outer check passes
				Specs: []ast.Spec{valueSpec},
			}
			declFile := &ast.File{
				Name:  ast.NewIdent("declfile"),
				Decls: []ast.Decl{genDecl},
			}

			declPkg := &packages.Package{
				Syntax: []*ast.File{declFile},
			}

			// getPkg returns the declaring package for the object's package path
			res, err := gast.ResolveTypeSpecFromExpr(mainPkg, &ast.File{Name: ast.NewIdent("f")}, ident, func(path string) (*packages.Package, error) {
				if path == "decl3/pg" {
					return declPkg, nil
				}
				return nil, fmt.Errorf("not found")
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not find TypeSpec for type"))
			Expect(res).To(Equal(gast.TypeSpecResolution{}))
		})
	})

	Context("ExtractConstValue", func() {

		It("Returns string value for types.String", func() {
			cv := constant.MakeFromLiteral(`"hello"`, token.STRING, 0)
			c := types.NewConst(token.NoPos, types.NewPackage("p", "p"), "C", types.Typ[types.String], cv)
			got := gast.ExtractConstValue(types.String, c)
			Expect(got).To(Equal("hello"))
		})

		It("Returns int64 for small Int kinds", func() {
			cv := constant.MakeInt64(42)
			c := types.NewConst(token.NoPos, types.NewPackage("p", "p"), "C", types.Typ[types.Int], cv)
			got := gast.ExtractConstValue(types.Int, c)
			Expect(got).To(Equal(int64(42)))
		})

		It("Returns nil for Int when value cannot be represented as int64", func() {
			// value > max int64
			cv := constant.MakeFromLiteral("9223372036854775808", token.INT, 0)
			c := types.NewConst(token.NoPos, types.NewPackage("p", "p"), "C", types.Typ[types.Int64], cv)
			got := gast.ExtractConstValue(types.Int64, c)
			Expect(got).To(BeNil())
		})

		It("Returns uint64 for small Uint kinds", func() {
			// Use a literal that is representable as uint64
			cv := constant.MakeFromLiteral("42", token.INT, 0)
			c := types.NewConst(token.NoPos, types.NewPackage("p", "p"), "UC", types.Typ[types.Uint], cv)
			got := gast.ExtractConstValue(types.Uint, c)
			Expect(got).To(Equal(uint64(42)))
		})

		It("Returns nil for Uint when value cannot be represented as uint64", func() {
			// value > max uint64
			cv := constant.MakeFromLiteral("18446744073709551616", token.INT, 0)
			c := types.NewConst(token.NoPos, types.NewPackage("p", "p"), "UC", types.Typ[types.Uint64], cv)
			got := gast.ExtractConstValue(types.Uint64, c)
			Expect(got).To(BeNil())
		})

		It("Returns float64 for Float kinds", func() {
			// deterministically construct the float constant
			cv := constant.MakeFromLiteral("3.14", token.FLOAT, 0)
			c := types.NewConst(token.NoPos, types.NewPackage("p", "p"), "F", types.Typ[types.Float64], cv)
			got := gast.ExtractConstValue(types.Float64, c)

			// Use numeric comparison with tolerance for safety
			Expect(got).ToNot(BeNil())
			Expect(got).To(BeNumerically("~", 3.14, 1e-9))
		})

		It("Returns nil for Float when value cannot be represented as float64", func() {
			// an extreme float that won't fit into a float64
			cv := constant.MakeFromLiteral("1e5000", token.FLOAT, 0)
			c := types.NewConst(token.NoPos, types.NewPackage("p", "p"), "F", types.Typ[types.Float64], cv)
			got := gast.ExtractConstValue(types.Float64, c)
			Expect(got).To(BeNil())
		})

		It("Returns bool for Bool kind", func() {
			// Use MakeBool to avoid nil constant.Value
			cv := constant.MakeBool(true)
			c := types.NewConst(token.NoPos, types.NewPackage("p", "p"), "B", types.Typ[types.Bool], cv)
			got := gast.ExtractConstValue(types.Bool, c)
			Expect(got).To(Equal(true))
		})

		It("Returns nil for unsupported basic kinds (default case)", func() {
			// Complex64 is not handled by the switch -> should fall through to nil
			cv := constant.MakeFromLiteral("1+2i", token.IMAG, 0)
			c := types.NewConst(token.NoPos, types.NewPackage("p", "p"), "Z", types.Typ[types.Complex64], cv)
			got := gast.ExtractConstValue(types.Complex64, c)
			Expect(got).To(BeNil())
		})
	})

	Context("FindConstSpecNode", func() {
		var pkg *packages.Package

		BeforeEach(func() {
			// Setup a dummy package with AST syntax for testing
			pkg = &packages.Package{
				Syntax: []*ast.File{},
			}
		})

		It("Returns nil if no files in package", func() {
			// No syntax files => no constants found
			Expect(gast.FindConstSpecNode(pkg, "MyConst")).To(BeNil())
		})

		It("Skips declarations that are not GenDecl", func() {
			file := &ast.File{
				Decls: []ast.Decl{
					// Use a fake decl type that is not *ast.GenDecl
					&ast.FuncDecl{},
				},
			}
			pkg.Syntax = []*ast.File{file}

			Expect(gast.FindConstSpecNode(pkg, "MyConst")).To(BeNil())
		})

		It("Skips GenDecls that are not const", func() {
			file := &ast.File{
				Decls: []ast.Decl{
					&ast.GenDecl{
						Tok: token.VAR, // not CONST
					},
				},
			}
			pkg.Syntax = []*ast.File{file}

			Expect(gast.FindConstSpecNode(pkg, "MyConst")).To(BeNil())
		})

		It("Skips specs that are not ValueSpec", func() {
			file := &ast.File{
				Decls: []ast.Decl{
					&ast.GenDecl{
						Tok: token.CONST,
						Specs: []ast.Spec{
							// fake spec that is not *ast.ValueSpec
							&ast.ImportSpec{},
						},
					},
				},
			}
			pkg.Syntax = []*ast.File{file}

			Expect(gast.FindConstSpecNode(pkg, "MyConst")).To(BeNil())
		})

		It("Returns the ValueSpec node when a const with matching name is found", func() {
			constName := "MyConst"
			valSpec := &ast.ValueSpec{
				Names: []*ast.Ident{
					ast.NewIdent(constName),
				},
			}
			file := &ast.File{
				Decls: []ast.Decl{
					&ast.GenDecl{
						Tok:   token.CONST,
						Specs: []ast.Spec{valSpec},
					},
				},
			}
			pkg.Syntax = []*ast.File{file}

			found := gast.FindConstSpecNode(pkg, constName)
			Expect(found).To(Equal(valSpec))
		})

		It("Returns nil when const with given name is not found", func() {
			valSpec := &ast.ValueSpec{
				Names: []*ast.Ident{
					ast.NewIdent("OtherConst"),
				},
			}
			file := &ast.File{
				Decls: []ast.Decl{
					&ast.GenDecl{
						Tok:   token.CONST,
						Specs: []ast.Spec{valSpec},
					},
				},
			}
			pkg.Syntax = []*ast.File{file}

			Expect(gast.FindConstSpecNode(pkg, "MyConst")).To(BeNil())
		})
	})

	Context("GetCommentsFromNode", func() {
		It("Returns nil when node is nil", func() {
			Expect(gast.GetCommentsFromNode(nil)).To(BeNil())
		})

		It("Returns comments for *ast.Field with Doc", func() {
			field := &ast.Field{
				Doc: &ast.CommentGroup{
					List: []*ast.Comment{{Text: "// field comment"}},
				},
			}
			Expect(gast.GetCommentsFromNode(field)).To(Equal([]string{"// field comment"}))
		})

		It("Returns nil for *ast.FuncDecl without Doc", func() {
			funcDecl := &ast.FuncDecl{}
			Expect(gast.GetCommentsFromNode(funcDecl)).To(BeNil())
		})

		It("Returns nil for unsupported node types", func() {
			badNode := &ast.BasicLit{}
			Expect(gast.GetCommentsFromNode(badNode)).To(BeNil())
		})
	})

	Context("GetCommentsFromTypeSpec", func() {
		It("Returns typeSpec.Doc comments when present", func() {
			typeSpec := &ast.TypeSpec{
				Doc: &ast.CommentGroup{
					List: []*ast.Comment{{Text: "// type comment"}},
				},
			}
			Expect(gast.GetCommentsFromTypeSpec(typeSpec, nil)).To(Equal([]string{"// type comment"}))
		})

		It("Returns owningGenDecl.Doc comments when typeSpec.Doc is nil", func() {
			typeSpec := &ast.TypeSpec{}
			genDecl := &ast.GenDecl{
				Doc: &ast.CommentGroup{
					List: []*ast.Comment{{Text: "// genDecl comment"}},
				},
			}
			Expect(gast.GetCommentsFromTypeSpec(typeSpec, genDecl)).To(Equal([]string{"// genDecl comment"}))
		})

		It("Returns empty slice when neither Doc is present", func() {
			typeSpec := &ast.TypeSpec{}
			genDecl := &ast.GenDecl{}
			Expect(gast.GetCommentsFromTypeSpec(typeSpec, genDecl)).To(Equal([]string{}))
		})
	})

	Context("DoesStructEmbedType", func() {
		It("Returns error when struct is not found in package", func() {
			typesPkg := types.NewPackage("example.com/mypkg", "mypkg")
			pkg := &packages.Package{
				PkgPath: "example.com/mypkg",
				Types:   typesPkg,
			}

			ok, err := gast.DoesStructEmbedType(pkg, "Missing", "other/pkg", "Embed")
			Expect(ok).To(BeFalse())
			Expect(err).To(MatchError(ContainSubstring("struct 'Missing' not found in package 'example.com/mypkg'")))
		})

		It("Returns error when found object is not a named type", func() {
			typesPkg := types.NewPackage("example.com/mypkg", "mypkg")

			// Insert a var (not a type) with the same name
			varObj := types.NewVar(token.NoPos, typesPkg, "NotNamed", types.Typ[types.Int])
			typesPkg.Scope().Insert(varObj)

			pkg := &packages.Package{
				PkgPath: "example.com/mypkg",
				Types:   typesPkg,
			}

			ok, err := gast.DoesStructEmbedType(pkg, "NotNamed", "", "Embed")
			Expect(ok).To(BeFalse())
			Expect(err).To(MatchError(ContainSubstring("type 'NotNamed' is not a named type")))
		})

		It("Returns error when named type's underlying is not a struct", func() {
			typesPkg := types.NewPackage("example.com/mypkg", "mypkg")

			tn := types.NewTypeName(token.NoPos, typesPkg, "MyAlias", nil)
			named := types.NewNamed(tn, types.Typ[types.Int], nil) // underlying is int
			// The NewNamed call automatically associates tn -> named

			typesPkg.Scope().Insert(named.Obj())

			pkg := &packages.Package{
				PkgPath: "example.com/mypkg",
				Types:   typesPkg,
			}

			ok, err := gast.DoesStructEmbedType(pkg, "MyAlias", "", "Embed")
			Expect(ok).To(BeFalse())
			Expect(err).To(MatchError(ContainSubstring("type 'MyAlias' is not a struct")))
		})

		It("Returns true when a field embeds the target type using fully-qualified package path", func() {
			// Create the embedding type in a separate package (some/pkg.Embed)
			embedPkg := types.NewPackage("some/pkg", "pkg")
			embedTn := types.NewTypeName(token.NoPos, embedPkg, "Embed", nil)
			embedNamed := types.NewNamed(embedTn, types.Typ[types.Int], nil)
			embedPkg.Scope().Insert(embedTn)

			// Create the struct type that embeds some/pkg.Embed
			typesPkg := types.NewPackage("example.com/mypkg", "mypkg")
			// embedded field: anonymous field of type some/pkg.Embed
			embeddedField := types.NewField(token.NoPos, embedPkg, "", embedNamed, true)
			structType := types.NewStruct([]*types.Var{embeddedField}, nil)

			tn := types.NewTypeName(token.NoPos, typesPkg, "HasEmbed", nil)
			_ = types.NewNamed(tn, structType, nil) // link TypeName -> Named
			typesPkg.Scope().Insert(tn)

			pkg := &packages.Package{
				PkgPath: "example.com/mypkg",
				Types:   typesPkg,
			}

			ok, err := gast.DoesStructEmbedType(pkg, "HasEmbed", "some/pkg", "Embed")
			Expect(err).ToNot(HaveOccurred())
			Expect(ok).To(BeTrue())
		})

		It("Returns false when embedded field does not match the requested type", func() {
			// Create a struct with a non-embedded field
			typesPkg := types.NewPackage("example.com/mypkg", "mypkg")
			nonEmbeddedField := types.NewVar(token.NoPos, typesPkg, "F", types.Typ[types.Int])
			structType := types.NewStruct([]*types.Var{nonEmbeddedField}, nil)

			tn := types.NewTypeName(token.NoPos, typesPkg, "NoEmbed", nil)
			_ = types.NewNamed(tn, structType, nil)
			typesPkg.Scope().Insert(tn)

			pkg := &packages.Package{
				PkgPath: "example.com/mypkg",
				Types:   typesPkg,
			}

			ok, err := gast.DoesStructEmbedType(pkg, "NoEmbed", "", "Embed")
			Expect(err).ToNot(HaveOccurred())
			Expect(ok).To(BeFalse())
		})
	})

	// Context block for IsEnumLike
	Context("IsEnumLike", func() {
		It("Returns false when spec.Type is not an *ast.Ident", func() {
			pkg := &packages.Package{Types: types.NewPackage("example.com/mypkg", "mypkg")}
			spec := &ast.TypeSpec{
				Name: ast.NewIdent("X"),
				Type: &ast.StructType{}, // not an Ident -> early false
			}
			Expect(gast.IsEnumLike(pkg, spec)).To(BeFalse())
		})

		It("Returns false when package lookup is not a *types.TypeName", func() {
			typesPkg := types.NewPackage("example.com/mypkg", "mypkg")
			// put a var named MyEnum in scope (not a TypeName)
			typesPkg.Scope().Insert(types.NewVar(token.NoPos, typesPkg, "MyEnum", types.Typ[types.Int]))

			pkg := &packages.Package{Types: typesPkg}
			spec := &ast.TypeSpec{
				Name: ast.NewIdent("MyEnum"),
				Type: ast.NewIdent("string"),
			}
			Expect(gast.IsEnumLike(pkg, spec)).To(BeFalse())
		})

		It("Returns false when underlying type is not basic", func() {
			typesPkg := types.NewPackage("example.com/mypkg", "mypkg")
			tn := types.NewTypeName(token.NoPos, typesPkg, "MyAlias", nil)
			named := types.NewNamed(tn, types.NewSlice(types.Typ[types.Int]), nil) // underlying is []int -> not basic
			_ = named                                                              // NewNamed links tn.Type -> named
			typesPkg.Scope().Insert(tn)

			// Also add a const of that same (non-basic) type to make sure the loop doesn't falsely match
			typesPkg.Scope().Insert(types.NewConst(token.NoPos, typesPkg, "C", named, constant.MakeInt64(1)))

			pkg := &packages.Package{Types: typesPkg}
			spec := &ast.TypeSpec{
				Name: ast.NewIdent("MyAlias"),
				Type: ast.NewIdent("[]int"),
			}
			Expect(gast.IsEnumLike(pkg, spec)).To(BeFalse())
		})

		It("Returns true when alias to a basic type has constants with matching alias type", func() {
			typesPkg := types.NewPackage("example.com/mypkg", "mypkg")
			tn := types.NewTypeName(token.NoPos, typesPkg, "MyEnum", nil)
			named := types.NewNamed(tn, types.Typ[types.String], nil)
			_ = named
			typesPkg.Scope().Insert(tn)

			// Insert a const whose type is the named alias
			typesPkg.Scope().Insert(types.NewConst(token.NoPos, typesPkg, "C1", named, constant.MakeInt64(1)))

			pkg := &packages.Package{Types: typesPkg}
			spec := &ast.TypeSpec{
				Name: ast.NewIdent("MyEnum"),
				Type: ast.NewIdent("string"),
			}

			Expect(gast.IsEnumLike(pkg, spec)).To(BeTrue())
		})

		It("Returns false when no constants of the alias type are present", func() {
			typesPkg := types.NewPackage("example.com/mypkg", "mypkg")
			tn := types.NewTypeName(token.NoPos, typesPkg, "MyEnum", nil)
			_ = types.NewNamed(tn, types.Typ[types.String], nil)
			typesPkg.Scope().Insert(tn)

			// Insert a const of a different (non-alias) type
			typesPkg.Scope().Insert(types.NewConst(token.NoPos, typesPkg, "Other", types.Typ[types.String], constant.MakeInt64(1)))

			pkg := &packages.Package{Types: typesPkg}
			spec := &ast.TypeSpec{
				Name: ast.NewIdent("MyEnum"),
				Type: ast.NewIdent("string"),
			}
			Expect(gast.IsEnumLike(pkg, spec)).To(BeFalse())
		})
	})

	Context("GetAstFileNameOrFallback", func() {
		When("File is nil", func() {
			It("returns the NIL fallback with provided fallback value", func() {
				fallback := "CUSTOM"
				result := gast.GetAstFileNameOrFallback(nil, &fallback)
				Expect(result).To(Equal("CUSTOM"))
			})

			It("returns the NIL fallback with generated value if fallback is nil", func() {
				result := gast.GetAstFileNameOrFallback(nil, nil)
				Expect(result).To(Equal("NIL_FILE"))
			})
		})

		When("File.Name is nil", func() {
			It("returns the UNNAMED fallback with provided fallback", func() {
				fallback := "X"
				file := &ast.File{Name: nil}
				result := gast.GetAstFileNameOrFallback(file, &fallback)
				Expect(result).To(Equal("X"))
			})

			It("returns the UNNAMED fallback with generated value if fallback is nil", func() {
				file := &ast.File{Name: nil}
				result := gast.GetAstFileNameOrFallback(file, nil)
				Expect(result).To(Equal("UNNAMED_FILE"))
			})
		})

		When("File.Name is present but empty", func() {
			It("returns the UNNAMED fallback with provided fallback", func() {
				fallback := "EMPTY"
				file := &ast.File{Name: &ast.Ident{Name: ""}}
				result := gast.GetAstFileNameOrFallback(file, &fallback)
				Expect(result).To(Equal("EMPTY"))
			})

			It("returns the UNNAMED fallback with generated value if fallback is nil", func() {
				file := &ast.File{Name: &ast.Ident{Name: ""}}
				result := gast.GetAstFileNameOrFallback(file, nil)
				Expect(result).To(Equal("UNNAMED_FILE"))
			})
		})

		When("File.Name is present and non-empty", func() {
			It("returns the package name directly", func() {
				file := &ast.File{Name: &ast.Ident{Name: "mypkg"}}
				result := gast.GetAstFileNameOrFallback(file, nil)
				Expect(result).To(Equal("mypkg"))
			})

			It("ignores fallback when package name is non-empty", func() {
				fallback := "SHOULD_NOT_USE"
				file := &ast.File{Name: &ast.Ident{Name: "mypkg"}}
				result := gast.GetAstFileNameOrFallback(file, &fallback)
				Expect(result).To(Equal("mypkg"))
			})
		})
	})
})

func TestUnits(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - AST")
}
