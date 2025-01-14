package swagen

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/haimkastner/gleece/definitions"
)

var objectType = &openapi3.Types{"object"}
var arrayType = &openapi3.Types{"array"}

func generateModelSpec(openapi *openapi3.T, model definitions.ModelMetadata) {
	schema := &openapi3.Schema{
		Title:       model.Name,
		Description: model.Description,
		Type:        objectType,
		Properties:  openapi3.Schemas{},
	}

	requiredFields := []string{}

	for _, field := range model.Fields {
		fieldSchemaRef := InterfaceToSchemaRef(openapi, field.Type)

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

// Fill the schema references in the components
func fillSchemaRef(openapi *openapi3.T) {
	// Iterate over all EmptyRefSchemas
	for _, schemaRefMap := range schemaRefMap {
		// Get the schema from the components
		schema := openapi.Components.Schemas[schemaRefMap.InterfaceType].Value
		// Set the schema reference to the schema
		schemaRefMap.SchemaRef.Value = schema
	}
}

func GenerateModelsSpec(openapi *openapi3.T, models []definitions.ModelMetadata) {
	for _, model := range models {
		generateModelSpec(openapi, model)
	}
	fillSchemaRef(openapi)
}
