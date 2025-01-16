package swagen

import (
	"context"
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/haimkastner/gleece/definitions"
	"github.com/haimkastner/gleece/infrastructure/logger"
)

type SecuritySchemeConfig struct {
	Description  string                         `json:"description"`
	SecurityName string                         `json:"name"`
	FieldName    string                         `json:"fieldName"`
	Type         definitions.SecuritySchemeType `json:"type"`
	In           definitions.SecuritySchemeIn   `json:"in"`
}

type OpenAPIGeneratorConfig struct {
	Info                 openapi3.Info               `json:"info"`
	BaseURL              string                      `json:"base_url"`
	SecuritySchemes      []SecuritySchemeConfig      `json:"securitySchemes"`
	DefaultRouteSecurity []definitions.RouteSecurity `json:"defaultSecurity"`
}

// GenerateSpec generates the OpenAPI specification
func GenerateSpec(config OpenAPIGeneratorConfig, defs []definitions.ControllerMetadata, models []definitions.ModelMetadata) ([]byte, error) {

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

	if err := GenerateSecuritySpec(openapi, &config.SecuritySchemes); err != nil {
		logger.Error("Failed to generate security spec:", err)
		return nil, err
	}
	logger.Info("Security spec generated successfully")

	if err := GenerateModelsSpec(openapi, models); err != nil {
		logger.Error("Failed to generate models spec:", err)
		return nil, err
	}
	logger.Info("Models spec generated successfully")

	if err := GenerateControllersSpec(openapi, &config, defs); err != nil {
		logger.Error("Failed to generate controllers spec:", err)
		return nil, err
	}
	logger.Info("Controllers spec generated successfully")

	// Validate the spec to ensure it meets OpenAPI requirements
	if err := openapi.Validate(context.Background()); err != nil {
		logger.Error("Validation failed:", err.Error())
		return nil, err
	}
	logger.Info("OpenAPI specification validated successfully")

	// Convert the spec to JSON with indentation for easy reading
	jsonBytes, err := json.MarshalIndent(openapi, "", "  ")
	if err != nil {
		logger.Error("Marshalling error:", err)
		return nil, err
	}

	logger.Info("OpenAPI specification generated successfully")
	return jsonBytes, nil

	// // Create the output directory if it doesn't exist
	// if err := os.MkdirAll("dist", os.ModePerm); err != nil {
	// 	logger.Error("Failed to create directory:", err)
	// 	return nil, err
	// }

	// // Write the JSON to the file
	// filePath := "dist/spec.json"
	// if err := os.WriteFile(filePath, jsonBytes, 0644); err != nil {
	// 	fmt.Println("Failed to write file:", err)
	// 	return nil, err
	// }

	// // Print the path to the generated JSON file
	// fmt.Println("OpenAPI specification written to:", filePath)

}
