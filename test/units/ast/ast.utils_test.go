package ast_test

import (
	"fmt"
	"go/ast"
	"go/token"
	"testing"

	"github.com/gopher-fleece/gleece/extractor"
	"github.com/gopher-fleece/gleece/generator/compilation"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/go/packages"
)

const auxTypesPackageName = "github.com/gopher-fleece/gleece/test/units"

const stdUnformattedCodeChunk = `
	package abc

		func TestFunc()(  string ,error )  {
      return "",        nil
		}
`

const invalidImportsCodeChunk = `
	package abc, def
	
	func TestFunc() (string, error) {
		return "", nil
	}
`

const stdFormattedCodeChunk = "package abc\nfunc TestFunc() (string, error) {\n\treturn \"\", nil\n}\n"

var typesPkgLoadOnly *packages.Package
var typesPkgFullSyntax *packages.Package

var _ = BeforeSuite(func() {
	typesPkgLoadOnly = utils.LoadPackageOrFail(auxTypesPackageName, packages.LoadFiles)
	typesPkgFullSyntax = utils.LoadPackageOrFail(auxTypesPackageName, packages.LoadAllSyntax)
})

var _ = Describe("Unit Tests", func() {
	Context("Compilation Utils", func() {
		Context("OptimizeImportsAndFormat", func() {
			It("Given correct input, formats without error and returns correct value", func() {
				formatted, err := compilation.OptimizeImportsAndFormat(stdUnformattedCodeChunk)
				Expect(err).To(BeNil())
				Expect(formatted).To(Equal(stdFormattedCodeChunk))
			})

			It("Given invalid imports, returns correct error", func() {
				formatted, err := compilation.OptimizeImportsAndFormat(invalidImportsCodeChunk)
				Expect(err).To(MatchError(ContainSubstring("failed to optimize imports")))
				Expect(err).To(MatchError(ContainSubstring("expected ';', found ','")))
				Expect(formatted).To(BeEmpty())
			})

			// Imports optimization performs gross syntax validation and as such,
			// it's borderline (or outright) impossible to get format.Source to break if imports.Process did not,
			// hence no third test here.
		})
	})

	Context("AST Utils", func() {
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
	})
})

func TestUnits(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests")
}
