package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	UserId       string `gorm:"unique;not null"`
	Email        string `gorm:"unique;not null"`
	Password     string `gorm:"not null"`
	Is_Verified  bool   `gorm:"not null"`
	Account_Type string `gorm:"not null"`
}

type User_Details struct {
	gorm.Model
	UserId   string `gorm:"unique;not null"`
	Email    string `gorm:"unique;not null"`
	Username string `gorm:"size:30;not null"`
	Phone    int    `gorm:"not null"`
	Twitter  string
	Discord  string
	Google   string
}

type Avatar struct {
	gorm.Model
	UserId string `gorm:"unique;not null"`
	Avatar string `gorm:"not null"`
}
