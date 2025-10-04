package resolver

import (
	"content-services/graph/model"
	"content-services/internal/repository"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// FlashcardSet is the resolver for the flashcardSet field.
func (r *queryResolver) FlashcardSet(ctx context.Context, id string) (*model.FlashcardSet, error) {
	if r.Flashcards == nil {
		return nil, gqlerror.Errorf("flashcard service not configured")
	}
	setID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid flashcard set id")
	}
	set, err := r.Flashcards.GetSetByID(ctx, setID)
	if err != nil {
		if errors.Is(err, repository.ErrFlashcardSetNotFound) {
			return nil, nil
		}
		return nil, mapFlashcardError(err)
	}
	return mapFlashcardSet(set), nil
}

// FlashcardSets is the resolver for the flashcardSets field.
func (r *queryResolver) FlashcardSets(ctx context.Context, filter *model.FlashcardSetFilterInput, page *int, pageSize *int, orderBy *model.FlashcardSetOrderInput) (*model.FlashcardSetList, error) {
	if r.Flashcards == nil {
		return nil, gqlerror.Errorf("flashcard service not configured")
	}

	setFilter, err := buildFlashcardSetFilter(filter)
	if err != nil {
		return nil, err
	}
	setSort := buildFlashcardSetOrder(orderBy)

	p := 1
	if page != nil && *page > 0 {
		p = *page
	}
	ps := 20
	if pageSize != nil && *pageSize > 0 {
		ps = *pageSize
	}

	sets, total, err := r.Flashcards.ListSets(ctx, setFilter, setSort, p, ps)
	if err != nil {
		return nil, mapFlashcardError(err)
	}

	items := make([]*model.FlashcardSet, 0, len(sets))
	for i := range sets {
		items = append(items, mapFlashcardSet(&sets[i]))
	}

	return &model.FlashcardSetList{
		Items:      items,
		TotalCount: int(total),
		Page:       p,
		PageSize:   ps,
	}, nil
}

// Flashcards is the resolver for the flashcards field.
func (r *queryResolver) Flashcards(ctx context.Context, setID string, filter *model.FlashcardFilterInput, page *int, pageSize *int, orderBy *model.FlashcardOrderInput) (*model.FlashcardCollection, error) {
	if r.Flashcards == nil {
		return nil, gqlerror.Errorf("flashcard service not configured")
	}

	id, err := uuid.Parse(setID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid setId")
	}

	cardFilter := buildFlashcardFilter(filter)
	cardSort := buildFlashcardOrder(orderBy)

	p := 1
	if page != nil && *page > 0 {
		p = *page
	}
	ps := 20
	if pageSize != nil && *pageSize > 0 {
		ps = *pageSize
	}

	cards, total, err := r.Flashcards.ListSetCards(ctx, id, cardFilter, cardSort, p, ps)
	if err != nil {
		return nil, mapFlashcardError(err)
	}

	return &model.FlashcardCollection{
		Items:      mapFlashcards(cards),
		TotalCount: int(total),
		Page:       p,
		PageSize:   ps,
	}, nil
}
