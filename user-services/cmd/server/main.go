package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-services/internal/api/repositories"
	"user-services/internal/api/services"
	"user-services/internal/cache"
	"user-services/internal/config"
	"user-services/internal/db"
	"user-services/internal/errors"
	"user-services/internal/queue"
	"user-services/internal/server"
	"user-services/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Create main context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("Starting User Services in %s mode", cfg.Environment)
	log.Printf("Server will listen on port %s", cfg.Server.Port)

	// Initialize dependencies
	deps, err := initializeDependencies(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	// Setup graceful shutdown cleanup
	defer cleanupDependencies(deps)

	// Start background workers
	if err := startBackgroundWorkers(ctx, cfg, deps); err != nil {
		log.Fatalf("Failed to start background workers: %v", err)
	}

	// Initialize and start server
	if err := startServer(ctx, cfg, deps); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server gracefully...")

	// Cancel context to signal shutdown
	cancel()

	// Give time for graceful shutdown
	time.Sleep(2 * time.Second)
	log.Println("Shutdown complete")
}

// Dependencies holds all application dependencies
type Dependencies struct {
	DB              interface{}
	RedisClient     interface{}
	RabbitConn      interface{}
	RabbitCh        interface{}
	OutboxProcessor interface{}
}

// initializeDependencies sets up all external connections and services
func initializeDependencies(ctx context.Context, cfg *config.Config) (*Dependencies, error) {
	deps := &Dependencies{}

	// Connect to PostgreSQL with GORM
	gormDB, err := db.ConnectPostgres()
	if err != nil {
		return nil, errors.ErrDatabaseConnection.WithCause(err)
	}
	deps.DB = gormDB

	// Get underlying SQL DB for configuration
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, errors.NewInternalError("Failed to get SQL DB instance").WithCause(err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxIdleTime(cfg.Database.MaxIdleTime)

	// Run database migrations
	if err := db.RunMigrations(gormDB, ""); err != nil {
		return nil, errors.NewInternalError("Failed to run database migrations").WithCause(err)
	}

	// Connect to Redis
	redisClient, err := cache.NewRedisClient(ctx)
	if err != nil {
		return nil, errors.ErrCacheConnection.WithCause(err)
	}
	deps.RedisClient = redisClient

	// Connect to RabbitMQ
	rabbitConn, rabbitCh, err := queue.NewRabbitMQ(ctx)
	if err != nil {
		return nil, errors.ErrQueueConnection.WithCause(err)
	}
	deps.RabbitConn = rabbitConn
	deps.RabbitCh = rabbitCh

	// Declare exchange
	err = rabbitCh.ExchangeDeclare(
		cfg.RabbitMQ.ExchangeName, // name
		"topic",                   // type
		true,                      // durable
		false,                     // auto-deleted
		false,                     // internal
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		return nil, errors.NewExternalServiceError("RabbitMQ", "Failed to declare exchange").WithCause(err)
	}

	log.Printf("Successfully connected to all external services")
	return deps, nil
}

// cleanupDependencies handles graceful cleanup of resources
func cleanupDependencies(deps *Dependencies) {
	log.Println("Cleaning up dependencies...")

	// Close RabbitMQ connection
	if rabbitCh, ok := deps.RabbitCh.(interface{ Close() error }); ok {
		if err := rabbitCh.Close(); err != nil {
			log.Printf("Error closing RabbitMQ channel: %v", err)
		}
	}

	if rabbitConn, ok := deps.RabbitConn.(interface{ Close() error }); ok {
		if err := rabbitConn.Close(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		}
	}

	// Close Redis connection
	if redisClient, ok := deps.RedisClient.(interface{ Close() error }); ok {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}

	// Close database connection
	if gormDB, ok := deps.DB.(interface{ DB() (*sql.DB, error) }); ok {
		if sqlDB, err := gormDB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Error closing database connection: %v", err)
			}
		}
	}
}

// startBackgroundWorkers initializes background workers
func startBackgroundWorkers(ctx context.Context, cfg *config.Config, deps *Dependencies) error {
	gormDB := deps.DB
	rabbitCh := deps.RabbitCh

	// Initialize Outbox Service
	outboxRepo := repositories.NewOutboxRepository(gormDB.(*gorm.DB))
	outboxService := services.NewOutboxService(outboxRepo, rabbitCh.(*amqp091.Channel), cfg.RabbitMQ.ExchangeName)

	// Start Outbox Processor
	outboxProcessor := worker.NewOutboxProcessor(
		outboxService,
		5*time.Second, // TODO: Make configurable
		10,            // TODO: Make configurable
	)

	// Run outbox processor in background
	go outboxProcessor.Start(ctx)
	deps.OutboxProcessor = outboxProcessor

	log.Println("Background workers started")
	return nil
}

// startServer initializes and starts the HTTP server
func startServer(ctx context.Context, cfg *config.Config, deps *Dependencies) error {
	// Initialize router with dependencies
	r := server.NewRouter(server.Deps{
		DB:          deps.DB.(*gorm.DB),
		RedisClient: deps.RedisClient.(*redis.Client),
	})

	// Configure server with timeouts from configuration
	r.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	// Start server in goroutine
	go func() {
		addr := ":" + cfg.Server.Port
		log.Printf("Server starting on %s", addr)

		if err := r.Run(addr); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	return nil
}
