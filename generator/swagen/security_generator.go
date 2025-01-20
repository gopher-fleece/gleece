package swagen

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gopher-fleece/gleece/definitions"
)

func GenerateSecuritySpec(openapi *openapi3.T, securityConfig *[]definitions.SecuritySchemeConfig) error {
	securitySchemes := openapi3.SecuritySchemes{}
	for _, scheme := range *securityConfig {
		securitySchemes[scheme.SecurityName] = &openapi3.SecuritySchemeRef{
			Value: &openapi3.SecurityScheme{
				Type:        string(scheme.Type),
				In:          string(scheme.In),
				Name:        scheme.FieldName,
				Description: scheme.Description,
			},
		}
	}

	openapi.Components.SecuritySchemes = securitySchemes
	return nil
}
