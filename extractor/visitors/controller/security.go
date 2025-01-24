package controller

import (
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor"
)

func (v ControllerVisitor) getDefaultSecurity() []definitions.RouteSecurity {
	return v.config.OpenAPIGeneratorConfig.DefaultRouteSecurity
}

func (v *ControllerVisitor) getSecurityFromContext(holder extractor.AttributesHolder) ([]definitions.RouteSecurity, error) {
	securities := []definitions.RouteSecurity{}

	normalSec := holder.GetAll(extractor.AttributeSecurity)
	if len(normalSec) > 0 {
		for _, secAttrib := range normalSec {
			schemaName := secAttrib.Value
			if len(schemaName) <= 0 {
				return securities, v.getFrozenError("a security schema's name cannot be empty")
			}

			definedScopes, err := extractor.GetCastProperty[[]string](secAttrib, extractor.PropertySecurityScopes)
			if err != nil {
				return securities, v.frozenError(err)
			}

			scopes := []string{}
			if definedScopes != nil && len(*definedScopes) > 0 {
				scopes = *definedScopes
			}

			securities = append(securities, definitions.RouteSecurity{
				SecurityAnnotation: []definitions.SecurityAnnotationComponent{{
					SchemaName: schemaName,
					Scopes:     scopes,
				}},
			})
		}
	}

	// AdvanceSecurity processing goes here

	return securities, nil
}

func (v *ControllerVisitor) getRouteSecurityWithInheritance(attributes extractor.AttributesHolder) ([]definitions.RouteSecurity, error) {
	explicitSec, err := v.getSecurityFromContext(attributes)
	if err != nil {
		return []definitions.RouteSecurity{}, v.frozenError(err)
	}

	if len(explicitSec) > 0 {
		return explicitSec, nil
	}

	return v.currentController.Security, nil
}
