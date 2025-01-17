package swagen

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/haimkastner/gleece/definitions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Spec Common", func() {

	Describe("InterfaceToSchemaRef", func() {
		var openapi *openapi3.T

		BeforeEach(func() {
			openapi = &openapi3.T{
				Components: &openapi3.Components{
					Schemas: openapi3.Schemas{},
				},
			}
		})

		It("should return a schema ref for a string type", func() {
			schemaRef := InterfaceToSchemaRef(openapi, "string")
			Expect(schemaRef.Value).To(Equal(openapi3.NewStringSchema()))
		})

		It("should return a schema ref for an object type", func() {
			openapi.Components.Schemas["testObject"] = &openapi3.SchemaRef{Value: openapi3.NewObjectSchema()}
			schemaRef := InterfaceToSchemaRef(openapi, "testObject")
			Expect(schemaRef.Ref).To(Equal("#/components/schemas/testObject"))
		})

		It("should handle nested schema references", func() {
			schemaRef := InterfaceToSchemaRef(openapi, "[]string")
			Expect(schemaRef.Value.Items.Value).To(Equal(openapi3.NewStringSchema()))
		})

		It("should handle nested-nested schema references", func() {
			schemaRef := InterfaceToSchemaRef(openapi, "[][]string")
			Expect(schemaRef.Value.Items.Value.Items.Value).To(Equal(openapi3.NewStringSchema()))
		})

		It("should handle nested-nested-nested int references", func() {
			schemaRef := InterfaceToSchemaRef(openapi, "[][][]int")
			Expect(schemaRef.Value.Items.Value.Items.Value.Items.Value).To(Equal(openapi3.NewIntegerSchema()))
		})

		It("should handle nested schema references", func() {
			schemaRef := InterfaceToSchemaRef(openapi, "[]testObject")
			Expect(schemaRef.Value.Items.Ref).To(Equal("#/components/schemas/testObject"))
		})
	})

	Describe("BuildSchemaValidation", func() {
		var schema *openapi3.SchemaRef

		BeforeEach(func() {
			schema = &openapi3.SchemaRef{
				Value: openapi3.NewSchema(),
			}
		})

		It("should apply email format validation", func() {
			BuildSchemaValidation(schema, "email", "string")
			Expect(schema.Value.Format).To(Equal("email"))
		})

		It("should apply email format validation while other irrelevant exists", func() {
			BuildSchemaValidation(schema, "email,required,other=5", "string")
			Expect(schema.Value.Format).To(Equal("email"))
		})

		It("should apply greater than validation", func() {
			BuildSchemaValidation(schema, "gt=10", "int")
			Expect(*schema.Value.Min).To(BeEquivalentTo(10))
		})

		It("should apply less than validation", func() {
			BuildSchemaValidation(schema, "lt=20", "int")
			Expect(*schema.Value.Max).To(BeEquivalentTo(20))
		})

		It("should apply min length validation for strings", func() {
			BuildSchemaValidation(schema, "min=5", "string")
			Expect(schema.Value.MinLength).To(BeEquivalentTo(5))
		})

		It("should apply max length validation for strings", func() {
			BuildSchemaValidation(schema, "max=10", "string")
			Expect(*schema.Value.MaxLength).To(BeEquivalentTo(10))
		})

		It("should apply pattern validation for strings", func() {
			BuildSchemaValidation(schema, "pattern=^[a-z]+$", "string")
			Expect(schema.Value.Pattern).To(Equal("^[a-z]+$"))
		})

		It("should apply min items validation for arrays", func() {
			BuildSchemaValidation(schema, "minItems=2", "[]string")
			Expect(schema.Value.MinItems).To(BeEquivalentTo(2))
		})

		It("should apply max items validation for arrays", func() {
			BuildSchemaValidation(schema, "maxItems=5", "[]int")
			Expect(*schema.Value.MaxItems).To(BeEquivalentTo(5))
		})

		It("should apply unique items validation for arrays", func() {
			BuildSchemaValidation(schema, "uniqueItems=true", "[]something")
			Expect(schema.Value.UniqueItems).To(BeTrue())
		})

		It("should apply enum validation for strings", func() {
			BuildSchemaValidation(schema, "enum=a|b|c", "string")
			Expect(schema.Value.Enum).To(ConsistOf("a", "b", "c"))
		})
	})

	Describe("IsPrimitiveType", func() {
		It("should return true for primitive types", func() {
			Expect(IsPrimitiveType("string")).To(BeTrue())
			Expect(IsPrimitiveType("int")).To(BeTrue())
			Expect(IsPrimitiveType("float64")).To(BeTrue())
		})

		It("should return false for non-primitive types", func() {
			Expect(IsPrimitiveType("complex")).To(BeFalse())
			Expect(IsPrimitiveType("[]string")).To(BeFalse())
		})
	})

	Describe("ToOpenApiType", func() {
		It("should convert Go types to OpenAPI types", func() {
			Expect(ToOpenApiType("string")).To(Equal("string"))
			Expect(ToOpenApiType("int")).To(Equal("integer"))
			Expect(ToOpenApiType("bool")).To(Equal("boolean"))
			Expect(ToOpenApiType("float64")).To(Equal("number"))
			Expect(ToOpenApiType("[]string")).To(Equal("array"))
			Expect(ToOpenApiType("customType")).To(Equal("object"))
		})
	})

	Describe("ToOpenApiSchema", func() {
		It("should create OpenAPI schema for string type", func() {
			schema := ToOpenApiSchema("string")
			Expect(schema.Type).To(Equal(openapi3.NewStringSchema().Type))
		})

		It("should create OpenAPI schema for integer type", func() {
			schema := ToOpenApiSchema("integer")
			Expect(schema.Type).To(Equal(openapi3.NewIntegerSchema().Type))
		})

		It("should create OpenAPI schema for boolean type", func() {
			schema := ToOpenApiSchema("boolean")
			Expect(schema.Type).To(Equal(openapi3.NewBoolSchema().Type))
		})

		It("should create OpenAPI schema for number type", func() {
			schema := ToOpenApiSchema("number")
			Expect(schema.Type).To(Equal(openapi3.NewFloat64Schema().Type))
		})

		It("should create OpenAPI schema for array type", func() {
			schema := ToOpenApiSchema("array")
			Expect(schema.Type).To(Equal(openapi3.NewArraySchema().Type))
		})

		It("should create OpenAPI schema for object type", func() {
			schema := ToOpenApiSchema("object")
			Expect(schema.Type).To(Equal(openapi3.NewObjectSchema().Type))
		})
	})

	Describe("ToOpenApiSchemaRef", func() {
		It("should create a schema reference for a string type", func() {
			schemaRef := ToOpenApiSchemaRef("string")
			Expect(schemaRef.Value.Type).To(Equal(openapi3.NewStringSchema().Type))
		})
	})

	Describe("IsSecurityNameInSecuritySchemes", func() {
		It("should return true if security name exists in schemes", func() {
			securitySchemes := []definitions.SecuritySchemeConfig{
				{SecurityName: "oauth2"},
				{SecurityName: "apiKey"},
			}
			Expect(IsSecurityNameInSecuritySchemes(securitySchemes, "oauth2")).To(BeTrue())
			Expect(IsSecurityNameInSecuritySchemes(securitySchemes, "apiKey")).To(BeTrue())
		})

		It("should return false if security name does not exist in schemes", func() {
			securitySchemes := []definitions.SecuritySchemeConfig{
				{SecurityName: "oauth2"},
				{SecurityName: "apiKey"},
			}
			Expect(IsSecurityNameInSecuritySchemes(securitySchemes, "basicAuth")).To(BeFalse())
		})
	})
})
