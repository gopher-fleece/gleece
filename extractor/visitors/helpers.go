package visitors

import (
	"fmt"
	"strings"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/extractor/arbitrators"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/runtime"
)

func getResponseStatusCodeAndDescription(
	attributes *annotations.AnnotationHolder,
	hasReturnValue bool,
) (runtime.HttpStatusCode, string, error) {
	// Set the success attrib code based on whether function returns a value or only error (200 vs 204)
	attrib := attributes.GetFirst(annotations.AttributeResponse)
	if attrib == nil {
		if hasReturnValue {
			return runtime.StatusOK, "", nil
		}

		return runtime.StatusNoContent, "", nil
	}

	var statusCode runtime.HttpStatusCode
	if len(attrib.Value) > 0 {
		code, err := definitions.ConvertToHttpStatus(attrib.Value)
		if err != nil {
			return 0, "", err
		}
		statusCode = code
	}

	return statusCode, attrib.Description, nil
}

// getSecurityFromContext Creates an array of RouteSecurity out of the given holder's attributes
func getSecurityFromContext(holder annotations.AnnotationHolder) ([]definitions.RouteSecurity, error) {
	securities := []definitions.RouteSecurity{}

	// Process @Security annotations. In the future, we'll support @AdvancedSecurity
	normalSec := holder.GetAll(annotations.AttributeSecurity)
	if len(normalSec) > 0 {
		for _, secAttrib := range normalSec {
			schemaName := secAttrib.Value
			if len(schemaName) <= 0 {
				return securities, fmt.Errorf("a security schema's name cannot be empty")
			}

			definedScopes, err := annotations.GetCastProperty[[]string](secAttrib, annotations.PropertySecurityScopes)
			if err != nil {
				return securities, err
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

func getTemplateContextMetadata(attributes *annotations.AnnotationHolder) (map[string]definitions.TemplateContext, error) {
	customAttributes := attributes.GetAll(annotations.AttributeTemplateContext)

	templateContext := map[string]definitions.TemplateContext{}

	for _, attr := range customAttributes {

		if _, exists := templateContext[attr.Value]; exists {
			return nil, fmt.Errorf("duplicate template context attribute '%s'", attr.Value)
		}

		if templateContext != nil {

		}
		templateContext[attr.Value] = definitions.TemplateContext{
			Options:     attr.Properties,
			Description: attr.Description,
		}
	}

	return templateContext, nil
}

func getDeprecationOpts(attributes *annotations.AnnotationHolder) definitions.DeprecationOptions {
	deprecationAttr := attributes.GetFirst(annotations.AttributeDeprecated)
	if deprecationAttr == nil {
		return definitions.DeprecationOptions{}
	}

	return definitions.DeprecationOptions{
		Deprecated:  true,
		Description: deprecationAttr.Description,
	}
}

func getMethodHideOpts(attributes *annotations.AnnotationHolder) definitions.MethodHideOptions {
	attr := attributes.GetFirst(annotations.AttributeHidden)
	if attr == nil {
		// No '@Hidden' attribute
		return definitions.MethodHideOptions{Type: definitions.HideMethodNever}
	}

	if attr.Properties == nil || len(attr.Properties) <= 0 {
		return definitions.MethodHideOptions{Type: definitions.HideMethodAlways}
	}

	// Technically a bit redundant since we know by length whether there's a condition defined
	// but nothing stops user from adding text to the comment so this mostly serves as a validation
	if len(attr.Value) <= 0 {
		// Standard '@Hidden' attribute; Always hide.
		return definitions.MethodHideOptions{Type: definitions.HideMethodAlways}
	}

	// A '@Hidden(condition)' attribute
	return definitions.MethodHideOptions{Type: definitions.HideMethodCondition, Condition: attr.Value}
}

// For now, all params are required, later we will support nil for pointers and slices params
func appendParamRequiredValidation(validation *string, isPointer bool, paramPassedIn definitions.ParamPassedIn) string {
	// For a pointer, we do allow to be optional and it's pending user decision via validate tag
	// BUT, as openapi specification, path params are always required
	if isPointer && paramPassedIn != definitions.PassedInPath {
		return *validation
	}
	if validation == nil || *validation == "" {
		return "required"
	}

	// Split the validation string into individual tags
	tags := strings.Split(*validation, ",")

	// Check if "required" is already present
	for _, tag := range tags {
		if tag == "required" {
			return *validation
		}
	}

	// Append "required" to the validation string
	return *validation + ",required"
}

func isAnErrorEmbeddingType(
	packagesFacade *arbitrators.PackagesFacade,
	meta definitions.TypeMetadata,
) (bool, error) {
	if meta.Name == "error" {
		return true, nil
	}

	// Universe types are leaf nodes and can never embed anything- no reason to check them
	if meta.IsUniverseType {
		return false, nil
	}

	pkg := gast.FilterPackageByFullName(packagesFacade.GetAllPackages(), meta.PkgPath)
	embeds, err := extractor.DoesStructEmbedType(pkg, meta.Name, "", "error")
	if err != nil {
		return false, err
	}

	return embeds, nil
}

// getDefaultSecurity Returns the default securities defined at the Gleece configuration file level
func getDefaultSecurity(config *definitions.GleeceConfig) []definitions.RouteSecurity {
	defaultSecurity := []definitions.RouteSecurity{}

	if config == nil {
		return nil
	}

	if config.OpenAPIGeneratorConfig.DefaultRouteSecurity == nil {
		return defaultSecurity
	}

	defaultSecurity = append(defaultSecurity, definitions.RouteSecurity{
		SecurityAnnotation: []definitions.SecurityAnnotationComponent{*config.OpenAPIGeneratorConfig.DefaultRouteSecurity},
	})
	return defaultSecurity
}
