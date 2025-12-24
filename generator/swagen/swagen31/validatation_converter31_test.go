package swagen31

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"go.yaml.in/yaml/v4"
)

var _ = Describe("Validation Converter Spec", func() {

	Describe("BuildSchemaValidationV31", func() {
		var schema *base.Schema

		BeforeEach(func() {
			schema = &base.Schema{}
		})

		It("should apply email format validation", func() {
			BuildSchemaValidationV31(schema, "email", "string")
			Expect(schema.Format).To(Equal("email"))
		})

		It("should apply email format validation while other irrelevant exists", func() {
			BuildSchemaValidationV31(schema, "email,required,other=5", "string")
			Expect(schema.Format).To(Equal("email"))
		})

		Context("String format validations", func() {
			It("should apply uuid format validation", func() {
				BuildSchemaValidationV31(schema, "uuid", "string")
				Expect(schema.Format).To(Equal("uuid"))
			})

			It("should apply ip format validation", func() {
				BuildSchemaValidationV31(schema, "ip", "string")
				Expect(schema.Format).To(Equal("ipv4"))
			})

			It("should apply ipv4 format validation", func() {
				BuildSchemaValidationV31(schema, "ipv4", "string")
				Expect(schema.Format).To(Equal("ipv4"))
			})

			It("should apply ipv6 format validation", func() {
				BuildSchemaValidationV31(schema, "ipv6", "string")
				Expect(schema.Format).To(Equal("ipv6"))
			})

			It("should apply hostname format validation", func() {
				BuildSchemaValidationV31(schema, "hostname", "string")
				Expect(schema.Format).To(Equal("hostname"))
			})

			It("should apply date format validation", func() {
				BuildSchemaValidationV31(schema, "date", "string")
				Expect(schema.Format).To(Equal("date"))
			})

			It("should apply datetime format validation", func() {
				BuildSchemaValidationV31(schema, "datetime", "string")
				Expect(schema.Format).To(Equal("date-time"))
			})
		})

		Context("Numeric validations", func() {
			It("should apply greater than validation", func() {
				BuildSchemaValidationV31(schema, "gt=10", "int")
				val := float64(10)
				Expect(schema.ExclusiveMinimum.B).To(Equal(val))
				Expect(schema.ExclusiveMinimum.N).To(Equal(1))
			})

			It("should apply greater than or equal validation", func() {
				BuildSchemaValidationV31(schema, "gte=10", "int")
				val := float64(10)
				Expect(schema.Minimum).To(Equal(&val))
			})

			It("should apply less than validation", func() {
				BuildSchemaValidationV31(schema, "lt=20", "int")
				val := float64(20)
				Expect(schema.ExclusiveMaximum.B).To(Equal(val))
			})

			It("should apply less than or equal validation", func() {
				BuildSchemaValidationV31(schema, "lte=20", "int")
				val := float64(20)
				Expect(schema.Maximum).To(Equal(&val))
			})

			It("should apply min value validation for numbers", func() {
				BuildSchemaValidationV31(schema, "min=5", "int")
				val := float64(5)
				Expect(schema.Minimum).To(Equal(&val))
			})

			It("should apply max value validation for numbers", func() {
				BuildSchemaValidationV31(schema, "max=10", "int")
				val := float64(10)
				Expect(schema.Maximum).To(Equal(&val))
			})

			It("should handle multiple numeric validations", func() {
				BuildSchemaValidationV31(schema, "gt=5,lt=10", "int")
				valMin := float64(5)
				valMax := float64(10)
				Expect(schema.ExclusiveMinimum.B).To(Equal(valMin))
				Expect(schema.ExclusiveMaximum.B).To(Equal(valMax))
			})
		})

		Context("String validations", func() {
			It("should apply min length validation for strings", func() {
				BuildSchemaValidationV31(schema, "min=5", "string")
				val := int64(5)
				Expect(schema.MinLength).To(Equal(&val))
			})

			It("should apply max length validation for strings", func() {
				BuildSchemaValidationV31(schema, "max=10", "string")
				val := int64(10)
				Expect(schema.MaxLength).To(Equal(&val))
			})

			It("should apply len validation for strings", func() {
				BuildSchemaValidationV31(schema, "len=8", "string")
				val := int64(8)
				Expect(schema.MinLength).To(Equal(&val))
				Expect(schema.MaxLength).To(Equal(&val))
			})

			It("should apply pattern validation for strings", func() {
				BuildSchemaValidationV31(schema, "pattern=^[a-z]+$", "string")
				Expect(schema.Pattern).To(Equal("^[a-z]+$"))
			})

			It("should not apply numeric validations to strings", func() {
				BuildSchemaValidationV31(schema, "gt=10", "string")
				Expect(schema.Minimum).To(BeNil())
				Expect(schema.ExclusiveMinimum).To(BeNil())
			})
		})

		Context("Array validations", func() {
			It("should apply min items validation for arrays", func() {
				BuildSchemaValidationV31(schema, "minItems=2", "[]string")
				val := int64(2)
				Expect(schema.MinItems).To(Equal(&val))
			})

			It("should apply max items validation for arrays", func() {
				BuildSchemaValidationV31(schema, "maxItems=5", "[]int")
				val := int64(5)
				Expect(schema.MaxItems).To(Equal(&val))
			})

			It("should apply unique items validation for arrays", func() {
				BuildSchemaValidationV31(schema, "uniqueItems=true", "[]something")
				val := true
				Expect(schema.UniqueItems).To(Equal(&val))
			})

			It("should not apply numeric validations to arrays", func() {
				BuildSchemaValidationV31(schema, "gt=10", "[]int")
				Expect(schema.Minimum).To(BeNil())
				Expect(schema.ExclusiveMinimum).To(BeNil())
			})
		})

		Context("Enum validations", func() {
			It("should apply enum validation for strings", func() {
				BuildSchemaValidationV31(schema, "enum=a|b|c", "string")
				expectedEnums := []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "a"},
					{Kind: yaml.ScalarNode, Value: "b"},
					{Kind: yaml.ScalarNode, Value: "c"},
				}
				Expect(schema.Enum).To(HaveLen(3))
				for i, enum := range schema.Enum {
					Expect(enum.Kind).To(Equal(expectedEnums[i].Kind))
					Expect(enum.Value).To(Equal(expectedEnums[i].Value))
				}
			})

			It("should handle empty enum values", func() {
				BuildSchemaValidationV31(schema, "enum=", "string")
				Expect(schema.Enum).To(BeEmpty())
			})
		})

		Context("Multiple validations", func() {
			It("should handle multiple validation rules correctly", func() {
				BuildSchemaValidationV31(schema, "min=5,max=10,pattern=^[a-z]+$", "string")
				valMin := int64(5)
				valMax := int64(10)
				Expect(schema.MinLength).To(Equal(&valMin))
				Expect(schema.MaxLength).To(Equal(&valMax))
				Expect(schema.Pattern).To(Equal("^[a-z]+$"))
			})
		})

		Context("OneOf validations", func() {
			It("should apply oneof validation for strings", func() {
				BuildSchemaValidationV31(schema, "oneof=pending active completed", "string")
				expectedEnums := []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "pending"},
					{Kind: yaml.ScalarNode, Value: "active"},
					{Kind: yaml.ScalarNode, Value: "completed"},
				}
				Expect(schema.Enum).To(HaveLen(3))
				for i, enum := range schema.Enum {
					Expect(enum.Kind).To(Equal(expectedEnums[i].Kind))
					Expect(enum.Value).To(Equal(expectedEnums[i].Value))
				}
			})

			It("should apply oneof validation for integers", func() {
				BuildSchemaValidationV31(schema, "oneof=1 2 3", "int")
				expectedEnums := []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "1", Tag: "!!int"},
					{Kind: yaml.ScalarNode, Value: "2", Tag: "!!int"},
					{Kind: yaml.ScalarNode, Value: "3", Tag: "!!int"},
				}
				Expect(schema.Enum).To(HaveLen(3))
				for i, enum := range schema.Enum {
					Expect(enum.Kind).To(Equal(expectedEnums[i].Kind))
					Expect(enum.Value).To(Equal(expectedEnums[i].Value))
					Expect(enum.Tag).To(Equal(expectedEnums[i].Tag))
				}
			})

			It("should apply oneof validation for numbers", func() {
				BuildSchemaValidationV31(schema, "oneof=1.5 2.5 3.5", "float64")
				expectedEnums := []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "1.5", Tag: "!!float"},
					{Kind: yaml.ScalarNode, Value: "2.5", Tag: "!!float"},
					{Kind: yaml.ScalarNode, Value: "3.5", Tag: "!!float"},
				}
				Expect(schema.Enum).To(HaveLen(3))
				for i, enum := range schema.Enum {
					Expect(enum.Kind).To(Equal(expectedEnums[i].Kind))
					Expect(enum.Value).To(Equal(expectedEnums[i].Value))
					Expect(enum.Tag).To(Equal(expectedEnums[i].Tag))
				}
			})

			It("should handle empty oneof values", func() {
				BuildSchemaValidationV31(schema, "oneof=", "string")
				Expect(schema.Enum).To(BeEmpty())
			})

			It("should handle oneof with single value", func() {
				BuildSchemaValidationV31(schema, "oneof=single", "string")
				expectedEnum := &yaml.Node{Kind: yaml.ScalarNode, Value: "single"}
				Expect(schema.Enum).To(HaveLen(1))
				Expect(schema.Enum[0].Kind).To(Equal(expectedEnum.Kind))
				Expect(schema.Enum[0].Value).To(Equal(expectedEnum.Value))
			})

			It("should handle oneof with multiple spaces", func() {
				BuildSchemaValidationV31(schema, "oneof=a   b     c", "string")
				expectedEnums := []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "a"},
					{Kind: yaml.ScalarNode, Value: "b"},
					{Kind: yaml.ScalarNode, Value: "c"},
				}
				Expect(schema.Enum).To(HaveLen(3))
				for i, enum := range schema.Enum {
					Expect(enum.Kind).To(Equal(expectedEnums[i].Kind))
					Expect(enum.Value).To(Equal(expectedEnums[i].Value))
				}
			})

			It("should handle oneof with mixed numeric values", func() {
				BuildSchemaValidationV31(schema, "oneof=1 invalid 3", "int")
				expectedEnums := []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "1", Tag: "!!int"},
					{Kind: yaml.ScalarNode, Value: "3", Tag: "!!int"},
				}
				Expect(schema.Enum).To(HaveLen(2))
				for i, enum := range schema.Enum {
					Expect(enum.Kind).To(Equal(expectedEnums[i].Kind))
					Expect(enum.Value).To(Equal(expectedEnums[i].Value))
					Expect(enum.Tag).To(Equal(expectedEnums[i].Tag))
				}
			})

			It("should handle oneof while other validations exist", func() {
				BuildSchemaValidationV31(schema, "required,oneof=a b c,min=1", "string")
				expectedEnums := []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "a"},
					{Kind: yaml.ScalarNode, Value: "b"},
					{Kind: yaml.ScalarNode, Value: "c"},
				}
				Expect(schema.Enum).To(HaveLen(3))
				for i, enum := range schema.Enum {
					Expect(enum.Kind).To(Equal(expectedEnums[i].Kind))
					Expect(enum.Value).To(Equal(expectedEnums[i].Value))
				}
			})

			It("should handle oneof validation for non-standard types", func() {
				// Using a type that doesn't match string/integer/number
				BuildSchemaValidationV31(schema, "oneof=true false maybe", "bool")

				expectedEnums := []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "true"},
					{Kind: yaml.ScalarNode, Value: "false"},
					{Kind: yaml.ScalarNode, Value: "maybe"},
				}

				Expect(schema.Enum).To(HaveLen(3))
				for i, enum := range schema.Enum {
					Expect(enum.Kind).To(Equal(expectedEnums[i].Kind))
					Expect(enum.Value).To(Equal(expectedEnums[i].Value))
					// Note: No Tag is set in the default case
					Expect(enum.Tag).To(BeEmpty())
				}
			})
		})
	})
})
