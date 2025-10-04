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

// Lessons is the resolver for the lessons field.
func (r *queryResolver) Lessons(ctx context.Context, filter *model.LessonFilterInput, page *int, pageSize *int, orderBy *model.LessonOrderInput) (*model.LessonCollection, error) {
	lessonService := r.Resolver.LessonService
	if lessonService == nil {
		return nil, gqlerror.Errorf("lesson service not configured")
	}

	lessonFilter, err := buildLessonFilter(filter)
	if err != nil {
		return nil, err
	}

	lessonSort := buildLessonOrder(orderBy)

	pageVal := 1
	if page != nil && *page > 0 {
		pageVal = *page
	}

	pageSizeVal := 20
	if pageSize != nil && *pageSize > 0 {
		pageSizeVal = *pageSize
	}

	lessons, total, err := lessonService.ListLessons(ctx, lessonFilter, lessonSort, pageVal, pageSizeVal)
	if err != nil {
		return nil, mapLessonError(err)
	}

	items := make([]*model.Lesson, 0, len(lessons))
	for i := range lessons {
		items = append(items, mapLesson(&lessons[i]))
	}

	return &model.LessonCollection{
		Items:      items,
		TotalCount: int(total),
		Page:       pageVal,
		PageSize:   pageSizeVal,
	}, nil
}

// LessonSections is the resolver for the lessonSections field.
func (r *queryResolver) LessonSections(ctx context.Context, lessonID string, filter *model.LessonSectionFilterInput, page *int, pageSize *int, orderBy *model.LessonSectionOrderInput) (*model.LessonSectionCollection, error) {
	lessonService := r.Resolver.LessonService
	if lessonService == nil {
		return nil, gqlerror.Errorf("lesson service not configured")
	}

	id, err := uuid.Parse(lessonID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid lesson ID: %v", err)
	}

	sectionFilter := buildLessonSectionFilter(filter)
	sectionSort := buildLessonSectionOrder(orderBy)

	pageVal := 1
	if page != nil && *page > 0 {
		pageVal = *page
	}
	pageSizeVal := 20
	if pageSize != nil && *pageSize > 0 {
		pageSizeVal = *pageSize
	}

	sections, total, err := lessonService.ListLessonSections(ctx, id, sectionFilter, sectionSort, pageVal, pageSizeVal)
	if err != nil {
		return nil, mapLessonSectionError(err)
	}

	return &model.LessonSectionCollection{
		Items:      mapLessonSections(sections),
		TotalCount: int(total),
		Page:       pageVal,
		PageSize:   pageSizeVal,
	}, nil
}
