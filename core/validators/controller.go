package validators

import (
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/arbitrators"
	"github.com/gopher-fleece/gleece/definitions"
)

func ValidateController(
	gleeceConfig *definitions.GleeceConfig,
	packagesFacade *arbitrators.PackagesFacade,
	meta definitions.ControllerMetadata,
) []error {
	routeErrors := common.Map(meta.Routes, func(route definitions.RouteMetadata) []error {
		return ValidateRoute(gleeceConfig, packagesFacade, route)
	})

	return common.Flatten(routeErrors)
}
