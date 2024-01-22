package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/JohnnyOhms/projectx/entity"
	"github.com/JohnnyOhms/projectx/services"
	"github.com/gin-gonic/gin"
)

// AuthController defines the methods for handling authentication-related operations.
type AuthController interface {
	LoginUser(ctx *gin.Context)
	SignUpUser(ctx *gin.Context)
	DiscordAuth(ctx *gin.Context)
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

	// Handle successful user creation
	ctx.JSON(http.StatusCreated, user)
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
	var UserInfo entity.UserData
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

	user, err := c.services.Find(FindUserInfo)
	if err != nil {
		ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	// "record not found"
	// work on here

	if user.Email != "" {
		// compare password
		err = c.services.ComparePassword([]byte(user.Password), []byte(FindUserInfo.Password))
		if err != nil {
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
		// set cookie
		ctx.SetSameSite(http.SameSiteLaxMode)
		ctx.SetCookie("Authorization", token, 3600*24*10, "", "", false, true)
		// Handle successful user retriever
		ctx.JSON(202, user)
		return
	}

	// Create the user From DB
	var CreateUserInfo entity.User = entity.User{
		Username: UserInfo.Username,
		Email:    UserInfo.Email,
		Password: UserInfo.ID,
		UserId:   c.services.GenerateUserId(),
	}
	user, err = c.services.Create(CreateUserInfo)
	if err != nil {
		ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Send the response body as part of the JSON response
	// ctx.JSON(http.StatusOK, gin.H{"message": "ok", "user_data": string(userResponseBody)})
	ctx.JSON(201, user)
}