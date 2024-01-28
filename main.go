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
	config.Loadenv()
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
		fmt.Println("Error creating 'avatar' directory:", err)
		return
	}

	// Load environment variables
	if err := config.Loadenv(); err != nil {
		fmt.Println("Failed to load environment variables:", err)
		return
	}

	// Check if PORT environment variable is set
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("PORT environment variable not set. Defaulting to port 9000.")
		port = "9000"
	} else {
		fmt.Println("Using PORT:", port)
	}

	r.Run(":" + port)
}
