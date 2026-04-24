package embed

import (
	_ "embed"
)

//go:embed auth.middleware.chi.hbs
var ChiAuthMiddleware string

//go:embed auth.middleware.echo.hbs
var EchoAuthMiddleware string

//go:embed auth.middleware.fiber.hbs
var FiberAuthMiddleware string

//go:embed auth.middleware.gin.hbs
var GinAuthMiddleware string

//go:embed auth.middleware.mux.hbs
var MuxAuthMiddleware string
