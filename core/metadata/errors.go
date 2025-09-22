package metadata

import (
	"errors"
)

type InvalidAnnotationError struct {
	error
	AnnotationName string
	Value          string
}

func NewInvalidAnnotationError(message string) InvalidAnnotationError {
	return InvalidAnnotationError{
		error: errors.New(message),
	}
}
