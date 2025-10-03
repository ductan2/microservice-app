package resolver

import (
	"content-services/graph/model"
	"content-services/internal/models"
	"content-services/internal/taxonomy"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// CreateLesson is the resolver for the createLesson field.
func (r *mutationResolver) CreateLesson(ctx context.Context, input model.CreateLessonInput) (*model.Lesson, error) {
	if r.LessonService == nil {
		return nil, gqlerror.Errorf("lesson service not configured")
	}

	if input.Title == "" {
		return nil, gqlerror.Errorf("title is required")
	}

	lesson := &models.Lesson{
		Title:       input.Title,
		Description: derefString(input.Description),
	}

	if input.Code != nil && *input.Code != "" {
		lesson.Code = *input.Code
	}

	if input.TopicID != nil {
		if r.Taxonomy != nil {
			_, err := r.Taxonomy.GetTopicByID(ctx, *input.TopicID)
			if err != nil {
				if errors.Is(err, taxonomy.ErrNotFound) {
					return nil, gqlerror.Errorf("topic not found: %s", *input.TopicID)
				}
				return nil, err
			}
		}
		topicID, err := uuid.Parse(*input.TopicID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid topic ID: %v", err)
		}
		lesson.TopicID = &topicID
	}

	if input.LevelID != nil {
		if r.Taxonomy != nil {
			_, err := r.Taxonomy.GetLevelByID(ctx, *input.LevelID)
			if err != nil {
				if errors.Is(err, taxonomy.ErrNotFound) {
					return nil, gqlerror.Errorf("level not found: %s", *input.LevelID)
				}
				return nil, err
			}
		}
		levelID, err := uuid.Parse(*input.LevelID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid level ID: %v", err)
		}
		lesson.LevelID = &levelID
	}

	if input.CreatedBy != nil {
		createdBy, err := uuid.Parse(*input.CreatedBy)
		if err != nil {
			return nil, gqlerror.Errorf("invalid created by ID: %v", err)
		}
		lesson.CreatedBy = &createdBy
	}

	createdLesson, err := r.LessonService.CreateLesson(ctx, lesson, nil)
	if err != nil {
		return nil, mapLessonError(err)
	}

	return mapLesson(createdLesson), nil
}

// UpdateLesson is the resolver for the updateLesson field.
func (r *mutationResolver) UpdateLesson(ctx context.Context, id string, input model.UpdateLessonInput) (*model.Lesson, error) {
	if r.LessonService == nil {
		return nil, gqlerror.Errorf("lesson service not configured")
	}

	lessonID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid lesson ID: %v", err)
	}

	updates := &models.Lesson{}

	if input.Title != nil {
		updates.Title = *input.Title
	}

	if input.Description != nil {
		updates.Description = *input.Description
	}

	if input.TopicID != nil {
		if *input.TopicID == "" {
			nilID := uuid.Nil
			updates.TopicID = &nilID
		} else {
			if r.Taxonomy != nil {
				_, err := r.Taxonomy.GetTopicByID(ctx, *input.TopicID)
				if err != nil {
					if errors.Is(err, taxonomy.ErrNotFound) {
						return nil, gqlerror.Errorf("topic not found: %s", *input.TopicID)
					}
					return nil, err
				}
			}
			topicID, err := uuid.Parse(*input.TopicID)
			if err != nil {
				return nil, gqlerror.Errorf("invalid topic ID: %v", err)
			}
			updates.TopicID = &topicID
		}
	}

	if input.LevelID != nil {
		if *input.LevelID == "" {
			nilID := uuid.Nil
			updates.LevelID = &nilID
		} else {
			if r.Taxonomy != nil {
				_, err := r.Taxonomy.GetLevelByID(ctx, *input.LevelID)
				if err != nil {
					if errors.Is(err, taxonomy.ErrNotFound) {
						return nil, gqlerror.Errorf("level not found: %s", *input.LevelID)
					}
					return nil, err
				}
			}
			levelID, err := uuid.Parse(*input.LevelID)
			if err != nil {
				return nil, gqlerror.Errorf("invalid level ID: %v", err)
			}
			updates.LevelID = &levelID
		}
	}

	updated, err := r.LessonService.UpdateLesson(ctx, lessonID, updates)
	if err != nil {
		return nil, mapLessonError(err)
	}

	return mapLesson(updated), nil
}

// PublishLesson is the resolver for the publishLesson field.
func (r *mutationResolver) PublishLesson(ctx context.Context, id string) (*model.Lesson, error) {
	if r.LessonService == nil {
		return nil, gqlerror.Errorf("lesson service not configured")
	}

	lessonID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid lesson ID: %v", err)
	}

	lesson, err := r.LessonService.PublishLesson(ctx, lessonID)
	if err != nil {
		return nil, mapLessonError(err)
	}

	return mapLesson(lesson), nil
}

// UnpublishLesson is the resolver for the unpublishLesson field.
func (r *mutationResolver) UnpublishLesson(ctx context.Context, id string) (*model.Lesson, error) {
	if r.LessonService == nil {
		return nil, gqlerror.Errorf("lesson service not configured")
	}

	lessonID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid lesson ID: %v", err)
	}

	lesson, err := r.LessonService.UnpublishLesson(ctx, lessonID)
	if err != nil {
		return nil, mapLessonError(err)
	}

	return mapLesson(lesson), nil
}

// CreateLessonSection is the resolver for the createLessonSection field.
func (r *mutationResolver) CreateLessonSection(ctx context.Context, lessonID string, input model.CreateLessonSectionInput) (*model.LessonSection, error) {
	if r.LessonService == nil {
		return nil, gqlerror.Errorf("lesson service not configured")
	}

	parsedLessonID, err := uuid.Parse(lessonID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid lesson ID: %v", err)
	}

	section := &models.LessonSection{
		Type: normalizeLessonSectionType(input.Type),
		Body: input.Body,
	}

	created, err := r.LessonService.AddSection(ctx, parsedLessonID, section)
	if err != nil {
		return nil, mapLessonSectionError(err)
	}

	return mapLessonSection(created), nil
}

// UpdateLessonSection is the resolver for the updateLessonSection field.
func (r *mutationResolver) UpdateLessonSection(ctx context.Context, id string, input model.UpdateLessonSectionInput) (*model.LessonSection, error) {
	if r.LessonService == nil {
		return nil, gqlerror.Errorf("lesson service not configured")
	}

	sectionID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid lesson section ID: %v", err)
	}

	updates := &models.LessonSection{}

	if input.Type != nil {
		updates.Type = normalizeLessonSectionType(*input.Type)
	}

	if input.Body != nil {
		updates.Body = input.Body
	}

	updated, err := r.LessonService.UpdateSection(ctx, sectionID, updates)
	if err != nil {
		return nil, mapLessonSectionError(err)
	}

	return mapLessonSection(updated), nil
}

// DeleteLessonSection is the resolver for the deleteLessonSection field.
func (r *mutationResolver) DeleteLessonSection(ctx context.Context, id string) (bool, error) {
	if r.LessonService == nil {
		return false, gqlerror.Errorf("lesson service not configured")
	}

	sectionID, err := uuid.Parse(id)
	if err != nil {
		return false, gqlerror.Errorf("invalid lesson section ID: %v", err)
	}

	if err := r.LessonService.DeleteSection(ctx, sectionID); err != nil {
		return false, mapLessonSectionError(err)
	}

	return true, nil
}
