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
func (r *queryResolver) FlashcardSets(ctx context.Context, topicID *string, levelID *string, page *int, pageSize *int) (*model.FlashcardSetList, error) {
	if r.Flashcards == nil {
		return nil, gqlerror.Errorf("flashcard service not configured")
	}

	var (
		topicUUID *uuid.UUID
		levelUUID *uuid.UUID
		err       error
	)
	if topicID != nil && *topicID != "" {
		var parsed uuid.UUID
		parsed, err = uuid.Parse(*topicID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid topicId")
		}
		topicUUID = &parsed
	}
	if levelID != nil && *levelID != "" {
		var parsed uuid.UUID
		parsed, err = uuid.Parse(*levelID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid levelId")
		}
		levelUUID = &parsed
	}

	p := 1
	if page != nil && *page > 0 {
		p = *page
	}
	ps := 20
	if pageSize != nil && *pageSize > 0 {
		ps = *pageSize
	}

	sets, total, err := r.Flashcards.ListSets(ctx, topicUUID, levelUUID, p, ps)
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
