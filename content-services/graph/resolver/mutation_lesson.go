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
