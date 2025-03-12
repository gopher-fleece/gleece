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
	highbaseSchema := &highbase.Schema{
		Title:       model.Name,
		Description: model.Description,
		Type:        []string{"object"},
		Properties:  orderedmap.New[string, *highbase.SchemaProxy](),
		Deprecated:  &isDeprecated,
	}

	requiredFields := []string{}

	for _, field := range model.Fields {
		fName := swagtool.GetJsonNameFromTag(field.Tag, field.Name)
		validationTag := swagtool.GetTagValue(field.Tag, "validate", "")

		if swagtool.IsFieldRequired(validationTag) {
			requiredFields = append(requiredFields, fName)
		}

		fieldSchemaRef := InterfaceToSchemaV3(doc, field.Type)

		innerSchema := fieldSchemaRef.Schema()

		if innerSchema != nil {
			BuildSchemaValidationV31(innerSchema, validationTag, field.Type)
			innerSchema.Description = field.Description
			isFieldDeprecated := swagtool.IsDeprecated(field.Deprecation)
			innerSchema.Deprecated = &isFieldDeprecated
		}
		highbaseSchema.Properties.Set(fName, fieldSchemaRef)
	}

	highbaseSchema.Required = requiredFields
	doc.Components.Schemas.Set(model.Name, highbase.CreateSchemaProxy(highbaseSchema))
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

func GenerateModelsSpec(doc *v3.Document, models *definitions.Models) error {
	for _, enum := range models.Enums {
		generateEnumsSpec(doc, enum)
	}

	for _, model := range models.Structs {
		generateStructsSpec(doc, model)
	}

	return nil
}
