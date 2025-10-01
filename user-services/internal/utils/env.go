package utils

import "os"

// GetEnv returns environment variable value or fallback if not set
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
