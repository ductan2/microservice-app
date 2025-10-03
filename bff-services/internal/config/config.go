package config

import (
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
