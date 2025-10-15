package ast_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopher-fleece/gleece/extractor"
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
			Expect(extractor.IsFuncDeclReceiverForStruct("StructWithReceivers", funcDecl)).To(BeFalse())
		})

		It("Returns true if the given FuncDecl is a value receiver", func() {
			funcDecl := utils.GetFunctionFromPackageOrFail(typesPkgFullSyntax, "ValueReceiverForStructWithReceivers")
			Expect(extractor.IsFuncDeclReceiverForStruct("StructWithReceivers", funcDecl)).To(BeTrue())
		})

		It("Returns true if the given FuncDecl is a pointer receiver", func() {
			funcDecl := utils.GetFunctionFromPackageOrFail(typesPkgFullSyntax, "PointerReceiverForStructWithReceivers")
			Expect(extractor.IsFuncDeclReceiverForStruct("StructWithReceivers", funcDecl)).To(BeTrue())
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

			Expect(extractor.IsFuncDeclReceiverForStruct("StructWithReceivers", funcDecl)).To(BeFalse())
		})
	})

	Context("GetDefaultPackageAlias", func() {
		It("Returns an error if the given ast.File has no name", func() {
			file := &ast.File{
				Name: nil,
			}
			alias, err := extractor.GetDefaultPackageAlias(file)
			Expect(err).To(MatchError(ContainSubstring("source file does not have a name")))
			Expect(alias).To(BeEmpty())
		})
	})

	Context("GetFullPackageName", func() {
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
			fullPkg, err := extractor.GetFullPackageName(astFile, fileSet)
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

			pkgName, err := extractor.GetFullPackageName(parsedFile, fileSet)
			Expect(err).To(BeNil())       // packages.Load succeeded
			Expect(pkgName).To(BeEmpty()) // but file wasn't found in the loaded package
		})

	})

	Context("LookupTypeName", func() {
		It("Returns correct error when given package has no type information", func() {
			typeName, err := extractor.LookupTypeName(typesPkgLoadOnly, "NA")
			Expect(typeName).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("does not have types or types scope")))
		})

		It("Returns no value and no error when given name does not exist in given package", func() {
			typeName, err := extractor.LookupTypeName(typesPkgFullSyntax, "ThisNameDoesNotExistInThisPackage")
			Expect(typeName).To(BeNil())
			Expect(err).To(BeNil())
		})

		It("Returns no value and no error when given name exist in given package but is not a TypeName", func() {
			typeName, err := extractor.LookupTypeName(typesPkgFullSyntax, "ConstA")
			Expect(typeName).To(BeNil())
			Expect(err).To(BeNil())
		})

		It("Returns correct value when given name exist in given package and is a TypeName", func() {
			typeName, err := extractor.LookupTypeName(typesPkgFullSyntax, "StructA")
			Expect(err).To(BeNil())
			Expect(typeName.Name()).ToNot(BeNil())

			typeName, err = extractor.LookupTypeName(typesPkgFullSyntax, "InterfaceA")
			Expect(err).To(BeNil())
			Expect(typeName.Name()).ToNot(BeNil())

			typeName, err = extractor.LookupTypeName(typesPkgFullSyntax, "EnumTypeA")
			Expect(err).To(BeNil())
			Expect(typeName.Name()).ToNot(BeNil())
		})
	})

	Context("GetTypeNameOrError", func() {
		It("Returns correct error when package does not have type information", func() {
			typeName, err := extractor.GetTypeNameOrError(typesPkgLoadOnly, "StructA")
			Expect(err).To(MatchError(ContainSubstring("does not have types or types scope")))
			Expect(typeName).To(BeNil())
		})

		It("Returns correct error given name does not exist in given package", func() {
			typeName, err := extractor.GetTypeNameOrError(typesPkgFullSyntax, "ThisNameDoesNotExistInThisPackage")
			Expect(err).To(MatchError(ContainSubstring("was not found in package")))
			Expect(typeName).To(BeNil())
		})
	})

	Context("FindTypesStructInPackage", func() {
		It("Returns correct error when package does not have type information", func() {
			typeName, err := extractor.FindTypesStructInPackage(typesPkgLoadOnly, "StructA")
			Expect(err).To(MatchError(ContainSubstring("does not have types or types scope")))
			Expect(typeName).To(BeNil())
		})

		It("Returns no value and no error when given name does not exist in given package", func() {
			typeName, err := extractor.FindTypesStructInPackage(typesPkgFullSyntax, "ThisNameDoesNotExistInThisPackage")
			Expect(err).To(BeNil())
			Expect(typeName).To(BeNil())
		})

		It("Returns correct error when given name exist in given package but is not a struct", func() {
			typeName, err := extractor.FindTypesStructInPackage(typesPkgFullSyntax, "InterfaceA")
			Expect(err).To(MatchError(ContainSubstring("is not a struct type")))
			Expect(typeName).To(BeNil())
		})
	})

	Context("GetUnderlyingTypeName", func() {
		It("correctly resolves basic and named types", func() {
			testStruct, err := extractor.FindTypesStructInPackage(typesPkgFullSyntax, "StructForGetUnderlyingTypeName")
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
				actual := extractor.GetUnderlyingTypeName(typ)
				Expect(actual).To(Equal(expected), fmt.Sprintf("field %s", name))
			}
		})
	})

	Context("GetFieldTypeString", func() {
		It("Correctly parses field AST expressions into string descriptions", func() {
			typeStruct, err := extractor.FindTypesStructInPackage(typesPkgFullSyntax, "StructForGetUnderlyingTypeName")
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
				actualStr := extractor.GetFieldTypeString(astField.Type)

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

			Expect(extractor.GetFieldTypeString(ellipsisExpr)).To(Equal("Variadic (...int)"))
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

			result := extractor.GetFieldTypeString(arrayExpr)
			Expect(result).To(Equal("[?]int"))
		})
	})

	Context("Given a non-pointer type", func() {
		It("should return the same type", func() {
			// Test with basic types
			intType := types.Typ[types.Int]
			result := extractor.UnwrapPointerType(intType)
			Expect(result).To(Equal(intType))

			stringType := types.Typ[types.String]
			result = extractor.UnwrapPointerType(stringType)
			Expect(result).To(Equal(stringType))

			boolType := types.Typ[types.Bool]
			result = extractor.UnwrapPointerType(boolType)
			Expect(result).To(Equal(boolType))
		})

		It("should return the same type for complex non-pointer types", func() {
			// Test with slice type
			sliceType := types.NewSlice(types.Typ[types.Int])
			result := extractor.UnwrapPointerType(sliceType)
			Expect(result).To(Equal(sliceType))

			// Test with array type
			arrayType := types.NewArray(types.Typ[types.String], 10)
			result = extractor.UnwrapPointerType(arrayType)
			Expect(result).To(Equal(arrayType))

			// Test with map type
			mapType := types.NewMap(types.Typ[types.String], types.Typ[types.Int])
			result = extractor.UnwrapPointerType(mapType)
			Expect(result).To(Equal(mapType))
		})
	})

	Context("Given a single-level pointer type", func() {
		It("should return the underlying type", func() {
			// Test with pointer to int
			intType := types.Typ[types.Int]
			ptrToInt := types.NewPointer(intType)
			result := extractor.UnwrapPointerType(ptrToInt)
			Expect(result).To(Equal(intType))

			// Test with pointer to string
			stringType := types.Typ[types.String]
			ptrToString := types.NewPointer(stringType)
			result = extractor.UnwrapPointerType(ptrToString)
			Expect(result).To(Equal(stringType))

			// Test with pointer to bool
			boolType := types.Typ[types.Bool]
			ptrToBool := types.NewPointer(boolType)
			result = extractor.UnwrapPointerType(ptrToBool)
			Expect(result).To(Equal(boolType))
		})

		It("should return the underlying type for complex pointer types", func() {
			// Test with pointer to slice
			sliceType := types.NewSlice(types.Typ[types.Int])
			ptrToSlice := types.NewPointer(sliceType)
			result := extractor.UnwrapPointerType(ptrToSlice)
			Expect(result).To(Equal(sliceType))

			// Test with pointer to array
			arrayType := types.NewArray(types.Typ[types.String], 5)
			ptrToArray := types.NewPointer(arrayType)
			result = extractor.UnwrapPointerType(ptrToArray)
			Expect(result).To(Equal(arrayType))

			// Test with pointer to map
			mapType := types.NewMap(types.Typ[types.String], types.Typ[types.Int])
			ptrToMap := types.NewPointer(mapType)
			result = extractor.UnwrapPointerType(ptrToMap)
			Expect(result).To(Equal(mapType))
		})
	})

	Context("Given a multi-level pointer type", func() {
		It("should recursively unwrap to the underlying type", func() {
			// Test with double pointer to int (**int)
			intType := types.Typ[types.Int]
			ptrToInt := types.NewPointer(intType)
			ptrToPtrToInt := types.NewPointer(ptrToInt)
			result := extractor.UnwrapPointerType(ptrToPtrToInt)
			Expect(result).To(Equal(intType))

			// Test with triple pointer to string (***string)
			stringType := types.Typ[types.String]
			ptrToString := types.NewPointer(stringType)
			ptrToPtrToString := types.NewPointer(ptrToString)
			ptrToPtrToPtrToString := types.NewPointer(ptrToPtrToString)
			result = extractor.UnwrapPointerType(ptrToPtrToPtrToString)
			Expect(result).To(Equal(stringType))
		})

		It("should handle deeply nested pointer types", func() {
			// Test with 5-level pointer nesting
			baseType := types.Typ[types.Float64]
			currentType := types.Type(baseType)

			// Create 5 levels of pointers
			for i := 0; i < 5; i++ {
				currentType = types.NewPointer(currentType)
			}

			result := extractor.UnwrapPointerType(currentType)
			Expect(result).To(Equal(baseType))
		})

		It("should handle multi-level pointers to complex types", func() {
			// Test with double pointer to slice (**[]int)
			sliceType := types.NewSlice(types.Typ[types.Int])
			ptrToSlice := types.NewPointer(sliceType)
			ptrToPtrToSlice := types.NewPointer(ptrToSlice)
			result := extractor.UnwrapPointerType(ptrToPtrToSlice)
			Expect(result).To(Equal(sliceType))

			// Test with triple pointer to map (***map[string]int)
			mapType := types.NewMap(types.Typ[types.String], types.Typ[types.Int])
			ptrToMap := types.NewPointer(mapType)
			ptrToPtrToMap := types.NewPointer(ptrToMap)
			ptrToPtrToPtrToMap := types.NewPointer(ptrToPtrToMap)
			result = extractor.UnwrapPointerType(ptrToPtrToPtrToMap)
			Expect(result).To(Equal(mapType))
		})
	})

	Context("Edge cases", func() {
		It("should handle nil type gracefully", func() {
			// While this might not be a realistic scenario in practice,
			// the function should handle nil input without panicking
			var nilType types.Type
			result := extractor.UnwrapPointerType(nilType)
			Expect(result).To(BeNil())
		})
	})
})

func TestUnits(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - AST")
}
