package cache

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"order-services/internal/config"
)

// NewRedisClient creates a redis client and PINGs it.
func NewRedisClient(ctx context.Context) (*redis.Client, error) {
	cfg := config.GetConfig()
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.RedisPassword,
		DB:       0,
	})

	ctxPing, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := client.Ping(ctxPing).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	log.Println("✅ Connected to Redis")
	return client, nil
}