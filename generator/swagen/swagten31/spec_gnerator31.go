package swagten31

import (
	"fmt"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

func GenerateSpec(config *definitions.OpenAPIGeneratorConfig, defs []definitions.ControllerMetadata, models []definitions.ModelMetadata) ([]byte, error) {
	// Create an OpenAPI 3.1 document
	doc := &v3.Document{
		Version: "3.1.0",
		Servers: []*v3.Server{
			{
				URL: config.BaseURL,
				// Description: , TODO: sully support it
			},
		},
		Info: &base.Info{
			// Summary:        config.Info.Title,
			Description:    config.Info.Description,
			TermsOfService: config.Info.TermsOfService,
			Title:          config.Info.Title,
			Version:        config.Info.Version,
			// TODO: support extensions
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
		logger.Error("Failed to generate security spec - %v", err)
		return nil, err
	}
	logger.Info("Security spec generated successfully")

	if err := GenerateModelsSpec(doc, models); err != nil {
		logger.Error("Failed to generate models spec - %v", err)
		return nil, err
	}
	logger.Info("Models spec generated successfully")

	jsonData, err := doc.RenderJSON("    ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
	}

	return jsonData, err
}
