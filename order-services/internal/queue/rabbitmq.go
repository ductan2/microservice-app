package queue

import (
	"context"
	"log"
	"time"

	"order-services/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Exchange and queue constants
const (
	// Exchanges
	OrdersExchange = "orders"
	EventsExchange = "events"

	// Routing keys
	OrderCreatedRoutingKey   = "order.created"
	OrderPaidRoutingKey      = "order.paid"
	OrderFailedRoutingKey    = "order.failed"
	OrderCancelledRoutingKey = "order.cancelled"
	OrderRefundedRoutingKey  = "order.refunded"

	// Queue names
	OrderEventsQueue = "order.events"
	LessonEventsQueue = "lesson.events"
	NotificationEventsQueue = "notification.events"
)

// NewRabbitMQ dials RabbitMQ and returns a connection and channel.
func NewRabbitMQ(ctx context.Context) (*amqp.Connection, *amqp.Channel, error) {
	cfg := config.GetConfig()
	url := cfg.RabbitMQURL()

	// Create connection with timeout
	conn, err := amqp.DialConfig(url, amqp.Config{
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

	log.Println("✅ Connected to RabbitMQ")
	return conn, ch, nil
}

// NewRabbitMQConnection creates a new RabbitMQ connection with retry logic
func NewRabbitMQConnection(ctx context.Context, maxRetries int, retryDelay time.Duration) (*amqp.Connection, error) {
	cfg := config.GetConfig()
	url := cfg.RabbitMQURL()

	var conn *amqp.Connection
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.DialConfig(url, amqp.Config{
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

// SetupExchangesAndQueues creates the necessary exchanges and queues
func SetupExchangesAndQueues(ch *amqp.Channel) error {
	// Declare orders exchange
	err := ch.ExchangeDeclare(
		OrdersExchange, // name
		"topic",        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return err
	}

	// Declare events exchange (general events)
	err = ch.ExchangeDeclare(
		EventsExchange, // name
		"topic",        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return err
	}

	// Declare queues
	queues := []string{
		OrderEventsQueue,
		LessonEventsQueue,
		NotificationEventsQueue,
	}

	for _, queueName := range queues {
		_, err := ch.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		if err != nil {
			return err
		}
	}

	// Bind queues to exchanges with appropriate routing keys
	// Order events queue gets all order events
	err = ch.QueueBind(
		OrderEventsQueue,         // queue name
		"order.*",                // routing key pattern
		OrdersExchange,           // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Lesson events queue gets only paid orders (for enrollment)
	err = ch.QueueBind(
		LessonEventsQueue,        // queue name
		OrderPaidRoutingKey,      // routing key
		OrdersExchange,           // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Notification events queue gets all order events
	err = ch.QueueBind(
		NotificationEventsQueue,  // queue name
		"order.*",                // routing key pattern
		OrdersExchange,           // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Println("✅ RabbitMQ exchanges and queues setup completed")
	return nil
}