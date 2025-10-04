package controllers

import (
	"net/http"
	"strings"

	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

// requireBearerToken extracts the bearer token from the Authorization header.
// It returns false and writes an error response if the token is missing or malformed.
func requireBearerToken(c *gin.Context) (string, bool) {
	header := c.GetHeader("Authorization")
	if header == "" {
		utils.Fail(c, "Authorization token is required", http.StatusUnauthorized, "missing authorization header")
		return "", false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		utils.Fail(c, "Invalid authorization header", http.StatusUnauthorized, "invalid bearer token")
		return "", false
	}

	return strings.TrimSpace(parts[1]), true
}

// getOptionalBearerToken extracts the bearer token if present; returns empty string otherwise.
func getOptionalBearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
