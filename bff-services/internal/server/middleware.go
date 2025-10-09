package server

import (
	"bff-services/internal/config"

	"github.com/gin-gonic/gin"
)

// setupGlobalMiddlewares configures global middlewares for the router
func setupGlobalMiddlewares(r *gin.Engine) {
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
}

// corsMiddleware returns a CORS middleware function
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := config.GetCORSOrigin()
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
