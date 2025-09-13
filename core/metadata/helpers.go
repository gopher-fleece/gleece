package metadata

import (
	"fmt"
	"strings"

	MapSet "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/runtime"
	"golang.org/x/tools/go/packages"
)

func GetMethodHideOpts(attributes *annotations.AnnotationHolder) definitions.MethodHideOptions {
	if attributes == nil {
		// No attributes, not hidden. A bit of an oxymoron as this isn't supposed to actually be called
		// for any node that doesn't have any annotations
		return definitions.MethodHideOptions{Type: definitions.HideMethodNever}
	}

	attr := attributes.GetFirst(annotations.GleeceAnnotationHidden)
	if attr == nil {
		// No '@Hidden' attribute
		return definitions.MethodHideOptions{Type: definitions.HideMethodNever}
	}

	// Here we can insert any expanded functionality.
	// There are several routes we can take here.
	// Listed below are a couple.
	//
	// 1. Env in Value, condition type + expected in props
	// @Hidden(ENV_VAR, {eq: "value"})
	//
	// 2. CONDITION keyword in Value, various options in props
	// @Hidden(CONDITION, { env: { var: "VAR_NAME", value: "EXPECTED_VALUE", operator: "EQUALS" }})
	return definitions.MethodHideOptions{Type: definitions.HideMethodAlways}
}

func GetDeprecationOpts(holder *annotations.AnnotationHolder) definitions.DeprecationOptions {
	if holder == nil {
		return definitions.DeprecationOptions{}
	}

	deprecationAttr := holder.GetFirst(annotations.GleeceAnnotationDeprecated)
	if deprecationAttr == nil {
		return definitions.DeprecationOptions{}
	}

	return definitions.DeprecationOptions{
		Deprecated:  true,
		Description: deprecationAttr.Description,
	}
}

