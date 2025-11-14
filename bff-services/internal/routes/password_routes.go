package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupPasswordRoutes configures password management routes
func SetupPasswordRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers == nil || controllers.Password == nil || sessionCache == nil {
		return
	}

	password := api.Group("/password")
	{
		// Public password reset routes
		password.POST("/reset/request", controllers.Password.RequestReset)
		password.POST("/reset/confirm", controllers.Password.ConfirmReset)
	}

	// Protected password change routes
	protectedPassword := password.Group("")
	protectedPassword.Use(middleware.AuthRequired(sessionCache))
	{
		protectedPassword.POST("/change", controllers.Password.ChangePassword)
	}
}
