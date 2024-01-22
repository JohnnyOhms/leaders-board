package routes

import (

	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) *gin.Engine {
	r.POST("/register",nil)
	r.POST("/login", nil)
	r.POST("/discord", nil)
	return r
}
