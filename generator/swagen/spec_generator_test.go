package swagen

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/haimkastner/gleece/definitions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var fullyFeaturesSpec = []byte(`{"components":{"schemas":{"ExampleSchema":{"description":"Example schema","properties":{"ExampleArrField":{"description":"Example array field","items":{"$ref":"#/components/schemas/ExampleSchema222"},"type":"array"},"ExampleArrStringField":{"description":"Example int arr field","items":{"items":{"items":{"items":{"$ref":"#/components/schemas/ExampleSchema222"},"type":"array"},"type":"array"},"type":"array"},"type":"array"},"ExampleField":{"description":"Example field","type":"string"},"ExampleObjField":{"$ref":"#/components/schemas/ExampleSchema222"}},"required":["ExampleField","ExampleObjField","ExampleArrField"],"title":"ExampleSchema","type":"object"},"ExampleSchema222":{"description":"Example object ref field","properties":{"MaxValue":{"description":"MaxValue DESCRIPTION","maximum":100,"minimum":1,"type":"integer"},"TheName":{"description":"TheName DESCRIPTION","format":"email","type":"string"}},"required":["TheName"],"title":"ExampleSchema222","type":"object"}},"securitySchemes":{"ApiKeyAuth":{"description":"API Key","in":"header","name":"X-API-Key2","type":"apiKey"},"ApiKeyAuth2":{"description":"API Key","in":"header","name":"X-API-Key2","type":"apiKey"}}},"info":{"contact":{"name":"John Doe"},"description":"This is a simple API?","license":{"name":"Apache 2.0","url":"https://www.apache.org/licenses/LICENSE-2.0.html"},"title":"My API","version":"1.0.0"},"openapi":"3.0.0","paths":{"/example-base/example-route":{"delete":{"description":"Example route","operationId":"exampleRouteDel","responses":{"204":{"description":"Example response OK for 204"},"500":{"content":{"application/json":{"schema":{"type":"object"}}},"description":"Internal server error"},"default":{"description":""}},"security":[{"ApiKeyAuth":["read"]}],"summary":"Example route","tags":["Example"]},"post":{"description":"Example route","operationId":"exampleRoute45","responses":{"200":{"content":{"application/json":{"schema":{"type":"integer"}}},"description":"Example response OK"},"500":{"content":{"application/json":{"schema":{"type":"object"}}},"description":"Internal server error"},"default":{"description":""}},"security":[{"ApiKeyAuth":["read"]}],"summary":"Example route","tags":["Example"]}},"/example-base/example-route/{my_path}":{"get":{"description":"Example route","operationId":"exampleRoute","parameters":[{"description":"Example query param","in":"query","name":"my_name","required":true,"schema":{"format":"email","type":"string"}},{"description":"Example query ARR param","in":"query","name":"my_names","schema":{"items":{"$ref":"#/components/schemas/ExampleSchema"},"type":"array"}},{"description":"Example Header param","in":"header","name":"my_header","required":true,"schema":{"type":"boolean"}},{"description":"Example Header num param","in":"header","name":"my_number","schema":{"maximum":100,"minimum":1,"type":"number"}},{"description":"Example Path param","in":"path","name":"my_path","required":true,"schema":{"type":"integer"}}],"requestBody":{"content":{"application/json":{"schema":{"format":"email","type":"string"}}},"description":"Example Body param","required":true},"responses":{"200":{"content":{"application/json":{"schema":{"items":{"$ref":"#/components/schemas/ExampleSchema"},"type":"array"}}},"description":""},"500":{"content":{"application/json":{"schema":{"type":"object"}}},"description":"Internal server error"},"default":{"description":""}},"security":[{"ApiKeyAuth":["read","write"],"ApiKeyAuth2":["write"]},{"ApiKeyAuth":["read"]}],"summary":"Example route","tags":["Example"]}}},"servers":[{"url":"http://localhost:8080"}]}`)

