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
		fieldSchemaRef := ToOpenApiSchemaRef(field.Type)
		if ToOpenApiType(field.Type) == "object" {
			// Handle other types or complex types as references to other schemas
			fieldSchemaRef = &openapi3.SchemaRef{
				Ref: "#/components/schemas/" + field.Type,
			}
		}

		if fieldSchemaRef.Value != nil {
			fieldSchemaRef.Value.Description = field.Description
		}

		BuildSchemaValidation(fieldSchemaRef, field.Validator, field.Type)

		// Add field to schema properties
		schema.Properties[field.Name] = fieldSchemaRef

		// If the field should be required, add its name to the requiredFields slice
		if IsFieldRequired(field.Validator) {
			requiredFields = append(requiredFields, field.Name)
		}
	}

	// Add required fields to the schema
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
