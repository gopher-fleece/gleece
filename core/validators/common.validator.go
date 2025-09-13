package validators

import (
	"fmt"
	"slices"
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

func (g CommonValidator) Validate() []diagnostics.ResolvedDiagnostic {
	return g.validateCommon()
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
				fmt.Sprintf("Unknown annotation '@%s'", attr.Name),
				diagnostics.DiagAnnotationUnknown,
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

	// Check if annotation is allowed on this entity.
	// Technically we can also stop here - no real reason to lint further
	diags = common.AppendIfNotNil(diags, g.validateAnnotationValidInContext(def, attr))

	// Check if the annotation requires/has a value
	diags = common.AppendIfNotNil(diags, g.validateRequiredAnnotationValue(def, attr))

	// Check if the annotation properties are as expected
	diags = common.AppendIfNotNil(diags, g.validateAnnotationProperties(def, attr))

	// Check if this annotation allows multiple instances
	diags = common.AppendIfNotNil(diags, g.validateDuplicateAnnotation(annotationCount, def, attr))

	// Check for mutually exclusive annotations
	diags = common.AppendIfNotNil(diags, g.validateMutuallyExclusive(annotationCount, def, attr))

	// Check for unique values across all annotation types that require uniqueness
	diags = common.AppendIfNotNil(diags, g.validateUniqueValue(uniqueValues, def, attr))

	diags = common.AppendIfNotNil(diags, g.validateAttribute(attr))

	return diags
}

func (g *CommonValidator) validateAnnotationValidInContext(
	def configuration.AnnotationConfigDefinition,
	attr annotations.Attribute,
) *diagnostics.ResolvedDiagnostic {
	if slices.Contains(def.Contexts, g.holder.Source()) {
		// Context is valid
		return nil
	}

	diag := g.getDiagnosticForAttribute(
		attr,
		fmt.Sprintf("Annotation '@%s' is not valid in the context of a %s", attr.Name, g.holder.Source()),
		diagnostics.DiagAnnotationInvalidInContext,
		diagnostics.DiagnosticWarning,
	)

	return &diag
}

func (g *CommonValidator) validateRequiredAnnotationValue(
	def configuration.AnnotationConfigDefinition,
	attr annotations.Attribute,
) *diagnostics.ResolvedDiagnostic {
	if !def.RequiresValue || attr.Value != "" {
		return nil
	}

	diag := g.getDiagnosticForAttribute(
		attr,
		fmt.Sprintf("Annotation '@%s' requires a value", attr.Name),
		diagnostics.DiagAnnotationValueMustExist,
		diagnostics.DiagnosticError,
	)

	return &diag
}

func (g *CommonValidator) validateAnnotationProperties(
	def configuration.AnnotationConfigDefinition,
	attr annotations.Attribute,
) *diagnostics.ResolvedDiagnostic {
	if def.AllowedProperties == nil {
		// If it's nil, any properties are allowed
		return nil
	}

	// If no properties are allowed, ensure none are provided
	if len(def.AllowedProperties) == 0 && len(attr.Properties) > 0 {
		diag := g.getDiagnosticForAttribute(
			attr,
			fmt.Sprintf("Annotation '@%s' does not support properties", attr.Name),
			diagnostics.DiagAnnotationPropertiesShouldNotExist,
			diagnostics.DiagnosticWarning,
		)

		return &diag
	}

	// Check each provided property
	for propName, propValue := range attr.Properties {
		// Check if the property is allowed
		propDef, allowed := def.AllowedProperties[propName]
		if !allowed {
			diag := g.getDiagnosticForAttribute(
				attr,
				fmt.Sprintf("Property '%s' is not allowed for annotation '@%s'", propName, attr.Name),
				diagnostics.DiagAnnotationPropertyShouldNotExist,
				diagnostics.DiagnosticWarning,
			)

			return &diag
		}

		// Check if the property value is of the expected type
		propValueDiag := g.validatePropertyType(attr, propName, propValue, propDef.Type)
		if propValueDiag != nil {
			return propValueDiag
		}

		// Check if the property value is among allowed values (if specified)
		if len(propDef.AllowedValues) > 0 {
			validValue := slices.Contains(propDef.AllowedValues, propValue)

			if !validValue {
				diag := g.getDiagnosticForAttribute(
					attr,
					fmt.Sprintf("Invalid value for property '%s'", propName), // Need a way to provide better hint as to what's expected
					diagnostics.DiagAnnotationPropertiesInvalidValueForKey,
					diagnostics.DiagnosticError,
				)

				return &diag
			}
		}
	}

	// Check for missing required properties
	for propName, propDef := range def.AllowedProperties {
		if propDef.Required {
			_, provided := attr.Properties[propName]
			if !provided {
				diag := g.getDiagnosticForAttribute(
					attr,
					fmt.Sprintf("Missing required property '%s'", propName),
					diagnostics.DiagAnnotationPropertiesMissingKey,
					diagnostics.DiagnosticError,
				)

				return &diag
			}
		}
	}

	return nil
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
			fmt.Sprintf("Multiple instances of '@%s' annotations are not allowed", attr.Name),
			diagnostics.DiagAnnotationDuplicate,
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
					fmt.Sprintf("Annotations '@%s' and '@%s' are mutually exclusive", attr.Name, exclusiveAttr),
					diagnostics.DiagAnnotationMutuallyExclusive,
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
		if _, exists := uniqueValues[attr.Value]; exists {
			diag := g.getDiagnosticForAttribute(
				attr,
				fmt.Sprintf(
					"Duplicate value '%s' referenced by multiple annotations",
					attr.Value,
				),
				diagnostics.DiagAnnotationDuplicateValue,
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
			diagnostics.DiagAnnotationValueInvalid,
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

func (g *CommonValidator) createMayNotHaveAnnotation(
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
		g.holder.FileName(),
		fmt.Sprintf("%s may not have @%s annotations", entity, attrib.Name),
		code,
		attrib.Comment.Range(),
	)
}

// validatePropertyType checks if a property value is of the expected type
// This method is far from ideal and should be updated to handle complex slice types and such.
func (g *CommonValidator) validatePropertyType(
	attr annotations.Attribute,
	propName string,
	value any,
	expectedType string,
) *diagnostics.ResolvedDiagnostic {

	var errMessage string
	switch expectedType {
	case "string":
		_, ok := value.(string)
		if !ok {
			errMessage = fmt.Sprintf("Property '%s' expected to be a string", propName)
		}
	case "number":
		_, okFloat := value.(float64)
		_, okInt := value.(int)
		if !okFloat && !okInt {
			errMessage = fmt.Sprintf("Property '%s' expected to be a number", propName)
		}
	case "boolean":
		_, ok := value.(bool)
		if !ok {
			errMessage = fmt.Sprintf("Property '%s' expected to be a boolean", propName)
		}
	case "array":
		_, ok := value.([]any)
		if !ok {
			errMessage = fmt.Sprintf("Property '%s' expected to be an array", propName)
		}
	case "object":
		_, ok := value.(map[string]any)
		if !ok {
			errMessage = fmt.Sprintf("Property '%s' expected to be an object", propName)
		}
	}

	if errMessage == "" {
		return nil
	}

	diag := g.getDiagnosticForAttribute(
		attr,
		errMessage,
		diagnostics.DiagAnnotationPropertiesInvalidValueForKey,
		diagnostics.DiagnosticWarning,
	)

	return &diag
}
