package sanity_test

import (
	"github.com/gopher-fleece/runtime"
)

// @Tag(Edge-Cases Controller Tag)
// @Route(/test/edge-cases)
// @Description Edge Cases Controller
type EdgeCasesController struct {
	runtime.GleeceController // Embedding the GleeceController to inherit its methods
}

// @Method(POST)
// @Route(/accepts-any)
// @Body(body)
func (ec *EdgeCasesController) ReceivesAny(body any) error {
	return nil
}

// @Method(POST)
// @Route(/returns-any)
func (ec *EdgeCasesController) ReturnsAny() (any, error) {
	return nil, nil
}
