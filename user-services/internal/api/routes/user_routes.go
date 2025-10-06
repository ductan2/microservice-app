package routes

import (
	"user-services/internal/api/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(rg *gin.RouterGroup, ctrl *controllers.UserController) {
	users := rg.Group("/users")
	{
		users.GET("", ctrl.ListUsers) // GET /users
	}
}
