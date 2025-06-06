package auth

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gopher-fleece/gleece/e2e/assets"
	"github.com/gopher-fleece/runtime"
)

func GleeceRequestAuthorization(ctx context.Context, fiberCtx *fiber.Ctx, check runtime.SecurityCheck) (context.Context, *runtime.SecurityError) {
	finalCtx := context.WithValue(ctx, assets.ContextAuth, "123")

	// Set the header for the test with the given LAST run scope.
	// Fiber gives you access to the underlying fasthttp request.
	fiberCtx.Request().Header.Set("x-test-scopes", check.SchemaName+check.Scopes[0])

	// Simulate auth failed
	authCode := 401

	// Retrieve the "fail-code" header. Convert the byte slice to string.
	failCodeStr := string(fiberCtx.Request().Header.Peek("fail-code"))
	if failCodeStr != "" {
		num, _ := strconv.Atoi(failCodeStr)
		authCode = num
	}

	// Check if the "fail-auth" header equals the schema name in the check.
	if string(fiberCtx.Request().Header.Peek("fail-auth")) == check.SchemaName {
		return finalCtx, &runtime.SecurityError{
			Message:    "Failed to authorize",
			StatusCode: runtime.HttpStatusCode(authCode),
		}
	}

	// Simulate auth failed with a custom error
	if string(fiberCtx.Request().Header.Peek("fail-auth-custom")) == check.SchemaName {
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
