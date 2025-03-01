package middlewares

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gopher-fleece/gleece/e2e/assets"
)

// MiddlewareBeforeOperation sets a header and optionally aborts the operation based on a request header.
func MiddlewareBeforeOperation(c *fiber.Ctx) bool {
	// Set the header for the test with the given LAST run scope.
	c.Response().Header.Set("X-pass-before-operation", "true")

	// Retrieve the "abort-before-operation" header value.
	abortBeforeOperation := string(c.Request().Header.Peek("abort-before-operation"))
	if abortBeforeOperation == "true" {
		// Send a JSON response with status 400.
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "abort-before-operation header is set to true",
		})
		return false
	}

	return true
}

// MiddlewareAfterOperationSuccess sets a header and optionally aborts after a successful operation.
func MiddlewareAfterOperationSuccess(c *fiber.Ctx) bool {
	c.Response().Header.Set("X-pass-after-succeed-operation", "true")

	// Retrieve the "abort-after-operation" header value.
	abortAfterOperationSuccess := string(c.Request().Header.Peek("abort-after-operation"))
	if abortAfterOperationSuccess == "true" {
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "abort-after-operation header is set to true",
		})
		return false
	}
	return true
}

// MiddlewareOnError sets a header and optionally aborts on error.
func MiddlewareOnError(c *fiber.Ctx, err error) bool {
	c.Response().Header.Set("X-pass-on-error", "true")

	abortOnError := string(c.Request().Header.Peek("abort-on-error"))
	if abortOnError == "true" {
		operationErr := ""
		// Handle different error types.
		switch e := err.(type) {
		case *fiber.Error:
			operationErr = e.Message
		case assets.CustomError:
			operationErr = e.Message
		default:
			operationErr = err.Error()
		}
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "abort-on-error header is set to true " + operationErr,
		})
		return false
	}
	return true
}

// MiddlewareOnError2 sets a header for error handling without aborting.
func MiddlewareOnError2(c *fiber.Ctx, err error) bool {
	c.Response().Header.Set("X-pass-on-error-2", "true")
	return true
}

func MiddlewareOnValidationError(c *fiber.Ctx, err error) bool {
	c.Response().Header.Set("X-pass-error-validation", "true")

	abortOnError := string(c.Request().Header.Peek("abort-on-error"))
	if abortOnError == "true" {
		operationErr := ""
		// Handle different error types
		switch e := err.(type) {
		case *fiber.Error:
			operationErr = e.Message
		case error:
			operationErr = e.Error()
		}
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "abort-on-error header is set to true " + operationErr,
		})
		return false
	}
	return true
}

func MiddlewareOnOutputValidationError(c *fiber.Ctx, err error) bool {
	c.Response().Header.Set("X-pass-output-validation", "true")

	returnNull := string(c.Request().Header.Peek("x-return-null"))
	if returnNull == "true" {
		c.Status(http.StatusOK).JSON(nil)
		return false
	}

	abortOnError := string(c.Request().Header.Peek("abort-on-error"))
	if abortOnError == "true" {
		operationErr := ""
		// Handle different error types
		switch e := err.(type) {
		case *fiber.Error:
			operationErr = e.Message
		case error:
			operationErr = e.Error()
		}
		c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "abort-on-error header is set to true " + operationErr,
		})
		return false
	}
	return true
}
