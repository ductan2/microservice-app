package service

import (
	"content-services/internal/models"
	"content-services/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type FlashcardService interface {
	// Sets
	CreateSet(ctx context.Context, set *models.FlashcardSet, tagIDs []uuid.UUID) (*models.FlashcardSet, error)
	GetSetByID(ctx context.Context, id uuid.UUID) (*models.FlashcardSet, error)
	ListSets(ctx context.Context, filter *repository.FlashcardSetFilter, sort *repository.SortOption, page, pageSize int) ([]models.FlashcardSet, int64, error)
	UpdateSet(ctx context.Context, id uuid.UUID, updates *models.FlashcardSet) (*models.FlashcardSet, error)
	DeleteSet(ctx context.Context, id uuid.UUID) error

	// Cards
	AddCard(ctx context.Context, setID uuid.UUID, card *models.Flashcard) (*models.Flashcard, error)
	GetCardByID(ctx context.Context, id uuid.UUID) (*models.Flashcard, error)
	UpdateCard(ctx context.Context, id uuid.UUID, updates *models.Flashcard) (*models.Flashcard, error)
	ReorderCards(ctx context.Context, setID uuid.UUID, cardIDs []uuid.UUID) ([]models.Flashcard, error)
	DeleteCard(ctx context.Context, id uuid.UUID) error
	GetSetCards(ctx context.Context, setID uuid.UUID) ([]models.Flashcard, error)
	ListSetCards(ctx context.Context, setID uuid.UUID, filter *repository.FlashcardFilter, sort *repository.SortOption, page, pageSize int) ([]models.Flashcard, int64, error)
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
	if set == nil {
		return nil, errors.New("flashcard set is nil")
	}
	if set.Title == "" {
		return nil, errors.New("flashcard set title is required")
	}
	if set.ID == uuid.Nil {
		set.ID = uuid.New()
	}
	if set.CreatedAt.IsZero() {
		set.CreatedAt = time.Now().UTC()
	}
	if err := s.setRepo.Create(ctx, set); err != nil {
		return nil, err
	}
	if len(tagIDs) > 0 && s.tagRepo != nil {
		for _, tagID := range tagIDs {
			if err := s.tagRepo.AddTagToContent(ctx, tagID, "flashcard_set", set.ID); err != nil {
				return nil, err
			}
		}
	}
	return set, nil
}

func (s *flashcardService) GetSetByID(ctx context.Context, id uuid.UUID) (*models.FlashcardSet, error) {
	return s.setRepo.GetByID(ctx, id)
}

func (s *flashcardService) ListSets(ctx context.Context, filter *repository.FlashcardSetFilter, sort *repository.SortOption, page, pageSize int) ([]models.FlashcardSet, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.setRepo.List(ctx, filter, sort, pageSize, offset)
}

func (s *flashcardService) UpdateSet(ctx context.Context, id uuid.UUID, updates *models.FlashcardSet) (*models.FlashcardSet, error) {
	current, err := s.setRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if updates.Title != "" {
		current.Title = updates.Title
	}
	current.Description = updates.Description
	current.TopicID = updates.TopicID
	current.LevelID = updates.LevelID
	current.CreatedBy = updates.CreatedBy
	if err := s.setRepo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *flashcardService) DeleteSet(ctx context.Context, id uuid.UUID) error {
	if err := s.setRepo.Delete(ctx, id); err != nil {
		return err
	}
	if s.tagRepo != nil {
		tags, err := s.tagRepo.GetContentTags(ctx, "flashcard_set", id)
		if err != nil {
			return err
		}
		for _, tag := range tags {
			if err := s.tagRepo.RemoveTagFromContent(ctx, tag.ID, "flashcard_set", id); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *flashcardService) AddCard(ctx context.Context, setID uuid.UUID, card *models.Flashcard) (*models.Flashcard, error) {
	if card == nil {
		return nil, errors.New("flashcard is nil")
	}
	card.SetID = setID
	if card.ID == uuid.Nil {
		card.ID = uuid.New()
	}
	if card.CreatedAt.IsZero() {
		card.CreatedAt = time.Now().UTC()
	}
	if err := s.cardRepo.Create(ctx, card); err != nil {
		return nil, err
	}
	return card, nil
}

func (s *flashcardService) GetCardByID(ctx context.Context, id uuid.UUID) (*models.Flashcard, error) {
	return s.cardRepo.GetByID(ctx, id)
}

func (s *flashcardService) UpdateCard(ctx context.Context, id uuid.UUID, updates *models.Flashcard) (*models.Flashcard, error) {
	current, err := s.cardRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if updates.FrontText != "" {
		current.FrontText = updates.FrontText
	}
	if updates.BackText != "" {
		current.BackText = updates.BackText
	}
	if updates.FrontMediaID != nil || current.FrontMediaID != nil {
		current.FrontMediaID = updates.FrontMediaID
	}
	if updates.BackMediaID != nil || current.BackMediaID != nil {
		current.BackMediaID = updates.BackMediaID
	}
	if updates.Ord != 0 {
		current.Ord = updates.Ord
	}
	if updates.Hints != nil {
		current.Hints = updates.Hints
	}
	if err := s.cardRepo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *flashcardService) ReorderCards(ctx context.Context, setID uuid.UUID, cardIDs []uuid.UUID) ([]models.Flashcard, error) {
	if err := s.cardRepo.Reorder(ctx, setID, cardIDs); err != nil {
		return nil, err
	}
	cards, _, err := s.cardRepo.ListBySetID(ctx, setID, nil, &repository.SortOption{Field: "ord", Direction: repository.SortAscending}, 0, 0)
	return cards, err
}

func (s *flashcardService) DeleteCard(ctx context.Context, id uuid.UUID) error {
	return s.cardRepo.Delete(ctx, id)
}

func (s *flashcardService) GetSetCards(ctx context.Context, setID uuid.UUID) ([]models.Flashcard, error) {
	cards, _, err := s.cardRepo.ListBySetID(ctx, setID, nil, &repository.SortOption{Field: "ord", Direction: repository.SortAscending}, 0, 0)
	return cards, err
}

func (s *flashcardService) ListSetCards(ctx context.Context, setID uuid.UUID, filter *repository.FlashcardFilter, sort *repository.SortOption, page, pageSize int) ([]models.Flashcard, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.cardRepo.ListBySetID(ctx, setID, filter, sort, pageSize, offset)
}
