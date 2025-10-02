package main

import (
	"content-services/graph/generated"
	gqlresolver "content-services/graph/resolver"
	"content-services/internal/config"
	"content-services/internal/db"
	"content-services/internal/server"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

func main() {
	// Determine port (defaults to 8001)
	port := config.GetPort()

	addr := ":" + port
	// Init Mongo
	mongoClient, err := db.NewMongoClient(context.Background())
	if err != nil {
		log.Fatalf("mongo connect error: %v", err)
	}
	database := db.GetDatabase(mongoClient)

	// Build GraphQL server
	resolver := &gqlresolver.Resolver{DB: database}
	gqlSrv := generated.NewExecutableSchema(generated.Config{Resolvers: resolver})
	graphqlHandler := handler.NewDefaultServer(gqlSrv)

	r := server.NewRouterWithGraphQL(graphqlHandler)
	if config.GetGraphQLPlaygroundEnabled() {
		// Expose playground at root
		r.GET("/", func(c *gin.Context) {
			playground.Handler("GraphQL", "/graphql").ServeHTTP(c.Writer, c.Request)
		})
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server in background
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
