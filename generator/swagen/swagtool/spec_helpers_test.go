package swagtool

import (
	"github.com/gopher-fleece/gleece/v2/definitions"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Spec Helpers", func() {

	It("Should add Rfc7807Error when default error in use", func() {
		models := []definitions.StructMetadata{}
		AppendErrorSchema(&models, true)
		Expect(len(models)).To(Equal(1))
	})
	It("Should not add error when default error not in use", func() {
		models := []definitions.StructMetadata{}
		AppendErrorSchema(&models, false)
		Expect(len(models)).To(Equal(0))
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
			Expect(ToOpenApiType("Time")).To(Equal("date-time"))
			Expect(ToOpenApiType("time.Time")).To(Equal("date-time"))
			Expect(ToOpenApiType("[]byte")).To(Equal("binary"))
			Expect(ToOpenApiType("customType")).To(Equal("object"))
		})

		It("should convert map types to 'map'", func() {
			Expect(ToOpenApiType("map[string]int")).To(Equal("map"))
			Expect(ToOpenApiType("map[string]string")).To(Equal("map"))
			Expect(ToOpenApiType("map[int]bool")).To(Equal("map"))
			Expect(ToOpenApiType("map[string][]int")).To(Equal("map"))
			Expect(ToOpenApiType("map[string]User")).To(Equal("map"))
			Expect(ToOpenApiType("map[string]interface{}")).To(Equal("map"))
			Expect(ToOpenApiType("map[int]map[string]bool")).To(Equal("map"))
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

	Describe("IsGenericObject", func() {
		It("should return false for map types", func() {
			Expect(IsGenericObject("map[string]int")).To(BeFalse())
			Expect(IsGenericObject("map[string]interface{}")).To(BeFalse())
			Expect(IsGenericObject("map[int]string")).To(BeFalse())
			Expect(IsGenericObject("map[string]any")).To(BeFalse())
		})

		It("should return false for non-generic types", func() {
			Expect(IsGenericObject("string")).To(BeFalse())
			Expect(IsGenericObject("[]string")).To(BeFalse())
			Expect(IsGenericObject("int")).To(BeFalse())
			Expect(IsGenericObject("model")).To(BeFalse())
			Expect(IsGenericObject("Model")).To(BeFalse())
		})

		It("should return false for partial map-like strings", func() {
			Expect(IsGenericObject("mapstring]")).To(BeFalse())
			Expect(IsGenericObject("[map]")).To(BeFalse())
		})

		It("should return true for generic types", func() {
			Expect(IsGenericObject("any")).To(BeTrue())
			Expect(IsGenericObject("interface{}")).To(BeTrue())
			Expect(IsGenericObject("")).To(BeTrue())
		})
	})

	Describe("IsFieldRequired", func() {
		It("should return true if 'required' is present in the validation string", func() {
			Expect(IsFieldRequired("required")).To(BeTrue())
			Expect(IsFieldRequired("min=1,required")).To(BeTrue())
		})

		It("should return false if 'required' is not present in the validation string", func() {
			Expect(IsFieldRequired("min=1,max=10")).To(BeFalse())
		})
	})

	Describe("GetArrayItemType", func() {
		It("should return the item type of an array", func() {
			Expect(GetArrayItemType("[]string")).To(Equal("string"))
			Expect(GetArrayItemType("[]int")).To(Equal("int"))
		})

		It("should return the sub array item type of an array", func() {
			Expect(GetArrayItemType("[][]string")).To(Equal("[]string"))
			Expect(GetArrayItemType("[][][]abc")).To(Equal("[][]abc"))
		})
	})

	Describe("GetMapItemType", func() {
		It("should return the value type of a map with string keys", func() {
			Expect(GetMapItemType("map[string]int")).To(Equal("int"))
			Expect(GetMapItemType("map[string]string")).To(Equal("string"))
			Expect(GetMapItemType("map[string]bool")).To(Equal("bool"))
		})

		It("should return the value type of a map with int keys", func() {
			Expect(GetMapItemType("map[int]string")).To(Equal("string"))
			Expect(GetMapItemType("map[int64]bool")).To(Equal("bool"))
			Expect(GetMapItemType("map[uint]float64")).To(Equal("float64"))
		})

		It("should return the value type for maps with complex values", func() {
			Expect(GetMapItemType("map[string][]int")).To(Equal("[]int"))
			Expect(GetMapItemType("map[string][]string")).To(Equal("[]string"))
			Expect(GetMapItemType("map[int][]bool")).To(Equal("[]bool"))
		})

		It("should return the value type for maps with custom object values", func() {
			Expect(GetMapItemType("map[string]User")).To(Equal("User"))
			Expect(GetMapItemType("map[int]CustomType")).To(Equal("CustomType"))
			Expect(GetMapItemType("map[string]models.Product")).To(Equal("models.Product"))
		})

		It("should handle nested map types", func() {
			Expect(GetMapItemType("map[string]map[int]string")).To(Equal("map[int]string"))
			Expect(GetMapItemType("map[int]map[string]bool")).To(Equal("map[string]bool"))
		})

		It("should handle maps with interface{} values", func() {
			Expect(GetMapItemType("map[string]interface{}")).To(Equal("interface{}"))
			Expect(GetMapItemType("map[string]any")).To(Equal("any"))
		})
	})

	Describe("GetJsonNameFromTag", func() {
		It("should extract simple json name correctly", func() {
			tag := `json:"userName"`
			value := GetJsonNameFromTag(tag, "default")
			Expect(value).To(Equal("userName"))
		})

		It("should handle json tag with omitempty", func() {
			tag := `json:"userName,omitempty"`
			value := GetJsonNameFromTag(tag, "default")
			Expect(value).To(Equal("userName"))
		})

		It("should handle json tag with multiple options", func() {
			tag := `json:"userName,omitempty,string"`
			value := GetJsonNameFromTag(tag, "default")
			Expect(value).To(Equal("userName"))
		})

		It("should return default name when json tag is empty", func() {
			tag := `json:""`
			value := GetJsonNameFromTag(tag, "default")
			Expect(value).To(Equal("default"))
		})

		It("should return default name when tag is missing", func() {
			tag := ``
			value := GetJsonNameFromTag(tag, "default")
			Expect(value).To(Equal("default"))
		})

		It("should handle tag with other fields", func() {
			tag := `json:"userName" validate:"required" binding:"required"`
			value := GetJsonNameFromTag(tag, "default")
			Expect(value).To(Equal("userName"))
		})

		It("should handle json tag with dash", func() {
			tag := `json:"-"`
			value := GetJsonNameFromTag(tag, "default")
			Expect(value).To(Equal("-"))
		})

		It("should handle json tag with special characters", func() {
			tag := `json:"user_name-123"`
			value := GetJsonNameFromTag(tag, "default")
			Expect(value).To(Equal("user_name-123"))
		})
	})

	Describe("HasEmbeddedField", func() {
		It("should return false for empty fields", func() {
			fields := []definitions.FieldMetadata{}
			Expect(HasEmbeddedField(fields)).To(BeFalse())
		})

		It("should return false when no fields are embedded", func() {
			fields := []definitions.FieldMetadata{
				{Name: "Field1", IsEmbedded: false},
				{Name: "Field2", IsEmbedded: false},
				{Name: "Field3", IsEmbedded: false},
			}
			Expect(HasEmbeddedField(fields)).To(BeFalse())
		})

		It("should return true when a single field is embedded", func() {
			fields := []definitions.FieldMetadata{
				{Name: "Field1", IsEmbedded: false},
				{Name: "Field2", IsEmbedded: true},
				{Name: "Field3", IsEmbedded: false},
			}
			Expect(HasEmbeddedField(fields)).To(BeTrue())
		})

		It("should return true when all fields are embedded", func() {
			fields := []definitions.FieldMetadata{
				{Name: "Field1", IsEmbedded: true},
				{Name: "Field2", IsEmbedded: true},
				{Name: "Field3", IsEmbedded: true},
			}
			Expect(HasEmbeddedField(fields)).To(BeTrue())
		})

		It("should return true when multiple fields are embedded", func() {
			fields := []definitions.FieldMetadata{
				{Name: "Field1", IsEmbedded: true},
				{Name: "Field2", IsEmbedded: false},
				{Name: "Field3", IsEmbedded: true},
			}
			Expect(HasEmbeddedField(fields)).To(BeTrue())
		})

		It("should return true when the first field is embedded", func() {
			fields := []definitions.FieldMetadata{
				{Name: "Field1", IsEmbedded: true},
				{Name: "Field2", IsEmbedded: false},
				{Name: "Field3", IsEmbedded: false},
			}
			Expect(HasEmbeddedField(fields)).To(BeTrue())
		})

		It("should return true when the last field is embedded", func() {
			fields := []definitions.FieldMetadata{
				{Name: "Field1", IsEmbedded: false},
				{Name: "Field2", IsEmbedded: false},
				{Name: "Field3", IsEmbedded: true},
			}
			Expect(HasEmbeddedField(fields)).To(BeTrue())
		})

		It("should return false when embedded field has type 'error'", func() {
			fields := []definitions.FieldMetadata{
				{Name: "Field1", Type: "string", IsEmbedded: false},
				{Name: "Field2", Type: "error", IsEmbedded: true},
				{Name: "Field3", Type: "int", IsEmbedded: false},
			}
			Expect(HasEmbeddedField(fields)).To(BeFalse())
		})

		It("should return true when embedded field has type other than 'error'", func() {
			fields := []definitions.FieldMetadata{
				{Name: "Field1", Type: "string", IsEmbedded: false},
				{Name: "Field2", Type: "CustomType", IsEmbedded: true},
				{Name: "Field3", Type: "int", IsEmbedded: false},
			}
			Expect(HasEmbeddedField(fields)).To(BeTrue())
		})

		It("should return false when all embedded fields are of type 'error'", func() {
			fields := []definitions.FieldMetadata{
				{Name: "Field1", Type: "error", IsEmbedded: true},
				{Name: "Field2", Type: "error", IsEmbedded: true},
				{Name: "Field3", Type: "int", IsEmbedded: false},
			}
			Expect(HasEmbeddedField(fields)).To(BeFalse())
		})

		It("should return true when at least one embedded field is not of type 'error'", func() {
			fields := []definitions.FieldMetadata{
				{Name: "Field1", Type: "error", IsEmbedded: true},
				{Name: "Field2", Type: "CustomType", IsEmbedded: true},
				{Name: "Field3", Type: "int", IsEmbedded: false},
			}
			Expect(HasEmbeddedField(fields)).To(BeTrue())
		})
	})
})
