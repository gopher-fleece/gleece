package swagen31

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

var _ = Describe("Spec V3.1 Common", func() {

	Describe("InterfaceToSchemaV31", func() {
		var doc *v3.Document

		BeforeEach(func() {
			doc = &v3.Document{
				Components: &v3.Components{
					Schemas: orderedmap.New[string, *base.SchemaProxy](),
				},
			}
		})

		It("should return a schema for a string type", func() {
			schema := InterfaceToSchemaV3(doc, "string")
			Expect(schema.Schema().Type).To(Equal([]string{"string"}))
		})

		It("should return a schema for an date-time type", func() {
			schema := InterfaceToSchemaV3(doc, "Time")
			Expect(schema.Schema().Type).To(Equal([]string{"string"}))
			Expect(schema.Schema().Format).To(Equal("date-time"))
		})

		It("should return a schema for a binary type", func() {
			schema := InterfaceToSchemaV3(doc, "[]byte")
			Expect(schema.Schema().Type).To(Equal([]string{"string"}))
			Expect(schema.Schema().Format).To(Equal("base64"))
		})

		It("should return a schema ref for an object type", func() {
			schemaProxy := InterfaceToSchemaV3(doc, "testObject")
			Expect(schemaProxy.GetReference()).To(Equal("#/components/schemas/testObject"))
		})

		It("should handle nested schema references", func() {
			schemaProxy := InterfaceToSchemaV3(doc, "[]string")
			arraySchema := schemaProxy.Schema()
			Expect(arraySchema.Type).To(Equal([]string{"array"}))
			Expect(arraySchema.Items.A.Schema().Type).To(Equal([]string{"string"}))
		})

		It("should handle nested-nested schema references", func() {
			schemaProxy := InterfaceToSchemaV3(doc, "[][]string")
			arraySchema := schemaProxy.Schema()
			Expect(arraySchema.Type).To(Equal([]string{"array"}))
			nestedArraySchema := arraySchema.Items.A.Schema()
			Expect(nestedArraySchema.Type).To(Equal([]string{"array"}))
			Expect(nestedArraySchema.Items.A.Schema().Type).To(Equal([]string{"string"}))
		})

		It("should handle nested-nested-nested int references", func() {
			schemaProxy := InterfaceToSchemaV3(doc, "[][][]int")
			arraySchema := schemaProxy.Schema()
			Expect(arraySchema.Type).To(Equal([]string{"array"}))
			nestedArraySchema := arraySchema.Items.A.Schema()
			Expect(nestedArraySchema.Type).To(Equal([]string{"array"}))
			deeperArraySchema := nestedArraySchema.Items.A.Schema()
			Expect(deeperArraySchema.Type).To(Equal([]string{"array"}))
			Expect(deeperArraySchema.Items.A.Schema().Type).To(Equal([]string{"integer"}))
		})

		It("should handle nested schema references with objects", func() {
			schemaProxy := InterfaceToSchemaV3(doc, "[]testObject")
			arraySchema := schemaProxy.Schema()
			Expect(arraySchema.Type).To(Equal([]string{"array"}))
			Expect(arraySchema.Items.A.GetReference()).To(Equal("#/components/schemas/testObject"))
		})
	})

	Describe("ToResponseDescription", func() {
		It("should return a space for empty description", func() {
			Expect(ToResponseDescription("")).To(Equal(" "))
		})

		It("should return the original description for non-empty strings", func() {
			Expect(ToResponseDescription("Success")).To(Equal("Success"))
			Expect(ToResponseDescription("Not Found")).To(Equal("Not Found"))
			Expect(ToResponseDescription("Some detailed description")).To(Equal("Some detailed description"))
		})

		It("should handle description with only whitespace", func() {
			Expect(ToResponseDescription("   ")).To(Equal("   "))
			Expect(ToResponseDescription("\t")).To(Equal("\t"))
			Expect(ToResponseDescription("\n")).To(Equal("\n"))
		})
	})
})
