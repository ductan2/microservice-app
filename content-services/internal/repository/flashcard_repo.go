package repository

import (
	"content-services/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FlashcardSetRepository interface {
	Create(ctx context.Context, set *models.FlashcardSet) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.FlashcardSet, error)
	List(ctx context.Context, topicID, levelID *uuid.UUID, limit, offset int) ([]models.FlashcardSet, int64, error)
	Update(ctx context.Context, set *models.FlashcardSet) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type FlashcardRepository interface {
	Create(ctx context.Context, card *models.Flashcard) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Flashcard, error)
	GetBySetID(ctx context.Context, setID uuid.UUID) ([]models.Flashcard, error)
	Update(ctx context.Context, card *models.Flashcard) error
	Reorder(ctx context.Context, setID uuid.UUID, cardIDs []uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type flashcardSetRepository struct {
	db *gorm.DB
}

type flashcardRepository struct {
	db *gorm.DB
}

func NewFlashcardSetRepository(db *gorm.DB) FlashcardSetRepository {
	return &flashcardSetRepository{db: db}
}

func NewFlashcardRepository(db *gorm.DB) FlashcardRepository {
	return &flashcardRepository{db: db}
}

// FlashcardSet implementations
func (r *flashcardSetRepository) Create(ctx context.Context, set *models.FlashcardSet) error {
	// TODO: implement
	return nil
}

func (r *flashcardSetRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.FlashcardSet, error) {
	// TODO: implement
	return nil, nil
}

func (r *flashcardSetRepository) List(ctx context.Context, topicID, levelID *uuid.UUID, limit, offset int) ([]models.FlashcardSet, int64, error) {
	// TODO: implement with filters
	return nil, 0, nil
}

func (r *flashcardSetRepository) Update(ctx context.Context, set *models.FlashcardSet) error {
	// TODO: implement
	return nil
}

func (r *flashcardSetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}

// Flashcard implementations
func (r *flashcardRepository) Create(ctx context.Context, card *models.Flashcard) error {
	// TODO: implement - auto-increment ord
	return nil
}

func (r *flashcardRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Flashcard, error) {
	// TODO: implement
	return nil, nil
}

func (r *flashcardRepository) GetBySetID(ctx context.Context, setID uuid.UUID) ([]models.Flashcard, error) {
	// TODO: implement - order by ord
	return nil, nil
}

func (r *flashcardRepository) Update(ctx context.Context, card *models.Flashcard) error {
	// TODO: implement
	return nil
}

func (r *flashcardRepository) Reorder(ctx context.Context, setID uuid.UUID, cardIDs []uuid.UUID) error {
	// TODO: implement
	return nil
}

func (r *flashcardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}
