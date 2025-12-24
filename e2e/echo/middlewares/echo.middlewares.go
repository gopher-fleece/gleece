package middlewares

import (
	"context"
	"net/http"

	"github.com/gopher-fleece/gleece/v2/e2e/assets"
	"github.com/labstack/echo/v4"
)

func MiddlewareBeforeOperation(ctx context.Context, echoCtx echo.Context) (context.Context, bool) {
	echoCtx.Response().Header().Set("X-pass-before-operation", "true")

	abortBeforeOperation := echoCtx.Request().Header.Get("abort-before-operation")
	if abortBeforeOperation == "true" {
		echoCtx.JSON(http.StatusBadRequest, map[string]string{"error": "abort-before-operation header is set to true"})
		return ctx, false
	}

	return context.WithValue(ctx, assets.ContextMiddleware, "pass"), true
}

func MiddlewareAfterOperationSuccess(ctx context.Context, echoCtx echo.Context) (context.Context, bool) {
	echoCtx.Response().Header().Set("X-pass-after-succeed-operation", "true")

	abortAfterOperationSuccess := echoCtx.Request().Header.Get("abort-after-operation")
	if abortAfterOperationSuccess == "true" {
		echoCtx.JSON(http.StatusBadRequest, map[string]string{"error": "abort-after-operation header is set to true"})
		return ctx, false
	}
	return ctx, true
}

func MiddlewareOnError(ctx context.Context, echoCtx echo.Context, err error) (context.Context, bool) {
	echoCtx.Response().Header().Set("X-pass-on-error", "true")

	abortOnError := echoCtx.Request().Header.Get("abort-on-error")
	if abortOnError == "true" {
		operationErr := ""
		switch e := err.(type) {
		case *echo.HTTPError:
			operationErr = e.Message.(string)
		case assets.CustomError:
			operationErr = e.Message
		default:
			operationErr = err.Error()
		}
		echoCtx.JSON(http.StatusBadRequest, map[string]string{"error": "abort-on-error header is set to true " + operationErr})
		return ctx, false
	}
	return ctx, true
}

func MiddlewareOnError2(ctx context.Context, echoCtx echo.Context, err error) (context.Context, bool) {
	echoCtx.Response().Header().Set("X-pass-on-error-2", "true")
	return ctx, true
}

func MiddlewareOnValidationError(ctx context.Context, echoCtx echo.Context, err error) (context.Context, bool) {
	echoCtx.Response().Header().Set("X-pass-error-validation", "true")

	abortOnError := echoCtx.Request().Header.Get("abort-on-error")
	if abortOnError == "true" {
		operationErr := ""
		switch e := err.(type) {
		case *echo.HTTPError:
			operationErr = e.Message.(string)
		case error:
			operationErr = e.Error()
		}
		echoCtx.JSON(http.StatusBadRequest, map[string]string{"error": "abort-on-error header is set to true " + operationErr})
		return ctx, false
	}
	return ctx, true
}

func MiddlewareOnOutputValidationError(ctx context.Context, echoCtx echo.Context, err error) (context.Context, bool) {
	echoCtx.Response().Header().Set("X-pass-output-validation", "true")

	returnNull := echoCtx.Request().Header.Get("x-return-null")
	if returnNull == "true" {
		echoCtx.JSON(http.StatusOK, nil)
		return ctx, false
	}
	abortOnError := echoCtx.Request().Header.Get("abort-on-error")
	if abortOnError == "true" {
		operationErr := ""
		switch e := err.(type) {
		case *echo.HTTPError:
			operationErr = e.Message.(string)
		case error:
			operationErr = e.Error()
		}
		echoCtx.JSON(http.StatusBadRequest, map[string]string{"error": "abort-on-error header is set to true " + operationErr})
		return ctx, false
	}
	return ctx, true
}
