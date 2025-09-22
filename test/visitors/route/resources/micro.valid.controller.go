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

// @Method(GET)
// @Route(/2)
// @Query(paramWithInvalidType)
func (rc *RouteVisitorTestController) Receiver2(paramWithInvalidType chan *string) {
	// Used to test parameter type error flows
}

// @Method(GET)
// @Route(/2)
func (rc *RouteVisitorTestController) Receiver3() (chan *string, error) {
	// Used to test return types error flows
	return nil, nil
}
