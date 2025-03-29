package swagen30

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gopher-fleece/gleece/definitions"
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

			generateStructSpec(openapi, model)

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
			Expect(*schemaRef.Value.Properties["field2"].Value.Min).To(Equal(10.0))
			Expect(schemaRef.Value.Properties["field2"].Value.ExclusiveMin).To(Equal(true))
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

			generateStructSpec(openapi, model1)
			generateStructSpec(openapi, model2)

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

	Describe("GenerateEnumSpec", func() {
		It("should generate a string enum specification correctly", func() {
			enumModel := definitions.EnumMetadata{
				Name:        "StringEnum",
				Description: "A string enum",
				Type:        "string",
				Values:      []string{"VALUE1", "VALUE2", "VALUE3"},
			}

			generateEnumSpec(openapi, enumModel)

			schemaRef := openapi.Components.Schemas["StringEnum"]
			Expect(schemaRef).NotTo(BeNil())
			Expect(schemaRef.Value.Title).To(Equal("StringEnum"))
			Expect(schemaRef.Value.Description).To(Equal("A string enum"))
			Expect(schemaRef.Value.Type).To(Equal(&openapi3.Types{"string"}))
			Expect(schemaRef.Value.Deprecated).To(BeFalse())
			Expect(schemaRef.Value.Enum).To(HaveLen(3))
			Expect(schemaRef.Value.Enum).To(ContainElements("VALUE1", "VALUE2", "VALUE3"))
		})

		It("should generate an integer enum specification correctly", func() {
			enumModel := definitions.EnumMetadata{
				Name:        "IntEnum",
				Description: "An integer enum",
				Type:        "int",
				Values:      []string{"1", "2", "3"},
			}

			generateEnumSpec(openapi, enumModel)

			schemaRef := openapi.Components.Schemas["IntEnum"]
			Expect(schemaRef).NotTo(BeNil())
			Expect(schemaRef.Value.Title).To(Equal("IntEnum"))
			Expect(schemaRef.Value.Description).To(Equal("An integer enum"))
			// Note: The actual output type depends on swagtool.ToOpenApiType implementation
			// Assuming it maps "int" to "integer" for OpenAPI spec
			Expect(schemaRef.Value.Enum).To(HaveLen(3))
			Expect(schemaRef.Value.Enum).To(ContainElements("1", "2", "3"))
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

			generateEnumSpec(openapi, enumModel)

			schemaRef := openapi.Components.Schemas["DeprecatedEnum"]
			Expect(schemaRef).NotTo(BeNil())
			Expect(schemaRef.Value.Deprecated).To(BeTrue())
		})

		It("should handle empty enum values", func() {
			enumModel := definitions.EnumMetadata{
				Name:        "EmptyEnum",
				Description: "An enum with no values",
				Type:        "string",
				Values:      []string{},
			}

			generateEnumSpec(openapi, enumModel)

			schemaRef := openapi.Components.Schemas["EmptyEnum"]
			Expect(schemaRef).NotTo(BeNil())
			Expect(schemaRef.Value.Enum).To(HaveLen(0))
		})

		It("should handle non-primitive types correctly", func() {
			enumModel := definitions.EnumMetadata{
				Name:        "ComplexEnum",
				Description: "An enum with complex type",
				Type:        "customType",
				Values:      []string{"CUSTOM1", "CUSTOM2"},
			}

			generateEnumSpec(openapi, enumModel)

			schemaRef := openapi.Components.Schemas["ComplexEnum"]
			Expect(schemaRef).NotTo(BeNil())
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

			err := GenerateModelsSpec(openapi, &definitions.Models{
				Structs: models,
			})
			Expect(err).To(BeNil())

			schemaRef1 := openapi.Components.Schemas["TestModel1"]
			Expect(schemaRef1).NotTo(BeNil())
			Expect(schemaRef1.Value.Title).To(Equal("TestModel1"))

			schemaRef2 := openapi.Components.Schemas["TestModel2"]
			Expect(schemaRef2).NotTo(BeNil())
			Expect(schemaRef2.Value.Title).To(Equal("TestModel2"))
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

			err := GenerateModelsSpec(openapi, &definitions.Models{
				Structs: models,
			})
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

		It("should load ref schema value from cache", func() {
			models := []definitions.StructMetadata{
				{
					Name:        "ModelC",
					Description: "Model C",
					Fields: []definitions.FieldMetadata{
						{Name: "refD", Type: "ModelD", Description: "Ref to Model C", Tag: ""},
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

			err := GenerateModelsSpec(openapi, &definitions.Models{
				Structs: models,
			})
			Expect(err).To(BeNil())

			schemaRefC := openapi.Components.Schemas["ModelC"]
			Expect(schemaRefC.Value.Properties["refD"].Ref).To(Equal("#/components/schemas/ModelD"))
			Expect(schemaRefC.Value.Properties["refD"].Value).ToNot(BeNil())
			Expect(schemaRefC.Value.Properties["refD"].Value.Properties["fieldD"].Value.Description).To(Equal("some field"))
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

			err := GenerateModelsSpec(openapi, &definitions.Models{
				Structs: structs,
				Enums:   enums,
			})
			Expect(err).To(BeNil())

			userSchema := openapi.Components.Schemas["User"]
			Expect(userSchema).NotTo(BeNil())
			Expect(userSchema.Value.Title).To(Equal("User"))

			statusSchema := openapi.Components.Schemas["Status"]
			Expect(statusSchema).NotTo(BeNil())
			Expect(statusSchema.Value.Title).To(Equal("Status"))
			Expect(statusSchema.Value.Enum).To(HaveLen(3))
			Expect(statusSchema.Value.Enum).To(ContainElements("ACTIVE", "INACTIVE", "SUSPENDED"))
		})
	})

	Describe("GenerateSchemaSpec with embedded fields", func() {
		It("should generate a model with embedded fields using allOf", func() {
			// First define a base model
			baseModel := definitions.StructMetadata{
				Name:        "BaseModel",
				Description: "A base model",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "ID",
						Type:        "string",
						Description: "Base ID field",
						Tag:         `json:"id" validate:"required"`,
					},
					{
						Name:        "CreatedAt",
						Type:        "time.Time",
						Description: "Creation timestamp",
						Tag:         `json:"created_at"`,
					},
				},
			}

			// Generate the base model schema
			generateStructSpec(openapi, baseModel)

			// Now create a model with an embedded field
			modelWithEmbedded := definitions.StructMetadata{
				Name:        "ExtendedModel",
				Description: "A model with embedded field",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "BaseModel",
						Type:        "BaseModel",
						Description: "",
						IsEmbedded:  true,
						Tag:         ``,
					},
					{
						Name:        "ExtraField",
						Type:        "string",
						Description: "An extra field",
						Tag:         `json:"extra_field" validate:"required"`,
					},
				},
			}

			// Generate the model with embedded field
			generateStructSpec(openapi, modelWithEmbedded)

			// Verify the schema structure
			schemaRef := openapi.Components.Schemas["ExtendedModel"]
			Expect(schemaRef).NotTo(BeNil())

			// Check that it has an allOf structure
			Expect(schemaRef.Value.AllOf).NotTo(BeNil())
			Expect(schemaRef.Value.AllOf).To(HaveLen(2))

			// First element in allOf should be the model's own properties schema
			ownPropsSchema := schemaRef.Value.AllOf[0].Value
			Expect(ownPropsSchema.Title).To(Equal("ExtendedModel"))
			Expect(ownPropsSchema.Properties).To(HaveKey("extra_field"))
			Expect(ownPropsSchema.Required).To(ContainElement("extra_field"))

			// Second element should be a reference to the embedded BaseModel
			embeddedRef := schemaRef.Value.AllOf[1]
			Expect(embeddedRef.Ref).To(Equal("#/components/schemas/BaseModel"))
		})

		It("should generate a model with multiple embedded fields", func() {
			// Define first embedded model
			firstEmbedded := definitions.StructMetadata{
				Name:        "Timestamps",
				Description: "Timestamp fields",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "CreatedAt",
						Type:        "time.Time",
						Description: "Creation timestamp",
						Tag:         `json:"created_at"`,
					},
					{
						Name:        "UpdatedAt",
						Type:        "time.Time",
						Description: "Update timestamp",
						Tag:         `json:"updated_at"`,
					},
				},
			}

			// Define second embedded model
			secondEmbedded := definitions.StructMetadata{
				Name:        "Identifiable",
				Description: "Identity fields",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "ID",
						Type:        "string",
						Description: "Unique identifier",
						Tag:         `json:"id" validate:"required"`,
					},
				},
			}

			// Generate schemas for embedded models
			generateStructSpec(openapi, firstEmbedded)
			generateStructSpec(openapi, secondEmbedded)

			// Define a model with multiple embedded fields
			modelWithMultipleEmbedded := definitions.StructMetadata{
				Name:        "ComplexModel",
				Description: "Model with multiple embedded types",
				Fields: []definitions.FieldMetadata{
					{
						Name:       "Timestamps",
						Type:       "Timestamps",
						IsEmbedded: true,
						Tag:        ``,
					},
					{
						Name:       "Identifiable",
						Type:       "Identifiable",
						IsEmbedded: true,
						Tag:        ``,
					},
					{
						Name:        "Name",
						Type:        "string",
						Description: "Model name",
						Tag:         `json:"name"`,
					},
				},
			}

			// Generate schema with multiple embedded fields
			generateStructSpec(openapi, modelWithMultipleEmbedded)

			// Verify the schema
			schemaRef := openapi.Components.Schemas["ComplexModel"]
			Expect(schemaRef).NotTo(BeNil())

			// Check allOf structure has own properties and two embedded types
			Expect(schemaRef.Value.AllOf).To(HaveLen(3))

			// First element should be own properties
			ownPropsSchema := schemaRef.Value.AllOf[0].Value
			Expect(ownPropsSchema.Properties).To(HaveKey("name"))

			// Other elements should be refs to embedded types
			embedRefs := []string{
				schemaRef.Value.AllOf[1].Ref,
				schemaRef.Value.AllOf[2].Ref,
			}
			Expect(embedRefs).To(ContainElement("#/components/schemas/Timestamps"))
			Expect(embedRefs).To(ContainElement("#/components/schemas/Identifiable"))
		})

		It("should generate a model with only embedded fields and no own properties", func() {
			// Define an embedded model
			embeddedModel := definitions.StructMetadata{
				Name:        "BasicFields",
				Description: "Basic fields model",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "Field1",
						Type:        "string",
						Description: "Basic field 1",
						Tag:         `json:"field1"`,
					},
				},
			}

			generateStructSpec(openapi, embeddedModel)

			// Create a model with only embedded fields
			onlyEmbeddedModel := definitions.StructMetadata{
				Name:        "EmbeddedOnly",
				Description: "Model with only embedded fields",
				Fields: []definitions.FieldMetadata{
					{
						Name:       "BasicFields",
						Type:       "BasicFields",
						IsEmbedded: true,
						Tag:        ``,
					},
				},
			}

			generateStructSpec(openapi, onlyEmbeddedModel)

			// Verify schema structure
			schemaRef := openapi.Components.Schemas["EmbeddedOnly"]
			Expect(schemaRef).NotTo(BeNil())

			// Should have allOf with two elements: own schema (empty props) and embedded ref
			Expect(schemaRef.Value.AllOf).To(HaveLen(2))

			// First schema is the model's own schema (should be empty)
			ownSchema := schemaRef.Value.AllOf[0].Value
			Expect(ownSchema.Properties).To(BeEmpty())

			// Second should be ref to embedded model
			Expect(schemaRef.Value.AllOf[1].Ref).To(Equal("#/components/schemas/BasicFields"))
		})

		It("should handle a model with embedded field containing deprecation", func() {
			// Create deprecated embedded model
			deprecatedModel := definitions.StructMetadata{
				Name:        "DeprecatedBase",
				Description: "A deprecated base model",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "OldField",
						Type:        "string",
						Description: "Old field",
						Tag:         `json:"old_field"`,
					},
				},
				Deprecation: definitions.DeprecationOptions{
					Deprecated:  true,
					Description: "Use NewBase instead",
				},
			}

			generateStructSpec(openapi, deprecatedModel)

			// Model embedding the deprecated model
			compositeModel := definitions.StructMetadata{
				Name:        "Composite",
				Description: "Composite model",
				Fields: []definitions.FieldMetadata{
					{
						Name:       "DeprecatedBase",
						Type:       "DeprecatedBase",
						IsEmbedded: true,
						Tag:        ``,
					},
					{
						Name:        "NewField",
						Type:        "string",
						Description: "New field",
						Tag:         `json:"new_field"`,
					},
				},
			}

			generateStructSpec(openapi, compositeModel)

			// Verify schema
			schemaRef := openapi.Components.Schemas["Composite"]
			Expect(schemaRef).NotTo(BeNil())

			// Check allOf structure
			Expect(schemaRef.Value.AllOf).To(HaveLen(2))

			// Second element should reference the deprecated model
			Expect(schemaRef.Value.AllOf[1].Ref).To(Equal("#/components/schemas/DeprecatedBase"))

			// Verify the deprecated model is actually marked as deprecated
			deprecatedRef := openapi.Components.Schemas["DeprecatedBase"]
			Expect(deprecatedRef.Value.Deprecated).To(BeTrue())
		})
	})
})
