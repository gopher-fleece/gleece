package annotations

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// PropertyDefinition defines the allowed properties for an annotation
type PropertyDefinition struct {
	Required      bool   // Whether the property is required
	DefaultValue  any    // Default value if not provided
	AllowedValues []any  // List of allowed values, if empty any value is allowed
	Type          string // Expected type (string, number, boolean, array, object)
}

// Validator contains validation logic for Gleece annotations
type Validator struct {
	// allowedAnnotations maps annotation names to their required parameters and contexts
	allowedAnnotations map[string]annotationDefinition
	createdAt          time.Time
	createdBy          string
}

type annotationDefinition struct {
	contexts            []CommentSource               // Where this annotation can be used (controller, route, schema, property)
	requiresValue       bool                          // Whether the annotation requires a basic parameter
	allowedProperties   map[string]PropertyDefinition // Allowed properties and their definitions
	allowsMultiple      bool                          // Whether multiple instances of this annotation are allowed
	mutuallyExclusive   []string                      // Names of annotations that cannot be used together with this annotation
	requiresUniqueValue bool                          // Whether the annotation value must be unique across all annotations
}

// NewValidator creates a new Gleece annotation validator
func NewValidator() *Validator {
	return &Validator{
		allowedAnnotations: initializeAnnotations(),
	}
}

// ValidateAnnotation checks if an annotation is valid according to Gleece specifications
func (v *Validator) ValidateAnnotation(attr Attribute, commentSource CommentSource) error {
	// Check if the annotation exists
	def, exists := v.allowedAnnotations[attr.Name]
	if !exists {
		return fmt.Errorf("unknown annotation @%s", attr.Name)
	}

	// Check if the annotation is valid in the given context
	validContext := slices.Contains(def.contexts, commentSource)
	if !validContext {
		return fmt.Errorf("annotation @%s is not valid in %s context", attr.Name, commentSource)
	}

	// Check if required value is provided
	if def.requiresValue && attr.Value == "" {
		return fmt.Errorf("annotation @%s requires a value", attr.Name)
	}

	// Validate properties
	if err := v.validateProperties(attr, def.allowedProperties); err != nil {
		return err
	}

	// Perform annotation-specific validation
	return v.validateSpecificAnnotation(attr)
}

// validateProperties checks if the provided properties are valid
func (v *Validator) validateProperties(attr Attribute, allowedProps map[string]PropertyDefinition) error {

	if allowedProps == nil {
		// If it's nil, any properties are allowed
		return nil
	}

	// If no properties are allowed, ensure none are provided
	if len(allowedProps) == 0 && len(attr.Properties) > 0 {
		return fmt.Errorf("annotation @%s does not support properties", attr.Name)
	}

	// Check each provided property
	for propName, propValue := range attr.Properties {
		// Check if the property is allowed
		propDef, allowed := allowedProps[propName]
		if !allowed {
			return fmt.Errorf("property %s is not allowed for annotation @%s", propName, attr.Name)
		}

		// Check if the property value is of the expected type
		if err := validatePropertyType(propName, propValue, propDef.Type); err != nil {
			return fmt.Errorf("invalid property %s for annotation @%s: %s", propName, attr.Name, err.Error())
		}

		// Check if the property value is among allowed values (if specified)
		if len(propDef.AllowedValues) > 0 {
			valid := false
			for _, allowedVal := range propDef.AllowedValues {
				if propValue == allowedVal {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid value for property %s of annotation @%s", propName, attr.Name)
			}
		}
	}

	// Check for missing required properties
	for propName, propDef := range allowedProps {
		if propDef.Required {
			_, provided := attr.Properties[propName]
			if !provided {
				return fmt.Errorf("required property %s is missing for annotation @%s", propName, attr.Name)
			}
		}
	}

	return nil
}

// validatePropertyType checks if a property value is of the expected type
func validatePropertyType(name string, value any, expectedType string) error {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		if !ok {
			return fmt.Errorf("property %s should be a string", name)
		}
	case "number":
		_, okFloat := value.(float64)
		_, okInt := value.(int)
		if !okFloat && !okInt {
			return fmt.Errorf("property %s should be a number", name)
		}
	case "boolean":
		_, ok := value.(bool)
		if !ok {
			return fmt.Errorf("property %s should be a boolean", name)
		}
	case "array":
		_, ok := value.([]any)
		if !ok {
			return fmt.Errorf("property %s should be an array", name)
		}
	case "object":
		_, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("property %s should be an object", name)
		}
	}
	return nil
}

// validateSpecificAnnotation performs specific validation for certain annotations
func (v *Validator) validateSpecificAnnotation(attr Attribute) error {
	switch attr.Name {
	case "Method":
		return validateMethod(attr.Value)
	case "Response", "ErrorResponse":
		return validateStatusCode(attr.Value)
	case "Security":
		return validateSecurity(attr)
	}
	return nil
}

