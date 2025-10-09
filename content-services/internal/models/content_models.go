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

// Folder groups media assets with up to 3 levels of nesting
type Folder struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id" bson:"_id"`
	Name      string     `gorm:"type:text;not null" json:"name" bson:"name"`
	ParentID  *uuid.UUID `gorm:"type:uuid;index:folders_parent_idx" json:"parent_id,omitempty" bson:"parent_id,omitempty"`
	Depth     int        `gorm:"not null;default:1;check:depth >= 1 AND depth <= 3" json:"depth" bson:"depth"` // 1=root, 2=child, 3=grandchild
	CreatedAt time.Time  `gorm:"default:now();not null" json:"created_at" bson:"created_at"`
	UpdatedAt time.Time  `gorm:"default:now();not null" json:"updated_at" bson:"updated_at"`
}

// MediaAsset for images and audio files
type MediaAsset struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id" bson:"_id"`
	StorageKey   string     `gorm:"type:text;uniqueIndex;not null" json:"storage_key" bson:"storage_key"` // S3/MinIO key
	Kind         string     `gorm:"type:text;not null;check:kind IN ('image','audio')" json:"kind" bson:"kind"`
	MimeType     string     `gorm:"type:text;not null" json:"mime_type" bson:"mime_type"`
	FolderID     *uuid.UUID `gorm:"type:uuid" json:"folder_id,omitempty" bson:"folder_id,omitempty"`
	OriginalName string     `gorm:"type:text;not null" json:"original_name" bson:"original_name"`
	ThumbnailURL string     `gorm:"type:text" json:"thumbnail_url,omitempty" bson:"thumbnail_url,omitempty"`
	Bytes        int        `json:"bytes,omitempty" bson:"bytes"`
	DurationMs   int        `json:"duration_ms,omitempty" bson:"duration_ms"` // for audio
	SHA256       string     `gorm:"type:text;not null;index:media_sha_idx" json:"sha256" bson:"sha256"`
	CreatedAt    time.Time  `gorm:"default:now();not null" json:"created_at" bson:"created_at"`
	UploadedBy   *uuid.UUID `gorm:"type:uuid" json:"uploaded_by,omitempty" bson:"uploaded_by,omitempty"` // logical FK to user
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

// Course bundles lessons into a structured learning path
type Course struct {
	ID            uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Title         string       `gorm:"type:text;not null" json:"title"`
	Description   string       `gorm:"type:text" json:"description,omitempty"`
	TopicID       *uuid.UUID   `gorm:"type:uuid;index:courses_topic_level_idx" json:"topic_id,omitempty"`
	LevelID       *uuid.UUID   `gorm:"type:uuid;index:courses_topic_level_idx" json:"level_id,omitempty"`
	InstructorID  *uuid.UUID   `gorm:"type:uuid;index:courses_instructor_idx" json:"instructor_id,omitempty"`
	ThumbnailURL  string       `gorm:"type:text" json:"thumbnail_url,omitempty"`
	IsPublished   bool         `gorm:"default:false;not null;index:courses_published_idx" json:"is_published"`
	IsFeatured    bool         `gorm:"default:false;not null;index:courses_featured_idx" json:"is_featured"`
	Price         float64      `gorm:"type:numeric(10,2)" json:"price,omitempty"`
	DurationHours int          `gorm:"type:int" json:"duration_hours,omitempty"`
	AverageRating float64      `gorm:"type:numeric(3,2);default:0" json:"average_rating"`
	ReviewCount   int          `gorm:"type:int;default:0" json:"review_count"`
	CreatedAt     time.Time    `gorm:"default:now();not null" json:"created_at"`
	UpdatedAt     time.Time    `gorm:"default:now();not null" json:"updated_at"`
	PublishedAt   sql.NullTime `json:"published_at,omitempty"`
}

// CourseLesson associates lessons to a course with ordering metadata
type CourseLesson struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	CourseID   uuid.UUID `gorm:"type:uuid;not null;index:course_lessons_course_idx" json:"course_id"`
	LessonID   uuid.UUID `gorm:"type:uuid;not null;index:course_lessons_lesson_idx" json:"lesson_id"`
	Ord        int       `gorm:"not null" json:"ord"`
	IsRequired bool      `gorm:"default:true;not null" json:"is_required"`
	CreatedAt  time.Time `gorm:"default:now();not null" json:"created_at"`
}

// CourseReview captures learner feedback for a course.
type CourseReview struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	CourseID  uuid.UUID `gorm:"type:uuid;not null" json:"course_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Rating    int       `gorm:"type:int;not null" json:"rating"`
	Comment   string    `gorm:"type:text" json:"comment"`
	CreatedAt time.Time `gorm:"default:now();not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now();not null" json:"updated_at"`
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
	ID          uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id" bson:"_id"`
	Title       string     `gorm:"type:text;not null" json:"title" bson:"title"`
	Description string     `gorm:"type:text" json:"description,omitempty" bson:"description,omitempty"`
	TopicID     *uuid.UUID `gorm:"type:uuid" json:"topic_id,omitempty" bson:"topic_id,omitempty"`
	LevelID     *uuid.UUID `gorm:"type:uuid" json:"level_id,omitempty" bson:"level_id,omitempty"`
	CreatedAt   time.Time  `gorm:"default:now();not null" json:"created_at" bson:"created_at"`
	CreatedBy   *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty" bson:"created_by,omitempty"`
}

// Flashcard individual flashcard
type Flashcard struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id" bson:"_id"`
	SetID        uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:flashcards_set_ord;constraint:OnDelete:CASCADE" json:"set_id" bson:"set_id"`
	FrontText    string     `gorm:"type:text;not null" json:"front_text" bson:"front_text"`
	BackText     string     `gorm:"type:text;not null" json:"back_text" bson:"back_text"`
	FrontMediaID *uuid.UUID `gorm:"type:uuid" json:"front_media_id,omitempty" bson:"front_media_id,omitempty"`
	BackMediaID  *uuid.UUID `gorm:"type:uuid" json:"back_media_id,omitempty" bson:"back_media_id,omitempty"`
	Ord          int        `gorm:"not null;uniqueIndex:flashcards_set_ord" json:"ord" bson:"ord"`
	Hints        []string   `gorm:"type:text[]" json:"hints,omitempty" bson:"hints,omitempty"`
	CreatedAt    time.Time  `gorm:"default:now();not null" json:"created_at" bson:"created_at"`
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
	TopicID     *uuid.UUID `gorm:"type:uuid;constraint:OnDelete:SET NULL;index:quizzes_topic_level_idx" json:"topic_id,omitempty"`
	LevelID     *uuid.UUID `gorm:"type:uuid;constraint:OnDelete:SET NULL;index:quizzes_topic_level_idx" json:"level_id,omitempty"`
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
