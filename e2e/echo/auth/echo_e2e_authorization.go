package auth

import (
	"strconv"

	"github.com/gopher-fleece/gleece/external"
	"github.com/labstack/echo/v4"
)

func GleeceRequestAuthorization(ctx echo.Context, check external.SecurityCheck) *external.SecurityError {
	// A WA to set the header for the test with the given LAST run scope
	ctx.Request().Header.Set("x-test-scopes", check.SchemaName+check.Scopes[0])
	// Simulate auth failed
	authCode := 401

	failCodeStr := ctx.Request().Header.Get("fail-code")
	if failCodeStr != "" {
		num, _ := strconv.Atoi(failCodeStr)
		authCode = num
	}

	if ctx.Request().Header.Get("fail-auth") == check.SchemaName {
		return &external.SecurityError{
			Message:    "Failed to authorize",
			StatusCode: external.HttpStatusCode(authCode),
		}
	}

	// Simulate auth failed with custom error
	if ctx.Request().Header.Get("fail-auth-custom") == check.SchemaName {
		return &external.SecurityError{
			Message:    "Failed to authorize",
			StatusCode: external.HttpStatusCode(authCode),
			CustomError: &external.CustomError{
				Payload: struct {
					Message     string `json:"message"`
					Description string `json:"description"`
				}{
					Message:     "Custom error message",
					Description: "Custom error description",
				},
			},
		}
	}
	return nil
}
