package services

import (
	"math/rand"
	"os"
	"time"

	"github.com/JohnnyOhms/projectx/config"
	"github.com/JohnnyOhms/projectx/entity"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService is an interface for user authentication services
type AuthService interface {
	Create(user entity.User) (entity.User, error)
	Find(user entity.LoginUser) (entity.User, error)
	HashPassword(pwd []byte) ([]byte, error)
	ComparePassword(userPwd []byte, pwd []byte) error
	GenearateToken(user entity.User) (string, error)
	GenerateUserId() string
	CreateDetails(details entity.User_Details) (entity.User_Details, error)
	FindDetails(userId entity.UserId) (entity.User_Details, error)
	SetAvatar(avatar entity.Avatar, filename string) (entity.Avatar, error)
}

// authservice is an implementation of UserAuthService
type authservice struct{}

// New creates and returns a new instance of UserAuthService
func New() AuthService {
	return &authservice{}
}

// find a user from the database by email
func (*authservice) Find(loginUser entity.LoginUser) (entity.User, error) {
	var foundUser entity.User

	result := config.DB.Where("email = ?", loginUser.Email).First(&foundUser)
	if result.Error != nil {
		return entity.User{}, result.Error
	}
	return foundUser, nil
}

// Add new user to the database
func (s *authservice) Create(user entity.User) (entity.User, error) {
	// Insert the new userId into the user body
	user.UserId = s.GenerateUserId()
	result := config.DB.Create(&user)
	if result.Error != nil {

		return entity.User{}, result.Error
	}
	return user, nil
}

// generate userID
func (s *authservice) GenerateUserId() string {
	const charaset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789~!@#$%^&*"
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, 30)
	for i := range b {
		b[i] = charaset[rand.Intn(len(charaset))]
	}
	return string(b)
}

// HashPassword hashes the given password
func (s *authservice) HashPassword(pwd []byte) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// compare the password sent by the req body
func (s *authservice) ComparePassword(userPwd []byte, pwd []byte) error {
	err := bcrypt.CompareHashAndPassword(userPwd, pwd)
	if err != nil {
		return err
	}
	return nil
}

// generate token
func (s *authservice) GenearateToken(user entity.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.UserId,
		"exp": time.Now().Add(time.Hour * 24 * 10).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// create user infomation form the database
func (s *authservice) CreateDetails(details entity.User_Details) (entity.User_Details, error) {
	result := config.DB.Create(&details)
	if result.Error != nil {
		return entity.User_Details{}, result.Error
	}
	return entity.User_Details{}, nil
}

// find the user information from the database
func (s *authservice) FindDetails(userId entity.UserId) (entity.User_Details, error) {
	var foundDetails entity.User_Details
	result := config.DB.Where("userid = ?", userId).First(&foundDetails)
	if result.Error != nil {
		return entity.User_Details{}, result.Error
	}
	return foundDetails, nil
}

func (s *authservice) SetAvatar(avatar entity.Avatar, filename string) (entity.Avatar, error) {
	// Check if the user ID already exists in the database
	var existingAvatar entity.Avatar
	result := config.DB.Where("user_id = ?", avatar.UserId).First(&existingAvatar)
	if result.Error == nil {
		// User ID already exists, update the record
		existingAvatar.Avatar = filename
		result := config.DB.Save(&existingAvatar)
		if result.Error != nil {
			return entity.Avatar{}, result.Error
		}
	} else if result.Error == gorm.ErrRecordNotFound {
		// User ID doesn't exist, create a new record
		result := config.DB.Create(&avatar)
		if result.Error != nil {
			return entity.Avatar{}, result.Error
		}
	} else {
		// Database error
		return entity.Avatar{}, result.Error
	}
	return existingAvatar, nil
}
