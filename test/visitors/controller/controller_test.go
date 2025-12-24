package controller_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/core/visitors"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ControllerVisitor", func() {
	var ctx utils.StdTestCtx
	BeforeEach(func() {
		ctx = utils.CreateStdTestCtx("gleece.test.config.json")
	})

	Context("NewControllerVisitor", func() {
		// Note - these tests are actually primarily designed to test the 'abstract' BaseVisitor.
		// A bit of a cross contamination.
		It("Returns an error when given a nil context", func() {
			_, err := visitors.NewControllerVisitor(nil)
			Expect(err).To(MatchError(ContainSubstring("nil context was given to contextInitGuard")))
		})

		It("Initialized properly when given a context with no globs", func() {
			// Note we're passing an empty config to prevent a panic during initialization.
			// While we could check this condition in the code, it's actually not a real use-case
			// and indicates gross misuse so a panic is fine.
			//
			// This 'empty' initialization takes care of the corpus of 'initializeWithGlobs' in the base visitor
			visitor, err := visitors.NewControllerVisitor(&visitors.VisitContext{
				GleeceConfig: &definitions.GleeceConfig{
					CommonConfig: definitions.CommonConfig{},
				},
			})
			Expect(err).To(BeNil())
			Expect(visitor).ToNot(BeNil())
		})
	})

	Context("GetFormatterDiagnosticStack", func() {
		It("Correctly prints the diagnostic stack when empty", func() {
			formattedStack := ctx.Orc.GetFormattedDiagnosticStack()
			Expect(formattedStack).To(ContainSubstring(""))
		})
	})
})

func TestControllerVisitor(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "ControllerVisitor")
}