// GetSecurityFromContext Creates an array of RouteSecurity out of the given holder's attributes
func GetSecurityFromContext(holder *annotations.AnnotationHolder) ([]definitions.RouteSecurity, error) {
	securities := []definitions.RouteSecurity{}

	// Process @Security annotations. In the future, we'll support @AdvancedSecurity
	normalSec := holder.GetAll(annotations.GleeceAnnotationSecurity)
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

func GetRouteSecurityWithInheritance(
	receiverAnnotations *annotations.AnnotationHolder,
	parentSecurity []definitions.RouteSecurity,
) ([]definitions.RouteSecurity, error) {
	explicitSec, err := GetSecurityFromContext(receiverAnnotations)
	if err != nil {
		return []definitions.RouteSecurity{}, err
	}

	if len(explicitSec) > 0 {
		return explicitSec, nil
	}

	return parentSecurity, nil
}

func GetTemplateContextMetadata(attributes *annotations.AnnotationHolder) (map[string]definitions.TemplateContext, error) {
	customAttributes := attributes.GetAll(annotations.GleeceAnnotationTemplateContext)

	templateContext := map[string]definitions.TemplateContext{}

	for _, attr := range customAttributes {

		if _, exists := templateContext[attr.Value]; exists {
			return nil, fmt.Errorf("duplicate template context attribute '%s'", attr.Value)
		}

		templateContext[attr.Value] = definitions.TemplateContext{
			Options:     attr.Properties,
			Description: attr.Description,
		}
	}

	return templateContext, nil
}

func GetResponseStatusCodeAndDescription(
	attributes *annotations.AnnotationHolder,
	hasReturnValue bool,
) (runtime.HttpStatusCode, string, error) {
	// Set the success attrib code based on whether function returns a value or only error (200 vs 204)
	attrib := attributes.GetFirst(annotations.GleeceAnnotationResponse)
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

func GetParamPassedIn(
	paramName string,
	paramAnnotations *annotations.AnnotationHolder,
) (definitions.ParamPassedIn, error) {
	if paramAnnotations == nil {
		return "", fmt.Errorf("parameter '%s' does not have any annotations", paramName)
	}

	paramAttrib := paramAnnotations.FindFirstByValue(paramName)
	if paramAttrib == nil {
		return definitions.PassedInHeader,
			NewInvalidAnnotationError(
				fmt.Sprintf("parameter '%s' does not have a matching documentation attribute", paramName),
			)

	}

	// Currently, only body param can be an object type
	switch strings.ToLower(paramAttrib.Name) {
	case "query":
		return definitions.PassedInQuery, nil
	case "header":
		return definitions.PassedInHeader, nil
	case "path":
		return definitions.PassedInPath, nil
	case "body":
		return definitions.PassedInBody, nil
	case "formfield":
		// Currently, form fields are the only supported form of form parameters,
		// in the future, a full form object may be supported too
		return definitions.PassedInForm, nil
	default:
		return definitions.PassedInHeader,
			fmt.Errorf(
				"parameter '%s' has an unexpected 'passed-in' annotation '%s'",
				paramName,
				paramAttrib.Name,
			)
	}
}

func GetParameterSchemaName(
	paramName string,
	paramAnnotations *annotations.AnnotationHolder,
) (string, error) {
	paramAttrib := paramAnnotations.FindFirstByValue(paramName)
	if paramAttrib == nil {
		return "", fmt.Errorf("parameter '%s' does not have a matching documentation attribute", paramName)
	}

	castName, err := annotations.GetCastProperty[string](paramAttrib, annotations.PropertyName)
	if err != nil {
		return "", err
	}

	if castName != nil && len(*castName) > 0 {
		return *castName, nil
	}

	return paramName, nil
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

func GetParamValidator(
	paramName string,
	paramAnnotations *annotations.AnnotationHolder,
	passedIn definitions.ParamPassedIn,
	isPointerParam bool,
) (string, error) {
	paramAttrib := paramAnnotations.FindFirstByValue(paramName)
	if paramAttrib == nil {
		return "", fmt.Errorf("parameter '%s' does not have a matching documentation attribute", paramName)
	}

	castValidator, err := annotations.GetCastProperty[string](paramAttrib, annotations.PropertyValidatorString)
	if err != nil {
		return "", err
	}

	validatorString := ""
	if castValidator != nil && len(*castValidator) > 0 {
		validatorString = *castValidator
	}

	return appendParamRequiredValidation(&validatorString, isPointerParam, passedIn), nil
}

func GetErrorResponses(routeAnnotations *annotations.AnnotationHolder) ([]definitions.ErrorResponse, error) {
	responseAttributes := routeAnnotations.GetAll(annotations.GleeceAnnotationErrorResponse)

	responses := []definitions.ErrorResponse{}
	encounteredCodes := MapSet.NewSet[runtime.HttpStatusCode]()

	for _, attr := range responseAttributes {
		code, err := definitions.ConvertToHttpStatus(attr.Value)
		if err != nil {
			return responses, err
		}

		if encounteredCodes.ContainsOne(code) {
			logger.Warn(
				"Status code '%d' appears multiple time on a controller receiver. Ignoring. Original Comment: %s",
				code,
				attr,
			)
			continue
		}
		responses = append(responses, definitions.ErrorResponse{HttpStatusCode: code, Description: attr.Description})
		encounteredCodes.Add(code)
	}

	return responses, nil
}

// GetDefaultSecurity Returns the default securities defined at the Gleece configuration file level
func GetDefaultSecurity(config *definitions.GleeceConfig) []definitions.RouteSecurity {
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

func isErrorEmbedding(typeName string, isUniverse bool, pkg *packages.Package) (bool, error) {
	if typeName == "error" {
		return true, nil
	}

	// Universe types are leaf nodes and can never embed anything- no reason to check them
	if isUniverse {
		return false, nil
	}

	embeds, err := gast.DoesStructEmbedType(pkg, typeName, "", "error")
	if err != nil {
		return false, err
	}

	return embeds, nil
}

func IsAnErrorEmbeddingTypeUsage(
	meta TypeUsageMeta,
	metaPackage *packages.Package,
) (bool, error) {
	return isErrorEmbedding(meta.Name, meta.IsUniverseType(), metaPackage)
}

func IsAnErrorEmbeddingType(
	meta definitions.TypeMetadata,
	metaPackage *packages.Package,
) (bool, error) {
	return isErrorEmbedding(meta.Name, meta.IsUniverseType, metaPackage)
}
