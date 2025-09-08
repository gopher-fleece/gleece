package validators

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/validators/configuration"
	"github.com/gopher-fleece/gleece/core/validators/diagnostics"
	"github.com/gopher-fleece/gleece/definitions"
)

type CommonValidator struct {
	holder *annotations.AnnotationHolder
}

func (v CommonValidator) Validate() []diagnostics.ResolvedDiagnostic {
	return v.validateCommon()
}

func (g *CommonValidator) validateCommon() []diagnostics.ResolvedDiagnostic {
	var diags []diagnostics.ResolvedDiagnostic

	// Track frequency of each annotation
	annotationCount := make(map[string]int)

	// Track used values for annotations that require uniqueness
	// This map tracks values across ALL annotation types that require unique values
	uniqueValues := make(map[string]string) // map[value]annotationName

	for _, attr := range g.holder.Attributes() {
		// Count each annotation type
		annotationCount[attr.Name]++

		// Get annotation definition
		def, exists := configuration.ValidatorConfigMap[attr.Name]
		if !exists {
			diags = append(diags, g.getDiagnosticForAttribute(
				attr,
				fmt.Sprintf("Unknown annotation \"@%s\"", attr.Name),
				"unknown-annotation",
				diagnostics.DiagnosticError,
			))

			// If there's no definition for hte annotation, there's no point in continuing hte validation chain
			continue
		}

		diags = append(diags, g.validateAnnotation(annotationCount, uniqueValues, def, attr)...)
	}

	return diags
}

func (g *CommonValidator) validateAnnotation(
	annotationCount map[string]int,
	uniqueValues map[string]string,
	def configuration.AnnotationConfigDefinition,
	attr annotations.Attribute,
) []diagnostics.ResolvedDiagnostic {
	var diags []diagnostics.ResolvedDiagnostic

	// Check if this annotation allows multiple instances
	diags = common.AppendIfNotNil(diags, g.validateDuplicateAnnotation(annotationCount, def, attr))

	// Check for mutually exclusive annotations
	diags = common.AppendIfNotNil(diags, g.validateMutuallyExclusive(annotationCount, def, attr))

	// Check for unique values across all annotation types that require uniqueness
	diags = common.AppendIfNotNil(diags, g.validateUniqueValue(uniqueValues, def, attr))

	diags = common.AppendIfNotNil(diags, g.validateAttribute(attr))

	return diags
}

func (g *CommonValidator) validateDuplicateAnnotation(
	annotationCount map[string]int,
	def configuration.AnnotationConfigDefinition,
	attr annotations.Attribute,
) *diagnostics.ResolvedDiagnostic {
	// Check if this annotation allows multiple instances
	if !def.AllowsMultiple && annotationCount[attr.Name] > 1 {
		diag := g.getDiagnosticForAttribute(
			attr,
			fmt.Sprintf("multiple instances of annotation @%s are not allowed", attr.Name),
			"duplicate-annotation",
			diagnostics.DiagnosticWarning,
		)
		return &diag
	}

	return nil
}

func (g *CommonValidator) validateMutuallyExclusive(
	annotationCount map[string]int,
	def configuration.AnnotationConfigDefinition,
	attr annotations.Attribute,
) *diagnostics.ResolvedDiagnostic {
	if len(def.MutuallyExclusive) > 0 {
		for _, exclusiveAttr := range def.MutuallyExclusive {
			if annotationCount[exclusiveAttr] > 0 {
				diag := g.getDiagnosticForAttribute(
					attr,
					fmt.Sprintf("annotations @%s and @%s cannot be used together", attr.Name, exclusiveAttr),
					"mutually-exclusive-annotations",
					diagnostics.DiagnosticError,
				)
				return &diag
			}
		}
	}
	return nil
}

func (g *CommonValidator) validateUniqueValue(
	uniqueValues map[string]string,
	def configuration.AnnotationConfigDefinition,
	attr annotations.Attribute,
) *diagnostics.ResolvedDiagnostic {
	var diagToReturn *diagnostics.ResolvedDiagnostic

	if def.RequiresUniqueValue && attr.Value != "" {
		// Check if this value has been used by any annotation type requiring uniqueness
		if previousAnnotation, exists := uniqueValues[attr.Value]; exists {
			diag := g.getDiagnosticForAttribute(
				attr,
				fmt.Sprintf(
					"duplicate value '%s' used in @%s and @%s annotations",
					attr.Value,
					previousAnnotation,
					attr.Name,
				),
				"duplicate-annotation-value",
				diagnostics.DiagnosticError,
			)
			diagToReturn = &diag
		}
	}

	// Mark this value as used by this annotation type
	uniqueValues[attr.Value] = attr.Name
	return diagToReturn
}

