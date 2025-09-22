package swagen31

import (
	"fmt"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

func GenerateSpec(config *definitions.OpenAPIGeneratorConfig, defs []definitions.ControllerMetadata, models *definitions.Models) ([]byte, error) {
	// Create an OpenAPI 3.1 document
	doc := &v3.Document{
		Version: "3.1.0",
		Servers: []*v3.Server{
			{
				URL: config.BaseURL,
			},
		},
		Info: &base.Info{
			Title:          config.Info.Title,
			Description:    config.Info.Description,
			TermsOfService: config.Info.TermsOfService,
			Version:        config.Info.Version,
		},
		Paths: &v3.Paths{
			PathItems: orderedmap.New[string, *v3.PathItem](),
		},
		Components: &v3.Components{
			Schemas:         orderedmap.New[string, *base.SchemaProxy](),
			SecuritySchemes: orderedmap.New[string, *v3.SecurityScheme](),
		},
	}

	if config.Info.License != nil {
		doc.Info.License = &base.License{
			Name: config.Info.License.Name,
			URL:  config.Info.License.URL,
		}
	}

	if config.Info.Contact != nil {
		doc.Info.Contact = &base.Contact{
			Name:  config.Info.Contact.Name,
			Email: config.Info.Contact.Email,
			URL:   config.Info.Contact.URL,
		}
	}

	if err := GenerateSecuritySpec(doc, &config.SecuritySchemes); err != nil {
		logger.Error("Failed to generate security v3.1 spec - %v", err)
		return nil, err
	}
	logger.Info("Security spec v3.1 generated successfully")

	if err := GenerateModelsSpec(doc, models); err != nil {
		logger.Error("Failed to generate models v3.1 spec - %v", err)
		return nil, err
	}
	logger.Info("Models spec v3.1 generated successfully")

	if err := GenerateControllersSpec(doc, config, defs); err != nil {
		logger.Error("Failed to generate controllers v3.1 spec - %v", err)
		return nil, err
	}
	logger.Info("Controllers spec v3.1 generated successfully")

	jsonData, err := doc.RenderJSON("    ")
	if err != nil {
		fmt.Println("Error marshaling v3.1 JSON:", err)
	}

	libopenapiDoc, docErr := libopenapi.NewDocument(jsonData)

	if docErr != nil {
		return nil, fmt.Errorf("Failed to build a valid libopenapi document %v", docErr.Error())
	}

	specValidator, validatorErrs := validator.NewValidator(libopenapiDoc)
	if validatorErrs != nil {
		validationText := FormatErrors(validatorErrs)
		return nil, fmt.Errorf("Failed to build a valid v3.1 specification %v", validationText)
	}

	succeeded, docValidatorErrs := specValidator.ValidateDocument()

	if !succeeded {
		docValidationText := FormatValidationErrors(docValidatorErrs)
		return nil, fmt.Errorf("Failed to build a valid v3.1 documentation specification %v", docValidationText)
	}
	logger.Info("OpenAPI specification validated successfully")

	return jsonData, err
}
