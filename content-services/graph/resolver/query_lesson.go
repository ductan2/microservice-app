package resolver

import (
	"content-services/graph/model"
	"context"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Lesson is the resolver for the lesson field.
func (r *queryResolver) Lesson(ctx context.Context, id string) (*model.Lesson, error) {
	lessonService := r.Resolver.LessonService
	if lessonService == nil {
		return nil, gqlerror.Errorf("lesson service not configured")
	}

	lessonID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid lesson ID: %v", err)
	}

	lessonDoc, err := lessonService.GetLessonByID(ctx, lessonID)
	if err != nil {
		return nil, mapLessonError(err)
	}

	return mapLesson(lessonDoc), nil
}

// LessonByCode is the resolver for the lessonByCode field.
func (r *queryResolver) LessonByCode(ctx context.Context, code string) (*model.Lesson, error) {
	lessonService := r.Resolver.LessonService
	if lessonService == nil {
		return nil, gqlerror.Errorf("lesson service not configured")
	}

	lessonDoc, err := lessonService.GetLessonByCode(ctx, code)
	if err != nil {
		return nil, mapLessonError(err)
	}

	return mapLesson(lessonDoc), nil
}
