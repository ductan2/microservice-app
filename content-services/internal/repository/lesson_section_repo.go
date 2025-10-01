package repository

import (
	"content-services/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LessonSectionRepository interface {
	Create(ctx context.Context, section *models.LessonSection) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.LessonSection, error)
	GetByLessonID(ctx context.Context, lessonID uuid.UUID) ([]models.LessonSection, error)
	Update(ctx context.Context, section *models.LessonSection) error
	Reorder(ctx context.Context, lessonID uuid.UUID, sectionIDs []uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type lessonSectionRepository struct {
	db *gorm.DB
}

func NewLessonSectionRepository(db *gorm.DB) LessonSectionRepository {
	return &lessonSectionRepository{db: db}
}

func (r *lessonSectionRepository) Create(ctx context.Context, section *models.LessonSection) error {
	// TODO: implement - auto-increment ord
	return nil
}

func (r *lessonSectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.LessonSection, error) {
	// TODO: implement
	return nil, nil
}

func (r *lessonSectionRepository) GetByLessonID(ctx context.Context, lessonID uuid.UUID) ([]models.LessonSection, error) {
	// TODO: implement - order by ord
	return nil, nil
}

func (r *lessonSectionRepository) Update(ctx context.Context, section *models.LessonSection) error {
	// TODO: implement
	return nil
}

func (r *lessonSectionRepository) Reorder(ctx context.Context, lessonID uuid.UUID, sectionIDs []uuid.UUID) error {
	// TODO: implement - update ord for each section
	return nil
}

func (r *lessonSectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}
