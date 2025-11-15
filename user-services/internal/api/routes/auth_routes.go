package routes

import (
	"user-services/internal/api/controllers"
	"user-services/internal/api/middleware"
	"user-services/internal/config"

	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers authentication-specific routes
func RegisterAuthRoutes(router *gin.RouterGroup, controller *controllers.TokenController, rateLimiter middleware.RateLimiter, cfg *config.Config) {
	auth := router.Group("/auth")
	{
		// Refresh token endpoint with rate limiting
		refreshConfig := middleware.RateLimitConfig{
			Requests: cfg.RateLimit.AuthRequestsPerMinute * 2, // Allow more refresh attempts
			Window:   cfg.RateLimit.AuthWindow,
		}

		auth.POST("/refresh",
			middleware.RateLimitMiddleware(rateLimiter, refreshConfig),
			controller.RefreshToken)

		// Revoke token requires authentication
		auth.Use(middleware.InternalAuthRequired())
		auth.POST("/revoke", controller.RevokeToken)
	}
}