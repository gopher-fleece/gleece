package swagten31

import (
	"github.com/gopher-fleece/gleece/definitions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	highbase "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

var _ = Describe("Swagten31", func() {
	var doc *v3.Document

	BeforeEach(func() {
		doc = &v3.Document{
			Components: &v3.Components{
				Schemas: orderedmap.New[string, *highbase.SchemaProxy](),
			},
		}
	})

	Describe("GenerateSchemaSpec", func() {
		It("should generate a model specification correctly", func() {
			model := definitions.ModelMetadata{
				Name:        "TestModel",
				Description: "A test model",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "Field1",
						Type:        "string",
						Description: "A string field",
						Tag:         `json:"field1" validate:"required"`,
					},
					{
						Name:        "field2",
						Type:        "int",
						Description: "An integer field",
						Tag:         `validate:"gt=10"`,
					},
				},
			}

			generateModelSpec(doc, model)

			schemaRef, found := doc.Components.Schemas.Get("TestModel")
			Expect(found).To(BeTrue())
			schema := schemaRef.Schema()
			Expect(schema.Title).To(Equal("TestModel"))
			Expect(schema.Description).To(Equal("A test model"))
			Expect(schema.Type).To(Equal(objectType))

			field1Schema, found := schema.Properties.Get("field1")
			Expect(found).To(BeTrue())
			Expect(field1Schema.Schema().Description).To(Equal("A string field"))

			field2Schema, found := schema.Properties.Get("field2")
			Expect(found).To(BeTrue())
			schema2 := field2Schema.Schema()
			Expect(schema2.Description).To(Equal("An integer field"))
			Expect(schema2.ExclusiveMinimum.B).To(Equal(10.0))

			Expect(schema.Required).To(ContainElement("field1"))
			Expect(schema.Required).NotTo(ContainElement("field2"))
		})

		It("should generate a model with references to other models", func() {
			model1 := definitions.ModelMetadata{
				Name:        "ModelA",
				Description: "Model A",
				Fields: []definitions.FieldMetadata{
					{Name: "id", Type: "string", Description: "ID of Model A", Tag: "required"},
				},
			}

			model2 := definitions.ModelMetadata{
				Name:        "ModelB",
				Description: "Model B",
				Fields: []definitions.FieldMetadata{
					{Name: "modelA", Type: "ModelA", Description: "Reference to Model A", Tag: "required"},
				},
			}

			generateModelSpec(doc, model1)
			generateModelSpec(doc, model2)

			schemaRef1, found := doc.Components.Schemas.Get("ModelA")
			Expect(found).To(BeTrue())
			Expect(schemaRef1.Schema().Title).To(Equal("ModelA"))

			schemaRef2, found := doc.Components.Schemas.Get("ModelB")
			Expect(found).To(BeTrue())
			Expect(schemaRef2.Schema().Title).To(Equal("ModelB"))

			modelARef, found := schemaRef2.Schema().Properties.Get("modelA")
			Expect(found).To(BeTrue())
			Expect(modelARef.GetReference()).To(Equal("#/components/schemas/ModelA"))
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
							Tag:         "required",
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
							Tag:         "",
						},
					},
				},
			}

			err := GenerateModelsSpec(doc, models)
			Expect(err).To(BeNil())

			schemaRef1, found := doc.Components.Schemas.Get("TestModel1")
			Expect(found).To(BeTrue())
			Expect(schemaRef1.Schema().Title).To(Equal("TestModel1"))

			schemaRef2, found := doc.Components.Schemas.Get("TestModel2")
			Expect(found).To(BeTrue())
			Expect(schemaRef2.Schema().Title).To(Equal("TestModel2"))
		})

		It("should generate specifications for models with references", func() {
			models := []definitions.ModelMetadata{
				{
					Name:        "ModelC",
					Description: "Model C",
					Fields: []definitions.FieldMetadata{
						{Name: "id", Type: "string", Description: "ID of Model C", Tag: "required"},
					},
				},
				{
					Name:        "ModelD",
					Description: "Model D",
					Fields: []definitions.FieldMetadata{
						{Name: "modelC", Type: "ModelC", Description: "Reference to Model C", Tag: "required"},
					},
				},
			}

			err := GenerateModelsSpec(doc, models)
			Expect(err).To(BeNil())

			schemaRefC, found := doc.Components.Schemas.Get("ModelC")
			Expect(found).To(BeTrue())
			Expect(schemaRefC.Schema().Title).To(Equal("ModelC"))

			schemaRefD, found := doc.Components.Schemas.Get("ModelD")
			Expect(found).To(BeTrue())
			Expect(schemaRefD.Schema().Title).To(Equal("ModelD"))

			modelCRef, found := schemaRefD.Schema().Properties.Get("modelC")
			Expect(found).To(BeTrue())
			Expect(modelCRef.GetReference()).To(Equal("#/components/schemas/ModelC"))
		})

		It("should handle references correctly", func() {
			models := []definitions.ModelMetadata{
				{
					Name:        "ModelC",
					Description: "Model C",
					Fields: []definitions.FieldMetadata{
						{Name: "refD", Type: "ModelD", Description: "Ref to Model D", Tag: ""},
					},
				},
				{
					Name:        "ModelD",
					Description: "Model D",
					Fields: []definitions.FieldMetadata{
						{Name: "fieldD", Type: "string", Description: "some field", Tag: ""},
					},
				},
			}

			err := GenerateModelsSpec(doc, models)
			Expect(err).To(BeNil())

			schemaRefC, found := doc.Components.Schemas.Get("ModelC")
			Expect(found).To(BeTrue())

			refDProp, found := schemaRefC.Schema().Properties.Get("refD")
			Expect(found).To(BeTrue())
			Expect(refDProp.GetReference()).To(Equal("#/components/schemas/ModelD"))

			// Access the referenced schema
			modelD, found := doc.Components.Schemas.Get("ModelD")
			Expect(found).To(BeTrue())
			fieldDProp, found := modelD.Schema().Properties.Get("fieldD")
			Expect(found).To(BeTrue())
			Expect(fieldDProp.Schema().Description).To(Equal("some field"))
		})
	})
})
