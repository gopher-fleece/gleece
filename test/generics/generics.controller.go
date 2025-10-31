package sanity_test

import (
	"github.com/gopher-fleece/gleece/test/types"
	"github.com/gopher-fleece/runtime"
)

type BodyWithPrimitiveMap struct {
	Dict map[string]int
}

type MonoGenericStruct[T any] struct {
	Value T
}

type MultiGenericStruct[TA, TB any] struct {
	ValueA TA
	ValueB TB
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

// @Method(POST)
// @Route(/primitive-map-body)
// @Body(body)
func (ec *GenericsController) RecvWithPrimitiveMapBody(body map[string]float32) error {
	return nil
}

// @Method(POST)
// @Route(/primitive-map-body)
// @Body(body)
func (ec *GenericsController) RecvWithNonPrimitiveMapBody(body map[string]types.HoldsVeryNestedStructs) error {
	return nil
}

// @Method(POST)
// @Route(/mono-generic-struct-body)
// @Body(body)
func (ec *GenericsController) RecvWithMonoGenericStructBody(body MonoGenericStruct[string]) error {
	return nil
}

// @Method(POST)
// @Route(/multi-generic-struct-body)
// @Body(body)
func (ec *GenericsController) RecvWithMultiGenericStructBody(body MultiGenericStruct[string, int]) error {
	return nil
}

// @Method(POST)
// @Route(/multi-generic-struct-body)
func (ec *GenericsController) RecvWithMultiGenericStructResponse() (MultiGenericStruct[string, int], error) {
	return MultiGenericStruct[string, int]{}, nil
}
