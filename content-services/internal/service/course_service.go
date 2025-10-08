package service

import (
	"content-services/internal/models"
	"content-services/internal/repository"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type CourseUpdate struct {
	Title         *string
	Description   *string
	TopicID       *uuid.UUID
	LevelID       *uuid.UUID
	InstructorID  *uuid.UUID
	ThumbnailURL  *string
	IsFeatured    *bool
	Price         *float64
	DurationHours *int
}

type CourseLessonUpdate struct {
	Ord        *int
	IsRequired *bool
}

type CourseService interface {
	CreateCourse(ctx context.Context, course *models.Course) (*models.Course, error)
	GetCourseByID(ctx context.Context, id uuid.UUID) (*models.Course, error)
	ListCourses(ctx context.Context, filter *repository.CourseFilter, sort *repository.SortOption, page, pageSize int) ([]models.Course, int64, error)
	UpdateCourse(ctx context.Context, id uuid.UUID, updates *CourseUpdate) (*models.Course, error)
	PublishCourse(ctx context.Context, id uuid.UUID) (*models.Course, error)
	UnpublishCourse(ctx context.Context, id uuid.UUID) (*models.Course, error)
	DeleteCourse(ctx context.Context, id uuid.UUID) error

	AddCourseLesson(ctx context.Context, courseID uuid.UUID, lesson *models.CourseLesson) (*models.CourseLesson, error)
	UpdateCourseLesson(ctx context.Context, id uuid.UUID, updates *CourseLessonUpdate) (*models.CourseLesson, error)
	ListCourseLessons(ctx context.Context, courseID uuid.UUID, filter *repository.CourseLessonFilter, sort *repository.SortOption, page, pageSize int) ([]models.CourseLesson, int64, error)
	ReorderCourseLessons(ctx context.Context, courseID uuid.UUID, lessonIDs []uuid.UUID) ([]models.CourseLesson, error)
	RemoveCourseLesson(ctx context.Context, id uuid.UUID) error
}

type courseService struct {
	courseRepo       repository.CourseRepository
	courseLessonRepo repository.CourseLessonRepository
	lessonRepo       repository.LessonRepository
}

func NewCourseService(courseRepo repository.CourseRepository, courseLessonRepo repository.CourseLessonRepository, lessonRepo repository.LessonRepository) CourseService {
	return &courseService{
		courseRepo:       courseRepo,
		courseLessonRepo: courseLessonRepo,
		lessonRepo:       lessonRepo,
	}
}

func (s *courseService) CreateCourse(ctx context.Context, course *models.Course) (*models.Course, error) {
	if course.ID == uuid.Nil {
		course.ID = uuid.New()
	}

	now := time.Now().UTC()
	course.CreatedAt = now
	course.UpdatedAt = now
	course.IsPublished = false
	course.PublishedAt = sql.NullTime{}

	if err := s.courseRepo.Create(ctx, course); err != nil {
		return nil, err
	}

	return course, nil
}

func (s *courseService) GetCourseByID(ctx context.Context, id uuid.UUID) (*models.Course, error) {
	return s.courseRepo.GetByID(ctx, id)
}

func (s *courseService) ListCourses(ctx context.Context, filter *repository.CourseFilter, sort *repository.SortOption, page, pageSize int) ([]models.Course, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.courseRepo.List(ctx, filter, sort, pageSize, offset)
}

func (s *courseService) UpdateCourse(ctx context.Context, id uuid.UUID, updates *CourseUpdate) (*models.Course, error) {
	course, err := s.courseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if updates.Title != nil {
		course.Title = *updates.Title
	}
	if updates.Description != nil {
		course.Description = *updates.Description
	}
	if updates.TopicID != nil {
		if *updates.TopicID == uuid.Nil {
			course.TopicID = nil
		} else {
			course.TopicID = updates.TopicID
		}
	}
	if updates.LevelID != nil {
		if *updates.LevelID == uuid.Nil {
			course.LevelID = nil
		} else {
			course.LevelID = updates.LevelID
		}
	}
	if updates.InstructorID != nil {
		if *updates.InstructorID == uuid.Nil {
			course.InstructorID = nil
		} else {
			course.InstructorID = updates.InstructorID
		}
	}
	if updates.ThumbnailURL != nil {
		course.ThumbnailURL = *updates.ThumbnailURL
	}
	if updates.IsFeatured != nil {
		course.IsFeatured = *updates.IsFeatured
	}
	if updates.Price != nil {
		course.Price = *updates.Price
	}
	if updates.DurationHours != nil {
		course.DurationHours = *updates.DurationHours
	}

	course.UpdatedAt = time.Now().UTC()

	if err := s.courseRepo.Update(ctx, course); err != nil {
		return nil, err
	}

	return course, nil
}

func (s *courseService) PublishCourse(ctx context.Context, id uuid.UUID) (*models.Course, error) {
	now := time.Now().UTC()
	if err := s.courseRepo.Publish(ctx, id, now); err != nil {
		return nil, err
	}
	return s.courseRepo.GetByID(ctx, id)
}

func (s *courseService) UnpublishCourse(ctx context.Context, id uuid.UUID) (*models.Course, error) {
	now := time.Now().UTC()
	if err := s.courseRepo.Unpublish(ctx, id, now); err != nil {
		return nil, err
	}
	return s.courseRepo.GetByID(ctx, id)
}

func (s *courseService) DeleteCourse(ctx context.Context, id uuid.UUID) error {
	if err := s.courseRepo.Delete(ctx, id); err != nil {
		return err
	}
	if s.courseLessonRepo != nil {
		_ = s.courseLessonRepo.DeleteByCourseID(ctx, id)
	}
	return nil
}

func (s *courseService) AddCourseLesson(ctx context.Context, courseID uuid.UUID, lesson *models.CourseLesson) (*models.CourseLesson, error) {
	if _, err := s.courseRepo.GetByID(ctx, courseID); err != nil {
		return nil, err
	}

	if _, err := s.lessonRepo.GetByID(ctx, lesson.LessonID); err != nil {
		return nil, err
	}

	if lesson.ID == uuid.Nil {
		lesson.ID = uuid.New()
	}
	lesson.CourseID = courseID
	if lesson.CreatedAt.IsZero() {
		lesson.CreatedAt = time.Now().UTC()
	}

	if err := s.courseLessonRepo.Create(ctx, lesson); err != nil {
		return nil, err
	}

	return lesson, nil
}

func (s *courseService) UpdateCourseLesson(ctx context.Context, id uuid.UUID, updates *CourseLessonUpdate) (*models.CourseLesson, error) {
	lesson, err := s.courseLessonRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if updates.Ord != nil {
		lesson.Ord = *updates.Ord
	}
	if updates.IsRequired != nil {
		lesson.IsRequired = *updates.IsRequired
	}

	if err := s.courseLessonRepo.Update(ctx, lesson); err != nil {
		return nil, err
	}

	return lesson, nil
}

func (s *courseService) ListCourseLessons(ctx context.Context, courseID uuid.UUID, filter *repository.CourseLessonFilter, sort *repository.SortOption, page, pageSize int) ([]models.CourseLesson, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.courseLessonRepo.ListByCourseID(ctx, courseID, filter, sort, pageSize, offset)
}

func (s *courseService) ReorderCourseLessons(ctx context.Context, courseID uuid.UUID, lessonIDs []uuid.UUID) ([]models.CourseLesson, error) {
	if err := s.courseLessonRepo.Reorder(ctx, courseID, lessonIDs); err != nil {
		return nil, err
	}
	lessons, _, err := s.courseLessonRepo.ListByCourseID(ctx, courseID, nil, &repository.SortOption{Field: "ord", Direction: repository.SortAscending}, 0, 0)
	if err != nil {
		return nil, err
	}
	return lessons, nil
}

func (s *courseService) RemoveCourseLesson(ctx context.Context, id uuid.UUID) error {
	if err := s.courseLessonRepo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}
