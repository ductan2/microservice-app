package repository

import (
	"content-services/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LessonFilter struct {
	TopicID     *uuid.UUID
	LevelID     *uuid.UUID
	IsPublished *bool
	Search      string
}

type LessonRepository interface {
	Create(ctx context.Context, lesson *models.Lesson) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Lesson, error)
	GetByCode(ctx context.Context, code string) (*models.Lesson, error)
	List(ctx context.Context, filter *LessonFilter, limit, offset int) ([]models.Lesson, int64, error)
	Update(ctx context.Context, lesson *models.Lesson) error
	Publish(ctx context.Context, id uuid.UUID) error
	Unpublish(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type lessonRepository struct {
	db *gorm.DB
}

func NewLessonRepository(db *gorm.DB) LessonRepository {
	return &lessonRepository{db: db}
}

func (r *lessonRepository) Create(ctx context.Context, lesson *models.Lesson) error {
	// TODO: implement
	return nil
}

func (r *lessonRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Lesson, error) {
	// TODO: implement
	return nil, nil
}

func (r *lessonRepository) GetByCode(ctx context.Context, code string) (*models.Lesson, error) {
	// TODO: implement
	return nil, nil
}

func (r *lessonRepository) List(ctx context.Context, filter *LessonFilter, limit, offset int) ([]models.Lesson, int64, error) {
	// TODO: implement with filtering
	return nil, 0, nil
}

func (r *lessonRepository) Update(ctx context.Context, lesson *models.Lesson) error {
	// TODO: implement
	return nil
}

func (r *lessonRepository) Publish(ctx context.Context, id uuid.UUID) error {
	// TODO: implement - set is_published=true, published_at=now()
	return nil
}

func (r *lessonRepository) Unpublish(ctx context.Context, id uuid.UUID) error {
	// TODO: implement - set is_published=false
	return nil
}

func (r *lessonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}
