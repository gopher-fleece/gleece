package swagen

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

type SchemaRefMap struct {
	InterfaceType string
	SchemaRef     *openapi3.SchemaRef
}

// Create map of SchemaRefMap to store the schema references
var schemaRefMap = []SchemaRefMap{} // Initialize as an empty slice

func InterfaceToSchemaRef(openapi *openapi3.T, interfaceType string) *openapi3.SchemaRef {
	openapiType := ToOpenApiType(interfaceType)
	fieldSchemaRef := ToOpenApiSchemaRef(openapiType)

	if openapiType == "object" && !IsMapObject(interfaceType) { // For now, ignore map objects, they will be handled later
		// Handle other types or complex types as references to other schemas
		fieldSchemaRef = &openapi3.SchemaRef{
			Ref: "#/components/schemas/" + interfaceType,
		}

		if openapi.Components.Schemas[interfaceType] != nil {
			fieldSchemaRef.Value = openapi.Components.Schemas[interfaceType].Value
		} else {
			// If the schema is not found in the components, add it to the EmptyRefSchemas slice so later we can fill it
			schemaRefMap = append(schemaRefMap, SchemaRefMap{InterfaceType: interfaceType, SchemaRef: fieldSchemaRef})
		}
	}
	if openapiType == "array" {
		// Handle array types
		itemType := GetArrayItemType(interfaceType)
		// Once the item type is determined, create a schema reference for it in a recursive manner
		itemSchemaRef := InterfaceToSchemaRef(openapi, itemType)
		fieldSchemaRef = &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:  arrayType,
				Items: itemSchemaRef,
			},
		}
	}
	return fieldSchemaRef
}

func BuildSchemaValidation(schema *openapi3.SchemaRef, validationString string, fieldInterface string) {

	// Parse and apply validation rules from the Validator field
	validationRules := strings.Split(validationString, ",")
	for _, rule := range validationRules {
		parts := strings.SplitN(rule, "=", 2)
		ruleName := parts[0]
		var ruleValue string
		if len(parts) > 1 {
			ruleValue = parts[1]
		}

		specType := ToOpenApiType(fieldInterface)
		switch ruleName {
		case "email":
			schema.Value.Format = "email"
		case "gt":
			// Greater than (only makes sense for numeric fields)
			if specType == "integer" || specType == "number" {
				// For gt, we need to set ExclusiveMinimum to true and use the value as minimum
				schema.Value.Min = ParseNumber(ruleValue)
				schema.Value.ExclusiveMin = true
			} else {
				logger.Warn("Validation rule 'gt' is only applicable to numeric fields")
			}
		case "gte":
			// Greater than or equal (only makes sense for numeric fields)
			if specType == "integer" || specType == "number" {
				// For gte, we just set the minimum value
				schema.Value.Min = ParseNumber(ruleValue)
				// ExclusiveMinimum should be false (default) or explicitly set to false
				schema.Value.ExclusiveMin = false
			} else {
				logger.Warn("Validation rule 'gte' is only applicable to numeric fields")
			}
		case "lt":
			// Less than (only makes sense for numeric fields)
			if specType == "integer" || specType == "number" {
				// For lt, we need to set ExclusiveMaximum to true and use the value as maximum
				schema.Value.Max = ParseNumber(ruleValue)
				schema.Value.ExclusiveMax = true
			} else {
				logger.Warn("Validation rule 'lt' is only applicable to numeric fields")
			}
		case "lte":
			// Less than or equal (only makes sense for numeric fields)
			if specType == "integer" || specType == "number" {
				// For lte, we just set the maximum value
				schema.Value.Max = ParseNumber(ruleValue)
				// ExclusiveMaximum should be false (default) or explicitly set to false
				schema.Value.ExclusiveMax = false
			} else {
				logger.Warn("Validation rule 'lte' is only applicable to numeric fields")
			}
		case "min":
			// Minimum length for strings, minimum value for numbers
			if specType == "string" {
				schema.Value.MinLength = *ParseInteger(ruleValue)
			} else if specType == "integer" || specType == "number" {
				schema.Value.Min = ParseNumber(ruleValue)
				schema.Value.ExclusiveMin = false
			} else {
				logger.Warn("Validation rule 'min' is only applicable to string or numeric fields")
			}
		case "max":
			// Maximum length for strings, maximum value for numbers
			if specType == "string" {
				schema.Value.MaxLength = ParseInteger(ruleValue)
			} else if specType == "integer" || specType == "number" {
				schema.Value.Max = ParseNumber(ruleValue)
				schema.Value.ExclusiveMax = false
			} else {
				logger.Warn("Validation rule 'max' is only applicable to string or numeric fields")
			}
		case "pattern":
			// Regular expression pattern for strings
			if specType == "string" {
				schema.Value.Pattern = ruleValue
			} else {
				logger.Warn("Validation rule 'pattern' is only applicable to string fields")
			}
		case "minItems":
			// Minimum number of items for arrays
			if specType == "array" {
				schema.Value.MinItems = *ParseInteger(ruleValue)
			} else {
				logger.Warn("Validation rule 'minItems' is only applicable to array fields")
			}
		case "maxItems":
			// Maximum number of items for arrays
			if specType == "array" {
				schema.Value.MaxItems = ParseInteger(ruleValue)
			} else {
				logger.Warn("Validation rule 'maxItems' is only applicable to array fields")
			}
		case "uniqueItems":
			// Ensure all items in the array are unique
			if specType == "array" {
				schema.Value.UniqueItems = *ParseBool(ruleValue)
			} else {
				logger.Warn("Validation rule 'uniqueItems' is only applicable to array fields")
			}
		case "enum":
			// Enum values for strings or numbers
			enumValues := strings.Split(ruleValue, "|")
			if enumValues == nil || len(enumValues) == 0 || enumValues[0] == "" {
				logger.Warn("Validation rule 'enum' must have at least one value")
				schema.Value.Enum = nil
			} else {
				for _, v := range enumValues {
					schema.Value.Enum = append(schema.Value.Enum, v)
				}
			}
		}
	}
}

