package swagen

import (
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

func createResponseSuccess(openapi *openapi3.T, route definitions.RouteMetadata) *openapi3.ResponseRef {
	var content openapi3.Content

	// Check if route.ResponseInterface.InterfaceName is a primitive type
	typeName := route.ResponseInterface.InterfaceName
	switch typeName {
	case "string":
		content = openapi3.NewContentWithJSONSchema(openapi3.NewStringSchema())
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		content = openapi3.NewContentWithJSONSchema(openapi3.NewIntegerSchema())
	case "bool":
		content = openapi3.NewContentWithJSONSchema(openapi3.NewBoolSchema())
	case "float32", "float64":
		content = openapi3.NewContentWithJSONSchema(openapi3.NewFloat64Schema())
	default:
		// If it's not a primitive type, create a reference to the schema
		schemaRef := &openapi3.SchemaRef{
			Ref:   "#/components/schemas/" + typeName,
			Value: openapi.Components.Schemas[typeName].Value,
		}
		content = openapi3.NewContentWithJSONSchemaRef(schemaRef)
	}

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
