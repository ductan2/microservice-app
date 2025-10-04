package resolver

import (
	"content-services/graph/model"
	"content-services/internal/models"
	"context"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// CreateFlashcardSet is the resolver for the createFlashcardSet field.
func (r *mutationResolver) CreateFlashcardSet(ctx context.Context, input model.CreateFlashcardSetInput) (*model.FlashcardSet, error) {
	if r.FlashcardService == nil {
		return nil, gqlerror.Errorf("flashcard service not configured")
	}

	var topicID *uuid.UUID
	if input.TopicID != nil && *input.TopicID != "" {
		id, err := uuid.Parse(*input.TopicID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid topicId")
		}
		topicID = &id
	}

	var levelID *uuid.UUID
	if input.LevelID != nil && *input.LevelID != "" {
		id, err := uuid.Parse(*input.LevelID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid levelId")
		}
		levelID = &id
	}

	var createdBy *uuid.UUID
	if input.CreatedBy != nil && *input.CreatedBy != "" {
		id, err := uuid.Parse(*input.CreatedBy)
		if err != nil {
			return nil, gqlerror.Errorf("invalid createdBy")
		}
		createdBy = &id
	}

	set := &models.FlashcardSet{
		Title:       input.Title,
		Description: derefString(input.Description),
		TopicID:     topicID,
		LevelID:     levelID,
		CreatedBy:   createdBy,
	}

	created, err := r.FlashcardService.CreateSet(ctx, set, nil)
	if err != nil {
		return nil, mapFlashcardError(err)
	}
	return mapFlashcardSet(created), nil
}

// AddFlashcard is the resolver for the addFlashcard field.
func (r *mutationResolver) AddFlashcard(ctx context.Context, input model.AddFlashcardInput) (*model.Flashcard, error) {
	if r.FlashcardService == nil {
		return nil, gqlerror.Errorf("flashcard service not configured")
	}

	setID, err := uuid.Parse(input.SetID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid setId")
	}

	card := &models.Flashcard{
		FrontText: input.FrontText,
		BackText:  input.BackText,
	}

	if input.FrontMediaID != nil && *input.FrontMediaID != "" {
		id, err := uuid.Parse(*input.FrontMediaID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid frontMediaId")
		}
		card.FrontMediaID = &id
	}

	if input.BackMediaID != nil && *input.BackMediaID != "" {
		id, err := uuid.Parse(*input.BackMediaID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid backMediaId")
		}
		card.BackMediaID = &id
	}

	if input.Hints != nil {
		card.Hints = make([]string, len(input.Hints))
		copy(card.Hints, input.Hints)
	}

	created, err := r.FlashcardService.AddCard(ctx, setID, card)
	if err != nil {
		return nil, mapFlashcardError(err)
	}
	return mapFlashcard(created), nil
}
