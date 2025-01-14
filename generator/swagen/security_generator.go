package swagen

import (
	"github.com/getkin/kin-openapi/openapi3"
)

func GenerateSecuritySpec(openapi *openapi3.T, securityConfig *[]SecuritySchemeConfig) {
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
}
