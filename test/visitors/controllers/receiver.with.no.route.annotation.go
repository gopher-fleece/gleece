package visitors_test

import (
	"github.com/gopher-fleece/runtime"
)

// @Tag(Receiver With No Route Annotation)
// @Route(/test/route)
// @Description Receiver With No Route Annotation
type ReceiverWithNoRouteAnnotation struct {
	runtime.GleeceController
}

// @Method(GET)
func (rc *ReceiverWithNoRouteAnnotation) NoRouteAnnotation() error {
	return nil
}
