package assets

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gopher-fleece/gleece/e2e/assets"
	"github.com/gopher-fleece/runtime"
)

func GleeceRequestAuthorization(ctx *gin.Context, check runtime.SecurityCheck) *runtime.SecurityError {

	// TODO: put this logic in the template, the "GleeceRequestAuthorization" should get the ctx and return ctx (along the SecurityError)
	reqCtx := ctx.Request.Context()
	finalCtx := context.WithValue(reqCtx, assets.ContextName, "123")
	// Replace gin's Request with a new one carrying the updated context
	ctx.Request = ctx.Request.WithContext(finalCtx)

	// A WA to set the header for the test with the given LAST run scope
	ctx.Request.Header.Set("x-test-scopes", check.SchemaName+check.Scopes[0])
	// Simulate auth failed
	authCode := 401

	failCodeStr := ctx.GetHeader("fail-code")
	if failCodeStr != "" {
		num, _ := strconv.Atoi(failCodeStr)
		authCode = num
	}

	if ctx.GetHeader("fail-auth") == check.SchemaName {
		return &runtime.SecurityError{
			Message:    "Failed to authorize",
			StatusCode: runtime.HttpStatusCode(authCode),
		}
	}

	// Simulate auth failed with custom error
	if ctx.GetHeader("fail-auth-custom") == check.SchemaName {
		return &runtime.SecurityError{
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
	return nil
}
