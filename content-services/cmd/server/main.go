package main

import (
	"content-services/graph/generated"
	gqlresolver "content-services/graph/resolver"
	"content-services/internal/config"
	"content-services/internal/db"
	"content-services/internal/repository"
	"content-services/internal/server"
	"content-services/internal/service"
	"content-services/internal/storage"
	"content-services/internal/taxonomy"
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

	// Prepare taxonomy store backed by MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	taxonomyStore, err := taxonomy.NewStore(ctx, database)
	cancel()
	if err != nil {
		log.Fatalf("taxonomy store init error: %v", err)
	}

	// Build GraphQL server
	mediaRepo := repository.NewMediaRepository(database)
	s3Client, err := storage.NewS3Client(context.Background(), storage.S3Config{
		Endpoint:        config.GetS3Endpoint(),
		Region:          config.GetS3Region(),
		Bucket:          config.GetS3Bucket(),
		AccessKeyID:     config.GetS3AccessKeyID(),
		SecretAccessKey: config.GetS3SecretAccessKey(),
		UsePathStyle:    config.GetS3UsePathStyle(),
		PresignExpires:  config.GetS3PresignTTL(),
	})
	if err != nil {
		log.Fatalf("s3 init error: %v", err)
	}
	mediaService := service.NewMediaService(mediaRepo, s3Client, config.GetS3PresignTTL())

	resolver := &gqlresolver.Resolver{DB: database, Taxonomy: taxonomyStore, Media: mediaService}
	gqlSrv := generated.NewExecutableSchema(generated.Config{Resolvers: resolver})
	graphqlHandler := handler.NewDefaultServer(gqlSrv)

	r := server.NewRouter(graphqlHandler)
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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
