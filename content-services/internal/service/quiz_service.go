package service

import (
	"content-services/internal/models"
	"content-services/internal/repository"
	"context"

	"github.com/google/uuid"
)

type QuizService interface {
	// Quiz
	CreateQuiz(ctx context.Context, quiz *models.Quiz, tagIDs []uuid.UUID) (*models.Quiz, error)
	GetQuizByID(ctx context.Context, id uuid.UUID) (*models.Quiz, error)
	ListQuizzes(ctx context.Context, lessonID *uuid.UUID, page, pageSize int) ([]models.Quiz, int64, error)
	UpdateQuiz(ctx context.Context, id uuid.UUID, updates *models.Quiz) (*models.Quiz, error)
	DeleteQuiz(ctx context.Context, id uuid.UUID) error

	// Questions
	AddQuestion(ctx context.Context, quizID uuid.UUID, question *models.QuizQuestion) (*models.QuizQuestion, error)
	UpdateQuestion(ctx context.Context, id uuid.UUID, updates *models.QuizQuestion) (*models.QuizQuestion, error)
	ReorderQuestions(ctx context.Context, quizID uuid.UUID, questionIDs []uuid.UUID) ([]models.QuizQuestion, error)
	DeleteQuestion(ctx context.Context, id uuid.UUID) error
	GetQuizQuestions(ctx context.Context, quizID uuid.UUID) ([]models.QuizQuestion, error)

	// Options
	AddOption(ctx context.Context, questionID uuid.UUID, option *models.QuestionOption) (*models.QuestionOption, error)
	UpdateOption(ctx context.Context, id uuid.UUID, updates *models.QuestionOption) (*models.QuestionOption, error)
	DeleteOption(ctx context.Context, id uuid.UUID) error
	GetQuestionOptions(ctx context.Context, questionID uuid.UUID) ([]models.QuestionOption, error)
}

type quizService struct {
	quizRepo     repository.QuizRepository
	questionRepo repository.QuizQuestionRepository
	optionRepo   repository.QuestionOptionRepository
	tagRepo      repository.TagRepository
	outboxRepo   repository.OutboxRepository
}

func NewQuizService(
	quizRepo repository.QuizRepository,
	questionRepo repository.QuizQuestionRepository,
	optionRepo repository.QuestionOptionRepository,
	tagRepo repository.TagRepository,
	outboxRepo repository.OutboxRepository,
) QuizService {
	return &quizService{
		quizRepo:     quizRepo,
		questionRepo: questionRepo,
		optionRepo:   optionRepo,
		tagRepo:      tagRepo,
		outboxRepo:   outboxRepo,
	}
}

func (s *quizService) CreateQuiz(ctx context.Context, quiz *models.Quiz, tagIDs []uuid.UUID) (*models.Quiz, error) {
	// TODO: implement
	return nil, nil
}

func (s *quizService) GetQuizByID(ctx context.Context, id uuid.UUID) (*models.Quiz, error) {
	// TODO: implement
	return nil, nil
}

func (s *quizService) ListQuizzes(ctx context.Context, lessonID *uuid.UUID, page, pageSize int) ([]models.Quiz, int64, error) {
	// TODO: implement
	return nil, 0, nil
}

func (s *quizService) UpdateQuiz(ctx context.Context, id uuid.UUID, updates *models.Quiz) (*models.Quiz, error) {
	// TODO: implement
	return nil, nil
}

func (s *quizService) DeleteQuiz(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}

func (s *quizService) AddQuestion(ctx context.Context, quizID uuid.UUID, question *models.QuizQuestion) (*models.QuizQuestion, error) {
	// TODO: implement - update quiz total_points
	return nil, nil
}

func (s *quizService) UpdateQuestion(ctx context.Context, id uuid.UUID, updates *models.QuizQuestion) (*models.QuizQuestion, error) {
	// TODO: implement
	return nil, nil
}

func (s *quizService) ReorderQuestions(ctx context.Context, quizID uuid.UUID, questionIDs []uuid.UUID) ([]models.QuizQuestion, error) {
	// TODO: implement
	return nil, nil
}

func (s *quizService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}

func (s *quizService) GetQuizQuestions(ctx context.Context, quizID uuid.UUID) ([]models.QuizQuestion, error) {
	// TODO: implement
	return nil, nil
}

func (s *quizService) AddOption(ctx context.Context, questionID uuid.UUID, option *models.QuestionOption) (*models.QuestionOption, error) {
	// TODO: implement
	return nil, nil
}

func (s *quizService) UpdateOption(ctx context.Context, id uuid.UUID, updates *models.QuestionOption) (*models.QuestionOption, error) {
	// TODO: implement
	return nil, nil
}

func (s *quizService) DeleteOption(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}

func (s *quizService) GetQuestionOptions(ctx context.Context, questionID uuid.UUID) ([]models.QuestionOption, error) {
	// TODO: implement
	return nil, nil
}
