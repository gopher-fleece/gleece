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

var fullyFeaturesSpec = []byte(`{"openapi":"3.1.0","info":{"title":"My API","description":"This is a simple API?","contact":{"name":"John Doe"},"license":{"name":"Apache 2.0","url":"https://www.apache.org/licenses/LICENSE-2.0.html"},"version":"1.0.0"},"servers":[{"url":"http://localhost:8080"}],"paths":{"/example-base/example-route/{my_path}":{"get":{"tags":["Example"],"summary":"Example route","description":"Example route","operationId":"exampleRoute","parameters":[{"name":"my_name","in":"query","description":"Example query param","required":true,"deprecated":true,"schema":{"type":"string","format":"email"}},{"name":"my_names","in":"query","description":"Example query ARR param","required":true,"schema":{"type":"array","items":{"$ref":"#/components/schemas/ExampleSchema"}}},{"name":"my_header","in":"header","description":"Example Header param","required":true,"schema":{"type":"boolean"}},{"name":"my_number","in":"header","description":"Example Header num param","required":true,"schema":{"exclusiveMaximum":100,"type":"number","minimum":18}},{"name":"my_path","in":"path","description":"Example Path param","required":true,"schema":{"type":"integer","enum":[1,2,3,4]}}],"requestBody":{"description":"Example Body param","content":{"application/json":{"schema":{"type":"string","format":"email"}}},"required":true},"responses":{"200":{"description":" ","content":{"application/json":{"schema":{"type":"array","items":{"$ref":"#/components/schemas/ExampleSchema"}}}}},"500":{"description":"Internal server error","content":{"application/json":{"schema":{"$ref":"#/components/schemas/Rfc7807Error"}}}}},"security":[{"ApiKeyAuth":["read","write"],"ApiKeyAuth2":["write"]},{"ApiKeyAuth":["read"]}]}},"/example-base/example-route":{"post":{"tags":["Example"],"summary":"Example route","description":"Example route","operationId":"exampleRoute45","parameters":[],"responses":{"200":{"description":"Example response OK","content":{"application/json":{"schema":{"type":"integer"}}}},"500":{"description":"Internal server error","content":{"application/json":{"schema":{"type":"string"}}}}},"security":[{"ApiKeyAuth":["read"]}]},"delete":{"tags":["Example"],"summary":"Example route","description":"Example route","operationId":"exampleRouteDel","parameters":[],"responses":{"204":{"description":"Example response OK for 204"},"500":{"description":"Internal server error","content":{"application/json":{"schema":{"type":"string"}}}}},"deprecated":true,"security":[{"ApiKeyAuth":["read"]}]}},"/example-base/post-enum":{"post":{"tags":["Example"],"summary":"Example enum route","description":"Example enum route","operationId":"exampleEnumRoute","parameters":[{"name":"my_enum","in":"query","description":"Example enum num param","required":true,"schema":{"$ref":"#/components/schemas/Status"}}],"requestBody":{"description":"Example Struct with Enum","content":{"application/json":{"schema":{"$ref":"#/components/schemas/ExampleSchemaWithEnum"}}},"required":true},"responses":{"200":{"description":" ","content":{"application/json":{"schema":{"type":"array","items":{"$ref":"#/components/schemas/ExampleSchemaWithEnum"}}}}},"500":{"description":"Internal server error","content":{"application/json":{"schema":{"$ref":"#/components/schemas/Rfc7807Error"}}}}},"security":[{"ApiKeyAuth":["read","write"],"ApiKeyAuth2":["write"]},{"ApiKeyAuth":["read"]}]}},"/example-base/post-alias":{"post":{"tags":["Example"],"summary":"Example alias route","description":"Example alias route","operationId":"exampleAliasRoute","parameters":[{"name":"my_alias","in":"query","description":"Example alias param","required":true,"schema":{"$ref":"#/components/schemas/ExampleAlias"}}],"requestBody":{"description":"Example Struct with Enum","content":{"application/json":{"schema":{"$ref":"#/components/schemas/StrctExampleAlias"}}},"required":true},"responses":{"200":{"description":" ","content":{"application/json":{"schema":{"type":"array","items":{"$ref":"#/components/schemas/StrctExampleAlias"}}}}},"500":{"description":"Internal server error","content":{"application/json":{"schema":{"$ref":"#/components/schemas/Rfc7807Error"}}}}},"security":[{"ApiKeyAuth":["read","write"],"ApiKeyAuth2":["write"]},{"ApiKeyAuth":["read"]}]}}},"components":{"schemas":{"Status":{"type":"string","title":"Status","enum":["ACTIVE","INACTIVE","SUSPENDED"],"description":"User status enum"},"Status2":{"type":"string","title":"Status2","enum":["ACTIVE2","INACTIVE2","SUSPENDED2"],"description":"User status enum"},"ExampleSchema222":{"type":"object","properties":{"MaxValue":{"type":"integer","maximum":100,"minimum":1,"description":"MaxValue DESCRIPTION"},"TheName":{"type":"string","format":"email","description":"TheName DESCRIPTION"}},"title":"ExampleSchema222","required":["TheName"],"description":"Example schema 222","deprecated":true},"ExampleSchemaWithEnum":{"type":"object","properties":{"TheStatus":{"$ref":"#/components/schemas/Status"},"TheStatus2":{"$ref":"#/components/schemas/Status2"}},"title":"ExampleSchemaWithEnum","required":["TheStatus"],"description":"Example enum schema"},"ExampleSchema":{"type":"object","properties":{"ExampleField":{"type":"string","enum":["one","two","three"],"description":"Example field","deprecated":true},"ExampleObjField":{"$ref":"#/components/schemas/ExampleSchema222"},"ExampleArrField":{"type":"array","items":{"$ref":"#/components/schemas/ExampleSchema222"},"description":"Example array field"},"ExampleArrStringField":{"type":"array","items":{"type":"array","items":{"type":"array","items":{"type":"array","items":{"$ref":"#/components/schemas/ExampleSchema222"}}}},"description":"Example int arr field"}},"title":"ExampleSchema","required":["ExampleField","ExampleObjField","ExampleArrField"],"description":"Example schema"},"StrctExampleAlias":{"type":"object","properties":{"AliasField":{"$ref":"#/components/schemas/ExampleAlias"},"IntAliasField":{"$ref":"#/components/schemas/ExampleIntAlias"}},"title":"StrctExampleAlias","required":["AliasField","IntAliasField"],"description":"Example struct with alias field"},"Rfc7807Error":{"type":"object","properties":{"type":{"type":"string","description":"A URI reference that identifies the problem type."},"title":{"type":"string","description":"A short, human-readable summary of the problem type."},"status":{"type":"integer","description":"The HTTP status code generated by the origin server for this occurrence of the problem."},"detail":{"type":"string","description":"A human-readable explanation specific to this occurrence of the problem."},"instance":{"type":"string","description":"A URI reference that identifies the specific occurrence of the problem."},"error":{"type":"string","description":"Error message"},"extensions":{"type":"object","description":"Additional metadata about the error."}},"title":"Rfc7807Error","required":["type","title","status"],"description":"A standard RFC-7807 error"},"ExampleAlias":{"type":"string","title":"ExampleAlias","description":"Example alias"},"ExampleIntAlias":{"type":"integer","title":"ExampleIntAlias","description":"Example int alias","deprecated":true}},"securitySchemes":{"ApiKeyAuth":{"type":"apiKey","description":"API Key","name":"X-API-Key2","in":"header"},"ApiKeyAuth2":{"type":"http","description":"API Key","scheme":"bearer"},"ApiKeyAuth3":{"type":"openIdConnect","description":"API Key","openIdConnectUrl":"https://example.com/auth"},"ApiKeyAuth4":{"type":"oauth2","description":"API Key","flows":{"implicit":{"authorizationUrl":"https://example.com/auth","refreshUrl":"https://example.com/refresh","scopes":{"read":"Read access","write":"Write access"}},"password":{"tokenUrl":"https://example.com/token","refreshUrl":"https://example.com/refresh","scopes":{"read":"Read access","write":"Write access"}},"clientCredentials":{"tokenUrl":"https://example.com/token","refreshUrl":"https://example.com/refresh","scopes":{"write":"Write access","read":"Read access"}},"authorizationCode":{"authorizationUrl":"https://example.com/auth","tokenUrl":"https://example.com/token","refreshUrl":"https://example.com/refresh","scopes":{"read":"Read access","write":"Write access"}}}}}}}`)
var formSpec = []byte(`{"components":{"schemas":{"Rfc7807Error":{"description":"A standard RFC-7807 error","properties":{"detail":{"description":"A human-readable explanation specific to this occurrence of the problem.","type":"string"},"error":{"description":"Error message","type":"string"},"extensions":{"description":"Additional metadata about the error.","type":"object"},"instance":{"description":"A URI reference that identifies the specific occurrence of the problem.","type":"string"},"status":{"description":"The HTTP status code generated by the origin server for this occurrence of the problem.","type":"integer"},"title":{"description":"A short, human-readable summary of the problem type.","type":"string"},"type":{"description":"A URI reference that identifies the problem type.","type":"string"}},"required":["type","title","status"],"title":"Rfc7807Error","type":"object"}},"securitySchemes":{"ApiKeyAuth":{"description":"API Key","in":"header","name":"X-API-Key2","type":"apiKey"}}},"info":{"contact":{"name":"John Doe"},"description":"This is a simple API?","license":{"name":"Apache 2.0","url":"https://www.apache.org/licenses/LICENSE-2.0.html"},"title":"My API","version":"1.0.0"},"openapi":"3.1.0","paths":{"/example-base/example-route":{"post":{"description":"Example form route","operationId":"exampleRoute","parameters":[],"requestBody":{"content":{"application/x-www-form-urlencoded":{"schema":{"properties":{"my_form":{"description":"Example my_form param","type":"string"},"my_form_number":{"description":"Example my_form_number param","exclusiveMaximum":100,"minimum":1,"type":"integer"},"my_form_option":{"description":"Example Header num param","type":"boolean"}},"required":["my_form","my_form_number"],"type":"object"}}}},"responses":{"200":{"description":" "},"500":{"content":{"application/json":{"schema":{"$ref":"#/components/schemas/Rfc7807Error"}}},"description":"Internal server error"}},"security":[{"ApiKeyAuth":["read"]}],"summary":"Example form route","tags":["Example"]}}},"servers":[{"url":"http://localhost:8080"}]}`)
var allOfSpecV31 = []byte(`{"openapi":"3.1.0","info":{"title":"AllOf API","description":"API with schema composition using allOf","contact":{"name":"API Support"},"license":{"name":"MIT"},"version":"1.0.0"},"servers":[{"url":"http://localhost:8080"}],"paths":{},"components":{"schemas":{"BaseModel":{"type":"object","properties":{"id":{"type":"string","description":"Unique identifier"},"created_at":{"type":"integer","description":"Creation timestamp"}},"title":"BaseModel","required":["id"],"description":"Base model with common fields"},"TaggableModel":{"type":"object","properties":{"tags":{"type":"array","items":{"type":"string"},"description":"Tags for categorization"}},"title":"TaggableModel","required":[],"description":"Model with tagging functionality"},"ChildModel":{"allOf":[{"type":"object","properties":{"name":{"type":"string","description":"Child name field"}},"title":"ChildModel","required":["name"],"description":"Child model that extends BaseModel"},{"$ref":"#/components/schemas/BaseModel"}]},"CompositeModel":{"allOf":[{"type":"object","properties":{"description":{"type":"string","description":"Model description"}},"title":"CompositeModel","required":[],"description":"Model with multiple embedded types"},{"$ref":"#/components/schemas/BaseModel"},{"$ref":"#/components/schemas/TaggableModel"}]},"Rfc7807Error":{"type":"object","properties":{"type":{"type":"string","description":"A URI reference that identifies the problem type."},"title":{"type":"string","description":"A short, human-readable summary of the problem type."},"status":{"type":"integer","description":"The HTTP status code generated by the origin server for this occurrence of the problem."},"detail":{"type":"string","description":"A human-readable explanation specific to this occurrence of the problem."},"instance":{"type":"string","description":"A URI reference that identifies the specific occurrence of the problem."},"error":{"type":"string","description":"Error message"},"extensions":{"type":"object","description":"Additional metadata about the error."}},"title":"Rfc7807Error","required":["type","title","status"],"description":"A standard RFC-7807 error"}}}}`)

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

		aliases := []definitions.NakedAliasMetadata{{
			Name:        "ExampleAlias",
			Type:        "string",
			Description: "Example alias",
		}, {
			Name:        "ExampleIntAlias",
			Type:        "int",
			Description: "Example int alias",
			Deprecation: definitions.DeprecationOptions{
				Description: "This alias is deprecated example",
				Deprecated:  true,
			},
		}}

		// Create an example controller and add it to the definitions
		defs = append(defs, definitions.ControllerMetadata{
			Tag: "Example",
			RestMetadata: definitions.RestMetadata{
				Path: "/example-base",
			},
			Description: "Example controller",
			Name:        "ExampleController",
			PkgPath:     "github.com/gopher-fleece/gleece/definitions",
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
								Name:    "[]ExampleSchema",
								PkgPath: "example",
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
							Deprecation: definitions.DeprecationOptions{
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
								Name:    "int",
								PkgPath: "example",
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
								Name:    "[]ExampleSchemaWithEnum",
								PkgPath: "example",
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
						Path: "/post-alias",
					},
					Description: "Example alias route",
					OperationId: "exampleAliasRoute",
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
								Name:    "[]StrctExampleAlias",
								PkgPath: "example",
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
									Name: "ExampleAlias",
								},
							},
							NameInSchema: "my_alias",
							PassedIn:     definitions.PassedInQuery,
							Description:  "Example alias param",
							Validator:    "required",
						},
						{
							ParamMeta: definitions.ParamMeta{
								Name: "the_content",
								TypeMeta: definitions.TypeMetadata{
									Name: "StrctExampleAlias",
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
			{
				Name:        "StrctExampleAlias",
				Description: "Example struct with alias field",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "AliasField",
						Type:        "ExampleAlias",
						Description: "Example alias field",
						Tag:         `validate:"required"`,
					},
					{
						Name:        "IntAliasField",
						Type:        "ExampleIntAlias",
						Description: "Example int alias field",
						Tag:         `validate:"required,gte=1,lte=100"`,
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
					Scheme:       definitions.HttpAuthSchemeBearer,
					Type:         definitions.HTTP,
					Description:  "API Key",
				},
				{
					SecurityName:     "ApiKeyAuth3",
					Type:             definitions.OpenIDConnect,
					OpenIdConnectUrl: "https://example.com/auth",
					Description:      "API Key",
				},
				{
					SecurityName: "ApiKeyAuth4",
					Type:         definitions.OAuth2,
					Description:  "API Key",
					Flows: &definitions.OAuthFlows{
						Implicit: &definitions.OAuthFlow{
							AuthorizationURL: "https://example.com/auth",
							RefreshURL:       "https://example.com/refresh",
							Scopes:           map[string]string{"read": "Read access", "write": "Write access"},
						},
						Password: &definitions.OAuthFlow{
							RefreshURL: "https://example.com/refresh",
							Scopes:     map[string]string{"read": "Read access", "write": "Write access"},
							TokenURL:   "https://example.com/token",
						},
						ClientCredentials: &definitions.OAuthFlow{
							RefreshURL: "https://example.com/refresh",
							Scopes:     map[string]string{"read": "Read access", "write": "Write access"},
							TokenURL:   "https://example.com/token",
						},
						AuthorizationCode: &definitions.OAuthFlow{
							AuthorizationURL: "https://example.com/auth",
							RefreshURL:       "https://example.com/refresh",
							Scopes:           map[string]string{"read": "Read access", "write": "Write access"},
							TokenURL:         "https://example.com/token",
						},
					},
				},
			},
			DefaultRouteSecurity: &definitions.SecurityAnnotationComponent{
				SchemaName: "ApiKeyAuth",
				Scopes:     []string{"read"},
			},
		}, defs, &definitions.Models{
			Structs: structs,
			Enums:   enums,
			Aliases: aliases,
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
			PkgPath:     "github.com/gopher-fleece/gleece/definitions",
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
								Name:    "[]string",
								PkgPath: "example",
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
			PkgPath:     "github.com/gopher-fleece/gleece/definitions",
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
								Name:    "[]string",
								PkgPath: "example",
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
			PkgPath:     "github.com/gopher-fleece/gleece/definitions",
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

	It("Should generate schemas with allOf for embedded fields in OpenAPI 3.1", func() {
		// Create models with embedded fields
		structs := []definitions.StructMetadata{
			{
				Name:        "BaseModel",
				Description: "Base model with common fields",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "ID",
						Type:        "string",
						Description: "Unique identifier",
						Tag:         `json:"id" validate:"required"`,
					},
					{
						Name:        "CreatedAt",
						Type:        "int",
						Description: "Creation timestamp",
						Tag:         `json:"created_at"`,
					},
				},
			},
			{
				Name:        "TaggableModel",
				Description: "Model with tagging functionality",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "Tags",
						Type:        "[]string",
						Description: "Tags for categorization",
						Tag:         `json:"tags"`,
					},
				},
			},
			{
				Name:        "ChildModel",
				Description: "Child model that extends BaseModel",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "BaseModel",
						Type:        "BaseModel",
						Description: "",
						IsEmbedded:  true,
						Tag:         ``,
					},
					{
						Name:        "Name",
						Type:        "string",
						Description: "Child name field",
						Tag:         `json:"name" validate:"required"`,
					},
				},
			},
			{
				Name:        "CompositeModel",
				Description: "Model with multiple embedded types",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "BaseModel",
						Type:        "BaseModel",
						Description: "",
						IsEmbedded:  true,
						Tag:         ``,
					},
					{
						Name:        "TaggableModel",
						Type:        "TaggableModel",
						Description: "",
						IsEmbedded:  true,
						Tag:         ``,
					},
					{
						Name:        "Description",
						Type:        "string",
						Description: "Model description",
						Tag:         `json:"description"`,
					},
				},
			},
		}

		swagtool.AppendErrorSchema(&structs, true)

		// Generate OpenAPI spec with empty controllers (we're only testing schemas)
		jsonBytes, err := GenerateSpec(&definitions.OpenAPIGeneratorConfig{
			Info: definitions.OpenAPIInfo{
				Title:       "AllOf API",
				Version:     "1.0.0",
				Description: "API with schema composition using allOf",
				Contact: &definitions.OpenAPIContact{
					Name: "API Support",
				},
				License: &definitions.OpenAPILicense{
					Name: "MIT",
				},
			},
			BaseURL: "http://localhost:8080",
		}, []definitions.ControllerMetadata{}, &definitions.Models{
			Structs: structs,
		})

		// If it fails, throw an error
		if err != nil {
			Fail("Failed to generate OpenAPI spec: " + err.Error())
		}

		// Write the generated spec to a file for debugging
		if err := os.MkdirAll("dist", os.ModePerm); err != nil {
			logger.Error("Failed to create directory - %v", err)
		}

		filePath := "dist/allof_generated_spec_v31.json"
		if err := os.WriteFile(filePath, jsonBytes, 0644); err != nil {
			logger.Error("Failed to write file - %v", err)
		}

		var prettyExpected bytes.Buffer
		json.Indent(&prettyExpected, allOfSpecV31, "", "  ")

		filePath2 := "dist/allof_expected_spec_v31.json"
		if err := os.WriteFile(filePath2, prettyExpected.Bytes(), 0644); err != nil {
			logger.Error("Failed to write file - %v", err)
		}

		// Parse both JSONs to compare schemas structure
		var generated map[string]interface{}
		var expected map[string]interface{}

		err = json.Unmarshal(jsonBytes, &generated)
		Expect(err).To(BeNil())

		err = json.Unmarshal(allOfSpecV31, &expected)
		Expect(err).To(BeNil())

		// Extract schemas for comparison
		generatedSchemas := generated["components"].(map[string]interface{})["schemas"].(map[string]interface{})

		// Check if ChildModel has allOf
		childModel := generatedSchemas["ChildModel"].(map[string]interface{})
		Expect(childModel).To(HaveKey("allOf"))
		allOf := childModel["allOf"].([]interface{})
		Expect(allOf).To(HaveLen(2))

		// First element should be an object with title "ChildModel"
		firstElem := allOf[0].(map[string]interface{})
		Expect(firstElem).To(HaveKey("title"))
		Expect(firstElem["title"]).To(Equal("ChildModel"))

		// Second element should be a reference to BaseModel
		secondElem := allOf[1].(map[string]interface{})
		Expect(secondElem).To(HaveKey("$ref"))
		Expect(secondElem["$ref"]).To(Equal("#/components/schemas/BaseModel"))

		// Check if CompositeModel has allOf with multiple schemas
		compositeModel := generatedSchemas["CompositeModel"].(map[string]interface{})
		Expect(compositeModel).To(HaveKey("allOf"))
		compositeAllOf := compositeModel["allOf"].([]interface{})
		Expect(compositeAllOf).To(HaveLen(3)) // Own properties + 2 embedded models

		// Collect references from allOf
		var refs []string
		for i := 1; i < len(compositeAllOf); i++ {
			elem := compositeAllOf[i].(map[string]interface{})
			if ref, ok := elem["$ref"].(string); ok {
				refs = append(refs, ref)
			}
		}

		// Check that both embedded models are referenced
		Expect(refs).To(ContainElement("#/components/schemas/BaseModel"))
		Expect(refs).To(ContainElement("#/components/schemas/TaggableModel"))

		// BaseModel should NOT have allOf since it doesn't have embedded fields
		baseModel := generatedSchemas["BaseModel"].(map[string]interface{})
		Expect(baseModel).NotTo(HaveKey("allOf"))

		// Overall, the schema structure should match our expectations
		areEqual, err := swagtool.AreJSONsIdentical(jsonBytes, allOfSpecV31)
		Expect(err).To(BeNil())
		Expect(areEqual).To(BeTrue())
	})
})
