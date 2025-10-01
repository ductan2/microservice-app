package server

import (
	"github.com/gin-gonic/gin"
	"content-services/internal/api/controllers"
)

// NewRouter configures routes and middleware and returns a Gin engine.
func NewRouter() *gin.Engine {
	r := gin.New()
	// Middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Routes
	r.GET("/health", handlers.Health)

	return r
}
