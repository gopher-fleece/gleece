package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gopher-fleece/gleece/e2e/assets"
)

func MiddlewareBeforeOperation(ctx context.Context, ginCtx *gin.Context) (context.Context, bool) {
	ginCtx.Header("X-pass-before-operation", "true")

	abortBeforeOperation := ginCtx.GetHeader("abort-before-operation")
	if abortBeforeOperation == "true" {
		ginCtx.JSON(400, gin.H{"error": "abort-before-operation header is set to true"})
		return ctx, false
	}

	return context.WithValue(ctx, assets.ContextMiddleware, "pass"), true
}

func MiddlewareAfterOperationSuccess(ctx context.Context, ginCtx *gin.Context) (context.Context, bool) {
	ginCtx.Header("X-pass-after-succeed-operation", "true")

	abortAfterOperationSuccess := ginCtx.GetHeader("abort-after-operation")
	if abortAfterOperationSuccess == "true" {
		ginCtx.JSON(400, gin.H{"error": "abort-after-operation header is set to true"})
		return ctx, false
	}
	return ctx, true
}

func MiddlewareOnError(ctx context.Context, ginCtx *gin.Context, err error) (context.Context, bool) {
	ginCtx.Header("X-pass-on-error", "true")

	abortOnError := ginCtx.GetHeader("abort-on-error")
	if abortOnError == "true" {
		operationErr := ""
		switch err.(type) {
		case assets.CustomError:
			customError := err.(assets.CustomError)
			operationErr = customError.Message
		case error:
			operationErr = err.Error()
		}
		ginCtx.JSON(400, gin.H{"error": "abort-on-error header is set to true " + operationErr})
		return ctx, false
	}
	return ctx, true
}

func MiddlewareOnError2(ctx context.Context, ginCtx *gin.Context, err error) (context.Context, bool) {
	ginCtx.Header("X-pass-on-error-2", "true")
	return ctx, true
}

func MiddlewareOnValidationError(ctx context.Context, ginCtx *gin.Context, err error) (context.Context, bool) {
	ginCtx.Header("X-pass-error-validation", "true")
	abortOnError := ginCtx.GetHeader("abort-on-error")
	if abortOnError == "true" {
		operationErr := ""
		switch err.(type) {
		case error:
			operationErr = err.Error()
		}
		ginCtx.JSON(400, gin.H{"error": "abort-on-error header is set to true " + operationErr})
		return ctx, false
	}
	return ctx, true
}

func MiddlewareOnOutputValidationError(ctx context.Context, ginCtx *gin.Context, err error) (context.Context, bool) {
	ginCtx.Header("X-pass-output-validation", "true")
	abortOnError := ginCtx.GetHeader("abort-on-error")

	returnNull := ginCtx.GetHeader("x-return-null")
	if returnNull == "true" {
		ginCtx.JSON(http.StatusOK, nil)
		return ctx, false
	}

	if abortOnError == "true" {
		operationErr := ""
		switch err.(type) {
		case error:
			operationErr = err.Error()
		}
		ginCtx.JSON(400, gin.H{"error": "abort-on-error header is set to true " + operationErr})
		return ctx, false
	}
	return ctx, true
}
