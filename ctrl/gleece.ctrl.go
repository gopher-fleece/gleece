package ctrl

import "context"

// GleeceController provides common functionality for controllers.
type GleeceController struct {
	statusCode *int
	headers    map[string]interface{}
	Ctx        context.Context
}

// SetStatus sets the status code for the GleeceController.
func (bc *GleeceController) SetStatus(statusCode int) {
	bc.statusCode = &statusCode
}

// GetStatus gets the status code for the GleeceController.
func (bc *GleeceController) GetStatus() *int {
	return bc.statusCode
}

// SetHeader sets a header for the GleeceController.
func (bc *GleeceController) SetHeader(name string, value interface{}) {
	bc.headers[name] = value
}
