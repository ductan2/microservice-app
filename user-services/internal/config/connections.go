package config

import (
	"fmt"
	"os"
)

// PostgresConfig holds database connection parameters.
type PostgresConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
	SSLMode  string
}

func GetPostgresConfig() PostgresConfig {
	return PostgresConfig{
		User:     getenv("DB_USER", "user"),
		Password: getenv("DB_PASSWORD", "password"),
		Host:     getenv("DB_HOST", "postgres"),
		Port:     getenv("DB_PORT", "5432"),
		DBName:   getenv("DB_NAME", "english_app"),
		SSLMode:  getenv("POSTGRES_SSLMODE", "disable"),
	}
}

// DSN returns a standard libpq connection string and URL.
func (c PostgresConfig) DSN() (connString string, url string) {
	connString = fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
	url = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
	return
}

// Redis

type RedisConfig struct {
	Addr     string
	Password string
}

func GetRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:     getenv("REDIS_HOST", "redis") + ":" + getenv("REDIS_PORT", "6379"),
		Password: os.Getenv("REDIS_PASSWORD"),
	}
}

// RabbitMQ

type RabbitConfig struct {
	URL string
}

func GetRabbitConfig() RabbitConfig {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		user := getenv("RABBITMQ_USER", "user")
		pass := getenv("RABBITMQ_PASSWORD", "password")
		host := getenv("RABBITMQ_HOST", "rabbitmq")
		port := getenv("RABBITMQ_PORT", "5672")
		vhost := getenv("RABBITMQ_VHOST", "/")
		url = fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, pass, host, port, vhost)
	}
	return RabbitConfig{URL: url}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
