package swagen

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/haimkastner/gleece/definitions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Swagen", func() {
	var openapi *openapi3.T

	BeforeEach(func() {
		openapi = &openapi3.T{
			Components: &openapi3.Components{
				Schemas: openapi3.Schemas{},
			},
		}
	})

	Describe("generateModelSpec", func() {
		It("should generate a model specification correctly", func() {
			model := definitions.ModelMetadata{
				Name:        "TestModel",
				Description: "A test model",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "field1",
						Type:        "string",
						Description: "A string field",
						Validator:   "required",
					},
					{
						Name:        "field2",
						Type:        "int",
						Description: "An integer field",
						Validator:   "",
					},
				},
			}

			generateModelSpec(openapi, model)

			schemaRef := openapi.Components.Schemas["TestModel"]
			Expect(schemaRef).NotTo(BeNil())
			Expect(schemaRef.Value.Title).To(Equal("TestModel"))
			Expect(schemaRef.Value.Description).To(Equal("A test model"))
			Expect(schemaRef.Value.Type).To(Equal(objectType))
			Expect(schemaRef.Value.Properties).To(HaveKey("field1"))
			Expect(schemaRef.Value.Properties).To(HaveKey("field2"))
			Expect(schemaRef.Value.Required).To(ContainElement("field1"))
			Expect(schemaRef.Value.Required).NotTo(ContainElement("field2"))
			Expect(schemaRef.Value.Properties["field1"].Value.Description).To(Equal("A string field"))
			Expect(schemaRef.Value.Properties["field2"].Value.Description).To(Equal("An integer field"))
		})

		It("should generate a model with references to other models", func() {
			model1 := definitions.ModelMetadata{
				Name:        "ModelA",
				Description: "Model A",
				Fields: []definitions.FieldMetadata{
					{Name: "id", Type: "string", Description: "ID of Model A", Validator: "required"},
				},
			}

			model2 := definitions.ModelMetadata{
				Name:        "ModelB",
				Description: "Model B",
				Fields: []definitions.FieldMetadata{
					{Name: "modelA", Type: "ModelA", Description: "Reference to Model A", Validator: "required"},
				},
			}

			generateModelSpec(openapi, model1)
			generateModelSpec(openapi, model2)

			schemaRef1 := openapi.Components.Schemas["ModelA"]
			Expect(schemaRef1).NotTo(BeNil())
			Expect(schemaRef1.Value.Title).To(Equal("ModelA"))

			schemaRef2 := openapi.Components.Schemas["ModelB"]
			Expect(schemaRef2).NotTo(BeNil())
			Expect(schemaRef2.Value.Title).To(Equal("ModelB"))
			Expect(schemaRef2.Value.Properties).To(HaveKey("modelA"))
			Expect(schemaRef2.Value.Properties["modelA"].Ref).To(Equal("#/components/schemas/ModelA"))
		})
	})

	Describe("GenerateModelsSpec", func() {
		It("should generate specifications for multiple models", func() {
			models := []definitions.ModelMetadata{
				{
					Name:        "TestModel1",
					Description: "First test model",
					Fields: []definitions.FieldMetadata{
						{
							Name:        "field1",
							Type:        "string",
							Description: "A string field",
							Validator:   "required",
						},
					},
				},
				{
					Name:        "TestModel2",
					Description: "Second test model",
					Fields: []definitions.FieldMetadata{
						{
							Name:        "field2",
							Type:        "int",
							Description: "An integer field",
							Validator:   "",
						},
					},
				},
			}

			err := GenerateModelsSpec(openapi, models)
			Expect(err).To(BeNil())

			schemaRef1 := openapi.Components.Schemas["TestModel1"]
			Expect(schemaRef1).NotTo(BeNil())
			Expect(schemaRef1.Value.Title).To(Equal("TestModel1"))

			schemaRef2 := openapi.Components.Schemas["TestModel2"]
			Expect(schemaRef2).NotTo(BeNil())
			Expect(schemaRef2.Value.Title).To(Equal("TestModel2"))
		})

		It("should generate specifications for models with references", func() {
			models := []definitions.ModelMetadata{
				{
					Name:        "ModelC",
					Description: "Model C",
					Fields: []definitions.FieldMetadata{
						{Name: "id", Type: "string", Description: "ID of Model C", Validator: "required"},
					},
				},
				{
					Name:        "ModelD",
					Description: "Model D",
					Fields: []definitions.FieldMetadata{
						{Name: "modelC", Type: "ModelC", Description: "Reference to Model C", Validator: "required"},
					},
				},
			}

			err := GenerateModelsSpec(openapi, models)
			Expect(err).To(BeNil())

			schemaRefC := openapi.Components.Schemas["ModelC"]
			Expect(schemaRefC).NotTo(BeNil())
			Expect(schemaRefC.Value.Title).To(Equal("ModelC"))

			schemaRefD := openapi.Components.Schemas["ModelD"]
			Expect(schemaRefD).NotTo(BeNil())
			Expect(schemaRefD.Value.Title).To(Equal("ModelD"))
			Expect(schemaRefD.Value.Properties).To(HaveKey("modelC"))
			Expect(schemaRefD.Value.Properties["modelC"].Ref).To(Equal("#/components/schemas/ModelC"))
		})
	})
})