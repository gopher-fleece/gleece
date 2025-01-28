package swagen30

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gopher-fleece/gleece/definitions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Swagen", func() {
	var openapi *openapi3.T
	var config *definitions.OpenAPIGeneratorConfig

	BeforeEach(func() {
		openapi = &openapi3.T{
			Components: &openapi3.Components{
				Schemas: openapi3.Schemas{},
			},
			Paths: openapi3.NewPaths(),
		}
		config = &definitions.OpenAPIGeneratorConfig{
			SecuritySchemes: []definitions.SecuritySchemeConfig{
				{
					SecurityName: "apiKeyAuth",
					Type:         "apiKey",
					In:           "header",
					FieldName:    "X-API-Key",
					Description:  "API Key Authentication",
				},
			},
			DefaultRouteSecurity: []definitions.RouteSecurity{
				{
					SecurityAnnotation: []definitions.SecurityAnnotationComponent{
						{
							SchemaName: "apiKeyAuth",
							Scopes:     []string{},
						},
					},
				},
			},
		}
	})

	Describe("createOperation", func() {
		It("should create an operation with correct metadata", func() {
			def := definitions.ControllerMetadata{Tag: "test"}
			route := definitions.RouteMetadata{
				Description: "Test route",
				OperationId: "testOperation",
			}
			operation := createOperation(def, route)

			Expect(operation.Summary).To(Equal("Test route"))
			Expect(operation.OperationID).To(Equal("testOperation"))
			Expect(operation.Tags).To(ConsistOf("test"))
		})
	})

	Describe("createErrorResponse", func() {
		It("should create an error response", func() {
			route := definitions.RouteMetadata{
				ResponseDescription: "Success1",
				Responses: []definitions.FuncReturnValue{
					{
						TypeMetadata: definitions.TypeMetadata{
							Name:        "int",
							Description: "Bla bla",
						},
					},
				},
				ResponseSuccessCode: 200,
			}
			errResp := definitions.ErrorResponse{
				Description:    "Error occurred",
				HttpStatusCode: 500,
			}
			responseRef := createErrorResponse(openapi, route, errResp)
			Expect(*responseRef.Value.Description).To(Equal("Error occurred"))
			Expect(responseRef.Value.Content).To(Equal(openapi3.NewContentWithJSONSchema(openapi3.NewIntegerSchema())))
		})
	})

	Describe("createResponseSuccess", func() {
		It("should create a success response", func() {
			route := definitions.RouteMetadata{
				ResponseDescription: "Success1",
				Responses: []definitions.FuncReturnValue{
					{
						TypeMetadata: definitions.TypeMetadata{
							Name:        "int",
							Description: "Bla bla",
						},
					},
					{
						TypeMetadata: definitions.TypeMetadata{
							Name: "error",
						},
					},
				},
				ResponseSuccessCode: 200,
			}
			responseRef := createResponseSuccess(openapi, route)

			Expect(*responseRef.Value.Description).To(Equal("Success1"))
			Expect(responseRef.Value.Content).To(Equal(openapi3.NewContentWithJSONSchemaRef(ToOpenApiSchemaRef("integer"))))
		})
	})

	Describe("buildSecurityMethod", func() {
		It("should build security requirement", func() {
			securityMethods := []definitions.SecurityAnnotationComponent{
				{SchemaName: "apiKeyAuth", Scopes: []string{}},
			}
			securityRequirement, err := buildSecurityMethod(config.SecuritySchemes, securityMethods)

			Expect(err).To(BeNil())
			Expect(*securityRequirement).To(HaveKey("apiKeyAuth"))
		})

		It("should return an error if security method name is not found", func() {
			securityMethods := []definitions.SecurityAnnotationComponent{
				{SchemaName: "unknownAuth", Scopes: []string{}},
			}
			_, err := buildSecurityMethod(config.SecuritySchemes, securityMethods)

			Expect(err).To(HaveOccurred())
		})

		It("should return an error if security operation is not valid", func() {
			route := definitions.RouteMetadata{
				OperationId: "testOperation",
				Security: []definitions.RouteSecurity{{
					SecurityAnnotation: []definitions.SecurityAnnotationComponent{{SchemaName: "unknownAuth"}}}},
			}
			operation := &openapi3.Operation{}

			err := generateOperationSecurity(operation, config, route)

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("generateOperationSecurity", func() {
		It("should add security requirements to the operation", func() {
			route := definitions.RouteMetadata{
				OperationId: "testOperation",
				Security:    []definitions.RouteSecurity{{SecurityAnnotation: []definitions.SecurityAnnotationComponent{{SchemaName: "apiKeyAuth"}}}},
			}
			operation := &openapi3.Operation{}
			err := generateOperationSecurity(operation, config, route)

			Expect(err).To(BeNil())
			Expect(operation.Security).NotTo(BeNil())
			Expect((*operation.Security)[0]).To(HaveKey("apiKeyAuth"))
		})
	})

	Describe("setNewRouteOperation", func() {
		It("should set the operation to the correct path", func() {
			def := definitions.ControllerMetadata{RestMetadata: definitions.RestMetadata{Path: "/api"}}
			route := definitions.RouteMetadata{RestMetadata: definitions.RestMetadata{Path: "/test"}, HttpVerb: "GET"}
			operation := &openapi3.Operation{Summary: "Test operation"}

			setNewRouteOperation(openapi, def, route, operation)

			pathItem := openapi.Paths.Value("/api/test")
			Expect(pathItem.Get).To(Equal(operation))
		})
	})

	Describe("generateControllerSpec", func() {
		It("should generate specifications for a controller", func() {
			def := definitions.ControllerMetadata{
				Tag: "test",
				Routes: []definitions.RouteMetadata{
					{
						HttpVerb:            "GET",
						Description:         "Test route",
						OperationId:         "testOperation",
						RestMetadata:        definitions.RestMetadata{Path: "/test"},
						ResponseSuccessCode: 200,
						ResponseDescription: "Success",
						Responses: []definitions.FuncReturnValue{
							{
								TypeMetadata: definitions.TypeMetadata{
									Name: "string",
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
			}
			err := generateControllerSpec(openapi, config, def)
			Expect(err).To(BeNil())

			pathItem := openapi.Paths.Value("/test")
			Expect(pathItem).NotTo(BeNil())
		})

		It("should abort generate specifications due to error", func() {
			def := definitions.ControllerMetadata{
				Tag: "test",
				Routes: []definitions.RouteMetadata{
					{
						HttpVerb:            "GET",
						Description:         "Test route",
						OperationId:         "testOperation",
						RestMetadata:        definitions.RestMetadata{Path: "/test"},
						ResponseSuccessCode: 200,
						ResponseDescription: "Success",
						Responses: []definitions.FuncReturnValue{
							{
								TypeMetadata: definitions.TypeMetadata{
									Name: "string",
								},
							},
							{
								TypeMetadata: definitions.TypeMetadata{
									Name: "error",
								},
							},
						},
						FuncParams: []definitions.FuncParam{},
						Security: []definitions.RouteSecurity{
							{
								SecurityAnnotation: []definitions.SecurityAnnotationComponent{
									{SchemaName: "unknownAuth"},
								},
							},
						},
					},
				},
			}
			err := generateControllerSpec(openapi, config, def)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("generateControllersSpec", func() {
		It("should generate specifications for controllers", func() {
			defs := []definitions.ControllerMetadata{
				{
					Tag: "test",
					Routes: []definitions.RouteMetadata{
						{
							HttpVerb:            "GET",
							Description:         "Test route",
							OperationId:         "testOperation",
							RestMetadata:        definitions.RestMetadata{Path: "/test"},
							ResponseSuccessCode: 200,
							ResponseDescription: "Success",
							Responses: []definitions.FuncReturnValue{
								{
									TypeMetadata: definitions.TypeMetadata{
										Name: "string",
									},
								},
								{
									TypeMetadata: definitions.TypeMetadata{
										Name: "error",
									},
								},
							},
						},
					},
				},
			}
			err := GenerateControllersSpec(openapi, config, defs)
			Expect(err).To(BeNil())

			pathItem := openapi.Paths.Value("/test")
			Expect(pathItem.Get).NotTo(BeNil())
		})

		It("should abort generate specifications due to error", func() {
			defs := []definitions.ControllerMetadata{
				{
					Tag: "test",
					Routes: []definitions.RouteMetadata{
						{
							HttpVerb:            "GET",
							Description:         "Test route",
							OperationId:         "testOperation",
							RestMetadata:        definitions.RestMetadata{Path: "/test"},
							ResponseSuccessCode: 200,
							ResponseDescription: "Success",
							Responses: []definitions.FuncReturnValue{
								{
									TypeMetadata: definitions.TypeMetadata{
										Name: "string",
									},
								},
								{
									TypeMetadata: definitions.TypeMetadata{
										Name: "error",
									},
								},
							},
							Security: []definitions.RouteSecurity{
								{
									SecurityAnnotation: []definitions.SecurityAnnotationComponent{
										{SchemaName: "unknownAuth"},
									},
								},
							},
						},
					},
				},
			}
			err := GenerateControllersSpec(openapi, config, defs)
			Expect(err).To(HaveOccurred())
		})
	})
})
