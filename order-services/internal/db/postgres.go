package db

import (
	"fmt"
	"log"
	"time"

	"order-services/internal/config"
	"order-services/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectPostgres creates a GORM connection to PostgreSQL
func ConnectPostgres() (*gorm.DB, error) {
	cfg := config.GetConfig()
	dsn := cfg.DatabaseURL()

	// Configure GORM logger based on environment
	logLevel := logger.Silent
	if cfg.IsDevelopment() {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Connected to PostgreSQL with GORM")
	return db, nil
}

// AutoMigrate runs GORM auto-migration for all models
func AutoMigrate(db *gorm.DB) error {
	// Enable UUID extension
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	// Get all models and auto-migrate
	allModels := models.AllModels()
	if err := db.AutoMigrate(allModels...); err != nil {
		return fmt.Errorf("auto-migrate failed: %w", err)
	}

	log.Println("✅ Database auto-migration completed")
	return nil
}