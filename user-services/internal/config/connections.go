package config

import (
	"fmt"
)

// GetPostgresConfig returns the PostgreSQL configuration from the global config
func GetPostgresConfig() DatabaseConfig {
	return GetConfig().Database
}

// DSN returns a standard libpq connection string and URL.
func (c DatabaseConfig) DSN() (connString string, url string) {
	connString = fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
	url = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
	return
}

// GetRedisConfig returns the Redis configuration from the global config
func GetRedisConfig() RedisConfig {
	return GetConfig().Redis
}

// GetRabbitConfig returns the RabbitMQ configuration from the global config
func GetRabbitConfig() RabbitConfig {
	return GetConfig().RabbitMQ
}
