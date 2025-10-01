package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Topic for content taxonomy
type Topic struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Slug      string    `gorm:"type:text;uniqueIndex;not null" json:"slug"`
	Name      string    `gorm:"type:text;not null" json:"name"`
	CreatedAt time.Time `gorm:"default:now();not null" json:"created_at"`
}

// Level represents CEFR levels (A1, A2, B1, B2, C1, C2)
type Level struct {
	ID   uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Code string    `gorm:"type:text;uniqueIndex;not null" json:"code"` // A1, A2, B1...
	Name string    `gorm:"type:text;not null" json:"name"`
}

// Tag for flexible content tagging
type Tag struct {
	ID   uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Slug string    `gorm:"type:text;uniqueIndex;not null" json:"slug"`
	Name string    `gorm:"type:text;not null" json:"name"`
}

// MediaAsset for images and audio files
type MediaAsset struct {
	ID         uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	StorageKey string     `gorm:"type:text;uniqueIndex;not null" json:"storage_key"` // S3/MinIO key
	Kind       string     `gorm:"type:text;not null;check:kind IN ('image','audio')" json:"kind"`
	MimeType   string     `gorm:"type:text;not null" json:"mime_type"`
	Bytes      int        `json:"bytes,omitempty"`
	DurationMs int        `json:"duration_ms,omitempty"` // for audio
	SHA256     string     `gorm:"type:text;not null;index:media_sha_idx" json:"sha256"`
	CreatedAt  time.Time  `gorm:"default:now();not null" json:"created_at"`
	UploadedBy *uuid.UUID `gorm:"type:uuid" json:"uploaded_by,omitempty"` // logical FK to users
}

// Lesson modular content with versioning
type Lesson struct {
	ID          uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Code        string       `gorm:"type:text;uniqueIndex" json:"code,omitempty"` // human-readable ID
	Title       string       `gorm:"type:text;not null" json:"title"`
	Description string       `gorm:"type:text" json:"description,omitempty"`
	TopicID     *uuid.UUID   `gorm:"type:uuid;index:lessons_topic_level_pub_idx" json:"topic_id,omitempty"`
	LevelID     *uuid.UUID   `gorm:"type:uuid;index:lessons_topic_level_pub_idx" json:"level_id,omitempty"`
	IsPublished bool         `gorm:"default:false;not null;index:lessons_topic_level_pub_idx" json:"is_published"`
	Version     int          `gorm:"default:1;not null" json:"version"`
	CreatedBy   *uuid.UUID   `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt   time.Time    `gorm:"default:now();not null" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"default:now();not null" json:"updated_at"`
	PublishedAt sql.NullTime `json:"published_at,omitempty"`
}

// LessonSection content blocks within a lesson
type LessonSection struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	LessonID  uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:lesson_sections_ord;constraint:OnDelete:CASCADE" json:"lesson_id"`
	Ord       int            `gorm:"not null;uniqueIndex:lesson_sections_ord" json:"ord"`
	Type      string         `gorm:"type:text;not null;check:type IN ('text','dialog','audio','image','exercise')" json:"type"`
	Body      map[string]any `gorm:"type:jsonb;default:'{}';not null" json:"body"`
	CreatedAt time.Time      `gorm:"default:now();not null" json:"created_at"`
}

// FlashcardSet collection of flashcards
type FlashcardSet struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Title       string     `gorm:"type:text;not null" json:"title"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	TopicID     *uuid.UUID `gorm:"type:uuid" json:"topic_id,omitempty"`
	LevelID     *uuid.UUID `gorm:"type:uuid" json:"level_id,omitempty"`
	CreatedAt   time.Time  `gorm:"default:now();not null" json:"created_at"`
	CreatedBy   *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
}

// Flashcard individual flashcard
type Flashcard struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	SetID        uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:flashcards_set_ord;constraint:OnDelete:CASCADE" json:"set_id"`
	FrontText    string     `gorm:"type:text;not null" json:"front_text"`
	BackText     string     `gorm:"type:text;not null" json:"back_text"`
	FrontMediaID *uuid.UUID `gorm:"type:uuid" json:"front_media_id,omitempty"`
	BackMediaID  *uuid.UUID `gorm:"type:uuid" json:"back_media_id,omitempty"`
	Ord          int        `gorm:"not null;uniqueIndex:flashcards_set_ord" json:"ord"`
	Hints        []string   `gorm:"type:text[]" json:"hints,omitempty"`
	CreatedAt    time.Time  `gorm:"default:now();not null" json:"created_at"`
}

// Quiz assessment with questions
type Quiz struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	LessonID    *uuid.UUID `gorm:"type:uuid;constraint:OnDelete:SET NULL" json:"lesson_id,omitempty"`
	Title       string     `gorm:"type:text;not null" json:"title"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	TotalPoints int        `gorm:"default:0;not null" json:"total_points"`
	TimeLimitS  int        `json:"time_limit_s,omitempty"`
	CreatedAt   time.Time  `gorm:"default:now();not null" json:"created_at"`
}

// QuizQuestion individual question in a quiz
type QuizQuestion struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	QuizID      uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:quiz_questions_ord;constraint:OnDelete:CASCADE" json:"quiz_id"`
	Ord         int            `gorm:"not null;uniqueIndex:quiz_questions_ord" json:"ord"`
	Type        string         `gorm:"type:text;not null;check:type IN ('mcq','multi_select','fill_blank','audio_transcribe','match','ordering')" json:"type"`
	Prompt      string         `gorm:"type:text;not null" json:"prompt"`
	PromptMedia *uuid.UUID     `gorm:"type:uuid" json:"prompt_media,omitempty"`
	Points      int            `gorm:"default:1;not null" json:"points"`
	Metadata    map[string]any `gorm:"type:jsonb;default:'{}';not null" json:"metadata"`
}

// QuestionOption answer options for questions
type QuestionOption struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	QuestionID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:question_options_ord;constraint:OnDelete:CASCADE" json:"question_id"`
	Ord        int       `gorm:"not null;uniqueIndex:question_options_ord" json:"ord"`
	Label      string    `gorm:"type:text;not null" json:"label"`
	IsCorrect  bool      `gorm:"default:false;not null" json:"is_correct"`
	Feedback   string    `gorm:"type:text" json:"feedback,omitempty"`
}

// ContentTag junction table for tagging
type ContentTag struct {
	TagID    uuid.UUID `gorm:"type:uuid;primaryKey;constraint:OnDelete:CASCADE" json:"tag_id"`
	Kind     string    `gorm:"type:text;primaryKey;not null;check:kind IN ('lesson','quiz','flashcard_set')" json:"kind"`
	ObjectID uuid.UUID `gorm:"type:uuid;primaryKey;not null" json:"object_id"`
}

// Outbox for event-driven architecture
type Outbox struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	AggregateID uuid.UUID      `gorm:"type:uuid;not null" json:"aggregate_id"`
	Topic       string         `gorm:"type:text;not null" json:"topic"` // content.events
	Type        string         `gorm:"type:text;not null" json:"type"`  // LessonPublished, QuizCreated
	Payload     map[string]any `gorm:"type:jsonb;not null" json:"payload"`
	CreatedAt   time.Time      `gorm:"default:now();not null" json:"created_at"`
	PublishedAt sql.NullTime   `gorm:"index:outbox_unpub_idx,where:published_at IS NULL" json:"published_at,omitempty"`
}
