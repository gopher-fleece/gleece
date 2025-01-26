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
	// Look for the requested tag
	prefix := tagName + ":\""

	// Find the start of the tag value
	start := strings.Index(tagStr, prefix)
	if start == -1 {
		return defaultValue
	}
	start += len(prefix)

	// Find the end of the tag value
	end := start
	for end < len(tagStr) && tagStr[end] != '"' {
		end++
	}

	// Extract and return the tag value
	if start < end {
		return tagStr[start:end]
	}

	// If tag value is empty, return default
	return defaultValue
}

func IsMapObject(typeName string) bool {
	return strings.HasPrefix(typeName, "map[")
}