// validateMethod checks if the HTTP method is valid
func validateMethod(method string) error {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	method = strings.ToUpper(method)

	for _, m := range validMethods {
		if m == method {
			return nil
		}
	}
	return fmt.Errorf("invalid HTTP method: %s", method)
}

// validateStatusCode checks if the status code is valid
func validateStatusCode(code string) error {
	// Convert to integer to check valid range
	_, err := strconv.Atoi(code)
	if err != nil {
		// This should never happen due to previous check
		return fmt.Errorf("invalid status code: %s", code)
	}

	return nil
}

// validateSecurity performs basic validation on security attributes
func validateSecurity(attr Attribute) error {
	// Could check for valid security scheme names or scope formats
	return nil
}

// ValidateAnnotationCollection validates a collection of annotations to ensure they conform to rules:
// 1. Annotations that don't allow multiple instances only appear once
// 2. Mutually exclusive annotations don't appear together
// 3. Annotations that require unique values have unique values across ALL annotations with uniqueness requirement
func (v *Validator) ValidateAnnotationCollection(attrs []Attribute, commentSource CommentSource) error {
	// Track frequency of each annotation
	annotationCount := make(map[string]int)

	// Track used values for annotations that require uniqueness
	// This map tracks values across ALL annotation types that require unique values
	uniqueValues := make(map[string]string) // map[value]annotationName

	for _, attr := range attrs {
		// Count each annotation type
		annotationCount[attr.Name]++

		// Get annotation definition
		def, _ := v.allowedAnnotations[attr.Name]

		// Check if this annotation allows multiple instances
		if !def.allowsMultiple && annotationCount[attr.Name] > 1 {
			return fmt.Errorf("multiple instances of annotation @%s are not allowed", attr.Name)
		}

		// Check for mutually exclusive annotations
		if len(def.mutuallyExclusive) > 0 {
			for _, exclusiveAttr := range def.mutuallyExclusive {
				if annotationCount[exclusiveAttr] > 0 {
					return fmt.Errorf("annotations @%s and @%s cannot be used together", attr.Name, exclusiveAttr)
				}
			}
		}

		// Check for unique values across all annotation types that require uniqueness
		if def.requiresUniqueValue && attr.Value != "" {
			// Check if this value has been used by any annotation type requiring uniqueness
			if previousAnnotation, exists := uniqueValues[attr.Value]; exists {
				return fmt.Errorf("duplicate value '%s' used in @%s and @%s annotations",
					attr.Value, previousAnnotation, attr.Name)
			}

			// Mark this value as used by this annotation type
			uniqueValues[attr.Value] = attr.Name
		}

		// Make sure Path annotations are used with Route URL
		if attr.Name == GleeceAnnotationPath {
			matchingURLPath := false
			pathName := attr.Value
			nameProp := attr.GetProperty(PropertyName)
			if nameProp != nil {
				if nameStr, ok := (*nameProp).(string); ok {
					pathName = nameStr
				}
			}
			expectedContains := fmt.Sprintf("{%s}", pathName)
			for _, otherAttr := range attrs {
				if otherAttr.Name == GleeceAnnotationRoute && strings.Contains(otherAttr.Value, expectedContains) {
					matchingURLPath = true
					break
				}
			}

			if !matchingURLPath {
				return fmt.Errorf("annotation @%s with name '%s' is not found in the route URL", attr.Name, pathName)
			}
		}

	}

	return nil
}

