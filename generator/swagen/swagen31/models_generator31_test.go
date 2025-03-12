package swagen31

import (
	"github.com/gopher-fleece/gleece/definitions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	highbase "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

var _ = Describe("swagen31", func() {
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
			model := definitions.StructMetadata{
				Name:        "TestModel",
				Description: "A test model",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "Field1",
						Type:        "string",
						Description: "A string field",
						Tag:         `json:"field1,omitempty" validate:"required"`,
					},
					{
						Name:        "field2",
						Type:        "int",
						Description: "An integer field",
						Tag:         `validate:"gt=10"`,
					},
				},
			}

			generateStructsSpec(doc, model)

			schemaRef, found := doc.Components.Schemas.Get("TestModel")
			Expect(found).To(BeTrue())
			schema := schemaRef.Schema()
			Expect(schema.Title).To(Equal("TestModel"))
			Expect(schema.Description).To(Equal("A test model"))
			Expect(schema.Type).To(Equal([]string{"object"}))

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
			model1 := definitions.StructMetadata{
				Name:        "ModelA",
				Description: "Model A",
				Fields: []definitions.FieldMetadata{
					{Name: "id", Type: "string", Description: "ID of Model A", Tag: "required"},
				},
			}

			model2 := definitions.StructMetadata{
				Name:        "ModelB",
				Description: "Model B",
				Fields: []definitions.FieldMetadata{
					{Name: "modelA", Type: "ModelA", Description: "Reference to Model A", Tag: "required"},
				},
			}

			generateStructsSpec(doc, model1)
			generateStructsSpec(doc, model2)

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

	Describe("GenerateEnumSpec", func() {
		It("should generate a string enum specification correctly", func() {
			enumModel := definitions.EnumMetadata{
				Name:        "StringEnum",
				Description: "A string enum",
				Type:        "string",
				Values:      []string{"VALUE1", "VALUE2", "VALUE3"},
			}

			generateEnumsSpec(doc, enumModel)

			schemaRef, found := doc.Components.Schemas.Get("StringEnum")
			Expect(found).To(BeTrue())
			schema := schemaRef.Schema()
			Expect(schema.Title).To(Equal("StringEnum"))
			Expect(schema.Description).To(Equal("A string enum"))
			Expect(schema.Type).To(Equal([]string{"string"}))

			isDeprecated := false
			Expect(schema.Deprecated).To(Equal(&isDeprecated))

			Expect(schema.Enum).To(HaveLen(3))
			enumValues := []string{}
			for _, node := range schema.Enum {
				enumValues = append(enumValues, node.Value)
			}
			Expect(enumValues).To(ContainElements("VALUE1", "VALUE2", "VALUE3"))
		})

		It("should generate an integer enum specification correctly", func() {
			enumModel := definitions.EnumMetadata{
				Name:        "IntEnum",
				Description: "An integer enum",
				Type:        "int",
				Values:      []string{"1", "2", "3"},
			}

			generateEnumsSpec(doc, enumModel)

			schemaRef, found := doc.Components.Schemas.Get("IntEnum")
			Expect(found).To(BeTrue())
			schema := schemaRef.Schema()
			Expect(schema.Title).To(Equal("IntEnum"))
			Expect(schema.Description).To(Equal("An integer enum"))

			// Check if the type is set correctly based on the swagtool.ToOpenApiType implementation
			// Assuming it maps "int" to "integer" for OpenAPI spec
			typeStr := schema.Type[0]
			Expect(typeStr == "integer" || typeStr == "number").To(BeTrue())

			Expect(schema.Enum).To(HaveLen(3))
			enumValues := []string{}
			for _, node := range schema.Enum {
				enumValues = append(enumValues, node.Value)
			}
			Expect(enumValues).To(ContainElements("1", "2", "3"))
		})

		It("should set deprecation flag correctly for enum", func() {
			deprecationInfo := "Deprecated since v2.0.0"
			enumModel := definitions.EnumMetadata{
				Name:        "DeprecatedEnum",
				Description: "A deprecated enum",
				Type:        "string",
				Values:      []string{"OLD_VALUE1", "OLD_VALUE2"},
				Deprecation: definitions.DeprecationOptions{
					Deprecated:  true,
					Description: deprecationInfo,
				},
			}

			generateEnumsSpec(doc, enumModel)

			schemaRef, found := doc.Components.Schemas.Get("DeprecatedEnum")
			Expect(found).To(BeTrue())
			schema := schemaRef.Schema()
			isDeprecated := true
			Expect(schema.Deprecated).To(Equal(&isDeprecated))
		})

		It("should handle empty enum values", func() {
			enumModel := definitions.EnumMetadata{
				Name:        "EmptyEnum",
				Description: "An enum with no values",
				Type:        "string",
				Values:      []string{},
			}

			generateEnumsSpec(doc, enumModel)

			schemaRef, found := doc.Components.Schemas.Get("EmptyEnum")
			Expect(found).To(BeTrue())
			schema := schemaRef.Schema()
			Expect(schema.Enum).To(HaveLen(0))
		})

		It("should handle non-primitive types correctly", func() {
			enumModel := definitions.EnumMetadata{
				Name:        "ComplexEnum",
				Description: "An enum with complex type",
				Type:        "customType",
				Values:      []string{"CUSTOM1", "CUSTOM2"},
			}

			generateEnumsSpec(doc, enumModel)

			_, found := doc.Components.Schemas.Get("ComplexEnum")
			Expect(found).To(BeTrue())
			// The actual type conversion depends on swagtool.ToOpenApiType implementation
		})
	})

	Describe("GenerateModelsSpec", func() {
		It("should generate specifications for multiple models", func() {
			models := []definitions.StructMetadata{
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

			err := GenerateModelsSpec(doc, &definitions.Models{
				Structs: models,
			})
			Expect(err).To(BeNil())

			schemaRef1, found := doc.Components.Schemas.Get("TestModel1")
			Expect(found).To(BeTrue())
			Expect(schemaRef1.Schema().Title).To(Equal("TestModel1"))

			schemaRef2, found := doc.Components.Schemas.Get("TestModel2")
			Expect(found).To(BeTrue())
			Expect(schemaRef2.Schema().Title).To(Equal("TestModel2"))
		})

		It("should generate specifications for models with references", func() {
			models := []definitions.StructMetadata{
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

			err := GenerateModelsSpec(doc, &definitions.Models{
				Structs: models,
			})
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
			models := []definitions.StructMetadata{
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

			err := GenerateModelsSpec(doc, &definitions.Models{
				Structs: models,
			})
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

		It("should generate specifications for both structs and enums", func() {
			structs := []definitions.StructMetadata{
				{
					Name:        "User",
					Description: "User model",
					Fields: []definitions.FieldMetadata{
						{Name: "id", Type: "string", Description: "User ID", Tag: ""},
						{Name: "status", Type: "Status", Description: "User status", Tag: ""},
					},
				},
			}

			enums := []definitions.EnumMetadata{
				{
					Name:        "Status",
					Description: "User status enum",
					Type:        "string",
					Values:      []string{"ACTIVE", "INACTIVE", "SUSPENDED"},
				},
			}

			err := GenerateModelsSpec(doc, &definitions.Models{
				Structs: structs,
				Enums:   enums,
			})
			Expect(err).To(BeNil())

			userSchema, found := doc.Components.Schemas.Get("User")
			Expect(found).To(BeTrue())
			Expect(userSchema.Schema().Title).To(Equal("User"))

			statusSchema, found := doc.Components.Schemas.Get("Status")
			Expect(found).To(BeTrue())
			Expect(statusSchema.Schema().Title).To(Equal("Status"))

			Expect(statusSchema.Schema().Enum).To(HaveLen(3))
			enumValues := []string{}
			for _, node := range statusSchema.Schema().Enum {
				enumValues = append(enumValues, node.Value)
			}
			Expect(enumValues).To(ContainElements("ACTIVE", "INACTIVE", "SUSPENDED"))
		})
	})
})
