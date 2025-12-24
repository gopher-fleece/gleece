package ast_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/v2/generator/compilation"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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

var _ = Describe("Unit Tests - Compilation", func() {
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

func TestUnits(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Compilation")
}
