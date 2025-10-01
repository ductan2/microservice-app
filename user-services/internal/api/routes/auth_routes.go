package routes

import (
	"user-services/internal/api/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, ctrl controllers.AuthController) {
	rg.POST("/register", ctrl.Register)
	rg.POST("/login", ctrl.Login)
	rg.POST("/logout", ctrl.Logout)
	rg.GET("/verify-email", ctrl.VerifyEmail)
}
