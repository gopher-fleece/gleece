package sanity_test

import (
	"github.com/gopher-fleece/runtime"
)

type BodyWithPrimitiveMap struct {
	Dict map[string]int
}

// @Tag(Generics Controller Tag)
// @Route(/test/generics)
// @Description Generics Controller
type GenericsController struct {
	runtime.GleeceController // Embedding the GleeceController to inherit its methods
}

// @Method(POST)
// @Route(/primitive-map-in-body)
// @Body(body)
func (ec *GenericsController) RecvWithPrimitiveMapInBody(body BodyWithPrimitiveMap) error {
	return nil
}

// @Method(POST)
// @Route(/primitive-map-return)
func (ec *GenericsController) RecvReturningAPrimitiveMap() (map[string]int, error) {
	return nil, nil
}

// This checks for composite de-duplication/diffing.
// Basically, all usages of an instantiated composite should have the same graph node
// @Method(POST)
// @Route(/other-primitive-map-return)
func (ec *GenericsController) RecvReturningAnotherPrimitiveMap() (map[string]string, error) {
	return nil, nil
}
