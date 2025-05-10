package visitors_test

import (
	"github.com/gopher-fleece/runtime"
)

// @Tag(Receiver With Void Return)
// @Route(/test/route)
// @Description Receiver With Void Return
type ReceiverWithVoidReturn struct {
	runtime.GleeceController
}

// @Method(GET)
// @Route(/)
func (rc *ReceiverWithVoidReturn) VoidReturn() {
}
