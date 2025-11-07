package config

import (
	"os"
	"time"
)

// JWTConfig contains signing secret and token lifetime.
type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

func GetJWTConfig() JWTConfig {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "change-me-dev-secret"
	}
	dur := os.Getenv("JWT_EXPIRES_IN")
	if dur == "" {
		dur = "24h"
	}
	exp, err := time.ParseDuration(dur)
	if err != nil {
		// Fallback if invalid duration
		exp = 24 * time.Hour
	}
	return JWTConfig{
		Secret:    secret,
		ExpiresIn: exp,
	}
}
