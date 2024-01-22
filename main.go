package main

import (
	"github.com/JohnnyOhms/projectx/config"
	"github.com/JohnnyOhms/projectx/controller"
	"github.com/JohnnyOhms/projectx/services"
	"github.com/gin-gonic/gin"
)

var (
	AuthService    services.AuthService      = services.New()
	AuthController controller.AuthController = controller.New(AuthService)
)

func init() {
	config.Loadenv()
	config.ConnectToDB()
	config.SyncDB()
}

func main() {
	r := gin.Default()

	r.POST("/api/auth/register", AuthController.SignUpUser)
	r.POST("/api/auth/login", AuthController.LoginUser)
	r.GET("/api/auth/discord/redirect", AuthController.DiscordAuth)

	r.Run(":9000")
}
