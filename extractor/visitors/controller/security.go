package controller

import (
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/annotations"
)

// getDefaultSecurity Returns the default securities defined at the Gleece configuration file level
func (v ControllerVisitor) getDefaultSecurity() []definitions.RouteSecurity {
	defaultSecurity := []definitions.RouteSecurity{}

	if v.config.OpenAPIGeneratorConfig.DefaultRouteSecurity == nil {
		return defaultSecurity
	}

	defaultSecurity = append(defaultSecurity, definitions.RouteSecurity{
		SecurityAnnotation: []definitions.SecurityAnnotationComponent{*v.config.OpenAPIGeneratorConfig.DefaultRouteSecurity},
	})
	return defaultSecurity
}

// getSecurityFromContext Creates an array of RouteSecurity out of the given holder's attributes
func (v *ControllerVisitor) getSecurityFromContext(holder annotations.AnnotationHolder) ([]definitions.RouteSecurity, error) {
	securities := []definitions.RouteSecurity{}

	// Process @Security annotations. In the future, we'll support @AdvancedSecurity
	normalSec := holder.GetAll(annotations.AttributeSecurity)
	if len(normalSec) > 0 {
		for _, secAttrib := range normalSec {
			schemaName := secAttrib.Value
			if len(schemaName) <= 0 {
				return securities, v.getFrozenError("a security schema's name cannot be empty")
			}

			definedScopes, err := annotations.GetCastProperty[[]string](secAttrib, annotations.PropertySecurityScopes)
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

// getRouteSecurityWithInheritance Gets the securities to be associated with the route annotated by the given AnnotationHolder.
// Security is hierarchial and uses a 'first-match' approach:
//
// 1. Explicit, receiver level annotations
// 2. Explicit, controller level annotations
// 3. Default securities in Gleece configuration file.
func (v *ControllerVisitor) getRouteSecurityWithInheritance(attributes annotations.AnnotationHolder) ([]definitions.RouteSecurity, error) {
	explicitSec, err := v.getSecurityFromContext(attributes)
	if err != nil {
		return []definitions.RouteSecurity{}, v.frozenError(err)
	}

	if len(explicitSec) > 0 {
		return explicitSec, nil
	}

	return v.currentController.Security, nil
}
