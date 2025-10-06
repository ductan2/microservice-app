package main

import (
	"bff-services/internal/cache"
	"bff-services/internal/config"
	"bff-services/internal/server"
	"bff-services/internal/services"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	port := config.GetPort()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Initialize Redis client
	redisConfig := config.GetRedisConfig()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       0,
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	} else {
		log.Println("Redis connected successfully")
	}

	// Initialize session cache
	sessionCache := cache.NewSessionCache(redisClient)

	userService := services.NewUserServiceClient(config.GetUserServiceURL(), nil)
	contentService := services.NewContentServiceClient(config.GetContentServiceURL(), nil)
	lessonService := services.NewLessonServiceClient(config.GetLessonServiceURL(), nil)

	addr := ":" + port
	r := server.NewRouter(server.Deps{
		UserService:    userService,
		ContentService: contentService,
		LessonService:  lessonService,
		SessionCache:   sessionCache,
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Printf("Starting server on %s", addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal then attempt graceful shutdown
	<-quit
	log.Println("Shutting down server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}
