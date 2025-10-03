package main

import (
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
)

func main() {
	port := config.GetPort()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	userService := services.NewUserServiceClient(config.GetUserServiceURL(), nil)

	addr := ":" + port
	r := server.NewRouter(server.Deps{
		UserService: userService,
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}
