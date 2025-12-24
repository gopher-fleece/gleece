package orchestrator_test

import (
	"go/ast"
	"testing"

	"github.com/gopher-fleece/gleece/v2/core/visitors"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	"github.com/gopher-fleece/gleece/v2/test/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("VisitorOrchestrator", func() {
	const testConfigPath = "./gleece.test.config.json"

	Describe("NewVisitorOrchestrator", func() {
		Context("When a nil VisitContext is provided", func() {
			It("Returns an error indicating a nil context was provided", func() {
				orchestrator, err := visitors.NewVisitorOrchestrator(nil)
				Expect(orchestrator).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(
					"nil context was provided to VisitorOrchestrator"))
			})
		})

		Context("When a VisitContext missing required fields is provided", func() {
			It("Returns a joined error listing missing VisitContext fields", func() {
				emptyVisitContext := visitors.VisitContext{}
				orchestrator, err := visitors.NewVisitorOrchestrator(&emptyVisitContext)
				Expect(orchestrator).To(BeNil())
				Expect(err).To(HaveOccurred())

				errorMessage := err.Error()
				Expect(errorMessage).To(ContainSubstring("arbitration provider"))
				Expect(errorMessage).To(ContainSubstring("Gleece Config"))
				Expect(errorMessage).To(ContainSubstring("graph builder"))
				Expect(errorMessage).To(ContainSubstring("metadata cache"))
				Expect(errorMessage).To(ContainSubstring("synchronized provider"))
			})
		})

		Context("When a valid VisitContext is provided", func() {
			It("Constructs the orchestrator without error and exposes expected methods",
				func() {
					orchestrator, err := buildOrchestratorFromConfig(testConfigPath)
					Expect(err).NotTo(HaveOccurred())
					Expect(orchestrator).NotTo(BeNil())

					fieldVisitor := orchestrator.GetFieldVisitor()
					Expect(fieldVisitor).NotTo(BeNil())

					allFiles := orchestrator.GetAllSourceFiles()
					// Ensure the returned value is a slice of *ast.File (may be empty)
					var sample []*ast.File
					Expect(allFiles).To(BeAssignableToTypeOf(sample))
				})
		})
	})

	Describe("Visit", func() {
		Context("When delegating to the internal controller visitor", func() {
			It("Returns a non-nil ast.Visitor for a basic AST node", func() {
				orchestrator, err := buildOrchestratorFromConfig(testConfigPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(orchestrator).NotTo(BeNil())

				testNode := &ast.Ident{Name: "SampleIdentifier"}
				returnedVisitor := orchestrator.Visit(testNode)
				Expect(returnedVisitor).NotTo(BeNil())
			})
		})
	})

	Describe("Diagnostic and error accessors", func() {
		Context("When inspecting last error and diagnostic stack", func() {
			It("Returns nil for last error and a string for formatted diagnostic stack",
				func() {
					orchestrator, err := buildOrchestratorFromConfig(testConfigPath)
					Expect(err).NotTo(HaveOccurred())
					Expect(orchestrator).NotTo(BeNil())

					lastError := orchestrator.GetLastError()
					Expect(lastError).To(BeNil())

					formattedStack := orchestrator.GetFormattedDiagnosticStack()
					// Ensure we get a string (may be empty)
					Expect(formattedStack).To(BeAssignableToTypeOf(""))
				})
		})
	})
})

// Constructs a valid VisitContext using the project test helper.
func buildValidVisitContextOrFail(relativeConfigPath string) visitors.VisitContext {
	return utils.GetVisitContextByRelativeConfigOrFail(relativeConfigPath)
}

// Builds an orchestrator from a relative config path.
func buildOrchestratorFromConfig(relativeConfigPath string) (*visitors.VisitorOrchestrator, error) {
	visitContext := buildValidVisitContextOrFail(relativeConfigPath)
	return visitors.NewVisitorOrchestrator(&visitContext)
}

func TestVisitorOrchestrator(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "VisitorOrchestrator")
}
