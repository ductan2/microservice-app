package config

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
	}
}

func GetPort() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8004"
}
