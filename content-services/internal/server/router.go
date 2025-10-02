package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter configures routes and middleware and returns a Gin engine.
func NewRouter() *gin.Engine {
	r := gin.New()
	// Middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}

// NewRouterWithGraphQL returns a Gin engine with the provided GraphQL handler mounted at /graphql
func NewRouterWithGraphQL(graphqlHandler http.Handler) *gin.Engine {
	r := NewRouter()
	if graphqlHandler != nil {
		r.Any("/graphql", gin.WrapH(graphqlHandler))
	}
	return r
}
