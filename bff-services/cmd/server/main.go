package main

import (
	"bff-services/internal/config"
	"bff-services/internal/server"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	port := config.GetPort()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)


	addr := ":" + port
	log.Printf("Starting server on %s", addr)
	if err := server.NewRouter(server.Deps{
	}).Run(addr); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
	
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server gracefully...")
}