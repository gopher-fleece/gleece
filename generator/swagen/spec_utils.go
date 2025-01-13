package swagen

import (
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/haimkastner/gleece/definitions"
)

func HttpStatusCodeToString(httpStatusCode definitions.HttpStatusCode) string {
	statusCode := uint64(httpStatusCode)
	return strconv.FormatUint(statusCode, 10)
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
	default:
		return openapi3.NewObjectSchema()
	}
}

func ToOpenApiSchemaRef(typeName string) *openapi3.SchemaRef {
	schema := ToOpenApiSchema(typeName)
	return &openapi3.SchemaRef{
		Value: schema,
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

// Helper function to parse numeric validation values
func ParseNumber(value string) *float64 {
	if v, err := strconv.ParseFloat(value, 64); err == nil {
		return &v
	}
	return nil
}

// Helper function to parse integer validation values
func ParseInteger(value string) *uint64 {
	if v, err := strconv.ParseUint(value, 10, 64); err == nil {
		return &v
	}
	return nil
}

// Helper function to parse boolean validation values
func ParseBool(value string) *bool {
	if v, err := strconv.ParseBool(value); err == nil {
		return &v
	}
	return nil
}

func IsFieldRequired(validationString string) bool {
	validationRules := strings.Split(validationString, ",")
	for _, rule := range validationRules {
		if rule == "required" {
			return true
		}
	}
	return false
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
				schema.Value.Min = ParseNumber(ruleValue)
			}
		case "lt":
			// Less than (only makes sense for numeric fields)
			if specType == "integer" || specType == "number" {
				schema.Value.Max = ParseNumber(ruleValue)
			}
		case "min":
			// Minimum length for strings, minimum value for numbers
			if specType == "string" {
				schema.Value.MinLength = *ParseInteger(ruleValue)
			} else if specType == "integer" || specType == "number" {
				schema.Value.Min = ParseNumber(ruleValue)
			}
		case "max":
			// Maximum length for strings, maximum value for numbers
			if specType == "string" {
				schema.Value.MaxLength = ParseInteger(ruleValue)
			} else if specType == "integer" || specType == "number" {
				schema.Value.Max = ParseNumber(ruleValue)
			}
		case "pattern":
			// Regular expression pattern for strings
			if specType == "string" {
				schema.Value.Pattern = ruleValue
			}
		case "minItems":
			// Minimum number of items for arrays
			if specType == "array" {
				schema.Value.MinItems = *ParseInteger(ruleValue)
			}
		case "maxItems":
			// Maximum number of items for arrays
			if specType == "array" {
				schema.Value.MaxItems = ParseInteger(ruleValue)
			}
		case "uniqueItems":
			// Ensure all items in the array are unique
			if specType == "array" {
				schema.Value.UniqueItems = *ParseBool(ruleValue)
			}
		case "enum":
			// Enum values for strings or numbers
			enumValues := strings.Split(ruleValue, "|")
			for _, v := range enumValues {
				schema.Value.Enum = append(schema.Value.Enum, v)
			}
		}
	}
}
