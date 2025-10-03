package resolver

import (
	"content-services/graph/generated"
	"content-services/graph/model"
	"context"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Quiz returns generated.QuizResolver implementation.
func (r *Resolver) Quiz() generated.QuizResolver { return &quizResolver{r} }

type quizResolver struct{ *Resolver }

// Questions is the resolver for the questions field.
func (r *quizResolver) Questions(ctx context.Context, obj *model.Quiz) ([]*model.QuizQuestion, error) {
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return nil, gqlerror.Errorf("quiz service not configured")
	}

	quizID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid quiz ID: %v", err)
	}

	questions, err := quizService.GetQuizQuestions(ctx, quizID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.QuizQuestion, 0, len(questions))
	for i := range questions {
		question := questions[i]
		result = append(result, mapQuizQuestion(&question))
	}

	return result, nil
}

// Tags is the resolver for the tags field.
func (r *quizResolver) Tags(ctx context.Context, obj *model.Quiz) ([]*model.Tag, error) {
	if r.TagRepo == nil {
		return []*model.Tag{}, nil
	}

	quizID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid quiz ID: %v", err)
	}

	tags, err := r.TagRepo.GetContentTags(ctx, "quiz", quizID)
	if err != nil {
		return nil, err
	}

	return mapRepositoryTags(tags), nil
}
