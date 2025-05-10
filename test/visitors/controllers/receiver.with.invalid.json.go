package visitors_test

import (
	"github.com/gopher-fleece/runtime"
)

// @Tag(Receiver With Invalid Json)
// @Route(/test/route)
// @Description Receiver With Invalid Json
type ReceiverWithNoComments struct {
	runtime.GleeceController
}

// @Query(id, {' Invalid JSON5 })
func (rc *ReceiverWithNoComments) HasInvalidJson() error {
	return nil
}
