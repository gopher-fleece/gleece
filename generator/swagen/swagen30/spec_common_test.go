package swagen30

import (
	"github.com/getkin/kin-openapi/openapi3"
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

		It("should create empty OpenAPI schema for unknown type", func() {
			schema := ToOpenApiSchema("unknown")
			Expect(schema.Type).To(Equal(openapi3.NewSchema().Type))
		})
	})

	Describe("ToOpenApiSchemaRef", func() {
		It("should create a schema reference for a string type", func() {
			schemaRef := ToOpenApiSchemaRef("string")
			Expect(schemaRef.Value.Type).To(Equal(openapi3.NewStringSchema().Type))
		})
	})

})
