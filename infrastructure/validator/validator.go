package validator

import (
	"reflect"
	"unicode"

	go_validator "github.com/go-playground/validator/v10"
	"github.com/haimkastner/gleece/definitions"
)

// Global validator instance
var Validate *go_validator.Validate

// Custom validation function to check if the slice is not nil
func validateNotNilSlice(fl go_validator.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() == reflect.Slice {
		return !field.IsNil()
	}
	return false
}

// Custom validation function to check if a string starts with a letter
func validateStartsWithLetter(fl go_validator.FieldLevel) bool {
	field := fl.Field().String()
	if field == "" {
		return false
	}
	firstChar := rune(field[0])
	return unicode.IsLetter(firstChar)
}

// registerEnumValidator creates a custom validation function for enum types
func registerEnumValidator(enumValues interface{}) go_validator.Func {
	values := reflect.ValueOf(enumValues)
	allowedValues := make(map[interface{}]struct{})

	for i := 0; i < values.Len(); i++ {
		allowedValues[values.Index(i).Interface()] = struct{}{}
	}

	return func(fl go_validator.FieldLevel) bool {
		field := fl.Field().Interface()
		_, ok := allowedValues[field]
		return ok
	}
}

func InitValidator() {
	// Initialize the validator instance
	Validate = go_validator.New()

	// Register custom validation functions globally
	Validate.RegisterValidation("not_nil_array", validateNotNilSlice)
	Validate.RegisterValidation("starts_with_letter", validateStartsWithLetter)

	// Register enum validation functions

	// SecuritySchemeIn
	Validate.RegisterValidation("security_schema_in", registerEnumValidator([]definitions.SecuritySchemeIn{definitions.InQuery, definitions.InHeader, definitions.InCookie}))
	// SecuritySchemeType
	Validate.RegisterValidation("security_schema_type", registerEnumValidator([]definitions.SecuritySchemeType{definitions.APIKey, definitions.OAuth2, definitions.OpenIDConnect, definitions.HTTP}))
}
