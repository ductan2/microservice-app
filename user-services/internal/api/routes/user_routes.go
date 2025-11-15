package routes

import (
	"user-services/internal/api/controllers"
	"user-services/internal/api/middleware"
	"user-services/internal/config"

	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes registers all user-related routes (auth, profile, user management)
func RegisterUserRoutes(router *gin.RouterGroup, controller *controllers.UserController, rateLimiter middleware.RateLimiter, cfg *config.Config) {
	users := router.Group("/users")
	{
		// Authentication routes (public) with rate limiting
		authConfig := middleware.RateLimitConfig{
			Requests: cfg.RateLimit.AuthRequestsPerMinute,
			Window:   cfg.RateLimit.AuthWindow,
		}

		users.POST("/register",
			middleware.AuthRateLimitMiddleware(rateLimiter, authConfig),
			controller.RegisterUser)
		users.POST("/login",
			middleware.AuthRateLimitMiddleware(rateLimiter, authConfig),
			controller.LoginUser)
		users.POST("/logout", controller.LogoutUser)
		users.GET("/verify-email", controller.VerifyUserEmail)

		// Profile routes (authenticated)
		profile := users.Group("/profile")
		// Use InternalAuthRequired for internal communication from BFF
		profile.Use(middleware.InternalAuthRequired())
		{
			profile.GET("", controller.GetUserProfile)
			profile.PUT("", controller.UpdateUserProfile)
		}

		// User management routes (authenticated)
		users.Use(middleware.InternalAuthRequired())
		{
			users.GET("", controller.ListAllUsers)
			users.GET("/:id", controller.GetUserByID)
			users.PUT("/:id/role", controller.UpdateUserRole)
			users.POST("/:id/lock", controller.LockAccount)
			users.POST("/:id/unlock", controller.UnlockAccount)
			users.DELETE("/:id/delete", controller.SoftDeleteAccount)
			users.POST("/:id/restore", controller.RestoreAccount)
		}
	}
}
