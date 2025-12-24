package diagnostics

import (
	"fmt"
	"slices"

	"github.com/gopher-fleece/gleece/v2/common"
)

type TextEdit struct {
	NewText  string
	FilePath string
	Range    common.ResolvedRange
}

func (t TextEdit) Key() string {
	return fmt.Sprintf("%s|%d:%d-%d:%d|%s",
		t.FilePath,
		t.Range.StartLine, t.Range.StartCol, t.Range.EndLine, t.Range.EndCol,
		t.NewText,
	)
}

type ResolvedDiagnostic struct {
	Message  string
	Severity DiagnosticSeverity
	FilePath string
	Range    common.ResolvedRange
	Code     string     // optional rule id
	Source   string     // "gleece"
	Fixes    []TextEdit // optional
}

func (d *ResolvedDiagnostic) Equal(other ResolvedDiagnostic) bool {
	if d.Message != other.Message {
		return false
	}

	if d.Severity != other.Severity {
		return false
	}

	if d.Code != other.Code {
		return false
	}

	if d.FilePath != other.FilePath {
		return false
	}

	if d.Range != other.Range {
		return false
	}

	if d.Source != other.Source {
		return false
	}

	return slices.Equal(d.Fixes, other.Fixes)
}

func NewDiagnostic(
	filePath, message string,
	code DiagnosticCode,
	severity DiagnosticSeverity,
	rng common.ResolvedRange,
) ResolvedDiagnostic {
	return ResolvedDiagnostic{
		Message:  message,
		Severity: severity,
		FilePath: filePath,
		Range:    rng,
		Code:     string(code),
		Source:   "gleece",
	}
}

func NewErrorDiagnostic(filePath, message string, code DiagnosticCode, rng common.ResolvedRange) ResolvedDiagnostic {
	return NewDiagnostic(filePath, message, code, DiagnosticError, rng)
}

func NewWarningDiagnostic(filePath, message string, code DiagnosticCode, rng common.ResolvedRange) ResolvedDiagnostic {
	return NewDiagnostic(filePath, message, code, DiagnosticWarning, rng)
}

func NewInfoDiagnostic(filePath, message string, code DiagnosticCode, rng common.ResolvedRange) ResolvedDiagnostic {
	return NewDiagnostic(filePath, message, code, DiagnosticInformation, rng)
}

func NewHintDiagnostic(filePath, message string, code DiagnosticCode, rng common.ResolvedRange) ResolvedDiagnostic {
	return NewDiagnostic(filePath, message, code, DiagnosticHint, rng)
}
