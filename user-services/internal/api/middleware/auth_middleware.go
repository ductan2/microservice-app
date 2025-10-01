package middleware

import (
	"net/http"
	"strings"

	"user-services/internal/utils"

	"github.com/gin-gonic/gin"
)

const (
	contextUserIDKey    = "userID"
	contextUserEmailKey = "userEmail"
)

// AuthRequired ensures requests include a valid Bearer access token.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := utils.ValidateJWT(strings.TrimSpace(parts[1]))
		if err != nil {
			utils.Fail(c, "Unauthorized", http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}

		c.Set(contextUserIDKey, claims.UserID)
		c.Set(contextUserEmailKey, claims.Email)

		c.Next()
	}
}

// ContextUserIDKey exposes the context key used to store the authenticated user ID.
func ContextUserIDKey() string {
	return contextUserIDKey
}

// ContextUserEmailKey exposes the context key used to store the authenticated user email.
func ContextUserEmailKey() string {
	return contextUserEmailKey
}
