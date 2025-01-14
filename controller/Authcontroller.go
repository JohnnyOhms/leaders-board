package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/JohnnyOhms/projectx/entity"
	"github.com/JohnnyOhms/projectx/services"
	"github.com/gin-gonic/gin"
)

// AuthController defines the methods for handling authentication-related operations.
type AuthController interface {
	LoginUser(ctx *gin.Context)
	SignUpUser(ctx *gin.Context)
	SetUserDetails(ctx *gin.Context)
	ReteriveUserDetails(ctx *gin.Context)
	DiscordAuth(ctx *gin.Context)
	UploadAvatar(ctx *gin.Context)
}

// controller is the implementation of AuthController.
type controller struct {
	services services.AuthService
}

// New creates a new instance of AuthController.
func New(services services.AuthService) AuthController {
	return &controller{
		services: services,
	}
}

// LoginUser handles the user login process.
func (c *controller) LoginUser(ctx *gin.Context) {
	// Get the req body
	var reqBody entity.LoginUser
	// Parse the req body
	if err := ctx.Bind(&reqBody); err != nil {
		ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	//find user
	user, err := c.services.Find(reqBody)
	if err != nil {
		ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	// compare password
	err = c.services.ComparePassword([]byte(user.Password), []byte(reqBody.Password))
	if err != nil {
		ctx.JSON(400, gin.H{
			"error": "Password incorrect",
		})
		return
	}
	// generate token for authorization
	token, err := c.services.GenearateToken(user)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": "Failed to Generate Token",
		})
		return
	}
	// set cookie
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("Authorization", token, 3600*24*10, "", "", false, true)
	// Handle successful user retriever
	ctx.JSON(202, user)
}

// SignUpUser handles the user login process.
func (c *controller) SignUpUser(ctx *gin.Context) {
	// Get the req body
	var reqBody entity.User
	// Parse the req body
	if err := ctx.Bind(&reqBody); err != nil {
		ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Hash the password
	hash, err := c.services.HashPassword([]byte(reqBody.Password))
	if err != nil {
		ctx.JSON(400, gin.H{
			"error": "failed in Hashing password",
		})
		return
	}
	reqBody.Password = string(hash)

	// Create new user
	user, err := c.services.Create(reqBody)
	if err != nil {
		ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	// generate token for authorization
	token, err := c.services.GenearateToken(user)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": "Failed to Generate Token",
		})
		return
	}
	// set cookie
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("Authorization", token, 3600*24*10, "", "", false, true)

	// Handle successful user creation
	ctx.JSON(http.StatusCreated, user)
}

