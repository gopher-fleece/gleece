package matchers

import (
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/validators/diagnostics"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func HaveMessage(msg string) types.GomegaMatcher {
	return WithTransform(func(d diagnostics.ResolvedDiagnostic) string {
		return d.Message
	}, Equal(msg))
}

func HaveSeverity(sev diagnostics.DiagnosticSeverity) types.GomegaMatcher {
	return WithTransform(func(d diagnostics.ResolvedDiagnostic) diagnostics.DiagnosticSeverity {
		return d.Severity
	}, Equal(sev))
}

func HaveFilePath(substr string) types.GomegaMatcher {
	return WithTransform(func(d diagnostics.ResolvedDiagnostic) string {
		return d.FilePath
	}, ContainSubstring(substr))
}

func HaveRange(r common.ResolvedRange) types.GomegaMatcher {
	return WithTransform(func(d diagnostics.ResolvedDiagnostic) common.ResolvedRange {
		return d.Range
	}, Equal(r))
}

func HaveCode(code diagnostics.DiagnosticCode) types.GomegaMatcher {
	return WithTransform(func(d diagnostics.ResolvedDiagnostic) string {
		return d.Code
	}, BeEquivalentTo(code))
}

func HaveCodeAndMessage(code diagnostics.DiagnosticCode, msg string) types.GomegaMatcher {
	type CodeAndMessage struct {
		Code    string
		Message string
	}

	return WithTransform(
		func(d diagnostics.ResolvedDiagnostic) CodeAndMessage {
			return CodeAndMessage{Code: d.Code, Message: d.Message}
		},
		Equal(CodeAndMessage{Code: string(code), Message: msg}),
	)
}

func HaveCodeAndMessageSubstring(code diagnostics.DiagnosticCode, substr string) types.GomegaMatcher {
	return SatisfyAll(
		WithTransform(func(d diagnostics.ResolvedDiagnostic) string {
			return d.Code
		}, Equal(code)),

		WithTransform(func(d diagnostics.ResolvedDiagnostic) string {
			return d.Message
		}, ContainSubstring(substr)),
	)
}

func BeDiagnosticErrorWithCodeAndMessage(code diagnostics.DiagnosticCode, message string) types.GomegaMatcher {
	return SatisfyAll(
		HaveSeverity(diagnostics.DiagnosticError),
		HaveCodeAndMessage(code, message),
	)
}
