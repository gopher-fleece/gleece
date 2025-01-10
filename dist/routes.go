package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func RegisterRoutes(engine *gin.Engine) {
	// ExtendedController
		engine.GET("/users/dont", func(ctx *gin.Context) {
			controller := ExtendedController{}
			controller.SetRequest(ctx)
			fgd := ctx.Query("fgd")
			result := controller.DontDoItPlease(fgd)
		})
			
			
			
			
			
			
			
			
			
			
			
		
}