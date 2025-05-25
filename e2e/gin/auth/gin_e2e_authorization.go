package assets

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gopher-fleece/gleece/e2e/assets"
	"github.com/gopher-fleece/runtime"
)

func GleeceRequestAuthorization(ctx context.Context, ginCtx *gin.Context, check runtime.SecurityCheck) (context.Context, *runtime.SecurityError) {

	finalCtx := context.WithValue(ctx, assets.ContextName, "123")

	// A WA to set the header for the test with the given LAST run scope
	ginCtx.Request.Header.Set("x-test-scopes", check.SchemaName+check.Scopes[0])
	// Simulate auth failed
	authCode := 401

	failCodeStr := ginCtx.GetHeader("fail-code")
	if failCodeStr != "" {
		num, _ := strconv.Atoi(failCodeStr)
		authCode = num
	}

	if ginCtx.GetHeader("fail-auth") == check.SchemaName {
		return finalCtx, &runtime.SecurityError{
			Message:    "Failed to authorize",
			StatusCode: runtime.HttpStatusCode(authCode),
		}
	}

	// Simulate auth failed with custom error
	if ginCtx.GetHeader("fail-auth-custom") == check.SchemaName {
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
