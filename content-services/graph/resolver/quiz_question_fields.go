package resolver

import (
	"content-services/graph/generated"
	"content-services/graph/model"
	"context"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// QuizQuestion returns generated.QuizQuestionResolver implementation.
func (r *Resolver) QuizQuestion() generated.QuizQuestionResolver { return &quizQuestionResolver{r} }

type quizQuestionResolver struct{ *Resolver }

// Options is the resolver for the options field.
func (r *quizQuestionResolver) Options(ctx context.Context, obj *model.QuizQuestion) ([]*model.QuestionOption, error) {
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return nil, gqlerror.Errorf("quiz service not configured")
	}

	questionID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid question ID: %v", err)
	}

	options, err := quizService.GetQuestionOptions(ctx, questionID)
	if err != nil {
		return nil, err
	}

	return mapQuestionOptions(options), nil
}
