package swagen31

import (
	"github.com/gopher-fleece/gleece/v2/definitions"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

func GenerateSecuritySpec(doc *v3.Document, securityConfig *[]definitions.SecuritySchemeConfig) error {
	securitySchemes := orderedmap.New[string, *v3.SecurityScheme]()
	for _, scheme := range *securityConfig {
		securityScheme := &v3.SecurityScheme{
			Type:        string(scheme.Type),
			In:          string(scheme.In),
			Name:        scheme.FieldName,
			Description: scheme.Description,
		}

		if scheme.Scheme != "" {
			securityScheme.Scheme = string(scheme.Scheme)
		}

		if scheme.OpenIdConnectUrl != "" {
			securityScheme.OpenIdConnectUrl = scheme.OpenIdConnectUrl
		}

		if scheme.Flows != nil {
			flows := &v3.OAuthFlows{}

			if scheme.Flows.Implicit != nil {
				scopesMap := orderedmap.New[string, string]()
				for k, v := range scheme.Flows.Implicit.Scopes {
					scopesMap.Set(k, v)
				}

				flows.Implicit = &v3.OAuthFlow{
					AuthorizationUrl: scheme.Flows.Implicit.AuthorizationURL,
					RefreshUrl:       scheme.Flows.Implicit.RefreshURL,
					Scopes:           scopesMap,
				}
			}

			if scheme.Flows.Password != nil {
				scopesMap := orderedmap.New[string, string]()
				for k, v := range scheme.Flows.Password.Scopes {
					scopesMap.Set(k, v)
				}

				flows.Password = &v3.OAuthFlow{
					TokenUrl:   scheme.Flows.Password.TokenURL,
					RefreshUrl: scheme.Flows.Password.RefreshURL,
					Scopes:     scopesMap,
				}
			}

			if scheme.Flows.ClientCredentials != nil {
				scopesMap := orderedmap.New[string, string]()
				for k, v := range scheme.Flows.ClientCredentials.Scopes {
					scopesMap.Set(k, v)
				}

				flows.ClientCredentials = &v3.OAuthFlow{
					TokenUrl:   scheme.Flows.ClientCredentials.TokenURL,
					RefreshUrl: scheme.Flows.ClientCredentials.RefreshURL,
					Scopes:     scopesMap,
				}
			}

			if scheme.Flows.AuthorizationCode != nil {
				scopesMap := orderedmap.New[string, string]()
				for k, v := range scheme.Flows.AuthorizationCode.Scopes {
					scopesMap.Set(k, v)
				}

				flows.AuthorizationCode = &v3.OAuthFlow{
					AuthorizationUrl: scheme.Flows.AuthorizationCode.AuthorizationURL,
					TokenUrl:         scheme.Flows.AuthorizationCode.TokenURL,
					RefreshUrl:       scheme.Flows.AuthorizationCode.RefreshURL,
					Scopes:           scopesMap,
				}
			}

			securityScheme.Flows = flows
		}

		securitySchemes.Set(scheme.SecurityName, securityScheme)
	}

	doc.Components.SecuritySchemes = securitySchemes
	return nil
}
