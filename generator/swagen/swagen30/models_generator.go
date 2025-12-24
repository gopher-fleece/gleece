package swagen30

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/generator/swagen/swagtool"
)

var objectType = &openapi3.Types{"object"}
var arrayType = &openapi3.Types{"array"}

func generateStructSpec(openapi *openapi3.T, model definitions.StructMetadata) {
	// The final schema that will be added to the components
	var modelSchema *openapi3.Schema

	// The schema that will be added to all fields (in case of embedded fields it will be part of the allOf array)
	schema := &openapi3.Schema{
		Title:       model.Name,
		Description: model.Description,
		Type:        objectType,
		Properties:  openapi3.Schemas{},
		Deprecated:  swagtool.IsDeprecated(&model.Deprecation),
	}

	var hasEmbeddedField bool
	var relevantFields []definitions.FieldMetadata

	// We need to ignore nested 'error' fields to preserve existing behavior.
	// Could probably be made better.
	// Perhaps switch the entire spec generation process to use the richer AST-containing representation?
	for _, field := range model.Fields {
		if field.IsEmbedded {
			if field.Type == "error" {
				continue
			}
			hasEmbeddedField = true
		}
		relevantFields = append(relevantFields, field)
	}

	if !hasEmbeddedField {
		// If the model has no embedded fields, the schema can be used directly
		modelSchema = schema
	} else {
		// If the model has embedded fields, the schema should be part of the allOf array
		modelSchema = &openapi3.Schema{}
		modelSchema.AllOf = make([]*openapi3.SchemaRef, 0)
		// Add the schema to the allOf array
		modelSchema.AllOf = append(modelSchema.AllOf, &openapi3.SchemaRef{
			Value: schema,
		})
	}

	requiredFields := []string{}

	for _, field := range relevantFields {
		fieldSchemaRef := InterfaceToSchemaRef(openapi, field.Type)

		if field.IsEmbedded {
			// If the field is embedded, add the embedded schema to the allOf array
			modelSchema.AllOf = append(modelSchema.AllOf, fieldSchemaRef)
			continue
		}
		validationTag := swagtool.GetTagValue(field.Tag, "validate", "")
		BuildSchemaValidation(fieldSchemaRef, validationTag, field.Type)

		// OpenAPI 3.0 does not support any extra field in the SchemaRef beside just ref, and we don't want to override the model properties themselves
		if fieldSchemaRef.Value != nil && fieldSchemaRef.Ref == "" {
			fieldSchemaRef.Value.Description = field.Description

			// If the schema marked as deprecated, the field / property should be marked as deprecated as well
			// Setting it as not deprecated (even if the field itself is not marked deprecated) will override the model deprecation
			if !fieldSchemaRef.Value.Deprecated {
				fieldSchemaRef.Value.Deprecated = swagtool.IsDeprecated(field.Deprecation)
			}
		}

		// Add field to schema properties
		fName := swagtool.GetJsonNameFromTag(field.Tag, field.Name)
		schema.Properties[fName] = fieldSchemaRef

		// If the field should be required, add its name to the requiredFields slice
		if swagtool.IsFieldRequired(validationTag) {
			requiredFields = append(requiredFields, fName)
		}
	}

	// Add required fields to the schema
	if len(requiredFields) > 0 {
		schema.Required = requiredFields
	}

	// Add schema to components
	openapi.Components.Schemas[model.Name] = &openapi3.SchemaRef{
		Value: modelSchema,
	}
}

func generateEnumSpec(openapi *openapi3.T, model definitions.EnumMetadata) {
	enumType := &openapi3.Types{swagtool.ToOpenApiType(model.Type)}

	// Create the enum schema
	schema := &openapi3.Schema{
		Title:       model.Name,
		Description: model.Description,
		Type:        enumType,
		Deprecated:  swagtool.IsDeprecated(&model.Deprecation),
	}

	// Add the possible enum values
	enumValues := []any{}
	for _, value := range model.Values {
		enumValues = append(enumValues, value)
	}

	schema.Enum = enumValues

	// Add schema to components
	openapi.Components.Schemas[model.Name] = &openapi3.SchemaRef{
		Value: schema,
	}
}

func generateAliasSpec(openapi *openapi3.T, alias definitions.NakedAliasMetadata) {
	aliasType := &openapi3.Types{swagtool.ToOpenApiType(alias.Type)}

	// Create the alias schema
	schema := &openapi3.Schema{
		Title:       alias.Name,
		Description: alias.Description,
		Type:        aliasType,
		Deprecated:  swagtool.IsDeprecated(&alias.Deprecation),
	}

	// Add schema to components
	openapi.Components.Schemas[alias.Name] = &openapi3.SchemaRef{
		Value: schema,
	}
}

// Fill the schema references in the components
func fillSchemaRef(openapi *openapi3.T) {
	// Iterate over all EmptyRefSchemas
	for _, schemaRefMap := range schemaRefMap {
		// Get the schema from the components
		schema := openapi.Components.Schemas[schemaRefMap.InterfaceType]
		// Check if the schema is not null
		if schema != nil && schema.Value != nil {
			// Set the schema reference to the schema
			schemaRefMap.SchemaRef.Value = schema.Value
		}
	}
}

func GenerateModelsSpec(openapi *openapi3.T, models *definitions.Models) error {
	for _, enum := range models.Enums {
		generateEnumSpec(openapi, enum)
	}

	for _, model := range models.Structs {
		generateStructSpec(openapi, model)
	}

	for _, alias := range models.Aliases {
		generateAliasSpec(openapi, alias)
	}

	fillSchemaRef(openapi)
	return nil
}
