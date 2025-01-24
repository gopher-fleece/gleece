package validation

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gopher-fleece/gleece/definitions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Define test structs to use in validation tests
type TestStruct struct {
	SliceField   []string                       `validate:"not_nil_array"`
	StringField  string                         `validate:"starts_with_letter"`
	RegexField   string                         `validate:"regex=^abc"`
	SecurityIn   definitions.SecuritySchemeIn   `validate:"security_schema_in"`
	SecurityType definitions.SecuritySchemeType `validate:"security_schema_type"`
}

var _ = Describe("Validation Utilities", func() {
	Describe("ValidateStruct", func() {
		It("should validate all fields correctly", func() {
			// Create a valid test struct
			validStruct := TestStruct{
				SliceField:   []string{},
				StringField:  "abc123",
				RegexField:   "abc123",
				SecurityIn:   definitions.InHeader,
				SecurityType: definitions.HTTP,
			}

			err := ValidateStruct(validStruct)
			Expect(err).To(BeNil())
		})

		It("should return errors for invalid fields", func() {
			// Create an invalid test struct
			invalidStruct := TestStruct{
				SliceField:   nil,
				StringField:  "123abc",
				RegexField:   "123abc",
				SecurityIn:   "invalid",
				SecurityType: "invalid",
			}

			// Validate the struct
			err := ValidateStruct(invalidStruct)
			Expect(err).To(HaveOccurred())

			// Check the validation errors
			validationErrors := err.(validator.ValidationErrors)
			Expect(validationErrors).To(HaveLen(5))

			Expect(validationErrors[0].Field()).To(Equal("SliceField"))
			Expect(validationErrors[1].Field()).To(Equal("StringField"))
			Expect(validationErrors[2].Field()).To(Equal("RegexField"))
			Expect(validationErrors[3].Field()).To(Equal("SecurityIn"))
			Expect(validationErrors[4].Field()).To(Equal("SecurityType"))
		})
	})

	Describe("ExtractValidationErrorMessage", func() {
		It("should return an empty string for no error", func() {
			err := error(nil)
			message := ExtractValidationErrorMessage(err, nil)
			Expect(message).To(Equal(""))
		})

		It("should return the error message for a non-validation error", func() {
			err := fmt.Errorf("some non-validation error")
			message := ExtractValidationErrorMessage(err, nil)
			Expect(message).To(Equal("some non-validation error"))
		})

		It("should return a readable message for validation errors", func() {
			// Create an invalid test struct
			invalidStruct := TestStruct{
				SliceField:   nil,
				StringField:  "123abc",
				RegexField:   "123abc",
				SecurityIn:   "invalid",
				SecurityType: "invalid",
			}

			// Validate the struct
			err := ValidateStruct(invalidStruct)
			Expect(err).To(HaveOccurred())

			// Extract readable validation error messages
			message := ExtractValidationErrorMessage(err, nil)
			Expect(message).To(ContainSubstring("Field 'SliceField' failed validation with tag 'not_nil_array'."))
			Expect(message).To(ContainSubstring("Field 'StringField' failed validation with tag 'starts_with_letter'."))
			Expect(message).To(ContainSubstring("Field 'RegexField' failed validation with tag 'regex'."))
			Expect(message).To(ContainSubstring("Field 'SecurityIn' failed validation with tag 'security_schema_in'."))
			Expect(message).To(ContainSubstring("Field 'SecurityType' failed validation with tag 'security_schema_type'."))
		})

		It("should return a readable message with overridden field name", func() {
			// Create an invalid test struct
			invalidStruct := TestStruct{
				SliceField:   nil,
				StringField:  "123abc",
				RegexField:   "123abc",
				SecurityIn:   "invalid",
				SecurityType: "invalid",
			}

			// Validate the struct
			err := ValidateStruct(invalidStruct)
			Expect(err).To(HaveOccurred())

			// Extract readable validation error messages with overridden field name
			overrideFieldName := "CustomField"
			message := ExtractValidationErrorMessage(err, &overrideFieldName)
			Expect(message).To(ContainSubstring("Field 'CustomField' failed validation with tag 'not_nil_array'."))
			Expect(message).To(ContainSubstring("Field 'CustomField' failed validation with tag 'starts_with_letter'."))
			Expect(message).To(ContainSubstring("Field 'CustomField' failed validation with tag 'regex'."))
			Expect(message).To(ContainSubstring("Field 'CustomField' failed validation with tag 'security_schema_in'."))
			Expect(message).To(ContainSubstring("Field 'CustomField' failed validation with tag 'security_schema_type'."))
		})
	})
})

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validation Suite")
}
