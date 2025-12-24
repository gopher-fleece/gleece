package sanity_test

import (
	"github.com/gopher-fleece/gleece/v2/test/units"
	"github.com/gopher-fleece/runtime"
)

// Some comment
// @Description This should be the actual description
type SimpleResponseModel struct {
	// A description for the value
	SomeValue int `validate:"required,min=0,max=10"`
}

// @Tag(Sanity Controller Tag)
// @Route(/test/sanity)
// @Description Sanity Controller
type SanityController struct {
	runtime.GleeceController // Embedding the GleeceController to inherit its methods
}

// A sanity test controller method
// @Method(POST)
// @Route(/valid-method-simple-route-query-and-header-params/{routeParamAlias})
// @Path(routeParam, {name: "routeParamAlias"})
// @Query(queryParam)
// @Header(headerParam)
// @Response(200) Description for HTTP 200
// @ErrorResponse(500) Code 500
// @ErrorResponse(502) Code 502
func (ec *SanityController) ValidMethodWithSimpleRouteQueryAndHeaderParameters(
	routeParam string,
	queryParam int,
	headerParam float32,
) (SimpleResponseModel, error) {
	return SimpleResponseModel{}, nil
}

// @Method(POST)
// @Route(/imports-from-other-package)
// @Body(body)
func (ec *SanityController) ImportsTypeFromOtherPackage(body []units.StructWithStructSlice) error {
	return nil
}
