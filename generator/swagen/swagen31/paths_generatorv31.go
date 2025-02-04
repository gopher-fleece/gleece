package swagen31

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/generator/swagen/swagtool"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	highbase "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

func createOperation(def definitions.ControllerMetadata, route definitions.RouteMetadata) *v3.Operation {
	isDeprecated := swagtool.IsDeprecated(&route.Deprecation)
	return &v3.Operation{
		Summary:     route.Description,
		Description: route.Description,
		OperationId: route.OperationId,
		Tags:        []string{def.Tag},
		Parameters:  []*v3.Parameter{},
		Deprecated:  &isDeprecated,
		Responses: &v3.Responses{
			Codes: orderedmap.New[string, *v3.Response](),
		},
	}
}

func createErrorResponse(doc *v3.Document, route definitions.RouteMetadata, errResp definitions.ErrorResponse) *v3.Response {
	errorReturnType := route.GetErrorReturnType()

	// Every vanilla error should be RFC7807
	// User can override it by inheriting from error and add it's own error schema (as any other schema)
	if errorReturnType.Name == "error" {
		errorReturnType.Name = definitions.Rfc7807ErrorName
	}

	content := createContentWithSchemaRef(doc, "", errorReturnType.Name)

	return &v3.Response{
		Description: ToResponseDescription(errResp.Description),
		Content:     content,
	}
}

func createContentWithSchemaRef(doc *v3.Document, validationString string, interfaceType string) *orderedmap.Map[string, *v3.MediaType] {
	schemaRef := InterfaceToSchemaV3(doc, interfaceType)
	if schemaRef.Schema() != nil {
		BuildSchemaValidationV31(schemaRef.Schema(), validationString, interfaceType)
	}

	content := orderedmap.New[string, *v3.MediaType]()
	content.Set("application/json", &v3.MediaType{
		Schema: schemaRef,
	})
	return content
}

func createResponseSuccess(doc *v3.Document, route definitions.RouteMetadata) *v3.Response {
	valueReturnType := route.GetValueReturnType()

	if valueReturnType == nil {
		return &v3.Response{
			Description: ToResponseDescription(route.ResponseDescription),
		}
	}

	content := createContentWithSchemaRef(doc, "", valueReturnType.Name)
	return &v3.Response{
		Description: ToResponseDescription(route.ResponseDescription),
		Content:     content,
	}
}

func buildSecurityMethod(securitySchemes []definitions.SecuritySchemeConfig, securityMethods []definitions.SecurityAnnotationComponent) (*highbase.SecurityRequirement, error) {
	securityRequirement := orderedmap.New[string, []string]()

	for _, securityMethod := range securityMethods {
		// Make sure the name exists in the openapi security schemes
		if !swagtool.IsSecurityNameInSecuritySchemes(securitySchemes, securityMethod.SchemaName) {
			errStr := fmt.Sprintf("Security method name %s does not exist in the defined security schemes %v",
				securityMethod.SchemaName, securitySchemes)
			return nil, errors.New(errStr)
		}
		securityRequirement.Set(securityMethod.SchemaName, securityMethod.Scopes)
	}

	return &highbase.SecurityRequirement{
		Requirements: securityRequirement,
	}, nil
}

func generateOperationSecurity(operation *v3.Operation, config *definitions.OpenAPIGeneratorConfig, route definitions.RouteMetadata) error {
	var securityRequirements []*highbase.SecurityRequirement

	routeSecurity := route.Security

	if len(routeSecurity) == 0 && config.DefaultRouteSecurity != nil {
		routeSecurity = []definitions.RouteSecurity{{SecurityAnnotation: []definitions.SecurityAnnotationComponent{*config.DefaultRouteSecurity}}}
	}

	for _, security := range routeSecurity {
		securityRequirement, err := buildSecurityMethod(config.SecuritySchemes, security.SecurityAnnotation)

		if err != nil {
			errStr := fmt.Sprintf("Building v3.1 security %s for %s failed: %s",
				route.OperationId, security.SecurityAnnotation, err.Error())
			return errors.New(errStr)
		}
		securityRequirements = append(securityRequirements, securityRequirement)
	}

	operation.Security = securityRequirements
	return nil
}

