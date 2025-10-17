package controllers

import (
	"strings"


	"github.com/gin-gonic/gin"
)

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
