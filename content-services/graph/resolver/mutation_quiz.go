package resolver

import (
	"content-services/graph/model"
	"content-services/internal/models"
	"content-services/internal/service"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.mongodb.org/mongo-driver/mongo"
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

	if input.TopicID != nil {
		topicID, err := uuid.Parse(*input.TopicID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid topic ID: %v", err)
		}
		quiz.TopicID = &topicID
	}

	if input.LevelID != nil {
		levelID, err := uuid.Parse(*input.LevelID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid level ID: %v", err)
		}
		quiz.LevelID = &levelID
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

// UpdateQuiz is the resolver for the updateQuiz field.
func (r *mutationResolver) UpdateQuiz(ctx context.Context, id string, input model.UpdateQuizInput) (*model.Quiz, error) {
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return nil, gqlerror.Errorf("quiz service not configured")
	}

	quizID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid quiz ID: %v", err)
	}

	updates := &models.Quiz{}

	if input.Title != nil {
		updates.Title = *input.Title
	}

	if input.Description != nil {
		updates.Description = *input.Description
	}

	if input.LessonID != nil {
		if *input.LessonID == "" {
			nilID := uuid.Nil
			updates.LessonID = &nilID
		} else {
			lessonID, err := uuid.Parse(*input.LessonID)
			if err != nil {
				return nil, gqlerror.Errorf("invalid lesson ID: %v", err)
			}
			updates.LessonID = &lessonID
		}
	}

	if input.TopicID != nil {
		if *input.TopicID == "" {
			updates.TopicID = nil
		} else {
			topicID, err := uuid.Parse(*input.TopicID)
			if err != nil {
				return nil, gqlerror.Errorf("invalid topic ID: %v", err)
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
				return nil, gqlerror.Errorf("invalid level ID: %v", err)
			}
			updates.LevelID = &levelID
		}
	}

	if input.TimeLimitS != nil {
		updates.TimeLimitS = *input.TimeLimitS
	}

	updated, err := quizService.UpdateQuiz(ctx, quizID, updates)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, gqlerror.Errorf("quiz not found: %s", id)
		}
		return nil, err
	}

	return mapQuiz(updated), nil
}

// DeleteQuiz is the resolver for the deleteQuiz field.
func (r *mutationResolver) DeleteQuiz(ctx context.Context, id string) (bool, error) {
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return false, gqlerror.Errorf("quiz service not configured")
	}

	quizID, err := uuid.Parse(id)
	if err != nil {
		return false, gqlerror.Errorf("invalid quiz ID: %v", err)
	}

	if err := quizService.DeleteQuiz(ctx, quizID); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, gqlerror.Errorf("quiz not found: %s", id)
		}
		return false, err
	}

	return true, nil
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

// AddQuestionOption is the resolver for the addQuestionOption field.
func (r *mutationResolver) AddQuestionOption(ctx context.Context, questionID string, input model.CreateQuestionOptionInput) (*model.QuestionOption, error) {
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return nil, gqlerror.Errorf("quiz service not configured")
	}

	qID, err := uuid.Parse(questionID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid question ID: %v", err)
	}

	option := &models.QuestionOption{
		QuestionID: qID,
		Ord:        input.Ord,
		Label:      input.Label,
		IsCorrect:  input.IsCorrect,
	}

	if input.Feedback != nil {
		option.Feedback = *input.Feedback
	}

	created, err := quizService.AddOption(ctx, qID, option)
	if err != nil {
		return nil, err
	}

	return mapQuestionOption(created), nil
}

// UpdateQuestionOption is the resolver for the updateQuestionOption field.
func (r *mutationResolver) UpdateQuestionOption(ctx context.Context, id string, input model.UpdateQuestionOptionInput) (*model.QuestionOption, error) {
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return nil, gqlerror.Errorf("quiz service not configured")
	}

	optionID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid option ID: %v", err)
	}

	updates := &service.QuestionOptionUpdate{}

	if input.Ord != nil {
		updates.Ord = input.Ord
	}

	if input.Label != nil {
		updates.Label = input.Label
	}

	if input.IsCorrect != nil {
		updates.IsCorrect = input.IsCorrect
	}

	if input.Feedback != nil {
		updates.Feedback = input.Feedback
	}

	updated, err := quizService.UpdateOption(ctx, optionID, updates)
	if err != nil {
		return nil, err
	}

	return mapQuestionOption(updated), nil
}

// DeleteQuestionOption is the resolver for the deleteQuestionOption field.
func (r *mutationResolver) DeleteQuestionOption(ctx context.Context, id string) (bool, error) {
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return false, gqlerror.Errorf("quiz service not configured")
	}

	optionID, err := uuid.Parse(id)
	if err != nil {
		return false, gqlerror.Errorf("invalid option ID: %v", err)
	}

	if err := quizService.DeleteOption(ctx, optionID); err != nil {
		return false, err
	}

	return true, nil
}
