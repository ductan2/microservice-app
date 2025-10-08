package resolver

import (
	"content-services/graph/model"
	"content-services/internal/models"
	"content-services/internal/service"
	"content-services/internal/taxonomy"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// CreateCourse is the resolver for the createCourse field.
func (r *mutationResolver) CreateCourse(ctx context.Context, input model.CreateCourseInput) (*model.Course, error) {
	if r.CourseService == nil {
		return nil, gqlerror.Errorf("course service not configured")
	}

	if input.Title == "" {
		return nil, gqlerror.Errorf("title is required")
	}

	course := &models.Course{
		Title:       input.Title,
		Description: derefString(input.Description),
	}

	if input.TopicID != nil {
		if *input.TopicID != "" {
			if r.Taxonomy != nil {
				if _, err := r.Taxonomy.GetTopicByID(ctx, *input.TopicID); err != nil {
					if errors.Is(err, taxonomy.ErrNotFound) {
						return nil, gqlerror.Errorf("topic not found: %s", *input.TopicID)
					}
					return nil, err
				}
			}
			id, err := uuid.Parse(*input.TopicID)
			if err != nil {
				return nil, gqlerror.Errorf("invalid topic ID: %v", err)
			}
			course.TopicID = &id
		}
	}

	if input.LevelID != nil {
		if *input.LevelID != "" {
			if r.Taxonomy != nil {
				if _, err := r.Taxonomy.GetLevelByID(ctx, *input.LevelID); err != nil {
					if errors.Is(err, taxonomy.ErrNotFound) {
						return nil, gqlerror.Errorf("level not found: %s", *input.LevelID)
					}
					return nil, err
				}
			}
			id, err := uuid.Parse(*input.LevelID)
			if err != nil {
				return nil, gqlerror.Errorf("invalid level ID: %v", err)
			}
			course.LevelID = &id
		}
	}

	if input.InstructorID != nil && *input.InstructorID != "" {
		id, err := uuid.Parse(*input.InstructorID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid instructor ID: %v", err)
		}
		course.InstructorID = &id
	}

	if input.ThumbnailURL != nil {
		course.ThumbnailURL = *input.ThumbnailURL
	}

	if input.IsFeatured != nil {
		course.IsFeatured = *input.IsFeatured
	}

	if input.Price != nil {
		course.Price = *input.Price
	}

	if input.DurationHours != nil {
		course.DurationHours = *input.DurationHours
	}

	created, err := r.CourseService.CreateCourse(ctx, course)
	if err != nil {
		return nil, mapCourseError(err)
	}

	return mapCourse(created), nil
}

// UpdateCourse is the resolver for the updateCourse field.
func (r *mutationResolver) UpdateCourse(ctx context.Context, id string, input model.UpdateCourseInput) (*model.Course, error) {
	if r.CourseService == nil {
		return nil, gqlerror.Errorf("course service not configured")
	}

	courseID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid course ID: %v", err)
	}

	updates := &service.CourseUpdate{}

	if input.Title != nil {
		updates.Title = input.Title
	}

	if input.Description != nil {
		updates.Description = input.Description
	}

	if input.TopicID != nil {
		if *input.TopicID == "" {
			nilID := uuid.Nil
			updates.TopicID = &nilID
		} else {
			if r.Taxonomy != nil {
				if _, err := r.Taxonomy.GetTopicByID(ctx, *input.TopicID); err != nil {
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
				if _, err := r.Taxonomy.GetLevelByID(ctx, *input.LevelID); err != nil {
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

	if input.InstructorID != nil {
		if *input.InstructorID == "" {
			nilID := uuid.Nil
			updates.InstructorID = &nilID
		} else {
			instructorID, err := uuid.Parse(*input.InstructorID)
			if err != nil {
				return nil, gqlerror.Errorf("invalid instructor ID: %v", err)
			}
			updates.InstructorID = &instructorID
		}
	}

	if input.ThumbnailURL != nil {
		updates.ThumbnailURL = input.ThumbnailURL
	}

	if input.IsFeatured != nil {
		updates.IsFeatured = input.IsFeatured
	}

	if input.Price != nil {
		updates.Price = input.Price
	}

	if input.DurationHours != nil {
		updates.DurationHours = input.DurationHours
	}

	updated, err := r.CourseService.UpdateCourse(ctx, courseID, updates)
	if err != nil {
		return nil, mapCourseError(err)
	}

	return mapCourse(updated), nil
}

// DeleteCourse is the resolver for the deleteCourse field.
func (r *mutationResolver) DeleteCourse(ctx context.Context, id string) (bool, error) {
	if r.CourseService == nil {
		return false, gqlerror.Errorf("course service not configured")
	}

	courseID, err := uuid.Parse(id)
	if err != nil {
		return false, gqlerror.Errorf("invalid course ID: %v", err)
	}

	if err := r.CourseService.DeleteCourse(ctx, courseID); err != nil {
		return false, mapCourseError(err)
	}

	return true, nil
}

// PublishCourse is the resolver for the publishCourse field.
func (r *mutationResolver) PublishCourse(ctx context.Context, id string) (*model.Course, error) {
	if r.CourseService == nil {
		return nil, gqlerror.Errorf("course service not configured")
	}

	courseID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid course ID: %v", err)
	}

	course, err := r.CourseService.PublishCourse(ctx, courseID)
	if err != nil {
		return nil, mapCourseError(err)
	}

	return mapCourse(course), nil
}

// UnpublishCourse is the resolver for the unpublishCourse field.
func (r *mutationResolver) UnpublishCourse(ctx context.Context, id string) (*model.Course, error) {
	if r.CourseService == nil {
		return nil, gqlerror.Errorf("course service not configured")
	}

	courseID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid course ID: %v", err)
	}

	course, err := r.CourseService.UnpublishCourse(ctx, courseID)
	if err != nil {
		return nil, mapCourseError(err)
	}

	return mapCourse(course), nil
}

// AddCourseLesson is the resolver for the addCourseLesson field.
func (r *mutationResolver) AddCourseLesson(ctx context.Context, courseID string, input model.AddCourseLessonInput) (*model.CourseLesson, error) {
	if r.CourseService == nil {
		return nil, gqlerror.Errorf("course service not configured")
	}

	id, err := uuid.Parse(courseID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid course ID: %v", err)
	}

	lessonID, err := uuid.Parse(input.LessonID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid lesson ID: %v", err)
	}

	if input.Ord <= 0 {
		return nil, gqlerror.Errorf("ord must be greater than 0")
	}

	lesson := &models.CourseLesson{
		LessonID:   lessonID,
		Ord:        input.Ord,
		IsRequired: true,
	}

	if input.IsRequired != nil {
		lesson.IsRequired = *input.IsRequired
	}

	created, err := r.CourseService.AddCourseLesson(ctx, id, lesson)
	if err != nil {
		return nil, mapCourseLessonError(err)
	}

	return mapCourseLesson(created), nil
}

// UpdateCourseLesson is the resolver for the updateCourseLesson field.
func (r *mutationResolver) UpdateCourseLesson(ctx context.Context, id string, input model.UpdateCourseLessonInput) (*model.CourseLesson, error) {
	if r.CourseService == nil {
		return nil, gqlerror.Errorf("course service not configured")
	}

	lessonID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid course lesson ID: %v", err)
	}

	updates := &service.CourseLessonUpdate{}

	if input.Ord != nil {
		if *input.Ord <= 0 {
			return nil, gqlerror.Errorf("ord must be greater than 0")
		}
		updates.Ord = input.Ord
	}

	if input.IsRequired != nil {
		updates.IsRequired = input.IsRequired
	}

	updated, err := r.CourseService.UpdateCourseLesson(ctx, lessonID, updates)
	if err != nil {
		return nil, mapCourseLessonError(err)
	}

	return mapCourseLesson(updated), nil
}

// ReorderCourseLessons is the resolver for the reorderCourseLessons field.
func (r *mutationResolver) ReorderCourseLessons(ctx context.Context, courseID string, lessonIDs []string) ([]*model.CourseLesson, error) {
	if r.CourseService == nil {
		return nil, gqlerror.Errorf("course service not configured")
	}

	if len(lessonIDs) == 0 {
		return nil, gqlerror.Errorf("lessonIds cannot be empty")
	}

	id, err := uuid.Parse(courseID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid course ID: %v", err)
	}

	parsed := make([]uuid.UUID, len(lessonIDs))
	for i, lessonID := range lessonIDs {
		parsedID, err := uuid.Parse(lessonID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid lesson ID at position %d: %v", i, err)
		}
		parsed[i] = parsedID
	}

	lessons, err := r.CourseService.ReorderCourseLessons(ctx, id, parsed)
	if err != nil {
		return nil, mapCourseLessonError(err)
	}

	return mapCourseLessons(lessons), nil
}

// RemoveCourseLesson is the resolver for the removeCourseLesson field.
func (r *mutationResolver) RemoveCourseLesson(ctx context.Context, id string) (bool, error) {
	if r.CourseService == nil {
		return false, gqlerror.Errorf("course service not configured")
	}

	lessonID, err := uuid.Parse(id)
	if err != nil {
		return false, gqlerror.Errorf("invalid course lesson ID: %v", err)
	}

	if err := r.CourseService.RemoveCourseLesson(ctx, lessonID); err != nil {
		return false, mapCourseLessonError(err)
	}

	return true, nil
}
