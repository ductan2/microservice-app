package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Outbox for cross-service events (transactional outbox pattern)
type Outbox struct {
	ID          int64        `gorm:"primaryKey;autoIncrement" json:"id"`
	AggregateID uuid.UUID    `gorm:"type:uuid;not null" json:"aggregate_id"`
	Topic       string       `gorm:"type:text;not null" json:"topic"`
	Type        string       `gorm:"type:text;not null" json:"type"`
	Payload     []byte       `gorm:"type:jsonb" json:"payload"`
	CreatedAt   time.Time    `gorm:"default:now();not null" json:"created_at"`
	PublishedAt sql.NullTime `gorm:"index:outbox_unpublished_idx,where:published_at IS NULL" json:"published_at,omitempty"`
}