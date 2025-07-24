package swagen

import (
	"os"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/generator/swagen/swagtool"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Spec Preparing", func() {

	It("Should output v3.0 spec file", func() {

		defs := []definitions.ControllerMetadata{}

		// Create an example controller and add it to the definitions
		defs = append(defs, definitions.ControllerMetadata{})

		models := []definitions.StructMetadata{}

		outputPath := "./dist/test-spec-pre-out.json"
		if swagtool.FileExists(outputPath) {
			os.Remove(outputPath)
		}

		GenerateAndOutputSpec(&definitions.OpenAPIGeneratorConfig{
			OpenAPI: "3.0.0",
			Info: definitions.OpenAPIInfo{
				Title:   "My API",
				Version: "1.0.0",
			},
			BaseURL: "http://localhost:8080",
			SpecGeneratorConfig: definitions.SpecGeneratorConfig{
				OutputPath: outputPath,
			},
		}, defs, &definitions.Models{
			Structs: models,
		}, false)

		Expect(swagtool.FileExists(outputPath)).To(BeTrue())
	})

	It("Should output v3.1 spec file", func() {

		defs := []definitions.ControllerMetadata{}

		// Create an example controller and add it to the definitions
		defs = append(defs, definitions.ControllerMetadata{})

		models := []definitions.StructMetadata{}

		outputPath := "./dist/test-spec-pre-v31-out.json"
		if swagtool.FileExists(outputPath) {
			os.Remove(outputPath)
		}

		GenerateAndOutputSpec(&definitions.OpenAPIGeneratorConfig{
			OpenAPI: "3.1.0",
			Info: definitions.OpenAPIInfo{
				Title:   "My API",
				Version: "1.0.0",
			},
			BaseURL: "http://localhost:8080",
			SpecGeneratorConfig: definitions.SpecGeneratorConfig{
				OutputPath: outputPath,
			},
		}, defs, &definitions.Models{
			Structs: models,
		}, false)

		Expect(swagtool.FileExists(outputPath)).To(BeTrue())
	})

	It("Should throw error for invalid 3.1 spec using 3.0 validation", func() {

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
						Path: "/example-route/{id}",
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

		swagtool.AppendErrorSchema(&models, true)

		_, err := GenerateSpec(&definitions.OpenAPIGeneratorConfig{
			OpenAPI: "3.1.0",
			Info: definitions.OpenAPIInfo{
				Title:   "My API",
				Version: "1.0.0",
			},
			BaseURL: "http://localhost:8080",
		}, defs, &definitions.Models{
			Structs: models,
		}, true)

		// Expect error not to be nil
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("invalid paths: operation GET /example-base/example-route/{id} must define exactly all path parameters (missing: [id])"))
	})
})
