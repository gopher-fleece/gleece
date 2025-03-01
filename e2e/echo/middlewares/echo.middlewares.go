package middlewares

import (
	"net/http"

	"github.com/gopher-fleece/gleece/e2e/assets"
	"github.com/labstack/echo/v4"
)

func MiddlewareBeforeOperation(c echo.Context) bool {
	c.Response().Header().Set("X-pass-before-operation", "true")

	abortBeforeOperation := c.Request().Header.Get("abort-before-operation")
	if abortBeforeOperation == "true" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "abort-before-operation header is set to true"})
		return false
	}

	return true
}

func MiddlewareAfterOperationSuccess(c echo.Context) bool {
	c.Response().Header().Set("X-pass-after-succeed-operation", "true")

	abortAfterOperationSuccess := c.Request().Header.Get("abort-after-operation")
	if abortAfterOperationSuccess == "true" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "abort-after-operation header is set to true"})
		return false
	}
	return true
}

func MiddlewareOnError(c echo.Context, err error) bool {
	c.Response().Header().Set("X-pass-on-error", "true")

	abortOnError := c.Request().Header.Get("abort-on-error")
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
		c.JSON(http.StatusBadRequest, map[string]string{"error": "abort-on-error header is set to true " + operationErr})
		return false
	}
	return true
}

func MiddlewareOnError2(c echo.Context, err error) bool {
	c.Response().Header().Set("X-pass-on-error-2", "true")
	return true
}

func MiddlewareOnValidationError(c echo.Context, err error) bool {
	c.Response().Header().Set("X-pass-error-validation", "true")
	
	abortOnError := c.Request().Header.Get("abort-on-error")
	if abortOnError == "true" {
		operationErr := ""
		switch e := err.(type) {
		case *echo.HTTPError:
			operationErr = e.Message.(string)
		case error:
			operationErr = e.Error()
		}
		c.JSON(http.StatusBadRequest, map[string]string{"error": "abort-on-error header is set to true " + operationErr})
		return false
	}
	return true
}

func MiddlewareOnOutputValidationError(c echo.Context, err error) bool {
	c.Response().Header().Set("X-pass-output-validation", "true")
	
	returnNull := c.Request().Header.Get("x-return-null")
	if returnNull == "true" {
		c.JSON(http.StatusOK, nil)
		return false
	}
	abortOnError := c.Request().Header.Get("abort-on-error")
	if abortOnError == "true" {
		operationErr := ""
		switch e := err.(type) {
		case *echo.HTTPError:
			operationErr = e.Message.(string)
		case error:
			operationErr = e.Error()
		}
		c.JSON(http.StatusBadRequest, map[string]string{"error": "abort-on-error header is set to true " + operationErr})
		return false
	}
	return true
}