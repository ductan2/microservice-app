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

// UpdateFlashcardSet is the resolver for the updateFlashcardSet field.
func (r *mutationResolver) UpdateFlashcardSet(ctx context.Context, id string, input model.UpdateFlashcardSetInput) (*model.FlashcardSet, error) {
	if r.FlashcardService == nil {
		return nil, gqlerror.Errorf("flashcard service not configured")
	}

	setID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid id")
	}

	current, err := r.FlashcardService.GetSetByID(ctx, setID)
	if err != nil {
		return nil, mapFlashcardError(err)
	}

	updates := &models.FlashcardSet{
		Description: current.Description,
		TopicID:     current.TopicID,
		LevelID:     current.LevelID,
		CreatedBy:   current.CreatedBy,
	}

	if input.Title != nil {
		updates.Title = derefString(input.Title)
	}
	if input.Description != nil {
		updates.Description = derefString(input.Description)
	}
	if input.TopicID != nil {
		if *input.TopicID == "" {
			updates.TopicID = nil
		} else {
			topicID, err := uuid.Parse(*input.TopicID)
			if err != nil {
				return nil, gqlerror.Errorf("invalid topicId")
			}
			updates.TopicID = &topicID
		}
	}
	if input.LevelID != nil {
		if *input.LevelID == "" {
			updates.LevelID = nil
		} else {
			levelID, err := uuid.Parse(*input.LevelID)
			if err != nil {
				return nil, gqlerror.Errorf("invalid levelId")
			}
			updates.LevelID = &levelID
		}
	}
	if input.CreatedBy != nil {
		if *input.CreatedBy == "" {
			updates.CreatedBy = nil
		} else {
			createdBy, err := uuid.Parse(*input.CreatedBy)
			if err != nil {
				return nil, gqlerror.Errorf("invalid createdBy")
			}
			updates.CreatedBy = &createdBy
		}
	}

	updated, err := r.FlashcardService.UpdateSet(ctx, setID, updates)
	if err != nil {
		return nil, mapFlashcardError(err)
	}
	return mapFlashcardSet(updated), nil
}

// DeleteFlashcardSet is the resolver for the deleteFlashcardSet field.
func (r *mutationResolver) DeleteFlashcardSet(ctx context.Context, id string) (bool, error) {
	if r.FlashcardService == nil {
		return false, gqlerror.Errorf("flashcard service not configured")
	}

	setID, err := uuid.Parse(id)
	if err != nil {
		return false, gqlerror.Errorf("invalid id")
	}

	if err := r.FlashcardService.DeleteSet(ctx, setID); err != nil {
		return false, mapFlashcardError(err)
	}
	return true, nil
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

// UpdateFlashcard is the resolver for the updateFlashcard field.
func (r *mutationResolver) UpdateFlashcard(ctx context.Context, id string, input model.UpdateFlashcardInput) (*model.Flashcard, error) {
	if r.FlashcardService == nil {
		return nil, gqlerror.Errorf("flashcard service not configured")
	}

	cardID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid id")
	}

	current, err := r.FlashcardService.GetCardByID(ctx, cardID)
	if err != nil {
		return nil, mapFlashcardError(err)
	}

	updates := &models.Flashcard{
		FrontMediaID: current.FrontMediaID,
		BackMediaID:  current.BackMediaID,
	}

	if input.FrontText != nil {
		updates.FrontText = derefString(input.FrontText)
	}
	if input.BackText != nil {
		updates.BackText = derefString(input.BackText)
	}
	if input.FrontMediaID != nil {
		if *input.FrontMediaID == "" {
			updates.FrontMediaID = nil
		} else {
			frontMediaID, err := uuid.Parse(*input.FrontMediaID)
			if err != nil {
				return nil, gqlerror.Errorf("invalid frontMediaId")
			}
			updates.FrontMediaID = &frontMediaID
		}
	}
	if input.BackMediaID != nil {
		if *input.BackMediaID == "" {
			updates.BackMediaID = nil
		} else {
			backMediaID, err := uuid.Parse(*input.BackMediaID)
			if err != nil {
				return nil, gqlerror.Errorf("invalid backMediaId")
			}
			updates.BackMediaID = &backMediaID
		}
	}
	if input.Ord != nil {
		updates.Ord = *input.Ord
	}
	if input.Hints != nil {
		updates.Hints = make([]string, len(input.Hints))
		copy(updates.Hints, input.Hints)
	}

	updated, err := r.FlashcardService.UpdateCard(ctx, cardID, updates)
	if err != nil {
		return nil, mapFlashcardError(err)
	}
	return mapFlashcard(updated), nil
}

// DeleteFlashcard is the resolver for the deleteFlashcard field.
func (r *mutationResolver) DeleteFlashcard(ctx context.Context, id string) (bool, error) {
	if r.FlashcardService == nil {
		return false, gqlerror.Errorf("flashcard service not configured")
	}

	cardID, err := uuid.Parse(id)
	if err != nil {
		return false, gqlerror.Errorf("invalid id")
	}

	if err := r.FlashcardService.DeleteCard(ctx, cardID); err != nil {
		return false, mapFlashcardError(err)
	}
	return true, nil
}
