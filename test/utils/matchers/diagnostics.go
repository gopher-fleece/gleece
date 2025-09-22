package matchers

import (
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/metadata"
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
		}, BeEquivalentTo(code)),

		WithTransform(func(d diagnostics.ResolvedDiagnostic) string {
			return d.Message
		}, ContainSubstring(substr)),
	)
}

func BeDiagnosticWarningWithCodeAndMessage(code diagnostics.DiagnosticCode, message string) types.GomegaMatcher {
	return SatisfyAll(
		HaveSeverity(diagnostics.DiagnosticWarning),
		HaveCodeAndMessage(code, message),
	)
}

func BeDiagnosticErrorWithCodeAndMessage(code diagnostics.DiagnosticCode, message string) types.GomegaMatcher {
	return SatisfyAll(
		HaveSeverity(diagnostics.DiagnosticError),
		HaveCodeAndMessage(code, message),
	)
}

func BeDiagnosticEntityOfReceiver(receiver *metadata.ReceiverMeta) types.GomegaMatcher {
	type EntityKindAndName struct {
		Kind string
		Name string
	}
	return WithTransform(func(d diagnostics.EntityDiagnostic) EntityKindAndName {
		return EntityKindAndName{
			Kind: d.EntityKind,
			Name: d.EntityName,
		}
	}, Equal(EntityKindAndName{
		Kind: "Receiver",
		Name: receiver.Name,
	}))
}

func HaveNoChildren() types.GomegaMatcher {
	return WithTransform(
		func(d diagnostics.EntityDiagnostic) int {
			return len(d.Children)
		},
		Equal(0),
	)
}

func BeChildlessDiagOfReceiver(receiver *metadata.ReceiverMeta) types.GomegaMatcher {
	return SatisfyAll(
		BeDiagnosticEntityOfReceiver(receiver),
		HaveNoChildren(),
	)
}

func BeAReturnSigDiagnostic() types.GomegaMatcher {
	return SatisfyAll(
		HaveCodeAndMessageSubstring(
			diagnostics.DiagReceiverRetValsInvalidSignature,
			"Expected method to return an error or a value and error tuple but found",
		),
		HaveSeverity(diagnostics.DiagnosticError),
	)
}
