package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"order-services/internal/models"
)

// WebhookEventRepository interface for webhook event data access
type WebhookEventRepository interface {
	Create(ctx context.Context, event *models.WebhookEvent) error
	IsProcessed(ctx context.Context, stripeEventID string) error
	MarkAsProcessed(ctx context.Context, stripeEventID string) error
	GetUnprocessedEvents(ctx context.Context, olderThan time.Time) ([]models.WebhookEvent, error)
	DeleteOldEvents(ctx context.Context, olderThan time.Time) error
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// webhookEventRepository implements WebhookEventRepository
type webhookEventRepository struct {
	db     dbQuerier
	origDB *sql.DB // Keep reference to original DB for BeginTx
}

// NewWebhookEventRepository creates a new webhook event repository
func NewWebhookEventRepository(db *sql.DB) WebhookEventRepository {
	return &webhookEventRepository{
		db:     db,
		origDB: db,
	}
}

// Create creates a new webhook event record
func (r *webhookEventRepository) Create(ctx context.Context, event *models.WebhookEvent) error {
	query := `
		INSERT INTO webhook_events (stripe_event_id, type, payload, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		event.StripeEventID,
		event.Type,
		event.Payload,
		time.Now(),
	).Scan(&event.ID)

	return err
}

// IsProcessed checks if a webhook event has already been processed
func (r *webhookEventRepository) IsProcessed(ctx context.Context, stripeEventID string) error {
	query := `
		SELECT processed FROM webhook_events
		WHERE stripe_event_id = $1`

	var processed bool
	err := r.db.QueryRowContext(ctx, query, stripeEventID).Scan(&processed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // Event doesn't exist, not processed
		}
		return err
	}

	if processed {
		return errors.New("webhook event already processed")
	}

	return nil
}

// MarkAsProcessed marks a webhook event as processed
func (r *webhookEventRepository) MarkAsProcessed(ctx context.Context, stripeEventID string) error {
	query := `
		UPDATE webhook_events
		SET processed = true, processed_at = $1
		WHERE stripe_event_id = $2`

	_, err := r.db.ExecContext(ctx, query, time.Now(), stripeEventID)
	return err
}

// GetUnprocessedEvents retrieves webhook events that haven't been processed
func (r *webhookEventRepository) GetUnprocessedEvents(ctx context.Context, olderThan time.Time) ([]models.WebhookEvent, error) {
	query := `
		SELECT id, stripe_event_id, type, payload, processed, processed_at, created_at, updated_at
		FROM webhook_events
		WHERE processed = false AND created_at < $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, olderThan)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.WebhookEvent
	for rows.Next() {
		var event models.WebhookEvent
		var processedAt sql.NullTime

		err := rows.Scan(
			&event.ID,
			&event.StripeEventID,
			&event.Type,
			&event.Payload,
			&event.Processed,
			&processedAt,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if processedAt.Valid {
			event.ProcessedAt = processedAt
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// DeleteOldEvents deletes old webhook events
func (r *webhookEventRepository) DeleteOldEvents(ctx context.Context, olderThan time.Time) error {
	query := `DELETE FROM webhook_events WHERE created_at < $1`
	_, err := r.db.ExecContext(ctx, query, olderThan)
	return err
}

// WithTx executes a function within a database transaction
func (r *webhookEventRepository) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := r.origDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		}
	}()

	// Create a new repository with the transaction
	txRepo := &webhookEventRepository{db: tx, origDB: r.origDB}

	// Create a new context with the transaction repository
	txCtx := context.WithValue(ctx, "webhookTxRepo", txRepo)

	// Execute the function
	err = fn(txCtx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetWebhookRepositoryFromContext gets the transaction repository from context
func GetWebhookRepositoryFromContext(ctx context.Context) WebhookEventRepository {
	if txRepo, ok := ctx.Value("webhookTxRepo").(WebhookEventRepository); ok {
		return txRepo
	}
	return nil
}
