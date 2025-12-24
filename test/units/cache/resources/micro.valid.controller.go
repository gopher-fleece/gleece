package controller_test

import (
	"github.com/gopher-fleece/gleece/v2/test/types"
	"github.com/gopher-fleece/runtime"
)

type SomeEnum int

const (
	Val1 SomeEnum = 1
	Val2 SomeEnum = 2
)

// @Tag(Controller for metadata cache)
// @Route(/test/metadata-cache)
// @Description Controller for metadata cache
type MetadataCacheTestController struct {
	runtime.GleeceController
}

// @Method(GET)
// @Route(/{v})
// @Path(v)
func (rc *MetadataCacheTestController) Receiver1(v SomeEnum) (types.ImportedWithDefaultAlias, error) {
	return types.ImportedWithDefaultAlias{}, nil
}
