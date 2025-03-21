package swagen30

import (
	"errors"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/generator/swagen/swagtool"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

func createOperation(def definitions.ControllerMetadata, route definitions.RouteMetadata) *openapi3.Operation {
	return &openapi3.Operation{
		Summary:     route.Description,
		Description: route.Description,
		Responses:   openapi3.NewResponses(),
		OperationID: route.OperationId,
		Tags:        []string{def.Tag},
		Parameters:  []*openapi3.ParameterRef{},
		Deprecated:  swagtool.IsDeprecated(&route.Deprecation),
	}
}

func createErrorResponse(openapi *openapi3.T, route definitions.RouteMetadata, errResp definitions.ErrorResponse) *openapi3.ResponseRef {
	errorReturnType := route.GetErrorReturnType()

	// Every vanilla error should be RFC7807
	// User can override it by inheriting from error and add it's own error schema (as any other schema)
	if errorReturnType.Name == "error" {
		errorReturnType.Name = definitions.Rfc7807ErrorName
	}

	content := createContentWithSchemaRef(openapi, "", errorReturnType.Name)
	errResString := errResp.Description
	response := &openapi3.Response{
		Description: &errResString,
		Content:     content,
	}
	return &openapi3.ResponseRef{
		Value: response,
	}
}

func createContentWithSchemaRef(openapi *openapi3.T, validationString string, interfaceType string) openapi3.Content {
	schemaRef := InterfaceToSchemaRef(openapi, interfaceType)
	BuildSchemaValidation(schemaRef, validationString, interfaceType)
	return openapi3.NewContentWithJSONSchemaRef(schemaRef)
}

func createResponseSuccess(openapi *openapi3.T, route definitions.RouteMetadata) *openapi3.ResponseRef {

	valueReturnType := route.GetValueReturnType()
	if valueReturnType == nil {
		return &openapi3.ResponseRef{
			Value: &openapi3.Response{
				Description: &route.ResponseDescription,
			},
		}
	}
	content := createContentWithSchemaRef(openapi, "", valueReturnType.Name)
	return &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: &route.ResponseDescription,
			Content:     content,
		},
	}
}

func buildSecurityMethod(securitySchemes []definitions.SecuritySchemeConfig, securityMethods []definitions.SecurityAnnotationComponent) (*openapi3.SecurityRequirement, error) {
	securityRequirement := openapi3.SecurityRequirement{}

	for _, securityMethod := range securityMethods {

		// Make sure the name is exist in the openapi security schemes
		if !swagtool.IsSecurityNameInSecuritySchemes(securitySchemes, securityMethod.SchemaName) {
			errStr := fmt.Sprintf("Security method name %s does not exist in the defined security schemes %v", securityMethod.SchemaName, securitySchemes)
			// Create error object and return it, add the method name that is not exist in the security schemes
			return nil, errors.New(errStr)
		}
		securityRequirement[securityMethod.SchemaName] = securityMethod.Scopes
	}

	return &securityRequirement, nil
}

func generateOperationSecurity(operation *openapi3.Operation, config *definitions.OpenAPIGeneratorConfig, route definitions.RouteMetadata) error {
	securityRequirements := openapi3.SecurityRequirements{}

	routeSecurity := route.Security

	if len(routeSecurity) == 0 && config.DefaultRouteSecurity != nil {
		routeSecurity = []definitions.RouteSecurity{{SecurityAnnotation: []definitions.SecurityAnnotationComponent{*config.DefaultRouteSecurity}}}
	}

	for _, security := range routeSecurity {
		securityRequirement, err := buildSecurityMethod(config.SecuritySchemes, security.SecurityAnnotation)

		if err != nil {
			errStr := fmt.Sprintf("Building security %s for %s failed: %s", route.OperationId, security.SecurityAnnotation, err.Error())
			return errors.New(errStr)
		}
		securityRequirements = append(securityRequirements, *securityRequirement)
	}

	operation.Security = &securityRequirements
	return nil
}

func setNewRouteOperation(openapi *openapi3.T, def definitions.ControllerMetadata, route definitions.RouteMetadata, operation *openapi3.Operation) {
	// Set the operation in the path item
	routePath := common.RemoveDuplicateSlash(def.RestMetadata.Path + route.RestMetadata.Path)

	// Set the path item in the openapi
	pathItem := openapi.Paths.Find(routePath)
	// If path item is nil, create a new path item
	if pathItem == nil {
		pathItem = &openapi3.PathItem{}
	}

	pathItem.SetOperation(string(route.HttpVerb), operation)
	openapi.Paths.Set(routePath, pathItem)
}

func handleRouteParamDeprecation(routeParam definitions.FuncParam, specParam *openapi3.ParameterRef) {
	if routeParam.Deprecation != nil && *&routeParam.Deprecation.Deprecated {
		specParam.Value.Deprecated = true
	}
}

