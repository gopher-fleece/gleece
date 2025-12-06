package swagen31

import (
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/generator/swagen/swagtool"
	highbase "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"gopkg.in/yaml.v3"
)

func generateStructsSpec(doc *v3.Document, model definitions.StructMetadata) {
	isDeprecated := swagtool.IsDeprecated(&model.Deprecation)

	// The schema that will hold all regular fields
	regularFieldsSchema := &highbase.Schema{
		Title:       model.Name,
		Description: model.Description,
		Type:        []string{"object"},
		Properties:  orderedmap.New[string, *highbase.SchemaProxy](),
		Deprecated:  &isDeprecated,
	}

	// Determine if we have any embedded fields
	hasEmbeddedField := swagtool.HasEmbeddedField(model.Fields)

	// The final schema to be added to components
	var finalSchema *highbase.Schema

	if !hasEmbeddedField {
		// If no embedded fields, use the regular schema directly
		finalSchema = regularFieldsSchema
	} else {
		// If there are embedded fields, create a schema with allOf
		finalSchema = &highbase.Schema{
			// Don't set Type for allOf schemas
			AllOf: []*highbase.SchemaProxy{},
		}

		// Add the regular fields schema to allOf
		finalSchema.AllOf = append(finalSchema.AllOf, highbase.CreateSchemaProxy(regularFieldsSchema))
	}

	requiredFields := []string{}

	for _, field := range model.Fields {
		if field.IsEmbedded {
			if field.Type == "error" {
				continue
			}
			// If the field is embedded, add it to the allOf array
			fieldSchemaRef := InterfaceToSchemaV3(doc, field.Type)
			finalSchema.AllOf = append(finalSchema.AllOf, fieldSchemaRef)
			continue
		}

		// Process regular fields as before
		fName := swagtool.GetJsonNameFromTag(field.Tag, field.Name)
		validationTag := swagtool.GetTagValue(field.Tag, "validate", "")

		if swagtool.IsFieldRequired(validationTag) {
			requiredFields = append(requiredFields, fName)
		}

		fieldSchemaRef := InterfaceToSchemaV3(doc, field.Type)

		innerSchema := fieldSchemaRef.Schema()

		// OpenAPI 3.0 does not support the any extra field in the SchemaRef beside just ref, and we don't want to override the model properties themselves
		// Hence, we here used 3.1, so it is possible, and it's in the TODO list to be implemented
		if innerSchema != nil && !fieldSchemaRef.IsReference() {
			BuildSchemaValidationV31(innerSchema, validationTag, field.Type)
			innerSchema.Description = field.Description
			isFieldDeprecated := swagtool.IsDeprecated(field.Deprecation)
			innerSchema.Deprecated = &isFieldDeprecated
		}

		regularFieldsSchema.Properties.Set(fName, fieldSchemaRef)
	}

	// Required fields are part of the regular schema
	regularFieldsSchema.Required = requiredFields

	// Add the final schema to components
	doc.Components.Schemas.Set(model.Name, highbase.CreateSchemaProxy(finalSchema))
}

func generateEnumsSpec(doc *v3.Document, model definitions.EnumMetadata) {
	isDeprecated := swagtool.IsDeprecated(&model.Deprecation)
	enumType := swagtool.ToOpenApiType(model.Type)

	// Create the enum schema
	highbaseSchema := &highbase.Schema{
		Title:       model.Name,
		Description: model.Description,
		Type:        []string{enumType},
		Deprecated:  &isDeprecated,
	}

	// Add the possible enum values as yaml.Node objects
	enumValues := []*yaml.Node{}
	for _, value := range model.Values {
		node := &yaml.Node{
			Kind: yaml.ScalarNode,
		}
		node.Value = value
		enumValues = append(enumValues, node)
	}

	highbaseSchema.Enum = enumValues

	// Add schema to components
	doc.Components.Schemas.Set(model.Name, highbase.CreateSchemaProxy(highbaseSchema))
}

func generateAliasSpec(doc *v3.Document, alias definitions.NakedAliasMetadata) {
	isDeprecated := swagtool.IsDeprecated(&alias.Deprecation)
	aliasType := swagtool.ToOpenApiType(alias.Type)

	// Create the alias schema
	highbaseSchema := &highbase.Schema{
		Title:       alias.Name,
		Description: alias.Description,
		Type:        []string{aliasType},
		Deprecated:  &isDeprecated,
	}

	// Add schema to components
	doc.Components.Schemas.Set(alias.Name, highbase.CreateSchemaProxy(highbaseSchema))
}

func GenerateModelsSpec(doc *v3.Document, models *definitions.Models) error {
	for _, enum := range models.Enums {
		generateEnumsSpec(doc, enum)
	}

	for _, model := range models.Structs {
		generateStructsSpec(doc, model)
	}

	for _, alias := range models.Aliases {
		generateAliasSpec(doc, alias)
	}

	return nil
}
