package service

import (
	"content-services/internal/models"
	"content-services/internal/repository"
	"context"

	"github.com/google/uuid"
)

type LessonService interface {
	CreateLesson(ctx context.Context, lesson *models.Lesson, tagIDs []uuid.UUID) (*models.Lesson, error)
	GetLessonByID(ctx context.Context, id uuid.UUID) (*models.Lesson, error)
	GetLessonByCode(ctx context.Context, code string) (*models.Lesson, error)
	ListLessons(ctx context.Context, filter *repository.LessonFilter, page, pageSize int) ([]models.Lesson, int64, error)
	UpdateLesson(ctx context.Context, id uuid.UUID, updates *models.Lesson) (*models.Lesson, error)
	PublishLesson(ctx context.Context, id uuid.UUID) (*models.Lesson, error)
	UnpublishLesson(ctx context.Context, id uuid.UUID) (*models.Lesson, error)
	DeleteLesson(ctx context.Context, id uuid.UUID) error

	// Sections
	AddSection(ctx context.Context, lessonID uuid.UUID, section *models.LessonSection) (*models.LessonSection, error)
	UpdateSection(ctx context.Context, id uuid.UUID, updates *models.LessonSection) (*models.LessonSection, error)
	ReorderSections(ctx context.Context, lessonID uuid.UUID, sectionIDs []uuid.UUID) ([]models.LessonSection, error)
	DeleteSection(ctx context.Context, id uuid.UUID) error
	GetLessonSections(ctx context.Context, lessonID uuid.UUID) ([]models.LessonSection, error)
}

type lessonService struct {
	lessonRepo  repository.LessonRepository
	sectionRepo repository.LessonSectionRepository
	tagRepo     repository.TagRepository
	outboxRepo  repository.OutboxRepository
}

func NewLessonService(
	lessonRepo repository.LessonRepository,
	sectionRepo repository.LessonSectionRepository,
	tagRepo repository.TagRepository,
	outboxRepo repository.OutboxRepository,
) LessonService {
	return &lessonService{
		lessonRepo:  lessonRepo,
		sectionRepo: sectionRepo,
		tagRepo:     tagRepo,
		outboxRepo:  outboxRepo,
	}
}

func (s *lessonService) CreateLesson(ctx context.Context, lesson *models.Lesson, tagIDs []uuid.UUID) (*models.Lesson, error) {
	// TODO: implement - create lesson, add tags, publish outbox event
	return nil, nil
}

func (s *lessonService) GetLessonByID(ctx context.Context, id uuid.UUID) (*models.Lesson, error) {
	// TODO: implement
	return nil, nil
}

func (s *lessonService) GetLessonByCode(ctx context.Context, code string) (*models.Lesson, error) {
	// TODO: implement
	return nil, nil
}

func (s *lessonService) ListLessons(ctx context.Context, filter *repository.LessonFilter, page, pageSize int) ([]models.Lesson, int64, error) {
	// TODO: implement pagination
	return nil, 0, nil
}

func (s *lessonService) UpdateLesson(ctx context.Context, id uuid.UUID, updates *models.Lesson) (*models.Lesson, error) {
	// TODO: implement
	return nil, nil
}

func (s *lessonService) PublishLesson(ctx context.Context, id uuid.UUID) (*models.Lesson, error) {
	// TODO: implement - mark as published, send outbox event
	return nil, nil
}

func (s *lessonService) UnpublishLesson(ctx context.Context, id uuid.UUID) (*models.Lesson, error) {
	// TODO: implement
	return nil, nil
}

func (s *lessonService) DeleteLesson(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}

func (s *lessonService) AddSection(ctx context.Context, lessonID uuid.UUID, section *models.LessonSection) (*models.LessonSection, error) {
	// TODO: implement
	return nil, nil
}

func (s *lessonService) UpdateSection(ctx context.Context, id uuid.UUID, updates *models.LessonSection) (*models.LessonSection, error) {
	// TODO: implement
	return nil, nil
}

func (s *lessonService) ReorderSections(ctx context.Context, lessonID uuid.UUID, sectionIDs []uuid.UUID) ([]models.LessonSection, error) {
	// TODO: implement
	return nil, nil
}

func (s *lessonService) DeleteSection(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}

func (s *lessonService) GetLessonSections(ctx context.Context, lessonID uuid.UUID) ([]models.LessonSection, error) {
	// TODO: implement
	return nil, nil
}
