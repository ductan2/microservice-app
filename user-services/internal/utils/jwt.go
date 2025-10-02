package utils

import (
	"errors"
	"fmt"
	"time"

	"user-services/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	SessionID uuid.UUID `json:"session_id"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a signed JWT for the given user id, email, and session id.
func GenerateJWT(userID uuid.UUID, email string, sessionID uuid.UUID) (string, error) {
	cfg := config.GetJWTConfig()

	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.ExpiresIn)),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(cfg.Secret))
}

// ValidateJWT parses and validates a JWT access token and returns its claims when valid.
func ValidateJWT(token string) (*Claims, error) {
	cfg := config.GetJWTConfig()

	claims := &Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
