package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func Loadenv() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
		log.Fatal("Error loading .env file")
	}
}
