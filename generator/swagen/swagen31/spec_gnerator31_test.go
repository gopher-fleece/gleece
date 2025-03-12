package swagen31

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/generator/swagen/swagtool"
	"github.com/gopher-fleece/gleece/infrastructure/logger"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var fullyFeaturesSpec = []byte(`{"openapi":"3.1.0","info":{"title":"My API","description":"This is a simple API?","contact":{"name":"John Doe"},"license":{"name":"Apache 2.0","url":"https://www.apache.org/licenses/LICENSE-2.0.html"},"version":"1.0.0"},"servers":[{"url":"http://localhost:8080"}],"paths":{"/example-base/example-route/{my_path}":{"get":{"tags":["Example"],"summary":"Example route","description":"Example route","operationId":"exampleRoute","parameters":[{"name":"my_name","in":"query","description":"Example query param","required":true,"deprecated":true,"schema":{"type":"string","format":"email"}},{"name":"my_names","in":"query","description":"Example query ARR param","required":true,"schema":{"type":"array","items":{"$ref":"#/components/schemas/ExampleSchema"}}},{"name":"my_header","in":"header","description":"Example Header param","required":true,"schema":{"type":"boolean"}},{"name":"my_number","in":"header","description":"Example Header num param","required":true,"schema":{"exclusiveMaximum":100,"type":"number","minimum":18}},{"name":"my_path","in":"path","description":"Example Path param","required":true,"schema":{"type":"integer","enum":[1,2,3,4]}}],"requestBody":{"description":"Example Body param","content":{"application/json":{"schema":{"type":"string","format":"email"}}},"required":true},"responses":{"200":{"description":" ","content":{"application/json":{"schema":{"type":"array","items":{"$ref":"#/components/schemas/ExampleSchema"}}}}},"500":{"description":"Internal server error","content":{"application/json":{"schema":{"$ref":"#/components/schemas/Rfc7807Error"}}}}},"security":[{"ApiKeyAuth":["read","write"],"ApiKeyAuth2":["write"]},{"ApiKeyAuth":["read"]}]}},"/example-base/example-route":{"post":{"tags":["Example"],"summary":"Example route","description":"Example route","operationId":"exampleRoute45","parameters":[],"responses":{"200":{"description":"Example response OK","content":{"application/json":{"schema":{"type":"integer"}}}},"500":{"description":"Internal server error","content":{"application/json":{"schema":{"type":"string"}}}}},"security":[{"ApiKeyAuth":["read"]}]},"delete":{"tags":["Example"],"summary":"Example route","description":"Example route","operationId":"exampleRouteDel","parameters":[],"responses":{"204":{"description":"Example response OK for 204"},"500":{"description":"Internal server error","content":{"application/json":{"schema":{"type":"string"}}}}},"deprecated":true,"security":[{"ApiKeyAuth":["read"]}]}},"/example-base/post-enum":{"post":{"tags":["Example"],"summary":"Example enum route","description":"Example enum route","operationId":"exampleEnumRoute","parameters":[{"name":"my_enum","in":"query","description":"Example enum num param","required":true,"schema":{"$ref":"#/components/schemas/Status"}}],"requestBody":{"description":"Example Struct with Enum","content":{"application/json":{"schema":{"$ref":"#/components/schemas/ExampleSchemaWithEnum"}}},"required":true},"responses":{"200":{"description":" ","content":{"application/json":{"schema":{"type":"array","items":{"$ref":"#/components/schemas/ExampleSchemaWithEnum"}}}}},"500":{"description":"Internal server error","content":{"application/json":{"schema":{"$ref":"#/components/schemas/Rfc7807Error"}}}}},"security":[{"ApiKeyAuth":["read","write"],"ApiKeyAuth2":["write"]},{"ApiKeyAuth":["read"]}]}}},"components":{"schemas":{"Status":{"type":"string","title":"Status","enum":["ACTIVE","INACTIVE","SUSPENDED"],"description":"User status enum"},"Status2":{"type":"string","title":"Status2","enum":["ACTIVE2","INACTIVE2","SUSPENDED2"],"description":"User status enum"},"ExampleSchema222":{"type":"object","properties":{"MaxValue":{"type":"integer","maximum":100,"minimum":1,"description":"MaxValue DESCRIPTION"},"TheName":{"type":"string","format":"email","description":"TheName DESCRIPTION"}},"title":"ExampleSchema222","required":["TheName"],"description":"Example schema 222","deprecated":true},"ExampleSchemaWithEnum":{"type":"object","properties":{"TheStatus":{"$ref":"#/components/schemas/Status"},"TheStatus2":{"$ref":"#/components/schemas/Status2"}},"title":"ExampleSchemaWithEnum","required":["TheStatus"],"description":"Example enum schema"},"ExampleSchema":{"type":"object","properties":{"ExampleField":{"type":"string","enum":["one","two","three"],"description":"Example field","deprecated":true},"ExampleObjField":{"$ref":"#/components/schemas/ExampleSchema222"},"ExampleArrField":{"type":"array","items":{"$ref":"#/components/schemas/ExampleSchema222"},"description":"Example array field"},"ExampleArrStringField":{"type":"array","items":{"type":"array","items":{"type":"array","items":{"type":"array","items":{"$ref":"#/components/schemas/ExampleSchema222"}}}},"description":"Example int arr field"}},"title":"ExampleSchema","required":["ExampleField","ExampleObjField","ExampleArrField"],"description":"Example schema"},"Rfc7807Error":{"type":"object","properties":{"type":{"type":"string","description":"A URI reference that identifies the problem type."},"title":{"type":"string","description":"A short, human-readable summary of the problem type."},"status":{"type":"integer","description":"The HTTP status code generated by the origin server for this occurrence of the problem."},"detail":{"type":"string","description":"A human-readable explanation specific to this occurrence of the problem."},"instance":{"type":"string","description":"A URI reference that identifies the specific occurrence of the problem."},"error":{"type":"string","description":"Error message"},"extensions":{"type":"object","description":"Additional metadata about the error."}},"title":"Rfc7807Error","required":["type","title","status"],"description":"A standard RFC-7807 error"}},"securitySchemes":{"ApiKeyAuth":{"type":"apiKey","description":"API Key","name":"X-API-Key2","in":"header"},"ApiKeyAuth2":{"type":"apiKey","description":"API Key","name":"X-API-Key2","in":"header"}}}}`)
