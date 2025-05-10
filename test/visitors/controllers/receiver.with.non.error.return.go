package visitors_test

import (
	"github.com/gopher-fleece/runtime"
)

// @Tag(Receiver With Non Error Return)
// @Route(/test/route)
// @Description Receiver With Non Error Return
type ReceiverWithNonErrorReturn struct {
	runtime.GleeceController
}

// @Method(GET)
// @Route(/)
func (rc *ReceiverWithNonErrorReturn) NonErrorReturn() bool {
	return true
}
