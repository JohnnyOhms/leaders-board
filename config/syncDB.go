package config

import "github.com/JohnnyOhms/projectx/model"

func SyncDB() {
	DB.AutoMigrate(&model.User{}, &model.User_Details{}, &model.Avatar{})
}
