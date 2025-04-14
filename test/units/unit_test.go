package units_test

import (
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
	typesPkgLoadOnly = utils.LoadPackage(auxTypesPackageName, packages.LoadFiles)
	typesPkgFullSyntax = utils.LoadPackage(auxTypesPackageName, packages.LoadAllSyntax)
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
	})
})

func TestUnits(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests")
}
