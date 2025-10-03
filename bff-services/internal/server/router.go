package server

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/services"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	UserService services.UserService
}

func NewRouter(deps Deps) *gin.Engine {
	r := gin.New()

	// Middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", controllers.Health)

	var authCtrl *controllers.AuthController
	if deps.UserService != nil {
		authCtrl = controllers.NewAuthController(deps.UserService)
	}

	api := r.Group("/api/v1")
	{
		api.GET("/health", controllers.Health)
		if authCtrl != nil {
			api.POST("/user/register", authCtrl.Register)
			api.POST("/user/login", authCtrl.Login)
		}
	}

	return r
}
