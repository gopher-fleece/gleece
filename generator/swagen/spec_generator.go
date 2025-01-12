package swagen

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/haimkastner/gleece/definitions"
)

type OpenAPIGeneratorConfig struct {
	Info    openapi3.Info
	BaseURL string
}

// GenerateSpec generates the OpenAPI specification
func GenerateSpec(config OpenAPIGeneratorConfig, defs []definitions.ControllerMetadata, models []definitions.ModelMetadata) {

	// Create a new OpenAPI specification using 3.0.0
	openapi := &openapi3.T{
		OpenAPI: "3.0.0",
		Info:    &config.Info,
		Servers: openapi3.Servers{
			{
				URL: config.BaseURL,
			},
		},
		Paths: openapi3.NewPaths(),
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{},
		},
	}

	GenerateModelsSpec(openapi, models)

	// Run on each route in definitions
	GenerateControllersSpec(openapi, defs)

	// Validate the spec to ensure it meets OpenAPI requirements
	if err := openapi.Validate(context.Background()); err != nil {
		fmt.Println("Validation failed:", err)
		// return
	}

	// Convert the spec to JSON with indentation for easy reading
	jsonBytes, err := json.MarshalIndent(openapi, "", "  ")
	if err != nil {
		fmt.Println("Marshalling error:", err)
		return
	}

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll("dist", os.ModePerm); err != nil {
		fmt.Println("Failed to create directory:", err)
		return
	}

	// Write the JSON to the file
	filePath := "dist/spec.json"
	if err := os.WriteFile(filePath, jsonBytes, 0644); err != nil {
		fmt.Println("Failed to write file:", err)
		return
	}

	// Print the path to the generated JSON file
	fmt.Println("OpenAPI specification written to:", filePath)

}
