package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes configures user management routes
func SetupUserRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache, userService services.UserService) {
	if controllers == nil || controllers.User == nil || sessionCache == nil {
		return
	}

	users := api.Group("/users")
	users.Use(middleware.AuthRequired(sessionCache))
	{
		// Regular user routes
		users.GET("/:id", controllers.User.GetUserById)
	}

	// Admin-only routes
	adminUsers := api.Group("/users")
	adminUsers.Use(middleware.AuthRequired(sessionCache))
	adminUsers.Use(middleware.AdminRequired(userService))
	{
		adminUsers.GET("", controllers.User.ListUsersWithProgress)
		adminUsers.PUT("/:id/role", controllers.User.UpdateUserRole)
		adminUsers.POST("/:id/lock", controllers.User.LockAccount)
		adminUsers.POST("/:id/unlock", controllers.User.UnlockAccount)
		adminUsers.DELETE("/:id/delete", controllers.User.SoftDeleteAccount)
		adminUsers.POST("/:id/restore", controllers.User.RestoreAccount)
	}
}
