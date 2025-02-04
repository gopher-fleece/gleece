package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/gopher-fleece/gleece/e2e/assets"
)

func MiddlewareBeforeOperation(ctx *gin.Context) bool {
	ctx.Header("X-pass-before-operation", "true")

	abortBeforeOperation := ctx.GetHeader("abort-before-operation")
	if abortBeforeOperation == "true" {
		ctx.JSON(400, gin.H{"error": "abort-before-operation header is set to true"})
		return false
	}

	return true
}

func MiddlewareAfterOperationSuccess(ctx *gin.Context) bool {
	ctx.Header("X-pass-after-succeed-operation", "true")

	abortAfterOperationSuccess := ctx.GetHeader("abort-after-operation")
	if abortAfterOperationSuccess == "true" {
		ctx.JSON(400, gin.H{"error": "abort-after-operation header is set to true"})
		return false
	}
	return true
}

func MiddlewareOnError(ctx *gin.Context, err error) bool {
	ctx.Header("X-pass-on-error", "true")

	abortOnError := ctx.GetHeader("abort-on-error")
	if abortOnError == "true" {
		operationErr := ""
		switch err.(type) {
		case assets.CustomError:
			customError := err.(assets.CustomError)
			operationErr = customError.Message
		case error:
			operationErr = err.Error()
		}
		ctx.JSON(400, gin.H{"error": "abort-on-error header is set to true " + operationErr})
		return false
	}
	return true
}

func MiddlewareOnError2(ctx *gin.Context, err error) bool {
	ctx.Header("X-pass-on-error-2", "true")
	return true
}
