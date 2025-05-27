package middlewares

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gopher-fleece/gleece/e2e/assets"
)

// MiddlewareBeforeOperation sets a header and optionally aborts the operation based on a request header.
func MiddlewareBeforeOperation(ctx context.Context, fiberCtx *fiber.Ctx) (context.Context, bool) {
	// Set the header for the test with the given LAST run scope.
	fiberCtx.Response().Header.Set("X-pass-before-operation", "true")

	// Retrieve the "abort-before-operation" header value.
	abortBeforeOperation := string(fiberCtx.Request().Header.Peek("abort-before-operation"))
	if abortBeforeOperation == "true" {
		// Send a JSON response with status 400.
		fiberCtx.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "abort-before-operation header is set to true",
		})
		return ctx, false
	}

	return context.WithValue(ctx, assets.ContextMiddleware, "pass"), true
}

// MiddlewareAfterOperationSuccess sets a header and optionally aborts after a successful operation.
func MiddlewareAfterOperationSuccess(ctx context.Context, fiberCtx *fiber.Ctx) (context.Context, bool) {
	fiberCtx.Response().Header.Set("X-pass-after-succeed-operation", "true")

	// Retrieve the "abort-after-operation" header value.
	abortAfterOperationSuccess := string(fiberCtx.Request().Header.Peek("abort-after-operation"))
	if abortAfterOperationSuccess == "true" {
		fiberCtx.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "abort-after-operation header is set to true",
		})
		return ctx, false
	}
	return ctx, true
}

// MiddlewareOnError sets a header and optionally aborts on error.
func MiddlewareOnError(ctx context.Context, fiberCtx *fiber.Ctx, err error) (context.Context, bool) {
	fiberCtx.Response().Header.Set("X-pass-on-error", "true")

	abortOnError := string(fiberCtx.Request().Header.Peek("abort-on-error"))
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
		fiberCtx.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "abort-on-error header is set to true " + operationErr,
		})
		return ctx, false
	}
	return ctx, true
}

// MiddlewareOnError2 sets a header for error handling without aborting.
func MiddlewareOnError2(ctx context.Context, fiberCtx *fiber.Ctx, err error) (context.Context, bool) {
	fiberCtx.Response().Header.Set("X-pass-on-error-2", "true")
	return ctx, true
}

func MiddlewareOnValidationError(ctx context.Context, fiberCtx *fiber.Ctx, err error) (context.Context, bool) {
	fiberCtx.Response().Header.Set("X-pass-error-validation", "true")

	abortOnError := string(fiberCtx.Request().Header.Peek("abort-on-error"))
	if abortOnError == "true" {
		operationErr := ""
		// Handle different error types
		switch e := err.(type) {
		case *fiber.Error:
			operationErr = e.Message
		case error:
			operationErr = e.Error()
		}
		fiberCtx.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "abort-on-error header is set to true " + operationErr,
		})
		return ctx, false
	}
	return ctx, true
}

func MiddlewareOnOutputValidationError(ctx context.Context, fiberCtx *fiber.Ctx, err error) (context.Context, bool) {
	fiberCtx.Response().Header.Set("X-pass-output-validation", "true")

	returnNull := string(fiberCtx.Request().Header.Peek("x-return-null"))
	if returnNull == "true" {
		fiberCtx.Status(http.StatusOK).JSON(nil)
		return ctx, false
	}

	abortOnError := string(fiberCtx.Request().Header.Peek("abort-on-error"))
	if abortOnError == "true" {
		operationErr := ""
		// Handle different error types
		switch e := err.(type) {
		case *fiber.Error:
			operationErr = e.Message
		case error:
			operationErr = e.Error()
		}
		fiberCtx.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "abort-on-error header is set to true " + operationErr,
		})
		return ctx, false
	}
	return ctx, true
}
