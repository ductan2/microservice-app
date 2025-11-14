package resolver

import (
	"content-services/graph/model"
	"content-services/internal/dto"
	"content-services/internal/mappers"
	"content-services/internal/models"
	"content-services/internal/utils"
	"content-services/internal/validators"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// GraphQL resolver-specific helper functions
// The mapping, validation, and utility functions have been moved to their respective packages

// ValidateAndParseUUID validates and parses a UUID string from GraphQL input
func (r *Resolver) ValidateAndParseUUID(idStr string) (uuid.UUID, error) {
	if idStr == "" {
		return uuid.Nil, gqlerror.Errorf("ID cannot be empty")
	}

	id, err := utils.ValidateUUID(idStr)
	if err != nil {
		return uuid.Nil, gqlerror.Errorf("invalid ID format")
	}

	return id, nil
}

// GetAuthenticatedUserID extracts user ID from GraphQL context with authentication check
func (r *Resolver) GetAuthenticatedUserID(ctx context.Context) (uuid.UUID, error) {
	return utils.UserIDFromContext(ctx)
}

// GetOptionalUserID extracts user ID from GraphQL context without authentication requirement
func (r *Resolver) GetOptionalUserID(ctx context.Context) (uuid.UUID, bool, error) {
	return utils.UserIDFromContextOptional(ctx)
}

// ValidateLessonInput validates lesson creation input using the validation layer
func (r *Resolver) ValidateLessonInput(input *model.CreateLessonInput) error {
	if input == nil {
		return gqlerror.Errorf("lesson input is required")
	}

	req := &dto.CreateLessonRequest{
		Code:        input.Code,
		Title:       input.Title,
		Description: input.Description,
		TopicID:     input.TopicID,
		LevelID:     input.LevelID,
		CreatedBy:   input.CreatedBy,
	}

	if err := validators.ValidateCreateLessonRequest(req); err != nil {
		return gqlerror.Errorf(err.Error())
	}

	return nil
}

// ValidateLessonUpdateInput validates lesson update input using the validation layer
func (r *Resolver) ValidateLessonUpdateInput(input *model.UpdateLessonInput) error {
	if input == nil {
		return gqlerror.Errorf("lesson update input is required")
	}

	req := &dto.UpdateLessonRequest{
		Title:       input.Title,
		Description: input.Description,
		TopicID:     input.TopicID,
		LevelID:     input.LevelID,
	}

	if err := validators.ValidateUpdateLessonRequest(req); err != nil {
		return gqlerror.Errorf(err.Error())
	}

	return nil
}

// ValidateCourseInput validates course creation input using the validation layer
func (r *Resolver) ValidateCourseInput(input *model.CreateCourseInput) error {
	if input == nil {
		return gqlerror.Errorf("course input is required")
	}

	req := &dto.CreateCourseRequest{
		Title:          input.Title,
		Description:    input.Description,
		TopicID:        input.TopicID,
		LevelID:        input.LevelID,
		InstructorID:   input.InstructorID,
		ThumbnailURL:   input.ThumbnailURL,
		Price:          input.Price,
		DurationHours:  input.DurationHours,
		IsPublished:    false, // Default value since not in GraphQL input
		IsFeatured:     input.IsFeatured != nil && *input.IsFeatured,
	}

	if err := validators.ValidateCreateCourseRequest(req); err != nil {
		return gqlerror.Errorf(err.Error())
	}

	return nil
}

// ValidateCourseReviewInput validates course review input using the validation layer
func (r *Resolver) ValidateCourseReviewInput(input *model.SubmitCourseReviewInput) error {
	if input == nil {
		return gqlerror.Errorf("course review input is required")
	}

	req := &dto.CreateCourseReviewRequest{
		CourseID: input.CourseID,
		Rating:   input.Rating,
		Comment:  input.Comment,
	}

	if err := validators.ValidateCourseReviewRequest(req); err != nil {
		return gqlerror.Errorf(err.Error())
	}

	return nil
}

// BuildFilterAndOrder constructs filter and order options using the mapper layer
func (r *Resolver) BuildFilterAndOrder(ctx context.Context, filterInput any, orderInput any, entityType string) (filter any, order any, err error) {
	switch entityType {
	case "lesson":
		lessonFilter, err := mappers.BuildLessonFilter(filterInput.(*model.LessonFilterInput))
		if err != nil {
			return nil, nil, err
		}
		lessonOrder := mappers.BuildLessonOrder(orderInput.(*model.LessonOrderInput))
		return lessonFilter, lessonOrder, nil

	case "course":
		courseFilter, err := mappers.BuildCourseFilter(filterInput.(*model.CourseFilterInput))
		if err != nil {
			return nil, nil, err
		}
		courseOrder := mappers.BuildCourseOrder(orderInput.(*model.CourseOrderInput))
		return courseFilter, courseOrder, nil

	case "lesson_section":
		sectionFilter := mappers.BuildLessonSectionFilter(filterInput.(*model.LessonSectionFilterInput))
		sectionOrder := mappers.BuildLessonSectionOrder(orderInput.(*model.LessonSectionOrderInput))
		return sectionFilter, sectionOrder, nil

	case "course_lesson":
		courseLessonFilter := mappers.BuildCourseLessonFilter(filterInput.(*model.CourseLessonFilterInput))
		courseLessonOrder := mappers.BuildCourseLessonOrder(orderInput.(*model.CourseLessonOrderInput))
		return courseLessonFilter, courseLessonOrder, nil

	default:
		return nil, nil, gqlerror.Errorf("unsupported entity type: %s", entityType)
	}
}

// MapEntitiesToGraphQL converts internal models to GraphQL models using the mapper layer
func (r *Resolver) MapEntitiesToGraphQL(ctx context.Context, entities interface{}, entityType string) (interface{}, error) {
	switch entityType {
	case "lesson":
		switch v := entities.(type) {
		case []models.Lesson:
			return mappers.LessonsToGraphQL(v), nil
		case *models.Lesson:
			return mappers.LessonToGraphQL(v), nil
		}

	case "course":
		switch v := entities.(type) {
		case []models.Course:
			courses := make([]*model.Course, len(v))
			for i, course := range v {
				courses[i] = mappers.CourseToGraphQL(&course)
			}
			return courses, nil
		case *models.Course:
			return mappers.CourseToGraphQL(v), nil
		}

	case "lesson_section":
		switch v := entities.(type) {
		case []models.LessonSection:
			return mappers.LessonSectionsToGraphQL(v), nil
		case *models.LessonSection:
			return mappers.LessonSectionToGraphQL(v), nil
		}

	case "flashcard_set":
		switch v := entities.(type) {
		case []models.FlashcardSet:
			sets := make([]*model.FlashcardSet, len(v))
			for i, set := range v {
				sets[i] = mappers.FlashcardSetToGraphQL(&set)
			}
			return sets, nil
		case *models.FlashcardSet:
			return mappers.FlashcardSetToGraphQL(v), nil
		}

	case "flashcard":
		switch v := entities.(type) {
		case []models.Flashcard:
			return mappers.FlashcardsToGraphQL(v), nil
		case *models.Flashcard:
			return mappers.FlashcardToGraphQL(v), nil
		}

	default:
		return nil, gqlerror.Errorf("unsupported entity type: %s", entityType)
	}

	return nil, gqlerror.Errorf("invalid entity data type")
}

// HandleRepositoryError converts repository errors to GraphQL-friendly errors
func (r *Resolver) HandleRepositoryError(err error, entityType string) error {
	if err == nil {
		return nil
	}

	// Use the error mapping utilities
	switch entityType {
	case "lesson":
		return utils.MapLessonError(err)
	case "lesson_section":
		return utils.MapLessonSectionError(err)
	case "course":
		return utils.MapCourseError(err)
	case "course_lesson":
		return utils.MapCourseLessonError(err)
	case "course_review":
		return utils.MapCourseReviewError(err)
	case "flashcard":
		return utils.MapFlashcardError(err)
	default:
		return err
	}
}

// HandleValidationError converts validation errors to GraphQL-friendly errors
func (r *Resolver) HandleValidationError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return gqlerror.Errorf("request timeout")
	}

	if errors.Is(err, context.Canceled) {
		return gqlerror.Errorf("request canceled")
	}

	return gqlerror.Errorf("validation error: %v", err)
}