func setNewRouteOperation(doc *v3.Document, def definitions.ControllerMetadata, route definitions.RouteMetadata, operation *v3.Operation) {
	// Set the operation in the path item
	routePath := def.RestMetadata.Path + route.RestMetadata.Path

	// Ensure Paths exists
	if doc.Paths == nil {
		doc.Paths = &v3.Paths{
			PathItems: orderedmap.New[string, *v3.PathItem](),
		}
	}

	// Get or create path item
	var pathItem *v3.PathItem
	existingPathItem, exists := doc.Paths.PathItems.Get(routePath)
	if exists && existingPathItem != nil {
		pathItem = existingPathItem
	} else {
		pathItem = &v3.PathItem{}
	}

	// Set the operation based on HTTP verb
	switch route.HttpVerb {
	case "GET":
		pathItem.Get = operation
	case "POST":
		pathItem.Post = operation
	case "PUT":
		pathItem.Put = operation
	case "DELETE":
		pathItem.Delete = operation
	case "PATCH":
		pathItem.Patch = operation
	case "HEAD":
		pathItem.Head = operation
	case "OPTIONS":
		pathItem.Options = operation
	case "TRACE":
		pathItem.Trace = operation
	}

	// Set the path item in the document
	doc.Paths.PathItems.Set(routePath, pathItem)
}

func handleRouteParamDeprecation(routeParam definitions.FuncParam, specParam *v3.Parameter) {
	if routeParam.Deprecation != nil && *&routeParam.Deprecation.Deprecated {
		specParam.Deprecated = true
	}
}

func createRouteParam(doc *v3.Document, param definitions.FuncParam) *v3.Parameter {
	schemaRef := InterfaceToSchemaV3(doc, param.TypeMeta.Name)
	if schemaRef.Schema() != nil {
		BuildSchemaValidationV31(schemaRef.Schema(), param.Validator, param.TypeMeta.Name)
	}
	isParamRequired := swagtool.IsFieldRequired(param.Validator)

	specParam := &v3.Parameter{
		Name:        param.NameInSchema,
		In:          strings.ToLower(string(param.PassedIn)),
		Description: param.Description,
		Required:    &isParamRequired,
		Schema:      schemaRef,
	}
	handleRouteParamDeprecation(param, specParam)
	return specParam
}

func createRequestBodyParam(doc *v3.Document, param definitions.FuncParam) *v3.RequestBody {
	content := createContentWithSchemaRef(doc, param.Validator, param.TypeMeta.Name)
	isBodyRequired := swagtool.IsFieldRequired(param.Validator)
	return &v3.RequestBody{
		Description: param.Description,
		Content:     content,
		Required:    &isBodyRequired,
	}
}

func generateParams(doc *v3.Document, route definitions.RouteMetadata, operation *v3.Operation) {
	// Iterate over FuncParams and create parameters
	for _, param := range route.FuncParams {
		if param.PassedIn == definitions.PassedInBody {
			operation.RequestBody = createRequestBodyParam(doc, param)
		} else {
			operation.Parameters = append(operation.Parameters, createRouteParam(doc, param))
		}
	}
}

// GenerateControllerSpec generates the specification for a controller
func generateControllerSpec(doc *v3.Document, config *definitions.OpenAPIGeneratorConfig, def definitions.ControllerMetadata) error {
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
			operation.Responses.Codes.Set(swagtool.HttpStatusCodeToString(errResp.HttpStatusCode), createErrorResponse(doc, route, errResp))
		}

		successResponse := createResponseSuccess(doc, route)
		operation.Responses.Codes.Set(swagtool.HttpStatusCodeToString(route.ResponseSuccessCode), successResponse)

		// operation.Responses.Default - for now, we do not support "default" response

		generateParams(doc, route, operation)

		// Add the security requirement to the operation
		if err := generateOperationSecurity(operation, config, route); err != nil {
			return err
		}

		// Finally, set the operation in the path item
		setNewRouteOperation(doc, def, route, operation)
	}
	return nil
}

func GenerateControllersSpec(doc *v3.Document, config *definitions.OpenAPIGeneratorConfig, defs []definitions.ControllerMetadata) error {
	// Iterate over the routes in the controller
	for _, def := range defs {
		if err := generateControllerSpec(doc, config, def); err != nil {
			errStr := fmt.Sprintf("Building v3.1 controller %s failed: %s", def.Name, err.Error())
			return errors.New(errStr)
		}
	}
	return nil
}