var _ = Describe("Spec Generator", func() {

	It("Should generate a fully featured OpenAPI spec", func() {

		defs := []definitions.ControllerMetadata{}

		// Create an example controller and add it to the definitions
		defs = append(defs, definitions.ControllerMetadata{
			Tag: "Example",
			RestMetadata: definitions.RestMetadata{
				Path: "/example-base",
			},
			Description: "Example controller",
			Name:        "ExampleController",
			Package:     "example",
			Routes: []definitions.RouteMetadata{
				{
					Security: []definitions.RouteSecurity{
						{
							SecurityMethod: []definitions.SecurityMethod{
								{
									Name:        "ApiKeyAuth",
									Permissions: []string{"read", "write"},
								},
								{
									Name:        "ApiKeyAuth2",
									Permissions: []string{"write"},
								},
							},
						},
						{
							SecurityMethod: []definitions.SecurityMethod{
								{
									Name:        "ApiKeyAuth",
									Permissions: []string{"read"},
								},
							},
						},
					},
					HttpVerb: "GET",
					RestMetadata: definitions.RestMetadata{
						Path: "/example-route/{my_path}",
					},
					Description: "Example route",
					OperationId: "exampleRoute",
					ErrorResponses: []definitions.ErrorResponse{
						{
							Description:    "Internal server error",
							HttpStatusCode: 500,
						},
					},
					ResponseSuccessCode: 200,
					ResponseInterface: definitions.ResponseMetadata{
						InterfaceName:         "[]ExampleSchema",
						Signature:             definitions.FuncRetValueAndError,
						FullyQualifiedPackage: "example",
					},
					FuncParams: []definitions.FuncParam{
						{
							Name:           "my_name",
							ParamType:      definitions.Query,
							Description:    "Example query param",
							ParamInterface: "string",
							Validator:      "required,email",
						},
						{
							Name:           "my_names",
							ParamType:      definitions.Query,
							Description:    "Example query ARR param",
							ParamInterface: "[]ExampleSchema",
							Validator:      "",
						},
						{
							Name:           "my_header",
							ParamType:      definitions.Header,
							Description:    "Example Header param",
							ParamInterface: "bool",
							Validator:      "required",
						},
						{
							Name:           "my_number",
							ParamType:      definitions.Header,
							Description:    "Example Header num param",
							ParamInterface: "float64",
							Validator:      "gt=1,lt=100",
						},
						{
							Name:           "my_path",
							ParamType:      definitions.Path,
							Description:    "Example Path param",
							ParamInterface: "int",
							Validator:      "required",
						},
						// {
						// 	Name:           "the_content",
						// 	ParamType:      definitions.Body,
						// 	Description:    "Example Body param",
						// 	ParamInterface: "[]ExampleSchema",
						// 	Validator:      "required",
						// },
						{
							Name:           "the_content",
							ParamType:      definitions.Body,
							Description:    "Example Body param",
							ParamInterface: "string",
							Validator:      "required,email",
						},
					},
				},
				{
					HttpVerb: "POST",
					RestMetadata: definitions.RestMetadata{
						Path: "/example-route",
					},
					Description: "Example route",
					OperationId: "exampleRoute45",
					ErrorResponses: []definitions.ErrorResponse{
						{
							Description:    "Internal server error",
							HttpStatusCode: 500,
						},
					},
					ResponseSuccessCode: 200,
					ResponseInterface: definitions.ResponseMetadata{
						InterfaceName:         "int",
						Signature:             definitions.FuncRetValueAndError,
						FullyQualifiedPackage: "example",
					},
					ResponseDescription: "Example response OK",
				},
				{
					HttpVerb: "DELETE",
					RestMetadata: definitions.RestMetadata{
						Path: "/example-route",
					},
					Description: "Example route",
					OperationId: "exampleRouteDel",
					ErrorResponses: []definitions.ErrorResponse{
						{
							Description:    "Internal server error",
							HttpStatusCode: 500,
						},
					},
					ResponseSuccessCode: 204,
					ResponseInterface: definitions.ResponseMetadata{
						InterfaceName:         "",
						Signature:             definitions.FuncRetError,
						FullyQualifiedPackage: "example",
					},
					ResponseDescription: "Example response OK for 204",
				},
			},
		})

		// Create an example models and add it to the definitions
		models := []definitions.ModelMetadata{
			{
				Name:        "ExampleSchema222",
				Package:     "example2",
				Description: "Example schema 222",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "MaxValue",
						Type:        "int",
						Description: "MaxValue DESCRIPTION",
						Validator:   "gt=1,lt=100",
					},
					{
						Name:        "TheName",
						Type:        "string",
						Description: "TheName DESCRIPTION",
						Validator:   "required,email",
					},
				},
			},
			{
				Name:        "ExampleSchema",
				Package:     "example",
				Description: "Example schema",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "ExampleField",
						Type:        "string",
						Description: "Example field",
						Validator:   "required",
					},
					{
						Name:        "ExampleObjField",
						Type:        "ExampleSchema222",
						Description: "Example object ref field",
						Validator:   "required",
					},
					{
						Name:        "ExampleArrField",
						Type:        "[]ExampleSchema222",
						Description: "Example array field",
						Validator:   "required",
					},
					{
						Name:        "ExampleArrStringField",
						Type:        "[][][][]ExampleSchema222",
						Description: "Example int arr field",
						Validator:   "",
					},
				},
			},
		}

		jsonBytes, err := GenerateSpec(OpenAPIGeneratorConfig{
			Info: openapi3.Info{
				Title:       "My API",
				Version:     "1.0.0",
				Description: "This is a simple API?",
				Contact: &openapi3.Contact{
					Name:  "John Doe",
					Email: "",
				},
				License: &openapi3.License{
					Name: "Apache 2.0",
					URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
				},
			},
			BaseURL: "http://localhost:8080",
			SecuritySchemes: []SecuritySchemeConfig{
				{
					SecurityName: "ApiKeyAuth",
					FieldName:    "X-API-Key2",
					Type:         definitions.APIKey,
					In:           definitions.InHeader,
					Description:  "API Key",
				},
				{
					SecurityName: "ApiKeyAuth2",
					FieldName:    "X-API-Key2",
					Type:         definitions.APIKey,
					In:           definitions.InHeader,
					Description:  "API Key",
				},
			},
			DefaultRouteSecurity: []definitions.RouteSecurity{
				{
					SecurityMethod: []definitions.SecurityMethod{
						{
							Name:        "ApiKeyAuth",
							Permissions: []string{"read"},
						},
					},
				},
			},
		}, defs, models)

		// If it fails, throw an error
		if err != nil {
			// throw error and print the error too
			Fail("Failed to generate OpenAPI spec: " + err.Error())
		}

		// Compare the generated spec with the expected spec
		areEqual, err := areJSONsIdentical(jsonBytes, fullyFeaturesSpec)
		if err != nil {
			Fail("Failed to compare JSONs: " + err.Error())
		}
		Expect(areEqual).To(BeTrue())
	})

	It("Should throw error for invalid spec", func() {

		defs := []definitions.ControllerMetadata{}

		// Create an example controller and add it to the definitions
		defs = append(defs, definitions.ControllerMetadata{
			Tag: "Example",
			RestMetadata: definitions.RestMetadata{
				Path: "/example-base",
			},
			Description: "Example controller",
			Name:        "ExampleController",
			Package:     "example",
			Routes: []definitions.RouteMetadata{
				{
					HttpVerb: "GET",
					RestMetadata: definitions.RestMetadata{
						Path: "/example-route/{my_path}",
					},
					Description: "Example route",
					OperationId: "exampleRoute",
					ErrorResponses: []definitions.ErrorResponse{
						{
							Description:    "Internal server error",
							HttpStatusCode: 500,
						},
					},
					ResponseSuccessCode: 200,
					ResponseInterface: definitions.ResponseMetadata{
						InterfaceName:         "[]string",
						Signature:             definitions.FuncRetValueAndError,
						FullyQualifiedPackage: "example",
					},
					FuncParams: []definitions.FuncParam{},
				},
			},
		})

		models := []definitions.ModelMetadata{}

		_, err := GenerateSpec(OpenAPIGeneratorConfig{
			Info: openapi3.Info{
				Title:   "My API",
				Version: "1.0.0",
			},
			BaseURL:              "http://localhost:8080",
			DefaultRouteSecurity: []definitions.RouteSecurity{},
		}, defs, models)

		// Expect error not to be nil
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("invalid paths: operation GET /example-base/example-route/{my_path} must define exactly all path parameters (missing: [my_path])"))
	})
})

func areJSONsIdentical(json1 []byte, json2 []byte) (bool, error) {
	var obj1, obj2 map[string]interface{}

	err := json.Unmarshal(json1, &obj1)
	if err != nil {
		return false, fmt.Errorf("invalid JSON 1: %v", err)
	}

	err = json.Unmarshal(json2, &obj2)
	if err != nil {
		return false, fmt.Errorf("invalid JSON 2: %v", err)
	}

	return reflect.DeepEqual(obj1, obj2), nil
}
