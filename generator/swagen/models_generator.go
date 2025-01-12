package swagen

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/haimkastner/gleece/definitions"
)

var objectType = &openapi3.Types{"object"}

func generateModelSpec(openapi *openapi3.T, model definitions.ModelMetadata) {
	schema := &openapi3.Schema{
		Title:       model.Name,
		Description: model.Description,
		Type:        objectType,
		Properties:  openapi3.Schemas{},
	}

	requiredFields := []string{}

	for _, field := range model.Fields {
		var fieldSchemaRef *openapi3.SchemaRef

		switch field.Type {
		case "string":
			fieldSchemaRef = &openapi3.SchemaRef{Value: openapi3.NewStringSchema()}
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
			fieldSchemaRef = &openapi3.SchemaRef{Value: openapi3.NewIntegerSchema()}
		case "bool":
			fieldSchemaRef = &openapi3.SchemaRef{Value: openapi3.NewBoolSchema()}
		case "float32", "float64":
			fieldSchemaRef = &openapi3.SchemaRef{Value: openapi3.NewFloat64Schema()}
		default:
			// Handle other types or complex types as references to other schemas
			fieldSchemaRef = &openapi3.SchemaRef{
				Ref: "#/components/schemas/" + field.Type,
			}
		}

		// fieldSchemaRef.Value.Description = field.Description

		// Add field to schema properties
		schema.Properties[field.Name] = fieldSchemaRef

		// If Validator indicates that the field is required, add it to the required list
		if field.Validator == "required" {
			requiredFields = append(requiredFields, field.Name)
		}
	}

	// Add required fields to schema
	if len(requiredFields) > 0 {
		schema.Required = requiredFields
	}

	// Add required fields to schema
	if len(requiredFields) > 0 {
		schema.Required = requiredFields
	}

	// Add schema to components
	openapi.Components.Schemas[model.Name] = &openapi3.SchemaRef{
		Value: schema,
	}
}

func fillSchemaRef(openapi *openapi3.T) {
	// Once building all models done, fill the missing references values (openapi.Components.Schemas)
	for _, schema := range openapi.Components.Schemas {
		for _, prop := range schema.Value.Properties {
			if prop.Ref != "" {
				// Get the name from the #/components/schemas/{name} format...
				propName := prop.Ref[len("#/components/schemas/"):]
				// ...and set the value to the actual schema
				prop.Value = openapi.Components.Schemas[propName].Value
			}
		}
	}
}

func GenerateModelsSpec(openapi *openapi3.T, models []definitions.ModelMetadata) {
	for _, model := range models {
		generateModelSpec(openapi, model)
	}
	fillSchemaRef(openapi)
}
