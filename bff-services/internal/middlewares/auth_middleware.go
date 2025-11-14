package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"bff-services/internal/cache"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	contextUserIDKey    = "userID"
	contextUserEmailKey = "userEmail"
	contextSessionIDKey = "sessionID"
	contextUserRoleKey  = "userRole"
)

const (
	RoleStudent    = "student"
	RoleTeacher    = "teacher"
	RoleAdmin      = "admin"
	RoleSuperAdmin = "super-admin"
)

// AuthRequired ensures requests include a valid Bearer access token and validates session in Redis.
func AuthRequired(sessionCache *cache.SessionCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			fmt.Println("missing authorization header")
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
		log.Println("claims.SessionID", claims.SessionID)
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

func GetUserContextFromMiddleware(c *gin.Context) (userID, email, sessionID string, ok bool) {
	userIDValue, exists := c.Get(ContextUserIDKey())
	if !exists {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "user context not found")
		return "", "", "", false
	}

	emailValue, exists := c.Get(ContextUserEmailKey())
	if !exists {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "email context not found")
		return "", "", "", false
	}

	sessionIDValue, exists := c.Get(ContextSessionIDKey())
	if !exists {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "session context not found")
		return "", "", "", false
	}

	// Convert UUID to string for internal communication
	userIDUUID, ok := userIDValue.(uuid.UUID)
	if !ok {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "invalid user ID type")
		return "", "", "", false
	}

	emailStr, ok := emailValue.(string)
	if !ok {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "invalid email type")
		return "", "", "", false
	}

	sessionIDUUID, ok := sessionIDValue.(uuid.UUID)
	if !ok {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "invalid session ID type")
		return "", "", "", false
	}

	return userIDUUID.String(), emailStr, sessionIDUUID.String(), true
}

// AdminRequired checks if the authenticated user has admin or super-admin role.
// This middleware must be used after AuthRequired middleware.
func AdminRequired(userService services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, email, sessionID, ok := GetUserContextFromMiddleware(c)
		if !ok {
			return
		}

		// Get user information from user-service to check role
		userResp, err := userService.GetUserById(c.Request.Context(), userID, email, sessionID, userID)
		if err != nil {
			utils.Fail(c, "Failed to verify user role", http.StatusInternalServerError, err.Error())
			c.Abort()
			return
		}

		if userResp.StatusCode != http.StatusOK {
			utils.Fail(c, "Failed to verify user role", userResp.StatusCode, "unable to fetch user data")
			c.Abort()
			return
		}

		// Parse user data to get role
		var userData struct {
			Data struct {
				Role string `json:"role"`
			} `json:"data"`
		}

		if err := json.Unmarshal(userResp.Body, &userData); err != nil {
			utils.Fail(c, "Failed to parse user data", http.StatusInternalServerError, err.Error())
			c.Abort()
			return
		}

		// Check if user is admin or super-admin
		if userData.Data.Role != RoleAdmin && userData.Data.Role != RoleSuperAdmin {
			utils.Fail(c, "Forbidden", http.StatusForbidden, "admin or super-admin role required")
			c.Abort()
			return
		}

		// Store role in context for later use if needed
		c.Set(contextUserRoleKey, userData.Data.Role)

		c.Next()
	}
}