var formSpec = []byte(`{"openapi":"3.1.0","info":{"title":"My API","description":"This is a simple API?","contact":{"name":"John Doe"},"license":{"name":"Apache 2.0","url":"https://www.apache.org/licenses/LICENSE-2.0.html"},"version":"1.0.0"},"servers":[{"url":"http://localhost:8080"}],"paths":{"/example-base/example-route":{"post":{"tags":["Example"],"summary":"Example form route","description":"Example form route","operationId":"exampleRoute","parameters":[],"requestBody":{"description":"Example my_form param","content":{"application/x-www-form-urlencoded":{"schema":{"type":"object","properties":{"my_form":{"type":"string"},"my_form_number":{"exclusiveMaximum":100,"type":"integer","minimum":1},"my_form_option":{"type":"boolean"}},"required":["my_form","my_form_number"]}}}},"responses":{"200":{"description":" "},"500":{"description":"Internal server error","content":{"application/json":{"schema":{"$ref":"#/components/schemas/Rfc7807Error"}}}}},"security":[{"ApiKeyAuth":["read"]}]}}},"components":{"schemas":{"Rfc7807Error":{"type":"object","properties":{"type":{"type":"string","description":"A URI reference that identifies the problem type."},"title":{"type":"string","description":"A short, human-readable summary of the problem type."},"status":{"type":"integer","description":"The HTTP status code generated by the origin server for this occurrence of the problem."},"detail":{"type":"string","description":"A human-readable explanation specific to this occurrence of the problem."},"instance":{"type":"string","description":"A URI reference that identifies the specific occurrence of the problem."},"error":{"type":"string","description":"Error message"},"extensions":{"type":"object","description":"Additional metadata about the error."}},"title":"Rfc7807Error","required":["type","title","status"],"description":"A standard RFC-7807 error"}},"securitySchemes":{"ApiKeyAuth":{"type":"apiKey","description":"API Key","name":"X-API-Key2","in":"header"}}}}`)

