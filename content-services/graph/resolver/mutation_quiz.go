package resolver

import (
	"content-services/graph/model"
	"content-services/internal/models"
	"context"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// CreateQuiz is the resolver for the createQuiz field.
func (r *mutationResolver) CreateQuiz(ctx context.Context, input model.CreateQuizInput) (*model.Quiz, error) {
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return nil, gqlerror.Errorf("quiz service not configured")
	}

	quiz := &models.Quiz{
		Title: input.Title,
	}

	if input.Description != nil {
		quiz.Description = *input.Description
	}

	if input.LessonID != nil {
		lessonID, err := uuid.Parse(*input.LessonID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid lesson ID: %v", err)
		}
		quiz.LessonID = &lessonID
	}

	if input.TimeLimitS != nil {
		quiz.TimeLimitS = *input.TimeLimitS
	}

	created, err := quizService.CreateQuiz(ctx, quiz, nil)
	if err != nil {
		return nil, err
	}

	return mapQuiz(created), nil
}

// AddQuizQuestion is the resolver for the addQuizQuestion field.
func (r *mutationResolver) AddQuizQuestion(ctx context.Context, quizID string, input model.CreateQuizQuestionInput) (*model.QuizQuestion, error) {
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return nil, gqlerror.Errorf("quiz service not configured")
	}

	id, err := uuid.Parse(quizID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid quiz ID: %v", err)
	}

	question := &models.QuizQuestion{
		Type:   input.Type,
		Prompt: input.Prompt,
	}

	if input.Points != nil {
		question.Points = *input.Points
	}

	if input.PromptMedia != nil {
		mediaID, err := uuid.Parse(*input.PromptMedia)
		if err != nil {
			return nil, gqlerror.Errorf("invalid media ID: %v", err)
		}
		question.PromptMedia = &mediaID
	}

	if input.Metadata != nil {
		question.Metadata = input.Metadata
	}

	created, err := quizService.AddQuestion(ctx, id, question)
	if err != nil {
		return nil, err
	}

	return mapQuizQuestion(created), nil
}
