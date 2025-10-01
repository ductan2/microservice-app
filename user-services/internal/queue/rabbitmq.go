package queue

import (
	"context"
	"time"

	"user-services/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

// NewRabbitMQ dials RabbitMQ and returns a connection and channel.
func NewRabbitMQ(ctx context.Context) (*amqp.Connection, *amqp.Channel, error) {
	cfg := config.GetRabbitConfig()

	// Create connection with timeout
	conn, err := amqp.DialConfig(cfg.URL, amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
	})
	if err != nil {
		return nil, nil, err
	}

	// Create channel with timeout
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	// Set QoS for better performance
	err = ch.Qos(10, 0, false)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}

// NewRabbitMQConnection creates a new RabbitMQ connection with retry logic
func NewRabbitMQConnection(ctx context.Context, maxRetries int, retryDelay time.Duration) (*amqp.Connection, error) {
	cfg := config.GetRabbitConfig()

	var conn *amqp.Connection
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.DialConfig(cfg.URL, amqp.Config{
			Heartbeat: 10 * time.Second,
			Locale:    "en_US",
		})

		if err == nil {
			return conn, nil
		}

		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay):
				// Continue to next retry
			}
		}
	}

	return nil, err
}

// NewRabbitMQChannel creates a new RabbitMQ channel with error handling
func NewRabbitMQChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Set QoS for better performance
	err = ch.Qos(10, 0, false)
	if err != nil {
		_ = ch.Close()
		return nil, err
	}

	return ch, nil
}
