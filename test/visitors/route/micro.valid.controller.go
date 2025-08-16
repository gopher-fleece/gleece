package route_test

import (
	"github.com/gopher-fleece/runtime"
)

// @Tag(Controller for Route Visitor Tests)
// @Route(/test/route)
// @Description Controller for Route Visitor Tests
type RouteVisitorTestController struct {
	runtime.GleeceController
}

// @Method(GET)
// @Route(/)
func (rc *RouteVisitorTestController) Receiver1() {
}