func IsPrimitiveType(typeName string) bool {
	switch typeName {
	case "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "bool", "float32", "float64":
		return true
	default:
		return false
	}
}

func ToOpenApiType(typeName string) string {
	switch typeName {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "integer"
	case "bool":
		return "boolean"
	case "float32", "float64":
		return "number"
	default:
		if strings.HasPrefix(typeName, "[]") {
			return "array"
		}
		return "object"
	}
}

func ToOpenApiSchema(typeName string) *openapi3.Schema {
	// Implement the conversion of the typeName to an OpenAPI schema
	switch typeName {
	case "string":
		return openapi3.NewStringSchema()
	case "integer":
		return openapi3.NewIntegerSchema()
	case "boolean":
		return openapi3.NewBoolSchema()
	case "number":
		return openapi3.NewFloat64Schema()
	case "array":
		return openapi3.NewArraySchema()
	case "object":
		return openapi3.NewObjectSchema()
	default:
		logger.Fatal("Unknown type: " + typeName)
		return openapi3.NewSchema()
	}
}

func ToOpenApiSchemaRef(typeName string) *openapi3.SchemaRef {
	schema := ToOpenApiSchema(typeName)
	return &openapi3.SchemaRef{
		Value: schema,
	}
}

func IsSecurityNameInSecuritySchemes(securitySchemes []definitions.SecuritySchemeConfig, securityName string) bool {
	for _, securityScheme := range securitySchemes {
		if securityScheme.SecurityName == securityName {
			return true
		}
	}
	return false
}

func IsHiddenAsset(hideOptions *definitions.MethodHideOptions) bool {
	if hideOptions == nil {
		return false
	}
	if hideOptions.Type == definitions.HideMethodNever {
		return false
	}
	if hideOptions.Type == definitions.HideMethodAlways {
		return true
	}

	// TODO: Check the condition...
	return false
}

func IsDeprecated(deprecationOptions *definitions.DeprecationOptions) bool {
	return deprecationOptions != nil && deprecationOptions.Deprecated
}

// GetTagValue extracts the value for a specific tag name from a struct tag string
// If the tag or value is not found, returns the default value
// Example usage:
//
//	tag := `json:"houseNumber" validate:"gte=1"`
//	jsonValue := GetTagValue(tag, "json", "default") // returns "houseNumber"
//	validateValue := GetTagValue(tag, "validate", "default") // returns "gte=1"
func GetTagValue(tagStr string, tagName string, defaultValue string) string {
	// Split the tag string into individual tag parts
	tags := strings.Split(tagStr, " ")

	// Look for the requested tag
	prefix := tagName + ":"
	for _, tag := range tags {
		// Remove any quotes around the tag
		tag = strings.Trim(tag, "`")

		if strings.HasPrefix(tag, prefix) {
			// Extract the value between the quotes if they exist
			value := strings.TrimPrefix(tag, prefix)
			value = strings.Trim(value, "\"")

			// If empty value, return default
			if value == "" {
				return defaultValue
			}
			return value
		}
	}

	// If tag not found, return default
	return defaultValue
}

func IsMapObject(typeName string) bool {
	return strings.HasPrefix(typeName, "map[")
}
