package imports_test

import (
	"github.com/gopher-fleece/runtime"
)

type SimpleCustomError struct {
	error
}

type ErrorDetails struct {
	Code   int
	Source string
}

type ComplexCustomError struct {
	Details        ErrorDetails
	AdditionalInfo string
	Epoch          uint64
	error
}

// @Tag(Errors Controller Tag)
// @Route(/test/errors)
type ErrorsController struct {
	runtime.GleeceController
}

// @Method(POST)
// @Route(/returns-a-simple-non-standard-error)
func (ec *ErrorsController) ReturnsASimpleCustomError() SimpleCustomError {
	return SimpleCustomError{}
}

// @Method(POST)
// @Route(/returns-a-complex-non-standard-error)
func (ec *ErrorsController) ReturnsAComplexCustomError() ComplexCustomError {
	return ComplexCustomError{}
}
