package ctrl

import "context"

// GleeceController provides common functionality for controllers.
type GleeceController struct {
	statusCode *int
	headers    map[string]interface{}
	Ctx        context.Context
	request    any
}

func (gc *GleeceController) SetRequest(request any) {
	gc.request = request
}

// SetStatus sets the status code for the GleeceController.
func (gc *GleeceController) SetStatus(statusCode int) {
	gc.statusCode = &statusCode
}

// GetStatus gets the status code for the GleeceController.
func (gc *GleeceController) GetStatus() *int {
	return gc.statusCode
}

// SetHeader sets a header for the GleeceController.
func (gc *GleeceController) SetHeader(name string, value interface{}) {
	gc.headers[name] = value
}
