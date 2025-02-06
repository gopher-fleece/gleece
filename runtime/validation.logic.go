package runtime

import (
	"fmt"

	validator "github.com/go-playground/validator/v10"
)

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
