package errorhandling_test

import (
	"github.com/gopher-fleece/runtime"
)

// @Tag(Dummy Controller Tag)
// @Route(/test/sanity)
// @Description Sanity Controller
type DummyController struct {
	runtime.GleeceController // Embedding the GleeceController to inherit its methods
}

// @Method(POST)
// @Route(/some/method)
func (ec *DummyController) EmptyMethod() error {
	return nil
}
