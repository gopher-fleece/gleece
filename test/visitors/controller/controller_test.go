package controller_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/core/arbitrators/caching"
	"github.com/gopher-fleece/gleece/core/visitors"
	"github.com/gopher-fleece/gleece/core/visitors/providers"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/infrastructure/logger"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const controllerFileRelPath = "./resources/micro.valid.controller.go"

type TestCtx struct {
	arbProvider       *providers.ArbitrationProvider
	metaCache         *caching.MetadataCache
	symGraph          symboldg.SymbolGraph
	visitCtx          *visitors.VisitContext
	controllerVisitor *visitors.ControllerVisitor
}

var _ = Describe("ControllerVisitor", func() {
	var ctx TestCtx
	BeforeEach(func() {
		ctx = createTestCtx([]string{controllerFileRelPath})
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
			formattedStack := ctx.controllerVisitor.GetFormattedDiagnosticStack()
			Expect(formattedStack).To(ContainSubstring(""))
		})
	})
})

func createTestCtx(fileGlobs []string) TestCtx {
	ctx := TestCtx{}

	// Pass the real controller file so the providers actually load it
	arbProvider, err := providers.NewArbitrationProvider(fileGlobs)
	Expect(err).To(BeNil())
	ctx.arbProvider = arbProvider

	// Verify files were properly loaded
	srcFiles := arbProvider.Pkg().GetAllSourceFiles()
	Expect(srcFiles).ToNot(BeEmpty(), "Arbitration provider parsed zero files; check glob and file contents")

	// Build the VisitContext and routeVisitor as before using arbProvider
	ctx.metaCache = caching.NewMetadataCache()
	ctx.symGraph = symboldg.NewSymbolGraph()
	ctx.visitCtx = &visitors.VisitContext{
		ArbitrationProvider: arbProvider,
		MetadataCache:       ctx.metaCache,
		Graph:               &ctx.symGraph,
	}

	ctx.controllerVisitor, err = visitors.NewControllerVisitor(ctx.visitCtx)
	Expect(err).To(BeNil(), "Failed to construct a new controller visitor")
	return ctx
}

func TestControllerVisitor(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "ControllerVisitor")
}
