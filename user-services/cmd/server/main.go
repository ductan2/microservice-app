package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-services/internal/api/repositories"
	"user-services/internal/api/services"
	"user-services/internal/config"
	"user-services/internal/db"
	"user-services/internal/queue"
	"user-services/internal/server"
	"user-services/internal/worker"
)

func main() {
	// Determine port (defaults to 8001)
	port := config.GetPort()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Create main context
	ctx := context.Background()

	// Connect to PostgreSQL with GORM
	gormDB, err := db.ConnectPostgres()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get underlying SQL DB for cleanup
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("Failed to get SQL DB: %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Auto-migrate database schema
	if err := db.AutoMigrate(gormDB); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Connect to RabbitMQ
	rabbitConn, rabbitCh, err := queue.NewRabbitMQ(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer func() {
		if rabbitCh != nil {
			_ = rabbitCh.Close()
		}
		if rabbitConn != nil {
			_ = rabbitConn.Close()
		}
	}()

	// Declare exchange
	exchangeName := getEnv("RABBITMQ_EXCHANGE", "notifications")
	err = rabbitCh.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	log.Printf("RabbitMQ connected and exchange '%s' declared", exchangeName)

	// Initialize Outbox Service
	outboxRepo := repositories.NewOutboxRepository(gormDB)
	outboxService := services.NewOutboxService(outboxRepo, rabbitCh, exchangeName)

	// Start Outbox Processor
	outboxProcessor := worker.NewOutboxProcessor(
		outboxService,
		5*time.Second, // process every 5 seconds
		10,            // batch size
	)

	// Run outbox processor in background
	go outboxProcessor.Start(ctx)
	log.Println("Outbox processor started")

	// Initialize router with dependencies
	r := server.NewRouter(server.Deps{DB: gormDB})

	// Start server
	addr := ":" + port
	log.Printf("Starting server on %s", addr)

	// Run server in goroutine
	go func() {
		if err := r.Run(addr); err != nil {
			log.Printf("Server error: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server gracefully...")

	// Stop outbox processor
	outboxProcessor.Stop()
	time.Sleep(1 * time.Second) // Give time for graceful shutdown
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
