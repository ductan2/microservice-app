package service

import (
	"content-services/internal/models"
	"content-services/internal/repository"
	"context"

	"github.com/google/uuid"
)

type FlashcardService interface {
	// Sets
	CreateSet(ctx context.Context, set *models.FlashcardSet, tagIDs []uuid.UUID) (*models.FlashcardSet, error)
	GetSetByID(ctx context.Context, id uuid.UUID) (*models.FlashcardSet, error)
	ListSets(ctx context.Context, topicID, levelID *uuid.UUID, page, pageSize int) ([]models.FlashcardSet, int64, error)
	UpdateSet(ctx context.Context, id uuid.UUID, updates *models.FlashcardSet) (*models.FlashcardSet, error)
	DeleteSet(ctx context.Context, id uuid.UUID) error

	// Cards
	AddCard(ctx context.Context, setID uuid.UUID, card *models.Flashcard) (*models.Flashcard, error)
	UpdateCard(ctx context.Context, id uuid.UUID, updates *models.Flashcard) (*models.Flashcard, error)
	ReorderCards(ctx context.Context, setID uuid.UUID, cardIDs []uuid.UUID) ([]models.Flashcard, error)
	DeleteCard(ctx context.Context, id uuid.UUID) error
	GetSetCards(ctx context.Context, setID uuid.UUID) ([]models.Flashcard, error)
}

type flashcardService struct {
	setRepo  repository.FlashcardSetRepository
	cardRepo repository.FlashcardRepository
	tagRepo  repository.TagRepository
}

func NewFlashcardService(
	setRepo repository.FlashcardSetRepository,
	cardRepo repository.FlashcardRepository,
	tagRepo repository.TagRepository,
) FlashcardService {
	return &flashcardService{
		setRepo:  setRepo,
		cardRepo: cardRepo,
		tagRepo:  tagRepo,
	}
}

func (s *flashcardService) CreateSet(ctx context.Context, set *models.FlashcardSet, tagIDs []uuid.UUID) (*models.FlashcardSet, error) {
	// TODO: implement
	return nil, nil
}

func (s *flashcardService) GetSetByID(ctx context.Context, id uuid.UUID) (*models.FlashcardSet, error) {
	// TODO: implement
	return nil, nil
}

func (s *flashcardService) ListSets(ctx context.Context, topicID, levelID *uuid.UUID, page, pageSize int) ([]models.FlashcardSet, int64, error) {
	// TODO: implement
	return nil, 0, nil
}

func (s *flashcardService) UpdateSet(ctx context.Context, id uuid.UUID, updates *models.FlashcardSet) (*models.FlashcardSet, error) {
	// TODO: implement
	return nil, nil
}

func (s *flashcardService) DeleteSet(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}

func (s *flashcardService) AddCard(ctx context.Context, setID uuid.UUID, card *models.Flashcard) (*models.Flashcard, error) {
	// TODO: implement
	return nil, nil
}

func (s *flashcardService) UpdateCard(ctx context.Context, id uuid.UUID, updates *models.Flashcard) (*models.Flashcard, error) {
	// TODO: implement
	return nil, nil
}

func (s *flashcardService) ReorderCards(ctx context.Context, setID uuid.UUID, cardIDs []uuid.UUID) ([]models.Flashcard, error) {
	// TODO: implement
	return nil, nil
}

func (s *flashcardService) DeleteCard(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}

func (s *flashcardService) GetSetCards(ctx context.Context, setID uuid.UUID) ([]models.Flashcard, error) {
	// TODO: implement
	return nil, nil
}
