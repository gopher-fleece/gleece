package visitors_test

import (
	"github.com/gopher-fleece/runtime"
)

// @Tag(Receiver With No Comments)
// @Route(/test/route)
// @Description Receiver With No Comments
type ReceiverWithInvalidJson struct {
	runtime.GleeceController
}

func (rc *ReceiverWithInvalidJson) NotAnApiMethod() error {
	return nil
}
