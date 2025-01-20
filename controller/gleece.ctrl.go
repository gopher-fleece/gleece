package ctrl

import (
	"github.com/gopher-fleece/gleece/external"
)

// GleeceController provides common functionality for controllers.
type GleeceController struct {
	statusCode *external.HttpStatusCode
	headers    map[string]interface{}
	// Request is the HTTP request from the underlying routing engine (gin, echo etc.)
	request any
}

func (gc *GleeceController) SetRequest(request any) {
	gc.request = request
}

// SetStatus sets the status code for the GleeceController.
func (gc *GleeceController) SetStatus(statusCode external.HttpStatusCode) {
	gc.statusCode = &statusCode
}

// GetStatus gets the status code for the GleeceController.
func (gc *GleeceController) GetStatus() *external.HttpStatusCode {
	return gc.statusCode
}

// SetHeader sets a header for the GleeceController.
func (gc *GleeceController) SetHeader(name string, value interface{}) {
	gc.headers[name] = value
}
