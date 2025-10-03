package server

import (
	"bff-services/internal/api/controllers"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	// RedisClient *redis.Client
}

func NewRouter(deps Deps) *gin.Engine {
	r := gin.New()

	// Middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", controllers.Health)


	api := r.Group("/api/v1")
	{
		api.GET("/health", controllers.Health)
	}

	return r
}
