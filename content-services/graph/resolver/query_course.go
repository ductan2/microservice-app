package resolver

import (
	"content-services/graph/model"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Course is the resolver for the course field.
func (r *queryResolver) Course(ctx context.Context, id string) (*model.Course, error) {
	if r.CourseService == nil {
		return nil, gqlerror.Errorf("course service not configured")
	}

	courseID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid course ID: %v", err)
	}

	course, err := r.CourseService.GetCourseByID(ctx, courseID)
	fmt.Println("course", course)
	fmt.Println("err", err)
	if err != nil {
		return nil, mapCourseError(err)
	}

	return mapCourse(course), nil
}

// Courses is the resolver for the courses field.
func (r *queryResolver) Courses(ctx context.Context, filter *model.CourseFilterInput, page *int, pageSize *int, orderBy *model.CourseOrderInput) (*model.CourseCollection, error) {
	if r.CourseService == nil {
		return nil, gqlerror.Errorf("course service not configured")
	}

	courseFilter, err := buildCourseFilter(filter)
	if err != nil {
		return nil, err
	}

	sortOption := buildCourseOrder(orderBy)

	pageVal := 1
	if page != nil && *page > 0 {
		pageVal = *page
	}

	pageSizeVal := 20
	if pageSize != nil && *pageSize > 0 {
		pageSizeVal = *pageSize
	}

	courses, total, err := r.CourseService.ListCourses(ctx, courseFilter, sortOption, pageVal, pageSizeVal)
	if err != nil {
		return nil, mapCourseError(err)
	}

	items := make([]*model.Course, 0, len(courses))
	for i := range courses {
		items = append(items, mapCourse(&courses[i]))
	}

	return &model.CourseCollection{
		Items:      items,
		TotalCount: int(total),
		Page:       pageVal,
		PageSize:   pageSizeVal,
	}, nil
}

// CourseLessons is the resolver for the courseLessons field.
func (r *queryResolver) CourseLessons(ctx context.Context, courseID string, filter *model.CourseLessonFilterInput, page *int, pageSize *int, orderBy *model.CourseLessonOrderInput) (*model.CourseLessonCollection, error) {
	if r.CourseService == nil {
		return nil, gqlerror.Errorf("course service not configured")
	}

	id, err := uuid.Parse(courseID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid course ID: %v", err)
	}

	lessonFilter := buildCourseLessonFilter(filter)
	sortOption := buildCourseLessonOrder(orderBy)

	pageVal := 1
	if page != nil && *page > 0 {
		pageVal = *page
	}

	pageSizeVal := 20
	if pageSize != nil && *pageSize > 0 {
		pageSizeVal = *pageSize
	}

	lessons, total, err := r.CourseService.ListCourseLessons(ctx, id, lessonFilter, sortOption, pageVal, pageSizeVal)
	if err != nil {
		return nil, mapCourseLessonError(err)
	}

	return &model.CourseLessonCollection{
		Items:      mapCourseLessons(lessons),
		TotalCount: int(total),
		Page:       pageVal,
		PageSize:   pageSizeVal,
	}, nil
}
