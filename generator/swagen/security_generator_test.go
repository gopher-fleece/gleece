package swagen

import (
	"github.com/getkin/kin-openapi/openapi3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Common Utilities", func() {
	Describe("GenerateSecuritySpec", func() {
		It("should generate security specifications correctly", func() {
			openapi := &openapi3.T{
				Components: &openapi3.Components{
					Schemas: openapi3.Schemas{},
				},
			}

			securityConfig := []SecuritySchemeConfig{
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

			err := GenerateSecuritySpec(openapi, &securityConfig)
			Expect(err).To(BeNil())
			Expect(openapi.Components.SecuritySchemes).To(HaveKey("apiKeyAuth"))
			Expect(openapi.Components.SecuritySchemes["apiKeyAuth"].Value.Type).To(Equal("apiKey"))
			Expect(openapi.Components.SecuritySchemes["apiKeyAuth"].Value.In).To(Equal("header"))
			Expect(openapi.Components.SecuritySchemes["apiKeyAuth"].Value.Name).To(Equal("X-API-Key"))
			Expect(openapi.Components.SecuritySchemes["apiKeyAuth"].Value.Description).To(Equal("API Key Authentication"))

			Expect(openapi.Components.SecuritySchemes).To(HaveKey("basicAuth"))
			Expect(openapi.Components.SecuritySchemes["basicAuth"].Value.Type).To(Equal("http"))
			Expect(openapi.Components.SecuritySchemes["basicAuth"].Value.In).To(Equal("header"))
			Expect(openapi.Components.SecuritySchemes["basicAuth"].Value.Name).To(Equal("Authorization"))
			Expect(openapi.Components.SecuritySchemes["basicAuth"].Value.Description).To(Equal("Basic HTTP Authentication"))
		})
	})
})
