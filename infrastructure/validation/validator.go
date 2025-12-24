package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/gopher-fleece/gleece/v2/definitions"
)

var validatorInstance *validator.Validate

// Custom validation function to check if the slice is not nil
func validateNotNilSlice(fl validator.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() == reflect.Slice {
		return !field.IsNil()
	}
	return false
}

// Custom validation function to check if a string starts with a letter
func validateStartsWithLetter(fl validator.FieldLevel) bool {
	field := fl.Field().String()
	if field == "" {
		return true // for empty validation, pass "required" tag too
	}
	firstChar := rune(field[0])
	return unicode.IsLetter(firstChar)
}

// registerEnumValidator creates a custom validation function for enum types
func registerEnumValidator(enumValues any) validator.Func {
	values := reflect.ValueOf(enumValues)
	allowedValues := make(map[any]struct{})

	for i := 0; i < values.Len(); i++ {
		allowedValues[values.Index(i).Interface()] = struct{}{}
	}

	return func(fl validator.FieldLevel) bool {
		field := fl.Field().Interface()
		_, ok := allowedValues[field]
		return ok
	}
}

func validateRegex(fl validator.FieldLevel) bool {
	// Get the regex pattern from the tag's parameter
	params := strings.SplitN(fl.Param(), ":", 2)
	if len(params) < 1 {
		return false // Invalid tag usage
	}

	// Compile the regex
	re, err := regexp.Compile(params[0])
	if err != nil {
		return false // Invalid regex
	}

	// Validate the field value
	value := fl.Field().String()
	return re.MatchString(value)
}

func initValidator() {
	// Initialize the validator instance
	validatorInstance = validator.New()

	// Register custom validation functions globally
	validatorInstance.RegisterValidation("not_nil_array", validateNotNilSlice)
	validatorInstance.RegisterValidation("starts_with_letter", validateStartsWithLetter)
	validatorInstance.RegisterValidation("regex", validateRegex)

	// Register enum validation functions

	// SecuritySchemeIn
	validatorInstance.RegisterValidation(
		"security_schema_in",
		registerEnumValidator(
			[]definitions.SecuritySchemeIn{
				definitions.Empty,
				definitions.InQuery,
				definitions.InHeader,
				definitions.InCookie,
			},
		),
	)

	// SecuritySchemeType
	validatorInstance.RegisterValidation(
		"security_schema_type",
		registerEnumValidator(
			[]definitions.SecuritySchemeType{
				definitions.APIKey,
				definitions.OAuth2,
				definitions.OpenIDConnect,
				definitions.HTTP,
			},
		),
	)
}

func ValidateStruct(s interface{}) error {
	if validatorInstance == nil {
		initValidator()
	}
	return validatorInstance.Struct(s)
}

func ExtractValidationErrorMessage(err error, fieldName *string) string {
	if err == nil {
		return ""
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err.Error()
	}

	var errStr string
	for _, validationErr := range validationErrors {
		fName := validationErr.Field()
		if fieldName != nil {
			fName = *fieldName
		}
		errStr += fmt.Sprintf("Field '%s' failed validation with tag '%s'. ", fName, validationErr.Tag())
	}

	return errStr
}
