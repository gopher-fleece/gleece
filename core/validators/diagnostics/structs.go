package diagnostics

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/gopher-fleece/gleece/common"
)

type TextEdit struct {
	NewText  string
	FilePath string
	Range    common.ResolvedRange
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

type EntityDiagnostic struct {
	EntityName  string
	EntityKind  string
	Diagnostics []ResolvedDiagnostic // Diagnostics directly attached to this entity
	Children    []*EntityDiagnostic  // Nested entity diagnostics
}

func NewEntityDiagnostic(context, name string) EntityDiagnostic {
	return EntityDiagnostic{
		EntityKind:  context,
		EntityName:  name,
		Diagnostics: []ResolvedDiagnostic{},
	}
}

func (d EntityDiagnostic) Empty() bool {
	return len(d.Diagnostics) <= 0 && len(d.Children) <= 0
}

func (d *EntityDiagnostic) AddDiagnostic(diag ResolvedDiagnostic) {
	if d.Diagnostics == nil {
		d.Diagnostics = []ResolvedDiagnostic{diag}
	} else {
		d.Diagnostics = append(d.Diagnostics, diag)
	}
}

func (d *EntityDiagnostic) AddDiagnostics(diags []ResolvedDiagnostic) {
	if len(diags) <= 0 {
		return
	}

	if d.Diagnostics == nil {
		d.Diagnostics = diags
	} else {
		d.Diagnostics = append(d.Diagnostics, diags...)
	}
}

func (d *EntityDiagnostic) AddDiagnosticIfNotNil(diag *ResolvedDiagnostic) {
	if diag == nil {
		return
	}

	d.AddDiagnostic(*diag)
}

// add child (short & safe)
func (d *EntityDiagnostic) AddChild(child *EntityDiagnostic) {
	if child == nil {
		return
	}

	if d.Children == nil {
		d.Children = []*EntityDiagnostic{child}
	} else {
		d.Children = append(d.Children, child)
	}
}

func GetDiagnosticsWithSeverity(diags []EntityDiagnostic, severities []DiagnosticSeverity) []EntityDiagnostic {
	matching := []EntityDiagnostic{}

	for _, diagEntity := range diags {
		for _, diag := range diagEntity.Diagnostics {
			if slices.Contains(severities, diag.Severity) {
				matching = append(matching, diagEntity)
			}
		}

		if len(diagEntity.Children) > 0 {
			// Got a bit of a ptr-value mess over here. Need to improve
			dereferencedChildren := common.DereferenceSliceElements(diagEntity.Children)
			matching = append(matching, GetDiagnosticsWithSeverity(dereferencedChildren, severities)...)
		}
	}

	return matching
}

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
		builder.WriteString(fmt.Sprintf("\t %s at %s - %s ", diag.Code, diag.FilePath, diag.Message))
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

func DiagnosticsToError(diags []EntityDiagnostic) error {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Entities with diagnostics: %d\n", len(diags)))
	for _, diag := range diags {
		builder.WriteString(fmt.Sprintf("%s %s:\n\t", diag.EntityKind, diag.EntityName))
		classified := ClassifyEntityDiags(diag)
		builder.WriteString(classified.String())
	}

	return errors.New(builder.String())
}
