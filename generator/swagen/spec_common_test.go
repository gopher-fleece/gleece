package swagen

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gopher-fleece/gleece/definitions"
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

	Describe("IsHiddenAsset", func() {
		It("should return false if hideOptions.Type is HideMethodNever", func() {
			hideOptions := definitions.MethodHideOptions{Type: definitions.HideMethodNever}
			Expect(IsHiddenAsset(&hideOptions)).To(BeFalse())
		})

		It("should return true if hideOptions.Type is HideMethodAlways", func() {
			hideOptions := definitions.MethodHideOptions{Type: definitions.HideMethodAlways}
			Expect(IsHiddenAsset(&hideOptions)).To(BeTrue())
		})

		It("should return false for other hideOptions.Type values (TODO: Check the condition)", func() {
			hideOptions := definitions.MethodHideOptions{Type: "someOtherType"}
			Expect(IsHiddenAsset(&hideOptions)).To(BeFalse())
		})

		It("should return false if no options passed", func() {
			Expect(IsHiddenAsset(nil)).To(BeFalse())
		})
	})

	Describe("IsDeprecated", func() {
		It("should return false if deprecationOptions is nil", func() {
			Expect(IsDeprecated(nil)).To(BeFalse())
		})

		It("should return false if deprecationOptions.Deprecated is false", func() {
			deprecationOptions := &definitions.DeprecationOptions{Deprecated: false}
			Expect(IsDeprecated(deprecationOptions)).To(BeFalse())
		})

		It("should return true if deprecationOptions.Deprecated is true", func() {
			deprecationOptions := &definitions.DeprecationOptions{Deprecated: true}
			Expect(IsDeprecated(deprecationOptions)).To(BeTrue())
		})
	})

	Describe("GetTagValue", func() {
		It("should extract json tag value correctly", func() {
			tag := `json:"houseNumber" validate:"gte=1"`
			value := GetTagValue(tag, "json", "default")
			Expect(value).To(Equal("houseNumber"))
		})

		It("should extract validate tag value correctly", func() {
			tag := `json:"houseNumber" validate:"gte=1"`
			value := GetTagValue(tag, "validate", "default")
			Expect(value).To(Equal("gte=1"))
		})

		It("should return default value when tag is not found", func() {
			tag := `json:"houseNumber" validate:"gte=1"`
			value := GetTagValue(tag, "nonexistent", "default")
			Expect(value).To(Equal("default"))
		})

		It("should handle empty tag value", func() {
			tag := `json:"" validate:"gte=1"`
			value := GetTagValue(tag, "json", "default")
			Expect(value).To(Equal("default"))
		})

		It("should handle empty tag string", func() {
			value := GetTagValue("", "json", "default")
			Expect(value).To(Equal("default"))
		})

		It("should handle tag with multiple values", func() {
			tag := `json:"house,omitempty" validate:"required"`
			value := GetTagValue(tag, "json", "default")
			Expect(value).To(Equal("house,omitempty"))
		})

		It("should handle tag value containing spaces", func() {
			tag := `json:"houseNumber" validate:"onefo=1 2 3"`
			value := GetTagValue(tag, "validate", "default")
			Expect(value).To(Equal("onefo=1 2 3"))
		})
	})

	Describe("IsMapObject", func() {
		It("should return true for map types", func() {
			Expect(IsMapObject("map[string]int")).To(BeTrue())
			Expect(IsMapObject("map[string]interface{}")).To(BeTrue())
			Expect(IsMapObject("map[int]string")).To(BeTrue())
		})

		It("should return false for non-map types", func() {
			Expect(IsMapObject("string")).To(BeFalse())
			Expect(IsMapObject("[]string")).To(BeFalse())
			Expect(IsMapObject("int")).To(BeFalse())
			Expect(IsMapObject("")).To(BeFalse())
		})

		It("should return false for partial map-like strings", func() {
			Expect(IsMapObject("mapstring]")).To(BeFalse())
			Expect(IsMapObject("[map]")).To(BeFalse())
		})
	})
})
