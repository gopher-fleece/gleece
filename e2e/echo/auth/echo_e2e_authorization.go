package auth

import (
	"github.com/gopher-fleece/gleece/external"
	"github.com/labstack/echo/v4"
)

func GleeceRequestAuthorization(ctx echo.Context, check external.SecurityCheck) *external.SecurityError {
	return nil
}
