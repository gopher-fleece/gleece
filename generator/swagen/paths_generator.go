package swagen

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/haimkastner/gleece/definitions"
)

func createOperation(def definitions.ControllerMetadata, route definitions.RouteMetadata) *openapi3.Operation {
	return &openapi3.Operation{
		Summary:     route.Description,
		Description: route.Description,
		Responses:   openapi3.NewResponses(),
		OperationID: route.OperationId,
		Tags:        []string{def.Tag},
		Parameters:  []*openapi3.ParameterRef{},
	}
}

func createErrorResponse(errResp definitions.ErrorResponse) *openapi3.ResponseRef {
	errResString := errResp.Description
	response := &openapi3.Response{
		Description: &errResString,
		Content:     openapi3.NewContentWithJSONSchema(openapi3.NewStringSchema()),
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
	content := createContentWithSchemaRef(openapi, "", route.ResponseInterface.InterfaceName)
	return &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: &route.ResponseDescription,
			Content:     content,
		},
	}
}

func setNewRouteOperation(openapi *openapi3.T, def definitions.ControllerMetadata, route definitions.RouteMetadata, operation *openapi3.Operation) {
	// Set the operation in the path item
	routePath := def.RestMetadata.Path + route.RestMetadata.Path
	// Set the path item in the openapi
	pathItem := openapi.Paths.Find(routePath)
	// If path item is nil, create a new path item
	if pathItem == nil {
		pathItem = &openapi3.PathItem{}
	}
	pathItem.SetOperation(string(route.HttpVerb), operation)
	openapi.Paths.Set(routePath, pathItem)
}

func createPrimitiveParam(openapi *openapi3.T, param definitions.FuncParam) *openapi3.ParameterRef {
	schemaRef := InterfaceToSchemaRef(openapi, param.ParamInterface)
	BuildSchemaValidation(schemaRef, param.Validator, param.ParamInterface)
	return &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:        param.Name,
			In:          strings.ToLower(string(param.ParamType)),
			Description: param.Description,
			Required:    IsFieldRequired(param.Validator),
			Schema:      schemaRef,
		},
	}
}

func createRequestBodyParam(openapi *openapi3.T, param definitions.FuncParam) *openapi3.RequestBodyRef {
	content := createContentWithSchemaRef(openapi, param.Validator, param.ParamInterface)
	return &openapi3.RequestBodyRef{
		Value: &openapi3.RequestBody{
			Description: param.Description,
			Content:     content,
			Required:    IsFieldRequired(param.Validator),
		},
	}
}

func generateParams(openapi *openapi3.T, route definitions.RouteMetadata, operation *openapi3.Operation) {
	// Iterate over FuncParams and create parameters
	for _, param := range route.FuncParams {
		if param.ParamType == definitions.Body {
			operation.RequestBody = createRequestBodyParam(openapi, param)
		} else {
			operation.Parameters = append(operation.Parameters, createPrimitiveParam(openapi, param))
		}
	}
}

// GenerateControllerSpec generates the specification for a controller
func generateControllerSpec(openapi *openapi3.T, def definitions.ControllerMetadata) {
	// Iterate over the routes in the controller
	for _, route := range def.Routes {
		// Create a new Operation for the route
		operation := createOperation(def, route)

		// Iterate over the error responses
		for _, errResp := range route.ErrorResponses {
			// Set the response using the Set method
			operation.Responses.Set(HttpStatusCodeToString(errResp.HttpStatusCode), createErrorResponse(errResp))
		}

		// Set the success response
		operation.Responses.Set(HttpStatusCodeToString(route.ResponseSuccessCode), createResponseSuccess(openapi, route))

		generateParams(openapi, route, operation)

		// Finally, set the operation in the path item
		setNewRouteOperation(openapi, def, route, operation)
	}
}

func GenerateControllersSpec(openapi *openapi3.T, defs []definitions.ControllerMetadata) {
	// Iterate over the routes in the controller
	for _, def := range defs {
		generateControllerSpec(openapi, def)
	}
}
