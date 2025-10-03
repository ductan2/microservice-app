package service

import (
	"content-services/internal/models"
	"content-services/internal/repository"
	"context"
	"time"

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
	outboxRepo  repository.OutboxRepository
}

func NewLessonService(
	lessonRepo repository.LessonRepository,
	sectionRepo repository.LessonSectionRepository,
	outboxRepo repository.OutboxRepository,
) LessonService {
	return &lessonService{
		lessonRepo:  lessonRepo,
		sectionRepo: sectionRepo,
		outboxRepo:  outboxRepo,
	}
}

func (s *lessonService) CreateLesson(ctx context.Context, lesson *models.Lesson, tagIDs []uuid.UUID) (*models.Lesson, error) {
	// Set defaults
	if lesson.ID == uuid.Nil {
		lesson.ID = uuid.New()
	}
	now := time.Now().UTC()
	lesson.CreatedAt = now
	lesson.UpdatedAt = now
	lesson.IsPublished = false
	lesson.Version = 1

	// Create lesson
	if err := s.lessonRepo.Create(ctx, lesson); err != nil {
		return nil, err
	}

	// TODO: Add tags to content_tags table if tagIDs provided

	// TODO: Publish outbox event for lesson created
	// event := &models.Outbox{
	// 	AggregateID: lesson.ID,
	// 	Topic:       "content.events",
	// 	Type:        "LessonCreated",
	// 	Payload:     map[string]any{"lesson_id": lesson.ID.String(), "title": lesson.Title},
	// 	CreatedAt:   now,
	// }
	// s.outboxRepo.Create(ctx, event)

	return lesson, nil
}

func (s *lessonService) GetLessonByID(ctx context.Context, id uuid.UUID) (*models.Lesson, error) {
	return s.lessonRepo.GetByID(ctx, id)
}

func (s *lessonService) GetLessonByCode(ctx context.Context, code string) (*models.Lesson, error) {
	return s.lessonRepo.GetByCode(ctx, code)
}

func (s *lessonService) ListLessons(ctx context.Context, filter *repository.LessonFilter, page, pageSize int) ([]models.Lesson, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.lessonRepo.List(ctx, filter, pageSize, offset)
}

func (s *lessonService) UpdateLesson(ctx context.Context, id uuid.UUID, updates *models.Lesson) (*models.Lesson, error) {
	// Get existing lesson
	existing, err := s.lessonRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.Title != "" {
		existing.Title = updates.Title
	}
	if updates.Description != "" {
		existing.Description = updates.Description
	}
	if updates.TopicID != nil {
		if *updates.TopicID == uuid.Nil {
			existing.TopicID = nil
		} else {
			existing.TopicID = updates.TopicID
		}
	}
	if updates.LevelID != nil {
		if *updates.LevelID == uuid.Nil {
			existing.LevelID = nil
		} else {
			existing.LevelID = updates.LevelID
		}
	}

	existing.UpdatedAt = time.Now().UTC()

	// Save updates
	if err := s.lessonRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	// TODO: Publish outbox event for lesson updated

	return existing, nil
}

func (s *lessonService) PublishLesson(ctx context.Context, id uuid.UUID) (*models.Lesson, error) {
	// Publish the lesson
	if err := s.lessonRepo.Publish(ctx, id); err != nil {
		return nil, err
	}

	// Get updated lesson
	lesson, err := s.lessonRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: Publish outbox event for lesson published
	// event := &models.Outbox{
	// 	AggregateID: id,
	// 	Topic:       "content.events",
	// 	Type:        "LessonPublished",
	// 	Payload: map[string]any{
	// 		"lesson_id": id.String(),
	// 		"title":     lesson.Title,
	// 		"published_at": lesson.PublishedAt.Time,
	// 	},
	// 	CreatedAt: time.Now().UTC(),
	// }
	// s.outboxRepo.Create(ctx, event)

	return lesson, nil
}

func (s *lessonService) UnpublishLesson(ctx context.Context, id uuid.UUID) (*models.Lesson, error) {
	if err := s.lessonRepo.Unpublish(ctx, id); err != nil {
		return nil, err
	}

	lesson, err := s.lessonRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: Publish outbox event for lesson unpublished

	return lesson, nil
}

func (s *lessonService) DeleteLesson(ctx context.Context, id uuid.UUID) error {
	// Delete lesson (sections will be cascade deleted)
	if err := s.lessonRepo.Delete(ctx, id); err != nil {
		return err
	}

	// TODO: Publish outbox event for lesson deleted
	// event := &models.Outbox{
	// 	AggregateID: id,
	// 	Topic:       "content.events",
	// 	Type:        "LessonDeleted",
	// 	Payload:     map[string]any{"lesson_id": id.String()},
	// 	CreatedAt:   time.Now().UTC(),
	// }
	// s.outboxRepo.Create(ctx, event)

	return nil
}

// ============= SECTION METHODS =============

func (s *lessonService) AddSection(ctx context.Context, lessonID uuid.UUID, section *models.LessonSection) (*models.LessonSection, error) {
	// Verify lesson exists
	if _, err := s.lessonRepo.GetByID(ctx, lessonID); err != nil {
		return nil, err
	}

	// Set defaults
	if section.ID == uuid.Nil {
		section.ID = uuid.New()
	}
	section.LessonID = lessonID
	section.CreatedAt = time.Now().UTC()

	// Get current max ord
	sections, err := s.sectionRepo.GetByLessonID(ctx, lessonID)
	if err != nil {
		return nil, err
	}

	maxOrd := 0
	for _, s := range sections {
		if s.Ord > maxOrd {
			maxOrd = s.Ord
		}
	}
	section.Ord = maxOrd + 1

	if err := s.sectionRepo.Create(ctx, section); err != nil {
		return nil, err
	}

	return section, nil
}

func (s *lessonService) UpdateSection(ctx context.Context, id uuid.UUID, updates *models.LessonSection) (*models.LessonSection, error) {
	// Get existing section
	existing, err := s.sectionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.Type != "" {
		existing.Type = updates.Type
	}
	if updates.Body != nil {
		existing.Body = updates.Body
	}

	if err := s.sectionRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *lessonService) ReorderSections(ctx context.Context, lessonID uuid.UUID, sectionIDs []uuid.UUID) ([]models.LessonSection, error) {
	if err := s.sectionRepo.Reorder(ctx, lessonID, sectionIDs); err != nil {
		return nil, err
	}

	return s.sectionRepo.GetByLessonID(ctx, lessonID)
}

func (s *lessonService) DeleteSection(ctx context.Context, id uuid.UUID) error {
	return s.sectionRepo.Delete(ctx, id)
}

func (s *lessonService) GetLessonSections(ctx context.Context, lessonID uuid.UUID) ([]models.LessonSection, error) {
	return s.sectionRepo.GetByLessonID(ctx, lessonID)
}
