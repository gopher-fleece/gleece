package metadata

import (
	"fmt"
	"strings"

	MapSet "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/runtime"
)

func GetMethodHideOpts(attributes *annotations.AnnotationHolder) definitions.MethodHideOptions {
	attr := attributes.GetFirst(annotations.GleeceAnnotationHidden)
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

func GetDeprecationOpts(attributes *annotations.AnnotationHolder) definitions.DeprecationOptions {
	deprecationAttr := attributes.GetFirst(annotations.GleeceAnnotationDeprecated)
	if deprecationAttr == nil {
		return definitions.DeprecationOptions{}
	}

	return definitions.DeprecationOptions{
		Deprecated:  true,
		Description: deprecationAttr.Description,
	}
}

// GetSecurityFromContext Creates an array of RouteSecurity out of the given holder's attributes
func GetSecurityFromContext(holder annotations.AnnotationHolder) ([]definitions.RouteSecurity, error) {
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
	controllerAnnotations *annotations.AnnotationHolder,
	receiverAnnotations *annotations.AnnotationHolder,
) ([]definitions.RouteSecurity, error) {
	explicitSec, err := GetSecurityFromContext(*receiverAnnotations)
	if err != nil {
		return []definitions.RouteSecurity{}, err
	}

	if len(explicitSec) > 0 {
		return explicitSec, nil
	}

	return GetSecurityFromContext(*controllerAnnotations)
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
	paramAttrib := paramAnnotations.FindFirstByValue(paramName)
	if paramAttrib == nil {
		return definitions.PassedInHeader, fmt.Errorf("parameter '%s' does not have a matching documentation attribute", paramName)
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
