package services

import (
	"math/rand"
	"os"
	"time"

	"github.com/JohnnyOhms/projectx/config"
	"github.com/JohnnyOhms/projectx/entity"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService is an interface for user authentication services
type AuthService interface {
	Create(user entity.User) (entity.User, error)
	Find(user entity.LoginUser) (entity.User, error)
	HashPassword(pwd []byte) ([]byte, error)
	ComparePassword(userPwd []byte, pwd []byte) error
	GenearateToken(user entity.User) (string, error)
	GenerateUserId() string
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

// Add new user to the db
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
	const charaset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
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
