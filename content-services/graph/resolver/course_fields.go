package resolver

import (
	"content-services/graph/generated"
	"content-services/graph/model"
	"content-services/internal/repository"
	"content-services/internal/taxonomy"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) Course() generated.CourseResolver { return &courseResolver{r} }

func (r *Resolver) CourseLesson() generated.CourseLessonResolver { return &courseLessonResolver{r} }

type courseResolver struct{ *Resolver }
type courseLessonResolver struct{ *Resolver }

// Topic is the resolver for the topic field.
func (r *courseResolver) Topic(ctx context.Context, obj *model.Course) (*model.Topic, error) {
	if obj.TopicID == nil || *obj.TopicID == "" {
		return nil, nil
	}

	if r.Taxonomy == nil {
		return nil, nil
	}

	topic, err := r.Taxonomy.GetTopicByID(ctx, *obj.TopicID)
	if err != nil {
		if errors.Is(err, taxonomy.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mapTopic(topic), nil
}

// Level is the resolver for the level field.
func (r *courseResolver) Level(ctx context.Context, obj *model.Course) (*model.Level, error) {
	if obj.LevelID == nil || *obj.LevelID == "" {
		return nil, nil
	}

	if r.Taxonomy == nil {
		return nil, nil
	}

	level, err := r.Taxonomy.GetLevelByID(ctx, *obj.LevelID)
	if err != nil {
		if errors.Is(err, taxonomy.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mapLevel(level), nil
}

// Lessons is the resolver for the lessons field.
func (r *courseResolver) Lessons(ctx context.Context, obj *model.Course) ([]*model.CourseLesson, error) {
	if r.CourseService == nil {
		return nil, gqlerror.Errorf("course service not configured")
	}

	courseID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid course ID: %v", err)
	}

	lessons, _, err := r.CourseService.ListCourseLessons(ctx, courseID, nil, &repository.SortOption{Field: "ord", Direction: repository.SortAscending}, 1, 1000)
	if err != nil {
		return nil, mapCourseLessonError(err)
	}

	return mapCourseLessons(lessons), nil
}

// Reviews is the resolver for the reviews field.
func (r *courseResolver) Reviews(ctx context.Context, obj *model.Course, page *int, pageSize *int) (*model.CourseReviewCollection, error) {
	if r.CourseReviewService == nil {
		return &model.CourseReviewCollection{Items: []*model.CourseReview{}, TotalCount: 0, Page: 1, PageSize: 20}, nil
	}

	courseID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid course ID: %v", err)
	}

	pageVal := 1
	if page != nil && *page > 0 {
		pageVal = *page
	}

	pageSizeVal := 20
	if pageSize != nil && *pageSize > 0 {
		pageSizeVal = *pageSize
	}

	reviews, total, err := r.CourseReviewService.ListReviews(ctx, courseID, pageVal, pageSizeVal)
	if err != nil {
		return nil, mapCourseReviewError(err)
	}

	return &model.CourseReviewCollection{
		Items:      mapCourseReviews(reviews),
		TotalCount: int(total),
		Page:       pageVal,
		PageSize:   pageSizeVal,
	}, nil
}

// MyReview is the resolver for the myReview field.
func (r *courseResolver) MyReview(ctx context.Context, obj *model.Course) (*model.CourseReview, error) {
	if r.CourseReviewService == nil {
		return nil, nil
	}

	userID, ok, err := userIDFromContextOptional(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	courseID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid course ID: %v", err)
	}

	review, err := r.CourseReviewService.GetReviewByUser(ctx, courseID, userID)
	if err != nil {
		return nil, mapCourseReviewError(err)
	}

	return mapCourseReview(review), nil
}

// Lesson is the resolver for the lesson field.
func (r *courseLessonResolver) Lesson(ctx context.Context, obj *model.CourseLesson) (*model.Lesson, error) {
	if r.LessonService == nil {
		return nil, nil
	}

	lessonID, err := uuid.Parse(obj.LessonID)
	if err != nil {
		return nil, nil
	}

	lesson, err := r.LessonService.GetLessonByID(ctx, lessonID)
	if err != nil {
		return nil, mapLessonError(err)
	}

	return mapLesson(lesson), nil
}
