package swagen31

import (
	"github.com/gopher-fleece/gleece/definitions"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

func GenerateSecuritySpec(doc *v3.Document, securityConfig *[]definitions.SecuritySchemeConfig) error {
	securitySchemes := orderedmap.New[string, *v3.SecurityScheme]()
	for _, scheme := range *securityConfig {
		securitySchemes.Set(scheme.SecurityName, &v3.SecurityScheme{
			Type:        string(scheme.Type),
			In:          string(scheme.In),
			Name:        scheme.FieldName,
			Description: scheme.Description,
		})
	}

	doc.Components.SecuritySchemes = securitySchemes
	return nil
}
