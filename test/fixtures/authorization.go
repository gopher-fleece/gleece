package fixtures

import (
	"github.com/gin-gonic/gin"
	"github.com/gopher-fleece/runtime"
)

func GleeceRequestAuthorization(ctx *gin.Context, check runtime.SecurityCheck) *runtime.SecurityError {
	return nil
}
