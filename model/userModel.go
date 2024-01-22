package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	UserId   string `gorm:"not null"`
	Email    string `gorm:"unique;not null"`
	Username string `gorm:"size:30;not null"`
	Password string `gorm:"not null"`
}
