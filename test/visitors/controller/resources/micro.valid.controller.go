package controller_test

import (
	"github.com/gopher-fleece/runtime"
)

// @Tag(Controller for Controller Visitor Tests)
// @Route(/test/controller)
// @Description Controller for controller Visitor Tests
type ControllerVisitorTestController struct {
	runtime.GleeceController
}

// @Method(GET)
// @Route(/)
func (rc *ControllerVisitorTestController) Receiver1() {
}
