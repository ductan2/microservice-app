package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env if present; non-fatal if missing (e.g., in production/containers)
	if err := godotenv.Load(); err != nil {
		// It's okay if .env is not present; PORT can still be provided by the environment
	}
}

// GetPort returns the port from the PORT env var. Falls back to 8001 if unset.
func GetPort() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8010"
}

func GetUserServiceURL() string {
	log.Println("USER_SERVICE_URL", os.Getenv("USER_SERVICE_URL"))
	if v := os.Getenv("USER_SERVICE_URL"); v != "" {

		return v
	}
	return "http://localhost:8001"
}

// GetCORSOrigin returns allowed CORS origin from env CORS_URL; default http://localhost:3000
func GetCORSOrigin() string {
	if v := os.Getenv("CORS_URL"); v != "" {
		return v
	}
	return "http://localhost:3000"
}
