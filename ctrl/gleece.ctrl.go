package ctrl

import (
	"context"

	"github.com/haimkastner/gleece/definitions"
)

// GleeceController provides common functionality for controllers.
type GleeceController struct {
	statusCode *definitions.HttpStatusCode
	headers    map[string]interface{}
	Ctx        context.Context
	request    any
}

func (gc *GleeceController) SetRequest(request any) {
	gc.request = request
}

// SetStatus sets the status code for the GleeceController.
func (gc *GleeceController) SetStatus(statusCode definitions.HttpStatusCode) {
	gc.statusCode = &statusCode
}

// GetStatus gets the status code for the GleeceController.
func (gc *GleeceController) GetStatus() *definitions.HttpStatusCode {
	return gc.statusCode
}

// SetHeader sets a header for the GleeceController.
func (gc *GleeceController) SetHeader(name string, value interface{}) {
	gc.headers[name] = value
}
