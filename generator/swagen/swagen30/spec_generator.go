package swagen30

import (
	"context"
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

// GenerateSpec generates the OpenAPI specification
func GenerateSpec(config *definitions.OpenAPIGeneratorConfig, defs []definitions.ControllerMetadata, models *definitions.Models) ([]byte, error) {

	// Create a new OpenAPI specification using 3.0.0
	openapi := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:          config.Info.Title,
			Description:    config.Info.Description,
			Version:        config.Info.Version,
			TermsOfService: config.Info.TermsOfService,
		},
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

	if config.Info.License != nil {
		openapi.Info.License = &openapi3.License{
			Name: config.Info.License.Name,
			URL:  config.Info.License.URL,
		}
	}

	if config.Info.Contact != nil {
		openapi.Info.Contact = &openapi3.Contact{
			Name:  config.Info.Contact.Name,
			Email: config.Info.Contact.Email,
			URL:   config.Info.Contact.URL,
		}
	}

	if err := GenerateSecuritySpec(openapi, &config.SecuritySchemes); err != nil {
		logger.Error("Failed to generate security spec - %v", err)
		return nil, err
	}
	logger.Info("Security spec generated successfully")

	if err := GenerateModelsSpec(openapi, models); err != nil {
		logger.Error("Failed to generate models spec - %v", err)
		return nil, err
	}
	logger.Info("Models spec generated successfully")

	if err := GenerateControllersSpec(openapi, config, defs); err != nil {
		logger.Error("Failed to generate controllers spec - %v", err)
		return nil, err
	}
	logger.Info("Controllers spec generated successfully")

	// Validate the spec to ensure it meets OpenAPI requirements
	if err := openapi.Validate(context.Background()); err != nil {
		logger.Error("Spec Validation failed - %v", err.Error())
		return nil, err
	}
	logger.Info("OpenAPI specification validated successfully")

	// Convert the spec to JSON with indentation for easy reading
	jsonBytes, err := json.MarshalIndent(openapi, "", "  ")
	if err != nil {
		logger.Error("Marshalling error:", err)
		return nil, err
	}

	logger.Info("OpenAPI specification generation completed successfully")
	return jsonBytes, nil

}
