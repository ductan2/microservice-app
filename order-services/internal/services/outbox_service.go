package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"order-services/internal/config"
	"order-services/internal/models"
	"order-services/internal/repositories"
)

var (
	ErrEventNotFound      = errors.New("event not found")
	ErrEventPublishFailed = errors.New("failed to publish event")
	ErrQueueConnection    = errors.New("failed to connect to message queue")
	ErrQueueChannel       = errors.New("failed to create channel")
)

// OutboxService defines the business logic interface for outbox pattern event publishing
type OutboxService interface {
	PublishEvents(ctx context.Context) error
	CreateEvent(ctx context.Context, aggregateID uuid.UUID, topic, eventType string, payload interface{}) error
	PublishEvent(ctx context.Context, event *models.Outbox) error
	GetUnpublishedEvents(ctx context.Context, limit int) ([]models.Outbox, error)
	MarkAsPublished(ctx context.Context, eventID int64) error
	StartEventPublisher(ctx context.Context) error
	StopEventPublisher()
}

// outboxService implements the outbox pattern business logic
type outboxService struct {
	outboxRepo repositories.OutboxRepository
	config     *config.Config
	conn       *amqp.Connection
	channel    *amqp.Channel
	stopChan   chan bool
	running    bool
}

// NewOutboxService creates a new outbox service instance
func NewOutboxService(
	outboxRepo repositories.OutboxRepository,
	config *config.Config,
) OutboxService {
	return &outboxService{
		outboxRepo: outboxRepo,
		config:     config,
		stopChan:   make(chan bool),
		running:    false,
	}
}

// PublishEvents processes and publishes unpublished events from the outbox
func (s *outboxService) PublishEvents(ctx context.Context) error {
	// Get unpublished events
	events, err := s.GetUnpublishedEvents(ctx, 100) // Process in batches of 100
	if err != nil {
		return fmt.Errorf("failed to get unpublished events: %w", err)
	}

	if len(events) == 0 {
		return nil // No events to publish
	}

	// Ensure RabbitMQ connection is established
	if err := s.ensureConnection(); err != nil {
		return fmt.Errorf("failed to ensure RabbitMQ connection: %w", err)
	}

	// Publish each event
	for _, event := range events {
		if err := s.PublishEvent(ctx, &event); err != nil {
			log.Printf("Failed to publish event %d: %v", event.ID, err)
			continue // Continue with other events
		}

		// Mark as published
		if err := s.MarkAsPublished(ctx, event.ID); err != nil {
			log.Printf("Failed to mark event %d as published: %v", event.ID, err)
		}
	}

	return nil
}

// CreateEvent creates a new outbox event
func (s *outboxService) CreateEvent(ctx context.Context, aggregateID uuid.UUID, topic, eventType string, payload interface{}) error {
	// Serialize payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize event payload: %w", err)
	}

	event := &models.Outbox{
		AggregateID: aggregateID,
		Topic:       topic,
		Type:        eventType,
		Payload:     payloadBytes,
	}

	return s.outboxRepo.Create(ctx, event)
}

// PublishEvent publishes a single event to RabbitMQ
func (s *outboxService) PublishEvent(ctx context.Context, event *models.Outbox) error {
	if s.channel == nil {
		return ErrQueueConnection
	}

	// Declare exchange if it doesn't exist
	err := s.channel.ExchangeDeclare(
		event.Topic, // name
		"topic",     // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Publish message
	err = s.channel.PublishWithContext(
		ctx,
		event.Topic, // exchange
		event.Type,  // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        event.Payload,
			MessageId:   fmt.Sprintf("%d", event.ID),
			Timestamp:   time.Now(),
			Headers: amqp.Table{
				"event_type":   event.Type,
				"aggregate_id": event.AggregateID.String(),
				"topic":        event.Topic,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEventPublishFailed, err)
	}

	return nil
}

// GetUnpublishedEvents retrieves events that haven't been published yet
func (s *outboxService) GetUnpublishedEvents(ctx context.Context, limit int) ([]models.Outbox, error) {
	return s.outboxRepo.GetUnpublishedEvents(ctx, limit)
}

// MarkAsPublished marks an event as published
func (s *outboxService) MarkAsPublished(ctx context.Context, eventID int64) error {
	return s.outboxRepo.MarkAsPublished(ctx, eventID)
}

// StartEventPublisher starts the background event publisher
func (s *outboxService) StartEventPublisher(ctx context.Context) error {
	if s.running {
		return fmt.Errorf("event publisher is already running")
	}

	// Initialize RabbitMQ connection
	if err := s.ensureConnection(); err != nil {
		return fmt.Errorf("failed to initialize RabbitMQ connection: %w", err)
	}

	s.running = true

	// Start background publisher goroutine
	go s.eventPublisherLoop(ctx)

	return nil
}

// StopEventPublisher stops the background event publisher
func (s *outboxService) StopEventPublisher() {
	if !s.running {
		return
	}

	s.running = false
	close(s.stopChan)

	// Close RabbitMQ connection
	if s.channel != nil {
		s.channel.Close()
		s.channel = nil
	}
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}

// eventPublisherLoop runs the background event publishing loop
func (s *outboxService) eventPublisherLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second) // Process events every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.PublishEvents(ctx); err != nil {
				log.Printf("Failed to publish events: %v", err)
			}
		}
	}
}

