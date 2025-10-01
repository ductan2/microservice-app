package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"user-services/internal/config"
)

// NewRedisClient creates a redis client and PINGs it.
func NewRedisClient(ctx context.Context) (*redis.Client, error) {
	cfg := config.GetRedisConfig()
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       0,
	})

	ctxPing, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := client.Ping(ctxPing).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}