func createRouteParam(openapi *openapi3.T, param definitions.FuncParam) *openapi3.ParameterRef {
	schemaRef := InterfaceToSchemaRef(openapi, param.TypeMeta.Name)
	BuildSchemaValidation(schemaRef, param.Validator, param.TypeMeta.Name)
	specParam := &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        param.NameInSchema,
			In:          strings.ToLower(string(param.PassedIn)),
			Description: param.Description,
			Required:    swagtool.IsFieldRequired(param.Validator),
			Schema:      schemaRef,
		},
	}
	handleRouteParamDeprecation(param, specParam)
	return specParam
}

func createRequestBodyParam(openapi *openapi3.T, param definitions.FuncParam) *openapi3.RequestBodyRef {
	content := createContentWithSchemaRef(openapi, param.Validator, param.TypeMeta.Name)
	return &openapi3.RequestBodyRef{
		Value: &openapi3.RequestBody{
			Description: param.Description,
			Content:     content,
			Required:    swagtool.IsFieldRequired(param.Validator),
		},
	}
}

func createRequestFormParam(openapi *openapi3.T, param definitions.FuncParam, operation *openapi3.Operation) {
	// Form parameters are always passed in the body, so we need to create a request body if it doesn't exist
	if operation.RequestBody == nil {
		// The body will be a object with the form parameters as properties
		schemaRef := ToOpenApiSchemaRef("object")
		operation.RequestBody = &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Content: openapi3.Content{
					string(definitions.ContentTypeFormURLEncoded): &openapi3.MediaType{
						Schema: schemaRef,
					},
				},
			},
		}
	}

	// Get the schema from the request body
	formSchema := operation.RequestBody.Value.Content[string(definitions.ContentTypeFormURLEncoded)].Schema
	// Create a new schema for the form parameter
	propertySchemaRef := InterfaceToSchemaRef(openapi, param.TypeMeta.Name)
	// Add the validation to the schema
	BuildSchemaValidation(propertySchemaRef, param.Validator, param.TypeMeta.Name)
	// Add the form parameter to the schema
	formSchema.Value.Properties[param.NameInSchema] = propertySchemaRef
	// Add the form parameter to the required list if it is required
	if swagtool.IsFieldRequired(param.Validator) {
		formSchema.Value.Required = append(formSchema.Value.Required, param.NameInSchema)
	}
}

func generateParams(openapi *openapi3.T, route definitions.RouteMetadata, operation *openapi3.Operation) {
	// Iterate over FuncParams and create parameters
	for _, param := range route.FuncParams {

		switch param.PassedIn {
		case definitions.PassedInBody:
			operation.RequestBody = createRequestBodyParam(openapi, param)
		case definitions.PassedInForm:
			createRequestFormParam(openapi, param, operation)
		default:
			operation.Parameters = append(operation.Parameters, createRouteParam(openapi, param))
		}
	}
}

// GenerateControllerSpec generates the specification for a controller
func generateControllerSpec(openapi *openapi3.T, config *definitions.OpenAPIGeneratorConfig, def definitions.ControllerMetadata) error {
	// Iterate over the routes in the controller
	for _, route := range def.Routes {

		if swagtool.IsHiddenAsset(&route.Hiding) {
			logger.Info(fmt.Sprintf("Skipping hidden route: %v %s (%s)", route.HttpVerb, route.RestMetadata.Path, route.OperationId))
			continue
		}

		// Create a new Operation for the route
		operation := createOperation(def, route)

		// Iterate over the error responses
		for _, errResp := range route.ErrorResponses {
			// Set the response using the Set method
			operation.Responses.Set(swagtool.HttpStatusCodeToString(errResp.HttpStatusCode), createErrorResponse(openapi, route, errResp))
		}

		operation.Responses.Set(swagtool.HttpStatusCodeToString(route.ResponseSuccessCode), createResponseSuccess(openapi, route))

		generateParams(openapi, route, operation)

		// Add the security requirement to the operation
		if err := generateOperationSecurity(operation, config, route); err != nil {
			return err
		}

		// Finally, set the operation in the path item
		setNewRouteOperation(openapi, def, route, operation)
	}
	return nil
}

func GenerateControllersSpec(openapi *openapi3.T, config *definitions.OpenAPIGeneratorConfig, defs []definitions.ControllerMetadata) error {
	// Iterate over the routes in the controller
	for _, def := range defs {
		if err := generateControllerSpec(openapi, config, def); err != nil {
			errStr := fmt.Sprintf("Building controller %s failed: %s", def.Name, err.Error())
			return errors.New(errStr)
		}
	}
	return nil
}