// validateAttribute performs specific validation for certain annotations
func (g *CommonValidator) validateAttribute(attr annotations.Attribute) *diagnostics.ResolvedDiagnostic {
	switch attr.Name {
	case annotations.GleeceAnnotationMethod:
		return g.validateMethodAttribute(attr)
	case annotations.GleeceAnnotationResponse, annotations.GleeceAnnotationErrorResponse:
		return g.validateStatusCodeBearingAttribute(attr)
	}
	return nil
}

// validateMethodAttribute checks if the HTTP verb is valid
func (g *CommonValidator) validateMethodAttribute(attribute annotations.Attribute) *diagnostics.ResolvedDiagnostic {
	isSupported := definitions.IsValidRouteHttpVerb(attribute.Value)

	if isSupported {
		return nil
	}

	isValid := definitions.IsValidHttpVerb(attribute.Value)

	var diag diagnostics.ResolvedDiagnostic
	supportedVerbsMsg := fmt.Sprintf(
		"Supported verbs are: %s",
		strings.Join(definitions.GetRouteSupportedHttpVerbs(), ", "),
	)

	if isValid {
		diag = g.getDiagnosticForAttributeValue(
			attribute,
			fmt.Sprintf(
				"HTTP verb '%s' is currently unsupported for @Method annotations. %s",
				attribute.Value,
				supportedVerbsMsg,
			),
			diagnostics.DiagFeatureUnsupported,
			diagnostics.DiagnosticError,
		)
	} else {
		diag = g.getDiagnosticForAttributeValue(
			attribute,
			fmt.Sprintf(
				"Invalid HTTP verb '%s'. %s",
				attribute.Value,
				supportedVerbsMsg,
			),
			diagnostics.DiagFeatureUnsupported,
			diagnostics.DiagnosticError,
		)
	}

	return &diag
}

// validateStatusCodeBearingAttribute checks if the status code is valid
func (g *CommonValidator) validateStatusCodeBearingAttribute(attribute annotations.Attribute) *diagnostics.ResolvedDiagnostic {
	parsed, err := strconv.ParseUint(attribute.Value, 10, 32)
	if err != nil {
		return common.Ptr(
			g.getDiagnosticForAttributeValue(
				attribute,
				fmt.Sprintf("Non-numeric HTTP status code '%s'", attribute.Value),
				diagnostics.DiagAnnotationValueInvalid,
				diagnostics.DiagnosticError,
			),
		)
	}

	parsedCode := uint(parsed)
	isKnownCode := definitions.IsValidHttpStatusCode(parsedCode)
	if isKnownCode {
		return nil
	}

	return common.Ptr(
		g.getDiagnosticForAttributeValue(
			attribute,
			fmt.Sprintf("Non-standard HTTP status code '%s'", attribute.Value),
			diagnostics.DiagAnnotationValueInvalid,
			diagnostics.DiagnosticWarning,
		),
	)
}

func (g *CommonValidator) getDiagnosticForAttribute(
	attribute annotations.Attribute,
	message string,
	code diagnostics.DiagnosticCode,
	severity diagnostics.DiagnosticSeverity,
) diagnostics.ResolvedDiagnostic {
	return diagnostics.NewDiagnostic(
		g.holder.FileName(),
		message,
		code,
		severity,
		common.ResolvedRange{
			StartLine: max(attribute.Comment.Position.StartLine, 0),
			StartCol:  max(attribute.Comment.Position.StartCol, 0),
			EndLine:   max(attribute.Comment.Position.EndLine, 0),
			EndCol:    max(attribute.Comment.Position.EndCol, 0),
		},
	)
}

func (g *CommonValidator) getDiagnosticForAttributeValue(
	attribute annotations.Attribute,
	message string,
	code diagnostics.DiagnosticCode,
	severity diagnostics.DiagnosticSeverity,
) diagnostics.ResolvedDiagnostic {
	return diagnostics.NewDiagnostic(
		g.holder.FileName(),
		message,
		code,
		severity,
		attribute.GetValueRange(),
	)
}

func (b *CommonValidator) createMayNotHaveAnnotation(
	entity string,
	attrib annotations.Attribute,
) diagnostics.ResolvedDiagnostic {
	var code diagnostics.DiagnosticCode

	if entity == "Controllers" {
		code = diagnostics.DiagControllerLevelAnnotationNotAllowed
	} else {
		code = diagnostics.DiagMethodLevelAnnotationNotAllowed
	}

	return diagnostics.NewErrorDiagnostic(
		b.holder.FileName(),
		fmt.Sprintf("%s may not have @%s annotations", entity, attrib.Name),
		code,
		attrib.Comment.Range(),
	)
}
