package swagen31

import (
	"github.com/gopher-fleece/gleece/v2/definitions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

var _ = Describe("Common Utilities", func() {
	Describe("GenerateSecuritySpec", func() {
		It("should generate security specifications correctly", func() {
			doc := &v3.Document{
				Components: &v3.Components{
					Schemas:         orderedmap.New[string, *base.SchemaProxy](),
					SecuritySchemes: orderedmap.New[string, *v3.SecurityScheme](),
				},
			}

			securityConfig := []definitions.SecuritySchemeConfig{
				{
					SecurityName: "apiKeyAuth",
					Type:         "apiKey",
					In:           "header",
					FieldName:    "X-API-Key",
					Description:  "API Key Authentication",
				},
				{
					SecurityName: "basicAuth",
					Type:         "http",
					In:           "header",
					FieldName:    "Authorization",
					Description:  "Basic HTTP Authentication",
				},
			}

			err := GenerateSecuritySpec(doc, &securityConfig)
			Expect(err).To(BeNil())

			apiKeyAuthScheme, _ := doc.Components.SecuritySchemes.Get("apiKeyAuth")
			Expect(apiKeyAuthScheme).ToNot(BeNil())
			Expect(apiKeyAuthScheme.Type).To(Equal("apiKey"))
			Expect(apiKeyAuthScheme.In).To(Equal("header"))
			Expect(apiKeyAuthScheme.Name).To(Equal("X-API-Key"))
			Expect(apiKeyAuthScheme.Description).To(Equal("API Key Authentication"))

			basicAuthScheme, _ := doc.Components.SecuritySchemes.Get("basicAuth")
			Expect(basicAuthScheme).ToNot(BeNil())
			Expect(basicAuthScheme.Type).To(Equal("http"))
			Expect(basicAuthScheme.In).To(Equal("header"))
			Expect(basicAuthScheme.Name).To(Equal("Authorization"))
			Expect(basicAuthScheme.Description).To(Equal("Basic HTTP Authentication"))
		})
	})
})
