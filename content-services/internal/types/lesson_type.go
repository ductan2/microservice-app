package types

import "errors"

var (
	// ErrLessonNotFound is returned when a lesson cannot be located.
	ErrLessonNotFound = errors.New("lesson: not found")
	// ErrDuplicateCode is returned when attempting to create a lesson with a duplicate code.
	ErrDuplicateCode = errors.New("lesson: duplicate code")
	// ErrAlreadyPublished is returned when attempting to publish an already published lesson.
	ErrAlreadyPublished = errors.New("lesson: already published")
	// ErrLessonSectionNotFound is returned when a lesson section cannot be located.
	ErrLessonSectionNotFound = errors.New("lesson section: not found")
)

// CreateLessonInput contains fields for creating a new lesson.
type CreateLessonInput struct {
	Code        *string
	Title       string
	Description string
	TopicID     *string
	LevelID     *string
	CreatedBy   *string
}

// UpdateLessonInput contains fields for updating an existing lesson.
type UpdateLessonInput struct {
	Title       *string
	Description *string
	TopicID     *string
	LevelID     *string
}