var _ = Describe("Spec v3.1 Generator", func() {

	It("Should generate a fully featured OpenAPI 3.1 spec", func() {

		defs := []definitions.ControllerMetadata{}

		enums := []definitions.EnumMetadata{{
			Name:        "Status",
			Description: "User status enum",
			Type:        "string",
			Values:      []string{"ACTIVE", "INACTIVE", "SUSPENDED"},
		}, {
			Name:        "Status2",
			Description: "User status enum",
			Type:        "string",
			Values:      []string{"ACTIVE2", "INACTIVE2", "SUSPENDED2"},
		}}

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
							SecurityAnnotation: []definitions.SecurityAnnotationComponent{
								{
									SchemaName: "ApiKeyAuth",
									Scopes:     []string{"read", "write"},
								},
								{
									SchemaName: "ApiKeyAuth2",
									Scopes:     []string{"write"},
								},
							},
						},
						{
							SecurityAnnotation: []definitions.SecurityAnnotationComponent{
								{
									SchemaName: "ApiKeyAuth",
									Scopes:     []string{"read"},
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
					ResponseDescription: "",
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
							NameInSchema: "my_name",
							PassedIn:     definitions.PassedInQuery,
							Description:  "Example query param",
							Validator:    "required,email",
							Deprecation: &definitions.DeprecationOptions{
								Description: "This query is deprecated example",
								Deprecated:  true,
							},
						},
						{
							ParamMeta: definitions.ParamMeta{
								Name: "my_names",
								TypeMeta: definitions.TypeMetadata{
									Name: "[]ExampleSchema",
								},
							},
							NameInSchema: "my_names",
							PassedIn:     definitions.PassedInQuery,
							Description:  "Example query ARR param",
							Validator:    "required",
						},
						{
							ParamMeta: definitions.ParamMeta{
								Name: "my_header",
								TypeMeta: definitions.TypeMetadata{
									Name: "bool",
								},
							},
							NameInSchema: "my_header",
							PassedIn:     definitions.PassedInHeader,
							Description:  "Example Header param",
							Validator:    "required",
						},
						{
							ParamMeta: definitions.ParamMeta{
								Name: "my_number",
								TypeMeta: definitions.TypeMetadata{
									Name: "float64",
								},
							},
							NameInSchema: "my_number",
							PassedIn:     definitions.PassedInHeader,
							Description:  "Example Header num param",
							Validator:    "required,gte=18,lt=100",
						},
						{
							ParamMeta: definitions.ParamMeta{
								Name: "my_path",
								TypeMeta: definitions.TypeMetadata{
									Name: "int",
								},
							},
							NameInSchema: "my_path",
							PassedIn:     definitions.PassedInPath,
							Description:  "Example Path param",
							Validator:    "required,oneof=1 2 3 4 ",
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
							NameInSchema: "the_content",
							PassedIn:     definitions.PassedInBody,
							Description:  "Example Body param",
							Validator:    "required,email",
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
								Name: "string",
							},
						},
					},
					ResponseDescription: "Example response OK",
				},
				{
					Deprecation: definitions.DeprecationOptions{
						Description: "This route is deprecated example",
						Deprecated:  true,
					},
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
								Name: "string",
							},
						},
					},
					ResponseDescription: "Example response OK for 204",
				},
				{
					HttpVerb: "PUT",
					RestMetadata: definitions.RestMetadata{
						Path: "/example-route",
					},
					Description: "NOT SHOWN",
					OperationId: "exampleRoutePut",
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
					Hiding: definitions.MethodHideOptions{
						Type:      definitions.HideMethodAlways,
						Condition: "",
					},
				},
				{
					Security: []definitions.RouteSecurity{
						{
							SecurityAnnotation: []definitions.SecurityAnnotationComponent{
								{
									SchemaName: "ApiKeyAuth",
									Scopes:     []string{"read", "write"},
								},
								{
									SchemaName: "ApiKeyAuth2",
									Scopes:     []string{"write"},
								},
							},
						},
						{
							SecurityAnnotation: []definitions.SecurityAnnotationComponent{
								{
									SchemaName: "ApiKeyAuth",
									Scopes:     []string{"read"},
								},
							},
						},
					},
					HttpVerb: "POST",
					RestMetadata: definitions.RestMetadata{
						Path: "/post-enum",
					},
					Description: "Example enum route",
					OperationId: "exampleEnumRoute",
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
								Name:                  "[]ExampleSchemaWithEnum",
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
								Name: "Status",
								TypeMeta: definitions.TypeMetadata{
									Name: "Status",
								},
							},
							NameInSchema: "my_enum",
							PassedIn:     definitions.PassedInQuery,
							Description:  "Example enum num param",
							Validator:    "required",
						},
						{
							ParamMeta: definitions.ParamMeta{
								Name: "the_content",
								TypeMeta: definitions.TypeMetadata{
									Name: "ExampleSchemaWithEnum",
								},
							},
							NameInSchema: "the_content",
							PassedIn:     definitions.PassedInBody,
							Description:  "Example Struct with Enum",
							Validator:    "required",
						},
					},
				},
			},
		})

		// Create an example structs and add it to the definitions
		structs := []definitions.StructMetadata{
			{
				Name:        "ExampleSchema222",
				Description: "Example schema 222",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "maxValue",
						Type:        "int",
						Description: "MaxValue DESCRIPTION",
						Tag:         `json:"MaxValue" validate:"gte=1,lte=100"`,
					},
					{
						Name:        "TheName",
						Type:        "string",
						Description: "TheName DESCRIPTION",
						Tag:         `validate:"required,email"`,
					},
				},
				Deprecation: definitions.DeprecationOptions{
					Description: "This model is deprecated example",
					Deprecated:  true,
				},
			},
			{
				Name:        "ExampleSchemaWithEnum",
				Description: "Example enum schema",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "TheStatus",
						Type:        "Status",
						Description: "The Schema status",
						Tag:         `validate:"required"`,
					},
					{
						Name:        "TheStatus2",
						Type:        "Status2",
						Description: "The Schema status2",
						Tag:         "",
					},
				},
			},
			{
				Name:        "ExampleSchema",
				Description: "Example schema",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "ExampleField",
						Type:        "string",
						Description: "Example field",
						Tag:         `validate:"required,oneof=one two three"`,
						Deprecation: &definitions.DeprecationOptions{
							Description: "This query is deprecated example",
							Deprecated:  true,
						},
					},
					{
						Name:        "ExampleObjField",
						Type:        "ExampleSchema222",
						Description: "Example object ref field",
						Tag:         `validate:"required"`,
					},
					{
						Name:        "ExampleArrField",
						Type:        "[]ExampleSchema222",
						Description: "Example array field",
						Tag:         `validate:"required"`,
					},
					{
						Name:        "ExampleArrStringField",
						Type:        "[][][][]ExampleSchema222",
						Description: "Example int arr field",
						Tag:         "",
					},
				},
			},
		}

		swagtool.AppendErrorSchema(&structs, true)

		jsonBytes, err := GenerateSpec(&definitions.OpenAPIGeneratorConfig{
			Info: definitions.OpenAPIInfo{
				Title:          "My API",
				Version:        "1.0.0",
				Description:    "This is a simple API?",
				TermsOfService: "",
				Contact: &definitions.OpenAPIContact{
					Name:  "John Doe",
					Email: "",
					URL:   "",
				},
				License: &definitions.OpenAPILicense{
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
			DefaultRouteSecurity: &definitions.SecurityAnnotationComponent{
				SchemaName: "ApiKeyAuth",
				Scopes:     []string{"read"},
			},
		}, defs, &definitions.Models{
			Structs: structs,
			Enums:   enums,
		})

		// If it fails, throw an error
		if err != nil {
			// throw error and print the error too
			Fail("Failed to generate OpenAPI spec: " + err.Error())
		}

		// Compare the generated spec with the expected spec
		areEqual, err := swagtool.AreJSONsIdentical(jsonBytes, fullyFeaturesSpec)
		if err != nil {
			Fail("Failed to compare JSONs: " + err.Error())
		}

		// Output the generated JSON and the expected JSON to files to easy testing troubleshooting
		// Create the output directory if it doesn't exist
		if err := os.MkdirAll("dist", os.ModePerm); err != nil {
			logger.Error("Failed to create directory - %v", err)
		}

		var prettyJSON bytes.Buffer
		json.Indent(&prettyJSON, fullyFeaturesSpec, "", "    ")

		// Write the JSON to the file
		filePath := "dist/org-spec-v3.json"
		if err := os.WriteFile(filePath, prettyJSON.Bytes(), 0644); err != nil {
			logger.Error("Failed to write file - %v", err)
		}

		// Write the JSON to the file
		filePath2 := "dist/test-spec-v3.json"
		if err := os.WriteFile(filePath2, jsonBytes, 0644); err != nil {
			logger.Error("Failed to write file - %v", err)
		}
		Expect(areEqual).To(BeTrue())
	})

	It("Should throw error for invalid 3.1 spec", func() {

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

		models := []definitions.StructMetadata{}

		_, err := GenerateSpec(&definitions.OpenAPIGeneratorConfig{
			Info: definitions.OpenAPIInfo{
				Title:   "My API",
				Version: "1.0.0",
			},
			BaseURL: "http://localhost:8080",
		}, defs, &definitions.Models{
			Structs: models,
		})

		// Expect error not to be nil
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("Failed to build a valid v3.1 specification component `#/components/schemas/Rfc7807Error` does not exist in the specification"))
	})

	It("Should throw error for invalid 3.1 spec content", func() {

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
						Path: "/example-route",
					},
					Description: "Example route",
					OperationId: "exampleRoute",
					ErrorResponses: []definitions.ErrorResponse{
						{
							Description:    "Internal server error",
							HttpStatusCode: 5000,
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

		models := []definitions.StructMetadata{}

		swagtool.AppendErrorSchema(&models, true)

		_, err := GenerateSpec(&definitions.OpenAPIGeneratorConfig{
			Info: definitions.OpenAPIInfo{
				Title:   "My API",
				Version: "1.0.0",
			},
			BaseURL: "http://localhost:8080",
		}, defs, &definitions.Models{
			Structs: models,
		})

		// Expect error not to be nil
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("Failed to build a valid v3.1 documentation specification Error: Document does not pass validation, Reason: OpenAPI document is not valid according to the 3.1.0 specification, Validation Errors: [Reason: validation failed, Location: /paths/~1example-base~1example-route/get/responses/5000]"))
	})

	It("Should generate OpenAPI 3.1 spec for x-www-form-urlencoded", func() {

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
							SecurityAnnotation: []definitions.SecurityAnnotationComponent{
								{
									SchemaName: "ApiKeyAuth",
									Scopes:     []string{"read"},
								},
							},
						},
					},
					HttpVerb: "POST",
					RestMetadata: definitions.RestMetadata{
						Path: "/example-route",
					},
					Description: "Example form route",
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
								Name: "error",
							},
						},
					},
					FuncParams: []definitions.FuncParam{
						{
							ParamMeta: definitions.ParamMeta{
								Name: "my_form",
								TypeMeta: definitions.TypeMetadata{
									Name: "string",
								},
							},
							NameInSchema: "my_form",
							PassedIn:     definitions.PassedInForm,
							Description:  "Example my_form param",
							Validator:    "required",
						},
						{
							ParamMeta: definitions.ParamMeta{
								Name: "my_form_number",
								TypeMeta: definitions.TypeMetadata{
									Name: "int",
								},
							},
							NameInSchema: "my_form_number",
							PassedIn:     definitions.PassedInForm,
							Description:  "Example my_form_number param",
							Validator:    "required,gte=1,lt=100",
						},
						{
							ParamMeta: definitions.ParamMeta{
								Name: "my_form_option",
								TypeMeta: definitions.TypeMetadata{
									Name: "bool",
								},
							},
							NameInSchema: "my_form_option",
							PassedIn:     definitions.PassedInForm,
							Description:  "Example Header num param",
							Validator:    "",
						},
					},
				},
			},
		})

		// Create an example models and add it to the definitions
		models := []definitions.StructMetadata{}

		swagtool.AppendErrorSchema(&models, true)

		jsonBytes, err := GenerateSpec(&definitions.OpenAPIGeneratorConfig{
			Info: definitions.OpenAPIInfo{
				Title:          "My API",
				Version:        "1.0.0",
				Description:    "This is a simple API?",
				TermsOfService: "",
				Contact: &definitions.OpenAPIContact{
					Name:  "John Doe",
					Email: "",
					URL:   "",
				},
				License: &definitions.OpenAPILicense{
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
			},
			DefaultRouteSecurity: &definitions.SecurityAnnotationComponent{
				SchemaName: "ApiKeyAuth",
				Scopes:     []string{"read"},
			},
		}, defs, &definitions.Models{
			Structs: models,
		})

		// If it fails, throw an error
		if err != nil {
			// throw error and print the error too
			Fail("Failed to generate OpenAPI spec: " + err.Error())
		}

		// Compare the generated spec with the expected spec
		areEqual, err := swagtool.AreJSONsIdentical(jsonBytes, formSpec)
		if err != nil {
			Fail("Failed to compare JSONs: " + err.Error())
		}

		// Output the generated JSON and the expected JSON to files to easy testing troubleshooting
		// Create the output directory if it doesn't exist
		if err := os.MkdirAll("dist", os.ModePerm); err != nil {
			logger.Error("Failed to create directory - %v", err)
		}

		var prettyJSON bytes.Buffer
		json.Indent(&prettyJSON, formSpec, "", "    ")

		// Write the JSON to the file
		filePath := "dist/org-form-spec-v3.json"
		if err := os.WriteFile(filePath, prettyJSON.Bytes(), 0644); err != nil {
			logger.Error("Failed to write file - %v", err)
		}

		// Write the JSON to the file
		filePath2 := "dist/test-form-spec-v3.json"
		if err := os.WriteFile(filePath2, jsonBytes, 0644); err != nil {
			logger.Error("Failed to write file - %v", err)
		}
		Expect(areEqual).To(BeTrue())
	})
})
