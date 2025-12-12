package diagnostics

import (
	"fmt"
	"strings"
)

type ClassifiedEntityDiags struct {
	Errors   []ResolvedDiagnostic
	Warnings []ResolvedDiagnostic
	Info     []ResolvedDiagnostic
	Hints    []ResolvedDiagnostic
}

func (d ClassifiedEntityDiags) formatSeverityClass(severity string, diags []ResolvedDiagnostic) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s: (Total %d)\n", severity, len(diags)))
	for _, diag := range diags {
		builder.WriteString(fmt.Sprintf(
			"\t %s at %s:%d:%d - %s\n",
			diag.Code,
			diag.FilePath,
			diag.Range.StartLine+1, // These are 0 based but IDE is generally 1 based.
			diag.Range.StartCol+1,  // Might have to re-think this and move back to 1-base
			diag.Message,
		))
	}

	return builder.String()
}

func (d ClassifiedEntityDiags) String() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s\n", d.formatSeverityClass("Errors", d.Errors)))
	builder.WriteString(fmt.Sprintf("%s\n", d.formatSeverityClass("Warnings", d.Warnings)))
	builder.WriteString(fmt.Sprintf("%s\n", d.formatSeverityClass("Info", d.Info)))
	builder.WriteString(fmt.Sprintf("%s\n", d.formatSeverityClass("Hints", d.Hints)))
	return builder.String()
}

func ClassifyEntityDiags(entityDiag EntityDiagnostic) ClassifiedEntityDiags {
	classified := ClassifiedEntityDiags{
		Errors:   []ResolvedDiagnostic{},
		Warnings: []ResolvedDiagnostic{},
		Info:     []ResolvedDiagnostic{},
		Hints:    []ResolvedDiagnostic{},
	}

	for _, diag := range entityDiag.Diagnostics {
		switch diag.Severity {
		case DiagnosticError:
			classified.Errors = append(classified.Errors, diag)
		case DiagnosticWarning:
			classified.Warnings = append(classified.Warnings, diag)
		case DiagnosticInformation:
			classified.Info = append(classified.Info, diag)
		case DiagnosticHint:
			classified.Hints = append(classified.Hints, diag)
		}
	}

	for _, childDiag := range entityDiag.Children {
		classifiedChild := ClassifyEntityDiags(*childDiag)
		classified.Errors = append(classified.Errors, classifiedChild.Errors...)
		classified.Warnings = append(classified.Warnings, classifiedChild.Warnings...)
		classified.Info = append(classified.Info, classifiedChild.Info...)
		classified.Hints = append(classified.Hints, classifiedChild.Hints...)
	}

	return classified
}
