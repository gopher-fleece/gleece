package specials_test

import (
	"github.com/gopher-fleece/gleece/test/types"
	"github.com/gopher-fleece/runtime"
)

type ObjectWithMapField struct {
	MapField map[string]int
}

type ObjectWithNonPrimitiveMapValueField struct {
	MapField map[string]types.SomeNestedStruct
}

// @Tag(Special Types Controller Tag)
// @Route(/test/special-types)
type SpecialTypesController struct {
	runtime.GleeceController // Embedding the GleeceController to inherit its methods
}

// @Method(POST)
// @Route(/body-with-map-field)
// @Body(body)
func (ec *SpecialTypesController) BodyWithMapField(body ObjectWithMapField) error {
	return nil
}

// @Method(POST)
// @Route(/body-with-non-primitive-map-field)
// @Body(body)
func (ec *SpecialTypesController) BodyWithNonPrimitiveMapField(body ObjectWithNonPrimitiveMapValueField) error {
	return nil
}
