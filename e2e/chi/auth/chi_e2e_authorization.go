package auth

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gopher-fleece/gleece/v2/e2e/assets"
	"github.com/gopher-fleece/runtime"
)

func GleeceRequestAuthorization(ctx context.Context, r *http.Request, check runtime.SecurityCheck) (context.Context, *runtime.SecurityError) {
	finalCtx := context.WithValue(ctx, assets.ContextAuth, "123")
	// A WA to set the header for the test with the given LAST run scope
	r.Header.Set("x-test-scopes", check.SchemaName+check.Scopes[0])
	// Simulate auth failed
	authCode := 401

	failCodeStr := r.Header.Get("fail-code")
	if failCodeStr != "" {
		num, _ := strconv.Atoi(failCodeStr)
		authCode = num
	}

	if r.Header.Get("fail-auth") == check.SchemaName {
		return finalCtx, &runtime.SecurityError{
			Message:    "Failed to authorize",
			StatusCode: runtime.HttpStatusCode(authCode),
		}
	}

	// Simulate auth failed with custom error
	if r.Header.Get("fail-auth-custom") == check.SchemaName {
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
