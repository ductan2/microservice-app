package db

import (
	"fmt"
	"log"
	"time"

	"user-services/internal/config"
	"user-services/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectPostgres creates a GORM connection to PostgreSQL
func ConnectPostgres() (*gorm.DB, error) {
	cfg := config.GetPostgresConfig()
	dsn, _ := cfg.DSN()

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
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

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.User{},
		&models.UserProfile{},
		&models.Session{},
		&models.RefreshToken{},
		&models.MFAMethod{},
		&models.LoginAttempt{},
		&models.PasswordReset{},
		&models.UserActivitySession{},
		&models.AuditLog{},
		&models.Outbox{},
	); err != nil {
		return fmt.Errorf("auto-migrate failed: %w", err)
	}

	log.Println("✅ Database auto-migration completed")
	return nil
}
