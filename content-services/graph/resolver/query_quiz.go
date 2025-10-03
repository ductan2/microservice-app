package resolver

import (
	"content-services/graph/model"
	"context"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Quiz is the resolver for the quiz field.
func (r *queryResolver) Quiz(ctx context.Context, id string) (*model.Quiz, error) {
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return nil, gqlerror.Errorf("quiz service not configured")
	}

	quizID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid quiz ID: %v", err)
	}

	quiz, err := quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	return mapQuiz(quiz), nil
}

// Quizzes is the resolver for the quizzes field.
func (r *queryResolver) Quizzes(ctx context.Context, lessonID *string, page *int, pageSize *int) (*model.QuizListResult, error) {
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return nil, gqlerror.Errorf("quiz service not configured")
	}

	var lessonUUID *uuid.UUID
	if lessonID != nil {
		parsed, err := uuid.Parse(*lessonID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid lesson ID: %v", err)
		}
		lessonUUID = &parsed
	}

	pg := 1
	if page != nil && *page > 0 {
		pg = *page
	}

	ps := 20
	if pageSize != nil && *pageSize > 0 {
		ps = *pageSize
	}

	quizzes, total, err := quizService.ListQuizzes(ctx, lessonUUID, pg, ps)
	if err != nil {
		return nil, err
	}

	items := make([]*model.Quiz, 0, len(quizzes))
	for i := range quizzes {
		quiz := quizzes[i]
		items = append(items, mapQuiz(&quiz))
	}

	return &model.QuizListResult{Items: items, TotalCount: int(total)}, nil
}
