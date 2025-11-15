package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port string

	// External services
	CourseServiceURL string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// RabbitMQ
	RabbitMQHost     string
	RabbitMQPort     string
	RabbitMQUser     string
	RabbitMQPassword string
	RabbitMQVHost    string

	// Stripe
	StripeSecretKey      string
	StripeWebhookSecret  string
	StripePublishableKey string

	// JWT
	JWTSecret string

	// Application
	AppName          string
	AppVersion       string
	LogLevel         string
	Environment      string
	OrderExpiresIn   int // order expiration in hours
	RefundWindowDays int
}

var cfg *Config

func init() {
	// Load .env if present; non-fatal if missing (e.g., in production/containers)
	if err := godotenv.Load(); err != nil {
		// It's okay if .env is not present; environment variables can be provided directly
	}
	cfg = loadConfig()
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	return &Config{
		// Server
		Port: getEnv("PORT", "8006"),

		// External services
		CourseServiceURL: getEnv("COURSE_SERVICE_URL", "http://localhost:8010"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "user"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "lms_order_serivecs"),

		// Redis
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", "redis_password"),

		// RabbitMQ
		RabbitMQHost:     getEnv("RABBITMQ_HOST", "localhost"),
		RabbitMQPort:     getEnv("RABBITMQ_PORT", "5672"),
		RabbitMQUser:     getEnv("RABBITMQ_USER", "user"),
		RabbitMQPassword: getEnv("RABBITMQ_PASSWORD", "password"),
		RabbitMQVHost:    getEnv("RABBITMQ_VHOST", "/"),

		// Stripe
		StripeSecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripePublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),

		// Application
		AppName:          getEnv("APP_NAME", "order-services"),
		AppVersion:       getEnv("APP_VERSION", "1.0.0"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		Environment:      getEnv("ENVIRONMENT", "development"),
		OrderExpiresIn:   getEnvInt("ORDER_EXPIRES_IN", 24), // 24 hours default
		RefundWindowDays: getEnvInt("REFUND_WINDOW_DAYS", 30),
	}
}

// GetConfig returns the application configuration
func GetConfig() *Config {
	return cfg
}

// DatabaseURL returns the PostgreSQL connection string
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

// RedisAddr returns the Redis address
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

// RabbitMQURL returns the RabbitMQ connection string
func (c *Config) RabbitMQURL() string {
	if c.RabbitMQVHost == "/" {
		return fmt.Sprintf("amqp://%s:%s@%s:%s/",
			c.RabbitMQUser, c.RabbitMQPassword, c.RabbitMQHost, c.RabbitMQPort)
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		c.RabbitMQUser, c.RabbitMQPassword, c.RabbitMQHost, c.RabbitMQPort, c.RabbitMQVHost)
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// ValidateConfig checks if required configuration is set
func (c *Config) ValidateConfig() error {
	if c.StripeSecretKey == "" {
		return fmt.Errorf("STRIPE_SECRET_KEY is required")
	}
	if c.StripeWebhookSecret == "" {
		return fmt.Errorf("STRIPE_WEBHOOK_SECRET is required")
	}
	if c.JWTSecret == "" || c.JWTSecret == "your-secret-key-change-in-production" {
		if c.IsProduction() {
			return fmt.Errorf("JWT_SECRET must be set in production")
		}
	}
	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
