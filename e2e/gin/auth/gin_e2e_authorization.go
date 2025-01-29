package assets

import (
	"github.com/gin-gonic/gin"
	"github.com/gopher-fleece/gleece/external"
)

func GleeceRequestAuthorization(ctx *gin.Context, check external.SecurityCheck) *external.SecurityError {
	return nil
}
