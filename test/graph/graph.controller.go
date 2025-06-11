package graph_test

import (
	"github.com/gopher-fleece/gleece/test/units"
	"github.com/gopher-fleece/runtime"
)

// @Tag(Graph Controller Tag)
// @Route(/test/sanity)
// @Description Graph Controller
type GraphController struct {
	runtime.GleeceController // Embedding the GleeceController to inherit its methods
}

// A symbol graph test controller method
// @Method(POST)
// @Route(/{routeParamAlias})
// @Path(routeParam, {name: "routeParamAlias"})
// @Query(queryParam)
// @Header(headerParam)
// @Response(200) Description for HTTP 200
// @ErrorResponse(500) Code 500
// @ErrorResponse(502) Code 502
func (ec *GraphController) ValidMethodWithSimpleRouteQueryAndHeaderParameters(
	routeParam string,
	queryParam int,
	headerParam float32,
) error {
	return nil
}

// @Method(POST)
// @Route(/r1)
// @Body(body)
func (ec *GraphController) MethodWithComplexExternalTypes(body units.StructWithStructSlice) (units.EnumTypeA, error) {
	return "", nil
}
