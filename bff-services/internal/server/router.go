package server

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/config"
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
	r.Use(func(c *gin.Context) {
		origin := config.GetCORSOrigin()
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

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
