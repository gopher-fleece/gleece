package errorhandling_test

import (
	"github.com/gopher-fleece/gleece/external"
)

// @Tag(Dummy Controller Tag)
// @Route(/test/sanity)
// @Description Sanity Controller
type DummyController struct {
	external.GleeceController // Embedding the GleeceController to inherit its methods
}

// @Method(POST)
// @Route(/some/method)
func (ec *DummyController) EmptyMethod() error {
	return nil
}
