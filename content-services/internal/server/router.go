package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter configures routes and middleware and returns a Gin engine.
func NewRouter(graphqlHandler http.Handler) *gin.Engine {
	r := gin.New()
	// Middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	if graphqlHandler != nil {
		r.Any("/graphql", gin.WrapH(graphqlHandler))
	}

	return r
}
