package config

import "github.com/JohnnyOhms/projectx/model"

func SyncDB() {
	DB.AutoMigrate(&model.User{})
}