// set the user details
func (c *controller) SetUserDetails(ctx *gin.Context) {
	// Get the req body from the user
	var UserBody entity.User_Details
	// Parse the req body
	if err := ctx.Bind(&UserBody); err != nil {
		ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	// create user details
	userDetails, err := c.services.CreateDetails(UserBody)
	if err != nil {
		ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusCreated, userDetails)
}

// retrieve the user details
func (c *controller) ReteriveUserDetails(ctx *gin.Context) {
	// Get the req body from the user
	var userId entity.UserId
	// Parse the req body
	if err := ctx.Bind(&userId); err != nil {
		ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	//find user details
	userDetails, err := c.services.FindDetails(userId)
	if err != nil {
		ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusCreated, userDetails)
}

// discord auth api
func (c *controller) DiscordAuth(ctx *gin.Context) {
	code := ctx.Query("code")
	if len(code) < 1 {
		ctx.JSON(400, gin.H{"error": "Missing 'code' parameter"})
		return
	}

	formData := url.Values{
		"client_id":     {os.Getenv("CLIENT_ID")},
		"client_secret": {os.Getenv("CLIENT_SECRET")},
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {os.Getenv("REDIRECT_URI")},
	}

	// Create a request to exchange the authorization code for an access token
	tokenReq, err := http.NewRequest("POST", "https://discord.com/api/v10/oauth2/token", strings.NewReader(formData.Encode()))
	if err != nil {
		fmt.Println("Error creating token request:", err)
		ctx.JSON(500, gin.H{"error": "Error creating token request, try again"})
		return
	}

	// Set the Content-Type header to application/x-www-form-urlencoded
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make the request to get the access token
	client := &http.Client{}
	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		fmt.Println("Error making token request:", err)
		ctx.JSON(500, gin.H{"error": "Error making token request, try again"})
		return
	}
	defer tokenResp.Body.Close()

	// Print the raw response body for debugging
	tokenResponseBody, err := ioutil.ReadAll(tokenResp.Body)
	if err != nil {
		fmt.Println("Error reading token response body:", err)
		ctx.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	// fmt.Println("Raw Token Response Body:", string(tokenResponseBody))

	// Parse the response body
	var tokenResponse entity.DiscordToken
	err = json.Unmarshal(tokenResponseBody, &tokenResponse)
	if err != nil {
		fmt.Println("Error decoding token JSON:", err)
		ctx.JSON(500, gin.H{"error": "Error decoding token JSON, try again"})
		return
	}
	// fmt.Println("Parsed Access Token:", tokenResponse.AccessToken)

	// Rewind the response body so it can be read again
	tokenResp.Body = ioutil.NopCloser(bytes.NewBuffer(tokenResponseBody))

	// Store the access token in a variable
	accessToken := tokenResponse.AccessToken

	// Create a request to get user information using the access token
	userReq, err := http.NewRequest("GET", "https://discord.com/api/v10/users/@me", nil)
	if err != nil {
		fmt.Println("Error creating user request:", err)
		ctx.JSON(500, gin.H{"error": "Error creating user request, try again"})
		return
	}

	// Set the Authorization header with the access token
	userReq.Header.Set("Authorization", "Bearer "+accessToken)

	// Make the request to get user information
	userClient := &http.Client{}
	userResp, err := userClient.Do(userReq)
	if err != nil {
		fmt.Println("Error making user request:", err)
		ctx.JSON(500, gin.H{"error": "Error making user request, try again"})
		return
	}
	defer userResp.Body.Close()

	// read the user body
	userResponseBody, err := ioutil.ReadAll(userResp.Body)
	if err != nil {
		fmt.Println("Error reading user response body:", err)
		ctx.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}

	//Parse the User Response body
	var UserInfo entity.UserDiscordData
	err = json.Unmarshal(userResponseBody, &UserInfo)
	if err != nil {
		fmt.Println("Error decoding token JSON:", err)
		ctx.JSON(500, gin.H{"error": "Error decoding userinfo JSON, try again"})
		return
	}

	// Find the User From DB
	var FindUserInfo entity.LoginUser = entity.LoginUser{
		Email:    UserInfo.Email,
		Password: UserInfo.ID,
	}

	// Create the user From DB
	var CreateUserInfo entity.User = entity.User{
		Email:    UserInfo.Email,
		Password: UserInfo.ID,
		UserId:   c.services.GenerateUserId(),
	}

	user, err := c.services.Find(FindUserInfo)
	if err != nil {
		if err.Error() == "record not found" {

			// Hash the password
			hash, err := c.services.HashPassword([]byte(CreateUserInfo.Password))
			if err != nil {
				ctx.JSON(400, gin.H{
					"error": "failed in Hashing password",
				})
				return
			}
			CreateUserInfo.Password = string(hash)
			user, err = c.services.Create(CreateUserInfo)
			if err != nil {
				ctx.JSON(400, gin.H{
					"error": err.Error(),
				})
				return
			}
			// generate token for authorization
			token, err := c.services.GenearateToken(user)
			if err != nil {
				ctx.JSON(500, gin.H{
					"error": "Failed to Generate Token",
				})
				return
			}
			// set cookie
			ctx.SetSameSite(http.SameSiteLaxMode)
			ctx.SetCookie("Authorization", token, 3600*24*10, "", "", false, true)
			ctx.JSON(202, user)
			return
		} else {

			ctx.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	// compare password
	err = c.services.ComparePassword([]byte(user.Password), []byte(FindUserInfo.Password))
	if err != nil {

		fmt.Println("Error:", err)
		ctx.JSON(400, gin.H{
			"error": "Password incorrect in the discord route",
		})
		return
	}

	// Generate Token and Set Cookie
	token, err := c.services.GenearateToken(user)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": "Failed to Generate Token",
		})
		return
	}
	user.Password = "NULL"
	// set cookie
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("Authorization", token, 3600*24*10, "", "", false, true)
	ctx.JSON(202, user)
}

// upload the avatar (profile pic) to the server
// func (c *controller) UploadAvatar(ctx *gin.Context) {
// 	// Parse the form data, including the uploaded file
// 	err := ctx.Request.ParseMultipartForm(1 << 20) // 10 MB limit for the uploaded file
// 	if err != nil {
// 		fmt.Println(err)
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Unable to parse form"})
// 		return
// 	}

// 	// Get the file from the request
// 	file, handler, err := ctx.Request.FormFile("avatar")
// 	if err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error retrieving the file"})
// 		return
// 	}
// 	defer file.Close()

// 	// Get the userId from the form data
// 	userId := ""

// 	// Create a new file on the server to save the uploaded file
// 	filename := filepath.Join("avatar", handler.Filename+userId)
// 	f, err := os.Create(filename)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create the file on the server"})
// 		return
// 	}
// 	defer f.Close()

// 	// Copy the file content to the new file
// 	_, err = io.Copy(f, file)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to copy the file content"})
// 		return
// 	}

// 	// Respond with a success message
// 	ctx.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("File %s uploaded successfully", handler.Filename)})
// }

// type Avatar struct {
// 	gorm.Model
// 	UserID uint
// 	Avatar string
// }

// UploadAvatar uploads the avatar (profile pic) to the server and stores the text in the database
func (c *controller) UploadAvatar(ctx *gin.Context) {
	// Parse the form data, including the uploaded file
	err := ctx.Request.ParseMultipartForm(1 << 20) // 1 MB limit for the uploaded file
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Unable to parse form"})
		return
	}

	// Get the file from the request
	file, handler, err := ctx.Request.FormFile("avatar")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error retrieving the file"})
		return
	}
	defer file.Close()

	// Get the userId from the the authorization middleware
	userID := "user id goes here"

	// Create a new file on the server to save the uploaded file
	filename := filepath.Join("avatars", fmt.Sprintf("%s_%s", userID, handler.Filename))
	f, err := os.Create(filename)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create the file on the server"})
		return
	}
	defer f.Close()

	// Copy the file content to the new file
	_, err = io.Copy(f, file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to copy the file content"})
		return
	}

	// Store the image text in the database
	fmt.Println("store to database")
	text := extractTextFromImage(filename) // Implement this function to extract text from the image
	avatar := entity.Avatar{UserId: userID, Avatar: text}

	_, err = c.services.SetAvatar(avatar, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to store image text in the database"})
		return
	}

	// Respond with a success message
	ctx.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("File %s uploaded successfully", handler.Filename)})
}

// extractTextFromImage extracts text from the image file
func extractTextFromImage(filename string) string {
	// Implement your logic to extract text from the image
	// This is just a placeholder
	return strings.ToUpper(filename) // For demonstration, return the filename in uppercase
}
