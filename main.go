package main

import (
	"fmt"
	"os"

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
	// config.Loadenv()
	config.ConnectToDB()
	config.SyncDB()
}

func main() {
	r := gin.Default()

	r.POST("/api/auth/register", AuthController.SignUpUser)
	r.POST("/api/auth/login", AuthController.LoginUser)
	r.POST("/api/auth/setdetails", AuthController.SetUserDetails)
	r.POST("/api/auth/getdetails", AuthController.ReteriveUserDetails)
	r.GET("/api/auth/discord/redirect", AuthController.DiscordAuth)
	r.POST("/api/upload", AuthController.UploadAvatar)

	// Create the "avatar" directory if it doesn't exist
	if err := os.MkdirAll("avatar", os.ModePerm); err != nil {
		fmt.Println("Error creating 'uploads' directory:", err)
		return
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	r.Run(":" + port)
}
