package swagen

import (
	"encoding/json"
	"fmt"
	"os"
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
					Responses: []definitions.FuncReturnValue{
						{
							TypeMetadata: definitions.TypeMetadata{
								Name:                  "[]ExampleSchema",
								FullyQualifiedPackage: "example",
							},
						},
						{
							TypeMetadata: definitions.TypeMetadata{
								Name: "error",
							},
						},
					},
					FuncParams: []definitions.FuncParam{
						{
							ParamMeta: definitions.ParamMeta{
								Name: "my_name",
								TypeMeta: definitions.TypeMetadata{
									Name: "string",
								},
							},
							PassedIn:    definitions.PassedInQuery,
							Description: "Example query param",
							Validator:   "required,email",
						},
						{
							ParamMeta: definitions.ParamMeta{
								Name: "my_names",
								TypeMeta: definitions.TypeMetadata{
									Name: "[]ExampleSchema",
								},
							},
							PassedIn:    definitions.PassedInQuery,
							Description: "Example query ARR param",
							Validator:   "",
						},
						{
							ParamMeta: definitions.ParamMeta{
								Name: "my_header",
								TypeMeta: definitions.TypeMetadata{
									Name: "bool",
								},
							},
							PassedIn:    definitions.PassedInHeader,
							Description: "Example Header param",
							Validator:   "required",
						},
						{
							ParamMeta: definitions.ParamMeta{
								Name: "my_number",
								TypeMeta: definitions.TypeMetadata{
									Name: "float64",
								},
							},
							PassedIn:    definitions.PassedInHeader,
							Description: "Example Header num param",
							Validator:   "gt=1,lt=100",
						},
						{
							ParamMeta: definitions.ParamMeta{
								Name: "my_path",
								TypeMeta: definitions.TypeMetadata{
									Name: "int",
								},
							},
							PassedIn:    definitions.PassedInPath,
							Description: "Example Path param",
							Validator:   "required",
						},
						// {
						// 	Name:           "the_content",
						// 	PassedIn:      definitions.Body,
						// 	Description:    "Example Body param",
						// 	ParamInterface: "[]ExampleSchema",
						// 	Validator:      "required",
						// },
						{
							ParamMeta: definitions.ParamMeta{
								Name: "the_content",
								TypeMeta: definitions.TypeMetadata{
									Name: "string",
								},
							},
							PassedIn:    definitions.PassedInBody,
							Description: "Example Body param",
							Validator:   "required,email",
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
					Responses: []definitions.FuncReturnValue{
						{
							TypeMetadata: definitions.TypeMetadata{
								Name:                  "int",
								FullyQualifiedPackage: "example",
							},
						},
						{
							TypeMetadata: definitions.TypeMetadata{
								Name: "error",
							},
						},
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
					Responses: []definitions.FuncReturnValue{
						{
							TypeMetadata: definitions.TypeMetadata{
								Name: "error",
							},
						},
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

		jsonBytes, err := GenerateSpec(&definitions.OpenAPIGeneratorConfig{
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
			SecuritySchemes: []definitions.SecuritySchemeConfig{
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

		// Output the generated JSON and the expected JSON to files to easy testing troubleshooting
		// Create the output directory if it doesn't exist
		if err := os.MkdirAll("dist", os.ModePerm); err != nil {
			fmt.Println("Failed to create directory:", err)
		}

		// Write the JSON to the file
		filePath := "dist/org-spec.json"
		if err := os.WriteFile(filePath, fullyFeaturesSpec, 0644); err != nil {
			fmt.Println("Failed to write file:", err)
		}

		// Write the JSON to the file
		filePath2 := "dist/test-spec.json"
		if err := os.WriteFile(filePath2, jsonBytes, 0644); err != nil {
			fmt.Println("Failed to write file:", err)
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
					Responses: []definitions.FuncReturnValue{
						{
							TypeMetadata: definitions.TypeMetadata{
								Name:                  "[]string",
								FullyQualifiedPackage: "example",
							},
						},
						{
							TypeMetadata: definitions.TypeMetadata{
								Name: "error",
							},
						},
					},
					FuncParams: []definitions.FuncParam{},
				},
			},
		})

		models := []definitions.ModelMetadata{}

		_, err := GenerateSpec(&definitions.OpenAPIGeneratorConfig{
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
