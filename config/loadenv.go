package config

import (
	"fmt"

	"github.com/joho/godotenv"
)

// Loadenv loads environment variables from a .env file.
func Loadenv() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	return nil
}
