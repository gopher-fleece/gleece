package errorhandling_test

import (
	. "github.com/gopher-fleece/gleece/test/types"
	"github.com/gopher-fleece/runtime"
)

// @Tag(Dummy Controller Tag)
// @Route(/test/sanity)
// @Description Sanity Controller
type UnScannedTypeController struct {
	runtime.GleeceController // Embedding the GleeceController to inherit its methods
}

// @Method(POST)
// @Route(/some/method)
func (ec *UnScannedTypeController) EmptyMethod() (HoldsVeryNestedStructs, error) {
	return HoldsVeryNestedStructs{}, nil
}
