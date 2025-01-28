package swagen30

import (
	"github.com/getkin/kin-openapi/openapi3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validation Converter Spec", func() {

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

		Context("String format validations", func() {
			It("should apply uuid format validation", func() {
				BuildSchemaValidation(schema, "uuid", "string")
				Expect(schema.Value.Format).To(Equal("uuid"))
			})

			It("should apply ip format validation", func() {
				BuildSchemaValidation(schema, "ip", "string")
				Expect(schema.Value.Format).To(Equal("ipv4"))
			})

			It("should apply ipv4 format validation", func() {
				BuildSchemaValidation(schema, "ipv4", "string")
				Expect(schema.Value.Format).To(Equal("ipv4"))
			})

			It("should apply ipv6 format validation", func() {
				BuildSchemaValidation(schema, "ipv6", "string")
				Expect(schema.Value.Format).To(Equal("ipv6"))
			})

			It("should apply hostname format validation", func() {
				BuildSchemaValidation(schema, "hostname", "string")
				Expect(schema.Value.Format).To(Equal("hostname"))
			})

			It("should apply date format validation", func() {
				BuildSchemaValidation(schema, "date", "string")
				Expect(schema.Value.Format).To(Equal("date"))
			})

			It("should apply datetime format validation", func() {
				BuildSchemaValidation(schema, "datetime", "string")
				Expect(schema.Value.Format).To(Equal("date-time"))
			})
		})

		Context("Numeric validations", func() {
			It("should apply greater than validation", func() {
				BuildSchemaValidation(schema, "gt=10", "int")
				Expect(*schema.Value.Min).To(BeEquivalentTo(10))
				Expect(schema.Value.ExclusiveMin).To(BeTrue())
			})

			It("should apply greater than or equal validation", func() {
				BuildSchemaValidation(schema, "gte=10", "int")
				Expect(*schema.Value.Min).To(BeEquivalentTo(10))
				Expect(schema.Value.ExclusiveMin).To(BeFalse())
			})

			It("should apply less than validation", func() {
				BuildSchemaValidation(schema, "lt=20", "int")
				Expect(*schema.Value.Max).To(BeEquivalentTo(20))
				Expect(schema.Value.ExclusiveMax).To(BeTrue())
			})

			It("should apply less than or equal validation", func() {
				BuildSchemaValidation(schema, "lte=20", "int")
				Expect(*schema.Value.Max).To(BeEquivalentTo(20))
				Expect(schema.Value.ExclusiveMax).To(BeFalse())
			})

			It("should apply min value validation for numbers", func() {
				BuildSchemaValidation(schema, "min=5", "int")
				Expect(*schema.Value.Min).To(BeEquivalentTo(5))
				Expect(schema.Value.ExclusiveMin).To(BeFalse())
			})

			It("should apply max value validation for numbers", func() {
				BuildSchemaValidation(schema, "max=10", "int")
				Expect(*schema.Value.Max).To(BeEquivalentTo(10))
				Expect(schema.Value.ExclusiveMax).To(BeFalse())
			})

			It("should handle multiple numeric validations", func() {
				BuildSchemaValidation(schema, "gt=5,lt=10", "int")
				Expect(*schema.Value.Min).To(BeEquivalentTo(5))
				Expect(*schema.Value.Max).To(BeEquivalentTo(10))
				Expect(schema.Value.ExclusiveMin).To(BeTrue())
				Expect(schema.Value.ExclusiveMax).To(BeTrue())
			})
		})

		Context("String validations", func() {
			It("should apply min length validation for strings", func() {
				BuildSchemaValidation(schema, "min=5", "string")
				Expect(schema.Value.MinLength).To(BeEquivalentTo(5))
			})

			It("should apply max length validation for strings", func() {
				BuildSchemaValidation(schema, "max=10", "string")
				Expect(*schema.Value.MaxLength).To(BeEquivalentTo(10))
			})

			It("should apply len validation for strings", func() {
				BuildSchemaValidation(schema, "len=8", "string")
				Expect(schema.Value.MinLength).To(BeEquivalentTo(8))
				Expect(*schema.Value.MaxLength).To(BeEquivalentTo(8))
			})

			It("should apply pattern validation for strings", func() {
				BuildSchemaValidation(schema, "pattern=^[a-z]+$", "string")
				Expect(schema.Value.Pattern).To(Equal("^[a-z]+$"))
			})

			It("should not apply numeric validations to strings", func() {
				BuildSchemaValidation(schema, "gt=10", "string")
				Expect(schema.Value.Min).To(BeNil())
				Expect(schema.Value.ExclusiveMin).To(BeFalse())
			})
		})

		Context("Array validations", func() {
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

			It("should not apply numeric validations to arrays", func() {
				BuildSchemaValidation(schema, "gt=10", "[]int")
				Expect(schema.Value.Min).To(BeNil())
				Expect(schema.Value.ExclusiveMin).To(BeFalse())
			})
		})

		Context("Enum validations", func() {
			It("should apply enum validation for strings", func() {
				BuildSchemaValidation(schema, "enum=a|b|c", "string")
				Expect(schema.Value.Enum).To(ConsistOf("a", "b", "c"))
			})

			It("should handle empty enum values", func() {
				BuildSchemaValidation(schema, "enum=", "string")
				Expect(schema.Value.Enum).To(BeEmpty())
			})
		})

		Context("Multiple validations", func() {
			It("should handle multiple validation rules correctly", func() {
				BuildSchemaValidation(schema, "min=5,max=10,pattern=^[a-z]+$", "string")
				Expect(schema.Value.MinLength).To(BeEquivalentTo(5))
				Expect(*schema.Value.MaxLength).To(BeEquivalentTo(10))
				Expect(schema.Value.Pattern).To(Equal("^[a-z]+$"))
			})
		})

		Context("OneOf validations", func() {
			It("should apply oneof validation for strings", func() {
				BuildSchemaValidation(schema, "oneof=pending active completed", "string")
				Expect(schema.Value.Enum).To(ConsistOf("pending", "active", "completed"))
			})

			It("should apply oneof validation for integers", func() {
				BuildSchemaValidation(schema, "oneof=1 2 3", "int")
				Expect(schema.Value.Enum).To(ConsistOf(int64(1), int64(2), int64(3)))
			})

			It("should apply oneof validation for numbers", func() {
				BuildSchemaValidation(schema, "oneof=1.5 2.5 3.5", "float64")
				Expect(schema.Value.Enum).To(ConsistOf(float64(1.5), float64(2.5), float64(3.5)))
			})

			It("should handle empty oneof values", func() {
				BuildSchemaValidation(schema, "oneof=", "string")
				Expect(schema.Value.Enum).To(BeEmpty())
			})

			It("should handle oneof with single value", func() {
				BuildSchemaValidation(schema, "oneof=single", "string")
				Expect(schema.Value.Enum).To(ConsistOf("single"))
			})

			It("should handle oneof with multiple spaces", func() {
				BuildSchemaValidation(schema, "oneof=a   b     c", "string")
				Expect(schema.Value.Enum).To(ConsistOf("a", "b", "c"))
			})

			It("should handle oneof with mixed numeric values", func() {
				BuildSchemaValidation(schema, "oneof=1 invalid 3", "int")
				// Should only include valid integers
				Expect(schema.Value.Enum).To(ConsistOf(int64(1), int64(3)))
			})

			It("should handle oneof while other validations exist", func() {
				BuildSchemaValidation(schema, "required,oneof=a b c,min=1", "string")
				Expect(schema.Value.Enum).To(ConsistOf("a", "b", "c"))
			})
		})
	})

})
