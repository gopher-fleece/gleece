package swagen31

import (
	"github.com/gopher-fleece/gleece/v2/definitions"
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

	Describe("GenerateStructsSpec with embedded fields", func() {
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
			generateStructsSpec(doc, baseModel)

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
			generateStructsSpec(doc, modelWithEmbedded)

			// Verify the schema structure
			schemaRef, found := doc.Components.Schemas.Get("ExtendedModel")
			Expect(found).To(BeTrue())

			schema := schemaRef.Schema()

			// Check that it has an allOf structure
			Expect(schema.AllOf).NotTo(BeNil())
			Expect(schema.AllOf).To(HaveLen(2))

			// First element in allOf should be the model's own properties schema
			ownPropsSchema := schema.AllOf[0].Schema()
			Expect(ownPropsSchema.Title).To(Equal("ExtendedModel"))

			// Check if extra_field exists in properties
			extraField, found := ownPropsSchema.Properties.Get("extra_field")
			Expect(found).To(BeTrue())
			Expect(extraField.Schema().Description).To(Equal("An extra field"))

			Expect(ownPropsSchema.Required).To(ContainElement("extra_field"))

			// Second element should be a reference to the embedded BaseModel
			embeddedRef := schema.AllOf[1]
			Expect(embeddedRef.GetReference()).To(Equal("#/components/schemas/BaseModel"))
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
			generateStructsSpec(doc, firstEmbedded)
			generateStructsSpec(doc, secondEmbedded)

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
			generateStructsSpec(doc, modelWithMultipleEmbedded)

			// Verify the schema
			schemaRef, found := doc.Components.Schemas.Get("ComplexModel")
			Expect(found).To(BeTrue())

			schema := schemaRef.Schema()

			// Check allOf structure has own properties and two embedded types
			Expect(schema.AllOf).To(HaveLen(3))

			// First element should be own properties with "name" field
			ownPropsSchema := schema.AllOf[0].Schema()
			nameField, found := ownPropsSchema.Properties.Get("name")
			Expect(found).To(BeTrue())
			Expect(nameField.Schema().Description).To(Equal("Model name"))

			// Other elements should be refs to embedded types
			// Collect references to check
			embedRefs := []string{
				schema.AllOf[1].GetReference(),
				schema.AllOf[2].GetReference(),
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

			generateStructsSpec(doc, embeddedModel)

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

			generateStructsSpec(doc, onlyEmbeddedModel)

			// Verify schema structure
			schemaRef, found := doc.Components.Schemas.Get("EmbeddedOnly")
			Expect(found).To(BeTrue())

			schema := schemaRef.Schema()

			// Should have allOf with two elements: own schema and embedded ref
			Expect(schema.AllOf).To(HaveLen(2))

			// First schema is the model's own schema (should be empty)
			ownSchema := schema.AllOf[0].Schema()
			Expect(ownSchema.Properties.Len()).To(Equal(0))

			// Second should be ref to embedded model
			Expect(schema.AllOf[1].GetReference()).To(Equal("#/components/schemas/BasicFields"))
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

			generateStructsSpec(doc, deprecatedModel)

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

			generateStructsSpec(doc, compositeModel)

			// Verify schema
			schemaRef, found := doc.Components.Schemas.Get("Composite")
			Expect(found).To(BeTrue())

			schema := schemaRef.Schema()

			// Check allOf structure
			Expect(schema.AllOf).To(HaveLen(2))

			// Second element should reference the deprecated model
			Expect(schema.AllOf[1].GetReference()).To(Equal("#/components/schemas/DeprecatedBase"))

			// Verify the deprecated model is actually marked as deprecated
			deprecatedRef, found := doc.Components.Schemas.Get("DeprecatedBase")
			Expect(found).To(BeTrue())
			Expect(*deprecatedRef.Schema().Deprecated).To(BeTrue())
		})

		It("should properly organize required fields in a model with embedded fields", func() {
			// Define base model with required fields
			baseModel := definitions.StructMetadata{
				Name:        "RequiredFieldsBase",
				Description: "Base model with required fields",
				Fields: []definitions.FieldMetadata{
					{
						Name:        "RequiredBaseField",
						Type:        "string",
						Description: "Required field in base",
						Tag:         `json:"required_base_field" validate:"required"`,
					},
				},
			}

			generateStructsSpec(doc, baseModel)

			// Create model with embedded base and its own required fields
			derivedModel := definitions.StructMetadata{
				Name:        "RequiredFieldsDerived",
				Description: "Derived model with more required fields",
				Fields: []definitions.FieldMetadata{
					{
						Name:       "RequiredFieldsBase",
						Type:       "RequiredFieldsBase",
						IsEmbedded: true,
						Tag:        ``,
					},
					{
						Name:        "RequiredDerivedField",
						Type:        "string",
						Description: "Required field in derived",
						Tag:         `json:"required_derived_field" validate:"required"`,
					},
					{
						Name:        "OptionalDerivedField",
						Type:        "string",
						Description: "Optional field in derived",
						Tag:         `json:"optional_derived_field"`,
					},
				},
			}

			generateStructsSpec(doc, derivedModel)

			// Verify schema
			derivedSchemaRef, found := doc.Components.Schemas.Get("RequiredFieldsDerived")
			Expect(found).To(BeTrue())

			// Check own schema has correct required field
			ownSchema := derivedSchemaRef.Schema().AllOf[0].Schema()
			Expect(ownSchema.Required).To(ContainElement("required_derived_field"))
			Expect(ownSchema.Required).NotTo(ContainElement("optional_derived_field"))

			// Base schema should still have its own required field
			baseSchemaRef, found := doc.Components.Schemas.Get("RequiredFieldsBase")
			Expect(found).To(BeTrue())
			baseSchema := baseSchemaRef.Schema()
			Expect(baseSchema.Required).To(ContainElement("required_base_field"))
		})
	})

	Describe("GenerateAliasSpec", func() {
		It("should generate a string alias specification correctly", func() {
			alias := definitions.NakedAliasMetadata{
				Name:        "UserID",
				Description: "A unique user identifier",
				Type:        "string",
				PkgPath:     "example.com/types",
			}

			generateAliasSpec(doc, alias)

			schemaRef, found := doc.Components.Schemas.Get("UserID")
			Expect(found).To(BeTrue())
			schema := schemaRef.Schema()
			Expect(schema.Title).To(Equal("UserID"))
			Expect(schema.Description).To(Equal("A unique user identifier"))
			Expect(schema.Type).To(Equal([]string{"string"}))
			Expect(*schema.Deprecated).To(BeFalse())
		})

		It("should generate an integer alias specification correctly", func() {
			alias := definitions.NakedAliasMetadata{
				Name:        "Count",
				Description: "A count of items",
				Type:        "int",
				PkgPath:     "example.com/types",
			}

			generateAliasSpec(doc, alias)

			schemaRef, found := doc.Components.Schemas.Get("Count")
			Expect(found).To(BeTrue())
			schema := schemaRef.Schema()
			Expect(schema.Title).To(Equal("Count"))
			Expect(schema.Description).To(Equal("A count of items"))
			Expect(schema.Type).To(Equal([]string{"integer"}))
		})

		It("should set deprecation flag correctly for alias", func() {
			deprecationInfo := "Use NewID instead"
			alias := definitions.NakedAliasMetadata{
				Name:        "OldID",
				Description: "A deprecated identifier",
				Type:        "string",
				PkgPath:     "example.com/types",
				Deprecation: definitions.DeprecationOptions{
					Deprecated:  true,
					Description: deprecationInfo,
				},
			}

			generateAliasSpec(doc, alias)

			schemaRef, found := doc.Components.Schemas.Get("OldID")
			Expect(found).To(BeTrue())
			schema := schemaRef.Schema()
			Expect(*schema.Deprecated).To(BeTrue())
		})

		It("should generate a float alias specification correctly", func() {
			alias := definitions.NakedAliasMetadata{
				Name:        "Price",
				Description: "A price value",
				Type:        "float64",
				PkgPath:     "example.com/types",
			}

			generateAliasSpec(doc, alias)

			schemaRef, found := doc.Components.Schemas.Get("Price")
			Expect(found).To(BeTrue())
			schema := schemaRef.Schema()
			Expect(schema.Title).To(Equal("Price"))
			Expect(schema.Description).To(Equal("A price value"))
			Expect(schema.Type).To(Equal([]string{"number"}))
		})

		It("should generate a boolean alias specification correctly", func() {
			alias := definitions.NakedAliasMetadata{
				Name:        "IsActive",
				Description: "Active status flag",
				Type:        "bool",
				PkgPath:     "example.com/types",
			}

			generateAliasSpec(doc, alias)

			schemaRef, found := doc.Components.Schemas.Get("IsActive")
			Expect(found).To(BeTrue())
			schema := schemaRef.Schema()
			Expect(schema.Title).To(Equal("IsActive"))
			Expect(schema.Description).To(Equal("Active status flag"))
			Expect(schema.Type).To(Equal([]string{"boolean"}))
		})

		It("should handle alias with empty description", func() {
			alias := definitions.NakedAliasMetadata{
				Name:    "EmptyDescAlias",
				Type:    "string",
				PkgPath: "example.com/types",
			}

			generateAliasSpec(doc, alias)

			schemaRef, found := doc.Components.Schemas.Get("EmptyDescAlias")
			Expect(found).To(BeTrue())
			schema := schemaRef.Schema()
			Expect(schema.Title).To(Equal("EmptyDescAlias"))
			Expect(schema.Description).To(Equal(""))
		})
	})

	Describe("GenerateModelsSpec with Aliases", func() {
		It("should generate specifications for models with aliases", func() {
			models := &definitions.Models{
				Structs: []definitions.StructMetadata{
					{
						Name:        "User",
						Description: "A user model",
						Fields: []definitions.FieldMetadata{
							{
								Name:        "ID",
								Type:        "UserID",
								Description: "User identifier",
								Tag:         `json:"id"`,
							},
							{
								Name:        "Name",
								Type:        "string",
								Description: "User name",
								Tag:         `json:"name"`,
							},
						},
					},
				},
				Enums: []definitions.EnumMetadata{
					{
						Name:        "Role",
						Description: "User role",
						Type:        "string",
						Values:      []string{"admin", "user"},
					},
				},
				Aliases: []definitions.NakedAliasMetadata{
					{
						Name:        "UserID",
						Description: "A unique user identifier",
						Type:        "string",
						PkgPath:     "example.com/types",
					},
					{
						Name:        "Count",
						Description: "A count value",
						Type:        "int",
						PkgPath:     "example.com/types",
					},
				},
			}

			err := GenerateModelsSpec(doc, models)
			Expect(err).To(BeNil())

			// Verify struct was generated
			userSchemaRef, found := doc.Components.Schemas.Get("User")
			Expect(found).To(BeTrue())
			Expect(userSchemaRef.Schema().Title).To(Equal("User"))

			// Verify enum was generated
			roleSchemaRef, found := doc.Components.Schemas.Get("Role")
			Expect(found).To(BeTrue())
			Expect(roleSchemaRef.Schema().Title).To(Equal("Role"))

			// Verify aliases were generated
			userIDSchemaRef, found := doc.Components.Schemas.Get("UserID")
			Expect(found).To(BeTrue())
			userIDSchema := userIDSchemaRef.Schema()
			Expect(userIDSchema.Title).To(Equal("UserID"))
			Expect(userIDSchema.Description).To(Equal("A unique user identifier"))
			Expect(userIDSchema.Type).To(Equal([]string{"string"}))

			countSchemaRef, found := doc.Components.Schemas.Get("Count")
			Expect(found).To(BeTrue())
			countSchema := countSchemaRef.Schema()
			Expect(countSchema.Title).To(Equal("Count"))
			Expect(countSchema.Description).To(Equal("A count value"))
			Expect(countSchema.Type).To(Equal([]string{"integer"}))
		})

		It("should handle empty aliases list", func() {
			models := &definitions.Models{
				Structs: []definitions.StructMetadata{
					{
						Name:        "Simple",
						Description: "Simple model",
						Fields:      []definitions.FieldMetadata{},
					},
				},
				Enums:   []definitions.EnumMetadata{},
				Aliases: []definitions.NakedAliasMetadata{},
			}

			err := GenerateModelsSpec(doc, models)
			Expect(err).To(BeNil())

			simpleSchemaRef, found := doc.Components.Schemas.Get("Simple")
			Expect(found).To(BeTrue())
			Expect(simpleSchemaRef.Schema().Title).To(Equal("Simple"))
		})
	})
})
