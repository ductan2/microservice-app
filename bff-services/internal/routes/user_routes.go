package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers.User == nil {
		return
	}

	if sessionCache != nil {
		users := api.Group("/users")
		users.Use(middleware.AuthRequired(sessionCache))
		{
			users.GET("", controllers.User.ListUsersWithProgress)
			users.GET("/:id", controllers.User.GetUserById)
			users.PUT("/:id/role", controllers.User.UpdateUserRole)
			users.POST("/:id/lock", controllers.User.LockAccount)
			users.POST("/:id/unlock", controllers.User.UnlockAccount)
			users.DELETE("/:id/delete", controllers.User.SoftDeleteAccount)
			users.POST("/:id/restore", controllers.User.RestoreAccount)
		}
	}
}
