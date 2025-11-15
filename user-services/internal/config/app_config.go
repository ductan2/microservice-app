package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	RabbitMQ    RabbitConfig
	JWT         JWTConfig
	Session     SessionConfig
	Email       EmailConfig
	Security    SecurityConfig
	RateLimit   RateLimitConfig
	Environment string
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
	SSLMode  string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  time.Duration
}

// RedisConfig contains Redis connection configuration
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// RabbitConfig contains RabbitMQ connection configuration
type RabbitConfig struct {
	URL              string
	User             string
	Password         string
	Host             string
	Port             string
	VHost            string
	ExchangeName     string
	ReconnectDelay   time.Duration
	Heartbeat        time.Duration
}

// JWTConfig contains JWT configuration
type JWTConfig struct {
	Secret           string
	ExpiresIn        time.Duration
	RefreshExpiresIn time.Duration
	Issuer           string
}

// SessionConfig contains session configuration
type SessionConfig struct {
	Expiry        time.Duration
	CleanupPeriod time.Duration
	MaxSessions   int
}

// EmailConfig contains email configuration
type EmailConfig struct {
	FrontendURL       string
	VerificationExpiry time.Duration
	PasswordResetExpiry time.Duration
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	PasswordMinLength     int
	PasswordRequireUpper   bool
	PasswordRequireLower   bool
	PasswordRequireDigit   bool
	PasswordRequireSpecial bool
	MaxLoginAttempts       int
	LoginAttemptWindow     time.Duration
	LockoutDuration        time.Duration
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	// Authentication endpoints
	AuthRequestsPerMinute    int           `env:"RATE_LIMIT_AUTH_REQUESTS" envDefault:"10"`
	AuthWindow               time.Duration `env:"RATE_LIMIT_AUTH_WINDOW" envDefault:"1m"`

	// Password reset endpoints
	PasswordResetPerHour     int           `env:"RATE_LIMIT_PASSWORD_RESET" envDefault:"3"`
	PasswordResetWindow      time.Duration `env:"RATE_LIMIT_PASSWORD_WINDOW" envDefault:"1h"`

	// Registration endpoints
	RegisterPerHour          int           `env:"RATE_LIMIT_REGISTER" envDefault:"5"`
	RegisterWindow           time.Duration `env:"RATE_LIMIT_REGISTER_WINDOW" envDefault:"1h"`

	// Account-specific limits
	AccountFailureThreshold  int           `env:"RATE_LIMIT_ACCOUNT_FAILURES" envDefault:"5"`
	AccountFailureWindow     time.Duration `env:"RATE_LIMIT_ACCOUNT_WINDOW" envDefault:"15m"`
	AccountBlockDuration     time.Duration `env:"RATE_LIMIT_ACCOUNT_BLOCK" envDefault:"30m"`

	// Progressive backoff settings
	EnableProgressiveBackoff bool          `env:"RATE_LIMIT_PROGRESSIVE_BACKOFF" envDefault:"true"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if present (non-fatal if missing)
	if err := godotenv.Load(); err != nil {
		// It's okay if .env is not present; environment variables can be provided directly
	}

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	// Load server configuration
	cfg.Server = ServerConfig{
		Port:         getEnv("PORT", "8001"),
		ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
		WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
	}

	// Load database configuration
	cfg.Database = DatabaseConfig{
		User:         getEnv("DB_USER", "user"),
		Password:     getEnv("DB_PASSWORD", "password"),
		Host:         getEnv("DB_HOST", "postgres"),
		Port:         getEnv("DB_PORT", "5432"),
		DBName:       getEnv("DB_NAME", "lms_user_services"),
		SSLMode:      getEnv("POSTGRES_SSLMODE", "disable"),
		MaxOpenConns: getIntEnv("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns: getIntEnv("DB_MAX_IDLE_CONNS", 5),
		MaxIdleTime:  getDurationEnv("DB_MAX_IDLE_TIME", 5*time.Minute),
	}

	// Load Redis configuration
	cfg.Redis = RedisConfig{
		Addr:         getEnv("REDIS_HOST", "redis") + ":" + getEnv("REDIS_PORT", "6379"),
		Password:     getEnv("REDIS_PASSWORD", ""),
		DB:           getIntEnv("REDIS_DB", 0),
		PoolSize:     getIntEnv("REDIS_POOL_SIZE", 10),
		MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 3),
		DialTimeout:  getDurationEnv("REDIS_DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout:  getDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second),
		WriteTimeout: getDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second),
	}

	// Load RabbitMQ configuration
	cfg.RabbitMQ = RabbitConfig{
		URL:            getEnv("RABBITMQ_URL", ""),
		User:           getEnv("RABBITMQ_USER", "user"),
		Password:       getEnv("RABBITMQ_PASSWORD", "password"),
		Host:           getEnv("RABBITMQ_HOST", "rabbitmq"),
		Port:           getEnv("RABBITMQ_PORT", "5672"),
		VHost:          getEnv("RABBITMQ_VHOST", "/"),
		ExchangeName:   getEnv("RABBITMQ_EXCHANGE", "notifications"),
		ReconnectDelay: getDurationEnv("RABBITMQ_RECONNECT_DELAY", 5*time.Second),
		Heartbeat:      getDurationEnv("RABBITMQ_HEARTBEAT", 10*time.Second),
	}

	// Build RabbitMQ URL if not provided
	if cfg.RabbitMQ.URL == "" {
		cfg.RabbitMQ.URL = fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
			cfg.RabbitMQ.User,
			cfg.RabbitMQ.Password,
			cfg.RabbitMQ.Host,
			cfg.RabbitMQ.Port,
			cfg.RabbitMQ.VHost,
		)
	}

	// Load JWT configuration
	cfg.JWT = JWTConfig{
		Secret:           getEnv("JWT_SECRET", "change-me-dev-secret"),
		ExpiresIn:        getDurationEnv("JWT_EXPIRES_IN", 24*time.Hour),
		RefreshExpiresIn: getDurationEnv("JWT_REFRESH_EXPIRES_IN", 60*24*time.Hour),
		Issuer:           getEnv("JWT_ISSUER", "user-services"),
	}

	// Load session configuration
	cfg.Session = SessionConfig{
		Expiry:        getDurationEnv("SESSION_EXPIRY", 30*24*time.Hour), // 30 days
		CleanupPeriod: getDurationEnv("SESSION_CLEANUP_PERIOD", 1*time.Hour),
		MaxSessions:   getIntEnv("SESSION_MAX_SESSIONS", 10),
	}

	// Load email configuration
	cfg.Email = EmailConfig{
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3001"),
		VerificationExpiry: getDurationEnv("EMAIL_VERIFICATION_EXPIRY", 24*time.Hour),
		PasswordResetExpiry: getDurationEnv("EMAIL_PASSWORD_RESET_EXPIRY", 2*time.Hour),
	}

	// Load security configuration
	cfg.Security = SecurityConfig{
		PasswordMinLength:     getIntEnv("SECURITY_PASSWORD_MIN_LENGTH", 8),
		PasswordRequireUpper:   getBoolEnv("SECURITY_PASSWORD_REQUIRE_UPPER", true),
		PasswordRequireLower:   getBoolEnv("SECURITY_PASSWORD_REQUIRE_LOWER", true),
		PasswordRequireDigit:   getBoolEnv("SECURITY_PASSWORD_REQUIRE_DIGIT", true),
		PasswordRequireSpecial: getBoolEnv("SECURITY_PASSWORD_REQUIRE_SPECIAL", true),
		MaxLoginAttempts:       getIntEnv("SECURITY_MAX_LOGIN_ATTEMPTS", 5),
		LoginAttemptWindow:     getDurationEnv("SECURITY_LOGIN_ATTEMPT_WINDOW", 15*time.Minute),
		LockoutDuration:        getDurationEnv("SECURITY_LOCKOUT_DURATION", 30*time.Minute),
	}

	// Load rate limiting configuration
	cfg.RateLimit = RateLimitConfig{
		AuthRequestsPerMinute:    getIntEnv("RATE_LIMIT_AUTH_REQUESTS", 10),
		AuthWindow:               getDurationEnv("RATE_LIMIT_AUTH_WINDOW", 1*time.Minute),
		PasswordResetPerHour:     getIntEnv("RATE_LIMIT_PASSWORD_RESET", 3),
		PasswordResetWindow:      getDurationEnv("RATE_LIMIT_PASSWORD_WINDOW", 1*time.Hour),
		RegisterPerHour:          getIntEnv("RATE_LIMIT_REGISTER", 5),
		RegisterWindow:           getDurationEnv("RATE_LIMIT_REGISTER_WINDOW", 1*time.Hour),
		AccountFailureThreshold:  getIntEnv("RATE_LIMIT_ACCOUNT_FAILURES", 5),
		AccountFailureWindow:     getDurationEnv("RATE_LIMIT_ACCOUNT_WINDOW", 15*time.Minute),
		AccountBlockDuration:     getDurationEnv("RATE_LIMIT_ACCOUNT_BLOCK", 30*time.Minute),
		EnableProgressiveBackoff: getBoolEnv("RATE_LIMIT_PROGRESSIVE_BACKOFF", true),
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Environment == "production" {
		if c.JWT.Secret == "change-me-dev-secret" {
			return fmt.Errorf("JWT_SECRET must be set in production")
		}
		if len(c.JWT.Secret) < 32 {
			return fmt.Errorf("JWT_SECRET must be at least 32 characters in production")
		}
	}

	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if c.Database.Password == "" && c.Environment != "development" {
		return fmt.Errorf("database password is required in production")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}