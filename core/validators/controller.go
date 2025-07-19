package validators

import (
	"fmt"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/arbitrators"
	"github.com/gopher-fleece/gleece/definitions"
)

func ValidateController(
	gleeceConfig *definitions.GleeceConfig,
	packagesFacade *arbitrators.PackagesFacade,
	meta definitions.ControllerMetadata,
) error {
	routeErrors := common.MapNonZero(meta.Routes, func(route definitions.RouteMetadata) error {
		return ValidateRoute(gleeceConfig, packagesFacade, route)
	})

	if len(routeErrors) > 0 {
		return &common.ContextualError{
			Context: fmt.Sprintf("Controller %s", meta.Name),
			Errors:  routeErrors,
		}
	}
	return nil
}
