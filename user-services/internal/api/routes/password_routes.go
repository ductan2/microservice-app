package routes

import (
	"user-services/internal/api/controllers"
	"user-services/internal/api/middleware"
	"user-services/internal/cache"

	"github.com/gin-gonic/gin"
)

func RegisterPasswordRoutes(router *gin.RouterGroup, controller *controllers.PasswordController, sessionCache *cache.SessionCache) {
	password := router.Group("/password")
	{
		password.POST("/reset/request", controller.RequestPasswordReset) // POST /password/reset/request
		password.POST("/reset/confirm", controller.ConfirmPasswordReset) // POST /password/reset/confirm

		// Change password requires authentication
		password.Use(middleware.AuthRequired(sessionCache))
		password.POST("/change", controller.ChangePassword) // POST /password/change (authenticated)
	}
}
