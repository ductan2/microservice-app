package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"user-services/internal/api/repositories"
	"user-services/internal/models"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/google/uuid"
)

type OutboxService interface {
	PublishUserEvent(ctx context.Context, aggregateID uuid.UUID, eventType string, payload map[string]any) error
	ProcessUnpublishedEvents(ctx context.Context, limit int) error
	CleanupPublishedEvents(ctx context.Context) error
}

type outboxService struct {
	outboxRepo repositories.OutboxRepository
	channel    *amqp.Channel
	exchange   string
}

func NewOutboxService(outboxRepo repositories.OutboxRepository, channel *amqp.Channel, exchange string) OutboxService {
	return &outboxService{
		outboxRepo: outboxRepo,
		channel:    channel,
		exchange:   exchange,
	}
}

func (s *outboxService) PublishUserEvent(ctx context.Context, aggregateID uuid.UUID, eventType string, payload map[string]any) error {
	event := &models.Outbox{
		AggregateID: aggregateID,
		Topic:       "user.events",
		Type:        eventType,
		Payload:     payload,
		CreatedAt:   time.Now(),
	}
	return s.outboxRepo.Create(ctx, event)
}

func (s *outboxService) ProcessUnpublishedEvents(ctx context.Context, limit int) error {
	// 1. Fetch unpublished events
	events, err := s.outboxRepo.GetUnpublished(ctx, limit)
	if err != nil {
		return fmt.Errorf("failed to get unpublished events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	log.Printf("Processing %d unpublished events", len(events))

	// 2. Publish each event to RabbitMQ
	for _, event := range events {
		if err := s.publishToRabbitMQ(ctx, event); err != nil {
			log.Printf("Failed to publish event %d: %v", event.ID, err)
			// Continue with other events instead of failing all
			continue
		}

		// 3. Mark as published
		if err := s.outboxRepo.MarkAsPublished(ctx, event.ID); err != nil {
			log.Printf("Failed to mark event %d as published: %v", event.ID, err)
			// Event was published but not marked - it will be retried
			// This is acceptable for at-least-once delivery
		}
	}

	return nil
}

func (s *outboxService) publishToRabbitMQ(ctx context.Context, event models.Outbox) error {
	// Create routing key from topic
	routingKey := event.Topic // e.g., "user.created", "user.events"

	// Marshal payload to JSON
	body, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	// Publish to RabbitMQ
	err = s.channel.PublishWithContext(
		ctx,
		s.exchange, // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Type:         event.Type,
			Headers: amqp.Table{
				"aggregate_id": event.AggregateID.String(),
				"event_type":   event.Type,
			},
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish to RabbitMQ: %w", err)
	}

	log.Printf("Published event %d (type=%s) to exchange=%s with routing_key=%s",
		event.ID, event.Type, s.exchange, routingKey)

	return nil
}

func (s *outboxService) CleanupPublishedEvents(ctx context.Context) error {
	// Delete published events older than 7 days
	// Get the last event ID before cutoff
	events, err := s.outboxRepo.GetUnpublished(ctx, 1)
	if err != nil {
		return err
	}

	if len(events) > 0 {
		// Delete all published events with ID less than this
		return s.outboxRepo.DeletePublished(ctx, events[0].ID)
	}

	return nil
}