// ensureConnection ensures RabbitMQ connection is established
func (s *outboxService) ensureConnection() error {
	// Check if connection is already established
	if s.conn != nil && !s.conn.IsClosed() {
		return nil
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial(s.config.RabbitMQURL())
	if err != nil {
		return fmt.Errorf("%w: %v", ErrQueueConnection, err)
	}
	s.conn = conn

	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		s.conn = nil
		return fmt.Errorf("%w: %v", ErrQueueChannel, err)
	}
	s.channel = channel

	// Set QoS to control how many messages are processed at once
	err = s.channel.Qos(
		100,   // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Printf("Warning: Failed to set QoS: %v", err)
	}

	// Declare default exchanges
	if err := s.declareDefaultExchanges(); err != nil {
		log.Printf("Warning: Failed to declare default exchanges: %v", err)
	}

	return nil
}

// declareDefaultExchanges declares the default exchanges used by the order service
func (s *outboxService) declareDefaultExchanges() error {
	exchanges := []string{
		"order.events",
		"payment.events",
		"coupon.events",
		"user.notifications", // For user notifications
	}

	for _, exchange := range exchanges {
		err := s.channel.ExchangeDeclare(
			exchange, // name
			"topic",  // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
		}
	}

	return nil
}

// CreateOrderEvent creates an order-related event
func (s *outboxService) CreateOrderEvent(ctx context.Context, eventType string, order *models.Order, additionalData map[string]interface{}) error {
	payload := map[string]interface{}{
		"event_id":   uuid.New().String(),
		"event_type": eventType,
		"order_id":   order.ID,
		"user_id":    order.UserID,
		"status":     order.Status,
		"amount":     order.TotalAmount,
		"currency":   order.Currency,
		"created_at": order.CreatedAt,
		"updated_at": order.UpdatedAt,
	}

	// Add additional data
	for k, v := range additionalData {
		payload[k] = v
	}

	return s.CreateEvent(ctx, order.ID, "order.events", eventType, payload)
}

// CreatePaymentEvent creates a payment-related event
func (s *outboxService) CreatePaymentEvent(ctx context.Context, eventType string, payment *models.Payment, order *models.Order) error {
	payload := map[string]interface{}{
		"event_id":          uuid.New().String(),
		"event_type":        eventType,
		"payment_id":        payment.ID,
		"order_id":          order.ID,
		"user_id":           order.UserID,
		"amount":            payment.Amount,
		"currency":          payment.Currency,
		"status":            payment.Status,
		"payment_intent_id": payment.StripePaymentIntentID,
		"created_at":        payment.CreatedAt,
	}

	return s.CreateEvent(ctx, order.ID, "payment.events", eventType, payload)
}

// CreateCouponEvent creates a coupon-related event
func (s *outboxService) CreateCouponEvent(ctx context.Context, eventType string, coupon *models.Coupon, userID uuid.UUID, orderID uuid.UUID) error {
	payload := map[string]interface{}{
		"event_id":    uuid.New().String(),
		"event_type":  eventType,
		"coupon_id":   coupon.ID,
		"user_id":     userID,
		"order_id":    orderID,
		"coupon_code": coupon.Code,
		"coupon_type": coupon.Type,
		"created_at":  time.Now(),
	}

	return s.CreateEvent(ctx, orderID, "coupon.events", eventType, payload)
}

// CreateUserNotificationEvent creates a user notification event
func (s *outboxService) CreateUserNotificationEvent(ctx context.Context, userID uuid.UUID, notificationType, title, message string, data map[string]interface{}) error {
	payload := map[string]interface{}{
		"event_id":   uuid.New().String(),
		"event_type": "user.notification",
		"user_id":    userID,
		"type":       notificationType,
		"title":      title,
		"message":    message,
		"data":       data,
		"created_at": time.Now(),
	}

	return s.CreateEvent(ctx, userID, "user.notifications", "user.notification", payload)
}

// GetEventStats returns statistics about outbox events
func (s *outboxService) GetEventStats(ctx context.Context) (*OutboxStats, error) {
	unpublishedCount, err := s.outboxRepo.CountUnpublishedEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count unpublished events: %w", err)
	}

	totalCount, err := s.outboxRepo.CountTotalEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count total events: %w", err)
	}

	stats := &OutboxStats{
		TotalEvents:       totalCount,
		UnpublishedEvents: unpublishedCount,
		PublishedEvents:   totalCount - unpublishedCount,
	}

	return stats, nil
}

// OutboxStats represents outbox event statistics
type OutboxStats struct {
	TotalEvents       int64 `json:"total_events"`
	UnpublishedEvents int64 `json:"unpublished_events"`
	PublishedEvents   int64 `json:"published_events"`
}

// CleanupOldEvents removes old published events from the outbox table
func (s *outboxService) CleanupOldEvents(ctx context.Context, olderThan time.Duration) error {
	cutoffTime := time.Now().Add(-olderThan)
	return s.outboxRepo.DeleteOldPublishedEvents(ctx, cutoffTime)
}
