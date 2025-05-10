package visitors_test

import (
	"github.com/gopher-fleece/runtime"
)

// @Tag(Receiver With No Method Annotation)
// @Route(/test/route)
// @Description Receiver With No Method Annotation
type ReceiverWithNoMethodAnnotation struct {
	runtime.GleeceController
}

// @Route(/)
func (rc *ReceiverWithNoMethodAnnotation) NoMethodAnnotation() error {
	return nil
}
