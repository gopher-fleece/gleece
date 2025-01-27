package swagten31

import (
	"github.com/gopher-fleece/gleece/definitions"
	highbase "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

var objectType = []string{"object"}
var arrayType = []string{"array"}
var stringType = []string{"string"}

func ToOpenApiSchemaV3(typeName string) *highbase.Schema {
	return &highbase.Schema{
		Type: []string{typeName},
	}
}

func InterfaceToSchemaV3(doc *v3.Document, interfaceType string) *highbase.SchemaProxy {

	openapiType := ToOpenApiType(interfaceType)
	fieldSchema := ToOpenApiSchemaV3(openapiType)

	if openapiType == "object" && !IsMapObject(interfaceType) { // For now, ignore map objects, they will be handled later
		return highbase.CreateSchemaProxyRef("#/components/schemas/" + interfaceType)
	}
	if openapiType == "array" {
		// Handle array types
		itemType := GetArrayItemType(interfaceType)
		// Once the item type is determined, create a schema reference for it in a recursive manner
		itemSchemaRef := InterfaceToSchemaV3(doc, itemType)
		fieldSchema.Items = &highbase.DynamicValue[*highbase.SchemaProxy, bool]{
			A: itemSchemaRef,
		}
	}
	return highbase.CreateSchemaProxy(fieldSchema)
}

func generateModelSpec(doc *v3.Document, model definitions.ModelMetadata) {
	isDeprecated := IsDeprecated(&model.Deprecation)
	highbaseSchema := &highbase.Schema{
		Title:       model.Name,
		Description: model.Description,
		Type:        objectType,
		Properties:  orderedmap.New[string, *highbase.SchemaProxy](),
		Deprecated:  &isDeprecated,
	}

	requiredFields := []string{}

	for _, field := range model.Fields {
		// openapiType := ToOpenApiType(field.Type)

		fName := GetTagValue(field.Tag, "json", field.Name)
		validationTag := GetTagValue(field.Tag, "validate", "")

		if IsFieldRequired(validationTag) {
			requiredFields = append(requiredFields, fName)
		}

		fieldSchemaRef := InterfaceToSchemaV3(doc, field.Type)

		innerSchema := fieldSchemaRef.Schema()

		if innerSchema != nil {
			BuildSchemaValidationV31(innerSchema, validationTag, field.Type)
			innerSchema.Description = field.Description
			isFieldDeprecated := IsDeprecated(field.Deprecation)
			innerSchema.Deprecated = &isFieldDeprecated
		}
		highbaseSchema.Properties.Set(fName, fieldSchemaRef)
	}

	highbaseSchema.Required = requiredFields
	doc.Components.Schemas.Set(model.Name, highbase.CreateSchemaProxy(highbaseSchema))
}

func GenerateModelsSpec(doc *v3.Document, models []definitions.ModelMetadata) error {
	for _, model := range models {
		generateModelSpec(doc, model)
	}
	return nil
}
