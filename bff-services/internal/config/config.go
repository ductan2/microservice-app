package config

import (
	"log"
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
	return "8010"
}

func GetUserServiceURL() string {
	log.Println("USER_SERVICE_URL", os.Getenv("USER_SERVICE_URL"))
	if v := os.Getenv("USER_SERVICE_URL"); v != "" {

		return v
	}
	return "http://localhost:8001"
}

func GetContentServiceURL() string {
	if v := os.Getenv("CONTENT_SERVICE_URL"); v != "" {
		return v
	}
	return "http://localhost/api/content"
}

func GetLessonServiceURL() string {
	if v := os.Getenv("LESSON_SERVICE_URL"); v != "" {
		return v
	}
	return "http://localhost:8005"
}

func GetNotificationServiceURL() string {
	if v := os.Getenv("NOTIFICATION_SERVICE_URL"); v != "" {
		return v
	}
	return "http://localhost:8003"
}

func GetOrderServiceURL() string {
	if v := os.Getenv("ORDER_SERVICE_URL"); v != "" {
		return v
	}
	return "http://localhost:8006"
}

// GetCORSOrigins returns allowed CORS origins from env CORS_URLS (comma-separated)
// Falls back to single env CORS_URL, then default http://localhost:3000.
func GetCORSOrigins() []string {
	// Highest priority: CORS_URLS (comma-separated)
	if v := os.Getenv("CORS_URLS"); v != "" {
		parts := strings.Split(v, ",")
		allowed := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				allowed = append(allowed, trimmed)
			}
		}
		if len(allowed) > 0 {
			return allowed
		}
	}

	// Next: legacy/single CORS_URL
	if v := os.Getenv("CORS_URL"); v != "" {
		return []string{v}
	}

	// Default
	return []string{"http://localhost:3000"}
}

// IsOriginAllowed checks whether the provided origin is in the configured allowlist.
func IsOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	for _, allowed := range GetCORSOrigins() {
		if origin == allowed {
			return true
		}
	}
	return false
}

// Redis
type RedisConfig struct {
	Addr     string
	Password string
}

func GetRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
	}
}