// initializeAnnotations creates a map of all valid annotations and their definitions
func initializeAnnotations() map[string]annotationDefinition {
	return map[string]annotationDefinition{
		// Controller (Class-Level) Annotations
		GleeceAnnotationTag: {
			contexts:            []CommentSource{"controller"},
			requiresValue:       true,
			allowedProperties:   map[string]PropertyDefinition{},
			allowsMultiple:      false,
			requiresUniqueValue: false,
		},
		GleeceAnnotationRoute: {
			contexts:            []CommentSource{"controller", "route"},
			requiresValue:       true,
			allowedProperties:   map[string]PropertyDefinition{},
			allowsMultiple:      false,
			requiresUniqueValue: false,
		},
		GleeceAnnotationSecurity: {
			contexts:      []CommentSource{"controller", "route"},
			requiresValue: true,
			allowedProperties: map[string]PropertyDefinition{
				"scopes": {
					Required:     false,
					Type:         "array",
					DefaultValue: []any{},
				},
			},
			allowsMultiple:      true,
			requiresUniqueValue: false,
		},
		GleeceAnnotationDescription: {
			contexts:            []CommentSource{"controller", "route", "schema", "property"},
			requiresValue:       false,
			allowedProperties:   map[string]PropertyDefinition{},
			allowsMultiple:      false,
			requiresUniqueValue: false,
		},
		GleeceAnnotationDeprecated: {
			contexts:            []CommentSource{"controller", "route", "schema", "property"},
			requiresValue:       false,
			allowedProperties:   map[string]PropertyDefinition{},
			allowsMultiple:      false,
			requiresUniqueValue: false,
		},

		// Route (Function-Level) Annotations
		GleeceAnnotationMethod: {
			contexts:            []CommentSource{"route"},
			requiresValue:       true,
			allowedProperties:   map[string]PropertyDefinition{},
			allowsMultiple:      false,
			requiresUniqueValue: false,
		},
		GleeceAnnotationQuery: {
			contexts:      []CommentSource{"route"},
			requiresValue: true,
			allowedProperties: map[string]PropertyDefinition{
				"name": {
					Required:     false,
					Type:         "string",
					DefaultValue: "",
				},
				"validate": {
					Required:     false,
					Type:         "string",
					DefaultValue: "",
				},
			},
			allowsMultiple:      true,
			requiresUniqueValue: true, // Values must be unique across all HTTP params annotations
		},
		GleeceAnnotationHeader: {
			contexts:      []CommentSource{"route"},
			requiresValue: true,
			allowedProperties: map[string]PropertyDefinition{
				"name": {
					Required:     false,
					Type:         "string",
					DefaultValue: "",
				},
				"validate": {
					Required:     false,
					Type:         "string",
					DefaultValue: "",
				},
			},
			allowsMultiple:      true,
			requiresUniqueValue: true, // Values must be unique across all HTTP params annotations
		},
		GleeceAnnotationPath: {
			contexts:      []CommentSource{"route"},
			requiresValue: true,
			allowedProperties: map[string]PropertyDefinition{
				"name": {
					Required:     false,
					Type:         "string",
					DefaultValue: "",
				},
				"validate": {
					Required:     false,
					Type:         "string",
					DefaultValue: "",
				},
			},
			allowsMultiple:      true,
			requiresUniqueValue: true, // Values must be unique across all HTTP params annotations
		},
		GleeceAnnotationBody: {
			contexts:      []CommentSource{"route"},
			requiresValue: true,
			allowedProperties: map[string]PropertyDefinition{
				"validate": {
					Required:     false,
					Type:         "string",
					DefaultValue: "",
				},
			},
			allowsMultiple:      false,
			mutuallyExclusive:   []string{GleeceAnnotationFormField},
			requiresUniqueValue: true, // Values must be unique across all HTTP params annotations
		},
		GleeceAnnotationFormField: {
			contexts:      []CommentSource{"route"},
			requiresValue: true,
			allowedProperties: map[string]PropertyDefinition{
				"name": {
					Required:     false,
					Type:         "string",
					DefaultValue: "",
				},
				"validate": {
					Required:     false,
					Type:         "string",
					DefaultValue: "",
				},
			},
			allowsMultiple:      true,
			mutuallyExclusive:   []string{GleeceAnnotationBody},
			requiresUniqueValue: true, // Values must be unique across all HTTP params annotations
		},
		GleeceAnnotationResponse: {
			contexts:            []CommentSource{"route"},
			requiresValue:       true,
			allowedProperties:   map[string]PropertyDefinition{},
			allowsMultiple:      true,
			requiresUniqueValue: false,
		},
		GleeceAnnotationErrorResponse: {
			contexts:            []CommentSource{"route"},
			requiresValue:       true,
			allowedProperties:   map[string]PropertyDefinition{},
			allowsMultiple:      true,
			requiresUniqueValue: false,
		},
		GleeceAnnotationHidden: {
			contexts:            []CommentSource{"route"},
			requiresValue:       false,
			allowedProperties:   map[string]PropertyDefinition{},
			allowsMultiple:      false,
			requiresUniqueValue: false,
		},
		GleeceAnnotationTemplateContext: {
			contexts:            []CommentSource{"route"},
			requiresValue:       true,
			allowedProperties:   nil, // Any properties are allowed
			allowsMultiple:      true,
			requiresUniqueValue: false,
		},
	}
}

// IsValidAnnotation is the exported function that checks if an annotation is valid
func IsValidAnnotation(attr Attribute, commentSource CommentSource) error {
	validator := NewValidator()
	return validator.ValidateAnnotation(attr, commentSource)
}

// IsValidAnnotationCollection is the exported function that validates a collection of annotations
// to ensure they follow the rules for multiple instances and mutual exclusivity
func IsValidAnnotationCollection(attrs []Attribute, commentSource CommentSource) error {
	validator := NewValidator()

	// First validate each annotation individually
	for _, attr := range attrs {
		if err := validator.ValidateAnnotation(attr, commentSource); err != nil {
			return err
		}
	}

	// Then validate the collection as a whole
	return validator.ValidateAnnotationCollection(attrs, commentSource)
}
