package diagnostics_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/pipeline"
	"github.com/gopher-fleece/gleece/core/validators/diagnostics"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Diagnostics", func() {
	var pipe pipeline.GleecePipeline

	BeforeEach(func() {
		pipe = utils.GetPipelineOrFail()
		Expect(pipe.GenerateGraph()).To(BeNil(), "Failed to prepare graph for tests")
	})

	It("Produces correct diagnostic structures", func() {
		diags, err := pipe.Validate()
		Expect(err).To(BeNil())
		Expect(diags).To(HaveLen(1))

		ctrlDiag := diags[0]

		// Controller level diagnostics
		Expect(ctrlDiag.EntityKind).To(Equal("Controller"))
		Expect(ctrlDiag.EntityName).To(Equal("DiagnosticsController"))
		Expect(ctrlDiag.Diagnostics).To(HaveLen(1))

		Expect(ctrlDiag.Diagnostics[0].Message).To(Equal("Controller 'DiagnosticsController' is lacking a @Tag annotation"))
		Expect(ctrlDiag.Diagnostics[0].Severity).To(Equal(diagnostics.DiagnosticWarning))
		Expect(ctrlDiag.Diagnostics[0].Code).To(Equal(string(diagnostics.DiagControllerLevelMissingTag)))
		Expect(ctrlDiag.Diagnostics[0].Range).To(Equal(common.ResolvedRange{
			StartLine: 4,
			StartCol:  0,
			EndLine:   5,
			EndCol:    28,
		}))

		Expect(ctrlDiag.Children).To(HaveLen(4))

		// Child #1
		Expect(ctrlDiag.Children[0].EntityKind).To(Equal("Receiver"))
		Expect(ctrlDiag.Children[0].EntityName).To(Equal("MethodWithNonStandardStatusCode"))
		Expect(ctrlDiag.Children[0].Diagnostics).To(HaveLen(1))
		Expect(ctrlDiag.Children[0].Diagnostics[0].Message).To(Equal("Non-standard HTTP status code '641'"))
		Expect(ctrlDiag.Children[0].Diagnostics[0].Severity).To(Equal(diagnostics.DiagnosticWarning))
		Expect(ctrlDiag.Children[0].Diagnostics[0].Code).To(Equal(string(diagnostics.DiagAnnotationValueInvalid)))
		Expect(ctrlDiag.Children[0].Diagnostics[0].Range).To(Equal(common.ResolvedRange{
			StartLine: 12,
			StartCol:  13,
			EndLine:   12,
			EndCol:    16,
		}))

		// Child #2
		Expect(ctrlDiag.Children[1].EntityKind).To(Equal("Receiver"))
		Expect(ctrlDiag.Children[1].EntityName).To(Equal("MethodWithUnAnnotatedParam"))
		Expect(ctrlDiag.Children[1].Diagnostics).To(HaveLen(1))
		Expect(ctrlDiag.Children[1].Diagnostics[0].Message).To(Equal("Function parameter 'id' is not referenced by a parameter annotation"))
		Expect(ctrlDiag.Children[1].Diagnostics[0].Severity).To(Equal(diagnostics.DiagnosticError))
		Expect(ctrlDiag.Children[1].Diagnostics[0].Code).To(Equal(string(diagnostics.DiagLinkerUnreferencedParameter)))
		Expect(ctrlDiag.Children[1].Diagnostics[0].Range).To(Equal(common.ResolvedRange{
			StartLine: 19,
			StartCol:  58,
			EndLine:   19,
			EndCol:    67,
		}))

		// Child #3
		Expect(ctrlDiag.Children[2].Children).To(BeEmpty())
		Expect(ctrlDiag.Children[2].EntityKind).To(Equal("Receiver"))
		Expect(ctrlDiag.Children[2].EntityName).To(Equal("MethodWithUnlinkedParam"))
		Expect(ctrlDiag.Children[2].Diagnostics).To(HaveLen(2))
		Expect(ctrlDiag.Children[2].Diagnostics[0].Message).To(Equal("Route parameter 'id' does not have a corresponding @Path annotation"))
		Expect(ctrlDiag.Children[2].Diagnostics[0].Severity).To(Equal(diagnostics.DiagnosticError))
		Expect(ctrlDiag.Children[2].Diagnostics[0].Code).To(Equal(string(diagnostics.DiagLinkerRouteMissingPath)))
		Expect(ctrlDiag.Children[2].Diagnostics[0].Range).To(Equal(common.ResolvedRange{
			StartLine: 23,
			StartCol:  26,
			EndLine:   23,
			EndCol:    30,
		}))

		Expect(ctrlDiag.Children[2].Diagnostics[1].Message).To(Equal("Function parameter 'id' is not referenced by a parameter annotation"))
		Expect(ctrlDiag.Children[2].Diagnostics[1].Severity).To(Equal(diagnostics.DiagnosticError))
		Expect(ctrlDiag.Children[2].Diagnostics[1].Code).To(Equal(string(diagnostics.DiagLinkerUnreferencedParameter)))
		Expect(ctrlDiag.Children[2].Diagnostics[1].Range).To(Equal(common.ResolvedRange{
			StartLine: 25,
			StartCol:  55,
			EndLine:   25,
			EndCol:    64,
		}))

		// Child #4
		Expect(ctrlDiag.Children[3].EntityKind).To(Equal("Receiver"))
		Expect(ctrlDiag.Children[3].EntityName).To(Equal("MethodWithUnlinkedAlias"))
		Expect(ctrlDiag.Children[3].Diagnostics).To(HaveLen(1))
		Expect(ctrlDiag.Children[3].Diagnostics[0].Message).To(Equal("Route parameter 'aliased' does not have a corresponding @Path annotation"))
		Expect(ctrlDiag.Children[3].Diagnostics[0].Severity).To(Equal(diagnostics.DiagnosticError))
		Expect(ctrlDiag.Children[3].Diagnostics[0].Code).To(Equal(string(diagnostics.DiagLinkerRouteMissingPath)))
		Expect(ctrlDiag.Children[3].Diagnostics[0].Range).To(Equal(common.ResolvedRange{
			StartLine: 29,
			StartCol:  26,
			EndLine:   29,
			EndCol:    35,
		}))
	})

	It("Correctly classifies and counts deep diagnostic structures", func() {
		diags, err := pipe.Validate()
		Expect(err).To(BeNil())
		Expect(diags).To(HaveLen(1))

		classified := diagnostics.ClassifyEntityDiags(diags[0])
		Expect(classified.Hints).To(HaveLen(0))
		Expect(classified.Info).To(HaveLen(0))
		Expect(classified.Warnings).To(HaveLen(2))
		Expect(classified.Errors).To(HaveLen(4))
	})

	It("Produces correct text when dumping diagnostic structures", func() {
		diags, err := pipe.Validate()
		Expect(err).To(BeNil())
		Expect(diags).To(HaveLen(1))

		classified := diagnostics.ClassifyEntityDiags(diags[0])
		Expect(classified.String()).To(MatchRegexp(
			`^Errors: \(Total 4\)\s+` +
				`linker-unreferenced-parameter .* - Function parameter 'id' is not referenced by a parameter annotation\s+` +
				`linker-route-missing-path-reference .* - Route parameter 'id' does not have a corresponding @Path annotation\s+` +
				`linker-unreferenced-parameter .* - Function parameter 'id' is not referenced by a parameter annotation\s+` +
				`linker-route-missing-path-reference .* - Route parameter 'aliased' does not have a corresponding @Path annotation\s+` +
				`Warnings: \(Total 4\)\s+` +
				`controller-missing-tag .* - Controller 'DiagnosticsController' is lacking a @Tag annotation\s+` +
				`annotation-value-invalid .* - Non-standard HTTP status code '641'\s+` +
				`Info: \(Total 4\)Hints: \(Total 4\)$`,
		))
	})

	It("Fails pipeline if any Error diagnostics exist", func() {
		pipe = utils.GetPipelineOrFail() // Use a new, clean graph
		_, err := pipe.Run()
		Expect(err).To(Not(BeNil()))
		errStr := err.Error()

		// Check total entities
		Expect(errStr).To(MatchRegexp(`(?m)^Entities with diagnostics: 4$`))

		// Check MethodWithUnAnnotatedParam block
		Expect(errStr).To(MatchRegexp(
			`Receiver MethodWithUnAnnotatedParam:\n\s+` +
				`Errors: \(Total 1\)\s+linker-unreferenced-parameter .* Function parameter 'id' is not referenced by a parameter annotation\s+` +
				`Warnings: \(Total 1\)Info: \(Total 1\)Hints: \(Total 1\)`))

		// Check MethodWithUnlinkedParam block
		Expect(errStr).To(MatchRegexp(
			`Receiver MethodWithUnlinkedParam:\n\s+` +
				`Errors: \(Total 2\)\s+` +
				`linker-route-missing-path-reference .* Route parameter 'id' does not have a corresponding @Path annotation\s+` +
				`linker-unreferenced-parameter .* Function parameter 'id' is not referenced by a parameter annotation\s+` +
				`Warnings: \(Total 2\)Info: \(Total 2\)Hints: \(Total 2\)`))

		// Check MethodWithUnlinkedAlias block
		Expect(errStr).To(MatchRegexp(
			`Receiver MethodWithUnlinkedAlias:\n\s+` +
				`Errors: \(Total 1\)\s+linker-route-missing-path-reference .* Route parameter 'aliased' does not have a corresponding @Path annotation\s+` +
				`Warnings: \(Total 1\)Info: \(Total 1\)Hints: \(Total 1\)`))

	})
})

func TestDiagnostics(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Diagnostics")
}
