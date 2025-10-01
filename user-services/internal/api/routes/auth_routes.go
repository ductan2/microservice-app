package routes

import (
	"github.com/gin-gonic/gin"
	"user-services/internal/api/controllers"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, ctrl controllers.AuthController) {
	rg.POST("/register", ctrl.Register)
	rg.POST("/login", ctrl.Login)
	rg.POST("/logout", ctrl.Logout)
}
