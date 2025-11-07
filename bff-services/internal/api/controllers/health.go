package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Health returns a simple health check response.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
