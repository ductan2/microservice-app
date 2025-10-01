package routes

import (
	"user-services/internal/api/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterPasswordRoutes(router *gin.RouterGroup, controller *controllers.PasswordController) {
	password := router.Group("/password")
	{
		password.POST("/reset/request", controller.RequestPasswordReset) // POST /password/reset/request
		password.POST("/reset/confirm", controller.ConfirmPasswordReset) // POST /password/reset/confirm
		password.POST("/change", controller.ChangePassword)              // POST /password/change (authenticated)
	}
}
