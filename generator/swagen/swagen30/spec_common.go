package swagen30

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gopher-fleece/gleece/generator/swagen/swagtool"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

type SchemaRefMap struct {
	InterfaceType string
	SchemaRef     *openapi3.SchemaRef
}

// Create map of SchemaRefMap to store the schema references
var schemaRefMap = []SchemaRefMap{} // Initialize as an empty slice

func InterfaceToSchemaRef(openapi *openapi3.T, interfaceType string) *openapi3.SchemaRef {
	openapiType := swagtool.ToOpenApiType(interfaceType)
	fieldSchemaRef := ToOpenApiSchemaRef(openapiType)

	if openapiType == "object" && !swagtool.IsMapObject(interfaceType) { // For now, ignore map objects, they will be handled later
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
		itemType := swagtool.GetArrayItemType(interfaceType)
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
