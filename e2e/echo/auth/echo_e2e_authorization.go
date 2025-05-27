package auth

import (
	"context"
	"strconv"

	"github.com/gopher-fleece/gleece/e2e/assets"
	"github.com/gopher-fleece/runtime"
	"github.com/labstack/echo/v4"
)

func GleeceRequestAuthorization(ctx context.Context, echoCtx echo.Context, check runtime.SecurityCheck) (context.Context, *runtime.SecurityError) {
	finalCtx := context.WithValue(ctx, assets.ContextAuth, "123")

	// A WA to set the header for the test with the given LAST run scope
	echoCtx.Request().Header.Set("x-test-scopes", check.SchemaName+check.Scopes[0])
	// Simulate auth failed
	authCode := 401

	failCodeStr := echoCtx.Request().Header.Get("fail-code")
	if failCodeStr != "" {
		num, _ := strconv.Atoi(failCodeStr)
		authCode = num
	}

	if echoCtx.Request().Header.Get("fail-auth") == check.SchemaName {
		return finalCtx, &runtime.SecurityError{
			Message:    "Failed to authorize",
			StatusCode: runtime.HttpStatusCode(authCode),
		}
	}

	// Simulate auth failed with custom error
	if echoCtx.Request().Header.Get("fail-auth-custom") == check.SchemaName {
		return finalCtx, &runtime.SecurityError{
			Message:    "Failed to authorize",
			StatusCode: runtime.HttpStatusCode(authCode),
			CustomError: &runtime.CustomError{
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
	return finalCtx, nil
}
