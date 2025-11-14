package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order-services/internal/config"
	"order-services/internal/controllers"
	"order-services/internal/db"
	"order-services/internal/repositories"
	"order-services/internal/router"
	"order-services/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Could not load .env file")
	}

	cfg := config.GetConfig()

	// Initialize dependencies
	engine, cleanup := buildServer(cfg)
	defer cleanup()

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Order Services server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func buildServer(cfg *config.Config) (*gin.Engine, func()) {
	gormDB, err := db.ConnectPostgres()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Repositories
	orderRepo := repositories.NewOrderRepository(gormDB)
	orderItemRepo := repositories.NewOrderItemRepository(gormDB)
	couponRepo := repositories.NewCouponRepository(gormDB)
	paymentRepo := repositories.NewPaymentRepository(gormDB)
	outboxRepo := repositories.NewOutboxRepository(sqlDB)
	webhookRepo := repositories.NewWebhookEventRepository(sqlDB)
	courseRepo := repositories.NewCourseRepository(cfg.CourseServiceURL)

	// Services
	orderService := services.NewOrderService(orderRepo, orderItemRepo, couponRepo, courseRepo, outboxRepo, cfg)
	couponService := services.NewCouponService(couponRepo, orderRepo)
	paymentService := services.NewPaymentService(orderRepo, paymentRepo, outboxRepo, webhookRepo, cfg)

	// Controllers
	orderController := controllers.NewOrderController(orderService)
	couponController := controllers.NewCouponController(couponService)
	paymentController := controllers.NewPaymentController(paymentService, cfg)

	engine := router.NewRouter(router.Dependencies{
		OrderController:   orderController,
		PaymentController: paymentController,
		CouponController:  couponController,
		JWTSecret:         cfg.JWTSecret,
	})

	cleanup := func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}

	return engine, cleanup
}
