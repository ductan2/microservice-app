package utils

import (
	"errors"
	"fmt"

	"content-services/internal/taxonomy"
	"content-services/internal/types"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// MapTaxonomyError maps taxonomy store errors to GraphQL errors
func MapTaxonomyError(resource string, err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, taxonomy.ErrDuplicate):
		return gqlerror.Errorf("%s already exists", resource)
	case errors.Is(err, taxonomy.ErrNotFound):
		return gqlerror.Errorf("%s not found", resource)
	default:
		return err
	}
}

// MapLessonError maps lesson store errors to GraphQL errors
func MapLessonError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, types.ErrLessonNotFound):
		return gqlerror.Errorf("lesson not found")
	case errors.Is(err, types.ErrDuplicateCode):
		return gqlerror.Errorf("lesson code already exists")
	case errors.Is(err, types.ErrAlreadyPublished):
		return gqlerror.Errorf("lesson is already published")
	default:
		return err
	}
}

// MapLessonSectionError maps lesson section errors to GraphQL errors
func MapLessonSectionError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, types.ErrLessonSectionNotFound):
		return gqlerror.Errorf("lesson section not found")
	default:
		return MapLessonError(err)
	}
}

// MapCourseError maps course errors to GraphQL errors
func MapCourseError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, types.ErrCourseNotFound):
		return gqlerror.Errorf("course not found")
	default:
		return err
	}
}

// MapCourseLessonError maps course lesson errors to GraphQL errors
func MapCourseLessonError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, types.ErrCourseLessonNotFound):
		return gqlerror.Errorf("course lesson not found")
	case errors.Is(err, types.ErrCourseLessonExists):
		return gqlerror.Errorf("course lesson already exists")
	case errors.Is(err, types.ErrLessonNotFound):
		return gqlerror.Errorf("lesson not found")
	default:
		return MapCourseError(err)
	}
}

// MapCourseReviewError maps course review errors to GraphQL errors
func MapCourseReviewError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, types.ErrCourseReviewInvalidRating):
		return gqlerror.Errorf("rating must be between 1 and 5")
	case errors.Is(err, types.ErrCourseReviewNotEnrolled):
		return gqlerror.Errorf("enrollment required to review this course")
	case errors.Is(err, types.ErrCourseReviewNotFound):
		return gqlerror.Errorf("course review not found")
	default:
		return err
	}
}

// MapFlashcardError converts flashcard repository errors to GraphQL-friendly errors
func MapFlashcardError(err error) error {
	if err == nil {
		return nil
	}

	// Import repository errors to avoid circular dependency
	// We'll need to move this to a more appropriate place or handle it differently
	return fmt.Errorf("flashcard error: %w", err)
}