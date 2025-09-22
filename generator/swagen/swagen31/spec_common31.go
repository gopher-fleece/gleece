package swagen31

import (
	"strings"

	"github.com/gopher-fleece/gleece/generator/swagen/swagtool"
	"github.com/pb33f/libopenapi-validator/errors"
	highbase "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func ToOpenApiSchemaV3(typeName string) *highbase.Schema {
	switch typeName {
	case "binary":
		return &highbase.Schema{
			Type:   []string{"string"},
			Format: "base64",
		}
	case "date-time":
		return &highbase.Schema{
			Type:   []string{"string"},
			Format: "date-time",
		}
	default:
		return &highbase.Schema{
			Type: []string{typeName},
		}
	}
}

func InterfaceToSchemaV3(doc *v3.Document, interfaceType string) *highbase.SchemaProxy {

	openapiType := swagtool.ToOpenApiType(interfaceType)
	fieldSchema := ToOpenApiSchemaV3(openapiType)

	if openapiType == "object" && !swagtool.IsMapObject(interfaceType) { // For now, ignore map objects, they will be handled later
		return highbase.CreateSchemaProxyRef("#/components/schemas/" + interfaceType)
	}
	if openapiType == "array" {
		// Handle array types
		itemType := swagtool.GetArrayItemType(interfaceType)
		// Once the item type is determined, create a schema reference for it in a recursive manner
		itemSchemaRef := InterfaceToSchemaV3(doc, itemType)
		fieldSchema.Items = &highbase.DynamicValue[*highbase.SchemaProxy, bool]{
			A: itemSchemaRef,
		}
	}
	return highbase.CreateSchemaProxy(fieldSchema)
}

func FormatValidationErrors(validationErrors []*errors.ValidationError) string {
	if len(validationErrors) == 0 {
		return ""
	}

	errorMsgs := make([]string, len(validationErrors))
	for i, err := range validationErrors {
		errorMsgs[i] = err.Error()
	}

	return strings.Join(errorMsgs, "\n")
}

func FormatErrors(validationErrors []error) string {
	if len(validationErrors) == 0 {
		return ""
	}

	errorMsgs := make([]string, len(validationErrors))
	for i, err := range validationErrors {
		errorMsgs[i] = err.Error()
	}

	return strings.Join(errorMsgs, "\n")
}

func ToResponseDescription(description string) string {
	// This "hack" is since if the description is empty,
	// the libopenapi will not generate the response description at all, when the description is a required field in the spec.
	if description == "" {
		description = " "
	}
	return description
}
