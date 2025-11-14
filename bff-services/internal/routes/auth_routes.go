package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes configures authentication-related routes
func SetupAuthRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers == nil || controllers.User == nil || sessionCache == nil {
		return
	}

	// Public authentication routes
	api.POST("/users/register", controllers.User.Register)
	api.POST("/users/login", controllers.User.Login)
	api.POST("/users/logout", controllers.User.Logout)
	api.GET("/users/verify-email", controllers.User.VerifyEmail)

	// Protected profile routes
	profile := api.Group("/users/profile")
	profile.Use(middleware.AuthRequired(sessionCache))
	{
		profile.GET("", controllers.User.GetProfile)
		profile.PUT("", controllers.User.UpdateProfile)
	}
}
