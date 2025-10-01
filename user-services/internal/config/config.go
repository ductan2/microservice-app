package config

import (
	"os"
	"strings"

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
	return "8001"
}

func GetAppName() string {
	return getEnv("APP_NAME", "Microservice App")
}

func GetSupportEmail() string {
	return getEnv("SUPPORT_EMAIL", "support@example.com")
}

func GetPublicAppURL() string {
	return getEnv("PUBLIC_APP_URL", "http://localhost:3000")
}

func GetPasswordResetPath() string {
	return getEnv("PASSWORD_RESET_PATH", "/reset-password")
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
