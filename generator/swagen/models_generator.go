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

func fillSchemaRef(openapi *openapi3.T) {
	// Once building all models is done, fill the missing references values (openapi.Components.Schemas)
	for _, schema := range openapi.Components.Schemas {
		for _, prop := range schema.Value.Properties {
			fillPropertyRef(openapi, prop)
		}
	}
}

func fillPropertyRef(openapi *openapi3.T, prop *openapi3.SchemaRef) {
	if prop.Ref != "" {
		// Get the name from the #/components/schemas/{name} format...
		propName := prop.Ref[len("#/components/schemas/"):]
		// ...and set the value to the actual schema
		prop.Value = openapi.Components.Schemas[propName].Value
	}

	if prop.Value != nil && prop.Value.Items != nil && prop.Value.Items.Ref != "" {
		// Handle array item references
		itemPropName := prop.Value.Items.Ref[len("#/components/schemas/"):]
		prop.Value.Items.Value = openapi.Components.Schemas[itemPropName].Value
	}

	// Recursively fill references for nested objects if they exist
	if prop.Value != nil && prop.Value.Properties != nil {
		for _, nestedProp := range prop.Value.Properties {
			fillPropertyRef(openapi, nestedProp)
		}
	}
}

func GenerateModelsSpec(openapi *openapi3.T, models []definitions.ModelMetadata) {
	for _, model := range models {
		generateModelSpec(openapi, model)
	}
	fillSchemaRef(openapi)
}
