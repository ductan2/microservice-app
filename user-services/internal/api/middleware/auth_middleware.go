package middleware

import (
	"log"
	"net/http"
	"strings"

	"user-services/internal/cache"
	"user-services/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	contextUserIDKey    = "userID"
	contextUserEmailKey = "userEmail"
	contextSessionIDKey = "sessionID"
)

// AuthRequired ensures requests include a valid Bearer access token and validates session in Redis.
func AuthRequired(sessionCache *cache.SessionCache) gin.HandlerFunc {
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

		// Check if session exists in Redis
		sessionData, err := sessionCache.GetSession(c.Request.Context(), claims.SessionID)
		if err != nil {
			utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "session not found or expired")
			c.Abort()
			return
		}

		// Additional validation: ensure userID in session matches JWT claims
		if sessionData.UserID != claims.UserID {
			utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "session user mismatch")
			c.Abort()
			return
		}

		c.Set(contextUserIDKey, claims.UserID)
		c.Set(contextUserEmailKey, claims.Email)
		c.Set(contextSessionIDKey, claims.SessionID)

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

// ContextSessionIDKey exposes the context key used to store the authenticated session ID.
func ContextSessionIDKey() string {
	return contextSessionIDKey
}

// InternalAuthRequired validates internal requests from BFF service.
// It extracts userID, email, and sessionID from headers set by BFF.
// This middleware is for internal microservice communication only.
func InternalAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		email := c.GetHeader("X-User-Email")
		sessionID := c.GetHeader("X-Session-ID")

		log.Println("userID", userID)
		log.Println("email", email)
		log.Println("sessionID", sessionID)
		if userID == "" || email == "" || sessionID == "" {
			utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "missing internal auth headers")
			c.Abort()
			return
		}

		// Parse UUID
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "invalid user ID format")
			c.Abort()
			return
		}

		parsedSessionID, err := uuid.Parse(sessionID)
		if err != nil {
			utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "invalid session ID format")
			c.Abort()
			return
		}

		// Set context values
		c.Set(contextUserIDKey, parsedUserID)
		c.Set(contextUserEmailKey, email)
		c.Set(contextSessionIDKey, parsedSessionID)

		c.Next()
	}
}
