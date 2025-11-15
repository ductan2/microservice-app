package routes

import (
	"user-services/internal/api/controllers"
	"user-services/internal/api/middleware"
	"user-services/internal/cache"
	"user-services/internal/config"

	"github.com/gin-gonic/gin"
)

func RegisterPasswordRoutes(router *gin.RouterGroup, controller *controllers.PasswordController, sessionCache *cache.SessionCache, rateLimiter middleware.RateLimiter, cfg *config.Config) {
	password := router.Group("/password")
	{
		// Password reset request with stricter rate limiting
		passwordResetConfig := middleware.RateLimitConfig{
			Requests: cfg.RateLimit.PasswordResetPerHour,
			Window:   cfg.RateLimit.PasswordResetWindow,
		}

		password.POST("/reset/request",
			middleware.AuthRateLimitMiddleware(rateLimiter, passwordResetConfig),
			controller.RequestPasswordReset) // POST /password/reset/request

		// Password reset confirmation (less restrictive since it requires valid token)
		passwordConfirmConfig := middleware.RateLimitConfig{
			Requests: cfg.RateLimit.AuthRequestsPerMinute * 2, // Allow more attempts for confirmation
			Window:   cfg.RateLimit.AuthWindow,
		}

		password.POST("/reset/confirm",
			middleware.RateLimitMiddleware(rateLimiter, passwordConfirmConfig),
			controller.ConfirmPasswordReset) // POST /password/reset/confirm

		// Change password requires authentication
		password.Use(middleware.AuthRequired(sessionCache))
		password.POST("/change", controller.ChangePassword) // POST /password/change (authenticated)
	}
}
