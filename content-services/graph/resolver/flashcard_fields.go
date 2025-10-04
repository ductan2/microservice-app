package resolver

import (
	"content-services/graph/generated"
	"content-services/graph/model"
	"context"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Cards is the resolver for the cards field.
func (r *flashcardSetResolver) Cards(ctx context.Context, obj *model.FlashcardSet) ([]*model.Flashcard, error) {
	if r.FlashcardService == nil {
		return nil, gqlerror.Errorf("flashcard service not configured")
	}

	setID, err := uuid.Parse(obj.ID)
	if err != nil {	
		return nil, gqlerror.Errorf("invalid flashcard set id")
	}

	cards, err := r.FlashcardService.GetSetCards(ctx, setID)
	if err != nil {
		return nil, mapFlashcardError(err)
	}

	return mapFlashcards(cards), nil
}

// FlashcardSet returns generated.FlashcardSetResolver implementation.
func (r *Resolver) FlashcardSet() generated.FlashcardSetResolver { return &flashcardSetResolver{r} }

type flashcardSetResolver struct{ *Resolver }

// Tags is the resolver for the tags field.
func (r *flashcardSetResolver) Tags(ctx context.Context, obj *model.FlashcardSet) ([]*model.Tag, error) {
	if r.TagRepo == nil {
		return []*model.Tag{}, nil
	}

	setID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid flashcard set id")
	}

	tags, err := r.TagRepo.GetContentTags(ctx, "flashcard_set", setID)
	if err != nil {
		return nil, err
	}

	return mapRepositoryTags(tags), nil
}
