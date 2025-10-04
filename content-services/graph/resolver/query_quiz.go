package resolver

import (
	"content-services/graph/model"
	"content-services/internal/repository"
	"context"
	"strings"

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
func (r *queryResolver) Quizzes(ctx context.Context, lessonID *string, search *string, page *int, pageSize *int, orderBy *model.QuizOrderInput) (*model.QuizCollection, error) {
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

	quizFilter := &repository.QuizFilter{LessonID: lessonUUID}
	if search != nil {
		quizFilter.Search = strings.TrimSpace(*search)
	}
	quizSort := buildQuizOrder(orderBy)

	pg := 1
	if page != nil && *page > 0 {
		pg = *page
	}

	ps := 20
	if pageSize != nil && *pageSize > 0 {
		ps = *pageSize
	}

	quizzes, total, err := quizService.ListQuizzes(ctx, quizFilter, quizSort, pg, ps)
	if err != nil {
		return nil, err
	}

	items := make([]*model.Quiz, 0, len(quizzes))
	for i := range quizzes {
		quiz := quizzes[i]
		items = append(items, mapQuiz(&quiz))
	}

	return &model.QuizCollection{
		Items:      items,
		TotalCount: int(total),
		Page:       pg,
		PageSize:   ps,
	}, nil
}

// QuizQuestions is the resolver for the quizQuestions field.
func (r *queryResolver) QuizQuestions(ctx context.Context, quizID string, filter *model.QuizQuestionFilterInput, page *int, pageSize *int, orderBy *model.QuizQuestionOrderInput) (*model.QuizQuestionCollection, error) {
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return nil, gqlerror.Errorf("quiz service not configured")
	}

	id, err := uuid.Parse(quizID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid quiz ID: %v", err)
	}

	questionFilter := buildQuizQuestionFilter(filter)
	questionSort := buildQuizQuestionOrder(orderBy)

	pg := 1
	if page != nil && *page > 0 {
		pg = *page
	}

	ps := 20
	if pageSize != nil && *pageSize > 0 {
		ps = *pageSize
	}

	questions, total, err := quizService.ListQuizQuestions(ctx, id, questionFilter, questionSort, pg, ps)
	if err != nil {
		return nil, err
	}

	items := make([]*model.QuizQuestion, 0, len(questions))
	for i := range questions {
		items = append(items, mapQuizQuestion(&questions[i]))
	}

	return &model.QuizQuestionCollection{
		Items:      items,
		TotalCount: int(total),
		Page:       pg,
		PageSize:   ps,
	}, nil
}
