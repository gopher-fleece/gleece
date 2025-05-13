package swagen30

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
		if scheme.Scheme != "" {
			securitySchemes[scheme.SecurityName].Value.Scheme = string(scheme.Scheme)
		}

		if scheme.OpenIdConnectUrl != "" {
			securitySchemes[scheme.SecurityName].Value.OpenIdConnectUrl = scheme.OpenIdConnectUrl
		}
		if scheme.Flows != nil {
			openAPIFlows := &openapi3.OAuthFlows{}

			if scheme.Flows.Implicit != nil {
				openAPIFlows.Implicit = &openapi3.OAuthFlow{
					AuthorizationURL: scheme.Flows.Implicit.AuthorizationURL,
					RefreshURL:       scheme.Flows.Implicit.RefreshURL,
					Scopes:           scheme.Flows.Implicit.Scopes,
				}
			}

			if scheme.Flows.Password != nil {
				openAPIFlows.Password = &openapi3.OAuthFlow{
					TokenURL:   scheme.Flows.Password.TokenURL,
					RefreshURL: scheme.Flows.Password.RefreshURL,
					Scopes:     scheme.Flows.Password.Scopes,
				}
			}

			if scheme.Flows.ClientCredentials != nil {
				openAPIFlows.ClientCredentials = &openapi3.OAuthFlow{
					TokenURL:   scheme.Flows.ClientCredentials.TokenURL,
					RefreshURL: scheme.Flows.ClientCredentials.RefreshURL,
					Scopes:     scheme.Flows.ClientCredentials.Scopes,
				}
			}

			if scheme.Flows.AuthorizationCode != nil {
				openAPIFlows.AuthorizationCode = &openapi3.OAuthFlow{
					AuthorizationURL: scheme.Flows.AuthorizationCode.AuthorizationURL,
					TokenURL:         scheme.Flows.AuthorizationCode.TokenURL,
					RefreshURL:       scheme.Flows.AuthorizationCode.RefreshURL,
					Scopes:           scheme.Flows.AuthorizationCode.Scopes,
				}
			}

			securitySchemes[scheme.SecurityName].Value.Flows = openAPIFlows
		}
	}

	openapi.Components.SecuritySchemes = securitySchemes
	return nil
}
