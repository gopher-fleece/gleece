package auth

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gopher-fleece/gleece/external"
)

func GleeceRequestAuthorization(c *fiber.Ctx, check external.SecurityCheck) *external.SecurityError {
	// Set the header for the test with the given LAST run scope.
	// Fiber gives you access to the underlying fasthttp request.
	c.Request().Header.Set("x-test-scopes", check.SchemaName+check.Scopes[0])

	// Simulate auth failed
	authCode := 401

	// Retrieve the "fail-code" header. Convert the byte slice to string.
	failCodeStr := string(c.Request().Header.Peek("fail-code"))
	if failCodeStr != "" {
		num, _ := strconv.Atoi(failCodeStr)
		authCode = num
	}

	// Check if the "fail-auth" header equals the schema name in the check.
	if string(c.Request().Header.Peek("fail-auth")) == check.SchemaName {
		return &external.SecurityError{
			Message:    "Failed to authorize",
			StatusCode: external.HttpStatusCode(authCode),
		}
	}

	// Simulate auth failed with a custom error
	if string(c.Request().Header.Peek("fail-auth-custom")) == check.SchemaName {
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
