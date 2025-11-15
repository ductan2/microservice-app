package repositories

import (
	"context"
	"database/sql"
	"time"

	"order-services/internal/models"

	"github.com/google/uuid"
)

// dbQuerier is an interface that both *sql.DB and *sql.Tx implement
type dbQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// OutboxRepository interface for outbox event data access
type OutboxRepository interface {
	Create(ctx context.Context, event *models.Outbox) error
	GetUnpublishedEvents(ctx context.Context, limit int) ([]models.Outbox, error)
	MarkAsPublished(ctx context.Context, eventID int64) error
	CountUnpublishedEvents(ctx context.Context) (int64, error)
	CountTotalEvents(ctx context.Context) (int64, error)
	DeleteOldPublishedEvents(ctx context.Context, olderThan time.Time) error
	GetEventsByAggregateID(ctx context.Context, aggregateID uuid.UUID) ([]models.Outbox, error)
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// outboxRepository implements OutboxRepository
type outboxRepository struct {
	db     dbQuerier
	origDB *sql.DB // Keep reference to original DB for BeginTx
}

// NewOutboxRepository creates a new outbox repository
func NewOutboxRepository(db *sql.DB) OutboxRepository {
	return &outboxRepository{
		db:     db,
		origDB: db,
	}
}

// Create creates a new outbox event
func (r *outboxRepository) Create(ctx context.Context, event *models.Outbox) error {
	query := `
		INSERT INTO outbox (aggregate_id, topic, type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		event.AggregateID,
		event.Topic,
		event.Type,
		event.Payload,
		time.Now(),
	).Scan(&event.ID)

	return err
}

// GetUnpublishedEvents retrieves unpublished events with limit
func (r *outboxRepository) GetUnpublishedEvents(ctx context.Context, limit int) ([]models.Outbox, error) {
	query := `
		SELECT id, aggregate_id, topic, type, payload, created_at, published_at
		FROM outbox
		WHERE published_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Outbox
	for rows.Next() {
		var event models.Outbox
		var publishedAt sql.NullTime

		err := rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.Topic,
			&event.Type,
			&event.Payload,
			&event.CreatedAt,
			&publishedAt,
		)
		if err != nil {
			return nil, err
		}

		if publishedAt.Valid {
			event.PublishedAt = publishedAt
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// MarkAsPublished marks an event as published
func (r *outboxRepository) MarkAsPublished(ctx context.Context, eventID int64) error {
	query := `UPDATE outbox SET published_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), eventID)
	return err
}

// CountUnpublishedEvents counts unpublished events
func (r *outboxRepository) CountUnpublishedEvents(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM outbox WHERE published_at IS NULL`
	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// CountTotalEvents counts all events
func (r *outboxRepository) CountTotalEvents(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM outbox`
	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// DeleteOldPublishedEvents deletes published events older than the specified time
func (r *outboxRepository) DeleteOldPublishedEvents(ctx context.Context, olderThan time.Time) error {
	query := `DELETE FROM outbox WHERE published_at IS NOT NULL AND published_at < $1`
	_, err := r.db.ExecContext(ctx, query, olderThan)
	return err
}

// GetEventsByAggregateID retrieves events for a specific aggregate
func (r *outboxRepository) GetEventsByAggregateID(ctx context.Context, aggregateID uuid.UUID) ([]models.Outbox, error) {
	query := `
		SELECT id, aggregate_id, topic, type, payload, created_at, published_at
		FROM outbox
		WHERE aggregate_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, aggregateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Outbox
	for rows.Next() {
		var event models.Outbox
		var publishedAt sql.NullTime

		err := rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.Topic,
			&event.Type,
			&event.Payload,
			&event.CreatedAt,
			&publishedAt,
		)
		if err != nil {
			return nil, err
		}

		if publishedAt.Valid {
			event.PublishedAt = publishedAt
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// WithTx executes a function within a database transaction
func (r *outboxRepository) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
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
	txRepo := &outboxRepository{db: tx, origDB: r.origDB}

	// Create a new context with the transaction repository
	txCtx := context.WithValue(ctx, "txRepo", txRepo)

	// Execute the function
	err = fn(txCtx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetRepositoryFromContext gets the transaction repository from context
func GetOutboxRepositoryFromContext(ctx context.Context) OutboxRepository {
	if txRepo, ok := ctx.Value("txRepo").(OutboxRepository); ok {
		return txRepo
	}
	// Return original repository if not in transaction
	return nil
}
