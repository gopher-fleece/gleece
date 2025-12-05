package diagnostics

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/gopher-fleece/gleece/common/linq"
)

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

// BaseKey is the diagnostic entity's identifier - a combination of entity kind and name.
// Note this does not consider 'contents' such as Diagnostics or Children
func (d EntityDiagnostic) BaseKey() string {
	return CreateEntityDiagKey(d.EntityKind, d.EntityName)
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

// AddChild appends a child EntityDiagnostic to this entity.
// 'Child' entity diagnostics are meant to represent a nested but distinct entity with an issue
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

// GetChild returns a reference to the first diagnostic entity that matches the given name/kind
// or nil if no match was found.
//
// This method is suboptimal as children may be duplicated.
// Need to revise this to a map perhaps.
func (d *EntityDiagnostic) GetChild(name, kind string) *EntityDiagnostic {
	match := linq.First(d.Children, func(child *EntityDiagnostic) bool {
		return child != nil && child.EntityName == name && child.EntityKind == kind
	})

	if match != nil {
		return *match
	}

	return nil
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
			dereferencedChildren := linq.DereferenceSliceElements(diagEntity.Children)
			matching = append(matching, GetDiagnosticsWithSeverity(dereferencedChildren, severities)...)
		}
	}

	return matching
}

func CreateEntityDiagKey(kind, name string) string {
	return fmt.Sprintf("%s-%s", kind, name)
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
