package middleware

import (
	"log"
	"net/http"
	"strings"

	"user-services/internal/cache"
	"user-services/internal/config"
	"user-services/internal/response"
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
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := strings.TrimSpace(parts[1])
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			// Don't expose specific JWT validation errors to prevent token probing
			if config.GetConfig().IsProduction() {
				response.Unauthorized(c, "Invalid authentication token")
			} else {
				// In development, provide more detailed error messages
				response.Unauthorized(c, "Invalid authentication token: "+err.Error())
			}
			c.Abort()
			return
		}

		// Check if session exists in Redis
		sessionData, err := sessionCache.GetSession(c.Request.Context(), claims.SessionID)
		if err != nil {
			response.Unauthorized(c, "Session has expired or is invalid")
			c.Abort()
			return
		}

		// Additional validation: ensure userID in session matches JWT claims
		if sessionData.UserID != claims.UserID {
			// This is a security issue - log it in production
			if !config.GetConfig().IsDevelopment() {
				// TODO: Add proper security logging for session hijacking attempts
			}
			response.Unauthorized(c, "Invalid session")
			c.Abort()
			return
		}

		// Set context values for downstream handlers
		c.Set(contextUserIDKey, claims.UserID)
		c.Set(contextUserEmailKey, claims.Email)
		c.Set(contextSessionIDKey, claims.SessionID)

		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		c.Next()
	}
}

// OptionalAuth provides optional authentication - doesn't abort if no auth provided
func OptionalAuth(sessionCache *cache.SessionCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Next()
			return
		}

		token := strings.TrimSpace(parts[1])
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			c.Next()
			return
		}

		// Check if session exists in Redis
		sessionData, err := sessionCache.GetSession(c.Request.Context(), claims.SessionID)
		if err != nil {
			c.Next()
			return
		}

		// Validate session matches JWT claims
		if sessionData.UserID != claims.UserID {
			c.Next()
			return
		}

		// Set context values
		c.Set(contextUserIDKey, claims.UserID)
		c.Set(contextUserEmailKey, claims.Email)
		c.Set(contextSessionIDKey, claims.SessionID)

		c.Next()
	}
}

// RequireRole middleware for role-based authorization (placeholder for future implementation)
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement role-based authorization once role system is in place
		// For now, just pass through
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
