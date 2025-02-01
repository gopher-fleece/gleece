package assets

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gopher-fleece/gleece/external"
)

func GleeceRequestAuthorization(ctx *gin.Context, check external.SecurityCheck) *external.SecurityError {
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
		return &external.SecurityError{
			Message:    "Failed to authorize",
			StatusCode: external.HttpStatusCode(authCode),
		}
	}

	// Simulate auth failed with custom error
	if ctx.GetHeader("fail-auth-custom") == check.SchemaName {
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
