package service

import (
	"content-services/internal/models"
	"content-services/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

type QuizService interface {
	// Quiz
	CreateQuiz(ctx context.Context, quiz *models.Quiz, tagIDs []uuid.UUID) (*models.Quiz, error)
	GetQuizByID(ctx context.Context, id uuid.UUID) (*models.Quiz, error)
	ListQuizzes(ctx context.Context, filter *repository.QuizFilter, sort *repository.SortOption, page, pageSize int) ([]models.Quiz, int64, error)
	UpdateQuiz(ctx context.Context, id uuid.UUID, updates *models.Quiz) (*models.Quiz, error)
	DeleteQuiz(ctx context.Context, id uuid.UUID) error

	// Questions
	AddQuestion(ctx context.Context, quizID uuid.UUID, question *models.QuizQuestion) (*models.QuizQuestion, error)
	UpdateQuestion(ctx context.Context, id uuid.UUID, updates *models.QuizQuestion) (*models.QuizQuestion, error)
	ReorderQuestions(ctx context.Context, quizID uuid.UUID, questionIDs []uuid.UUID) ([]models.QuizQuestion, error)
	DeleteQuestion(ctx context.Context, id uuid.UUID) error
	GetQuizQuestions(ctx context.Context, quizID uuid.UUID) ([]models.QuizQuestion, error)
	ListQuizQuestions(ctx context.Context, quizID uuid.UUID, filter *repository.QuizQuestionFilter, sort *repository.SortOption, page, pageSize int) ([]models.QuizQuestion, int64, error)

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
	if quiz == nil {
		return nil, errors.New("quiz: nil quiz")
	}

	if quiz.ID == uuid.Nil {
		quiz.ID = uuid.New()
	}

	now := time.Now().UTC()
	if quiz.CreatedAt.IsZero() {
		quiz.CreatedAt = now
	}

	if err := s.quizRepo.Create(ctx, quiz); err != nil {
		return nil, err
	}

	// TODO: handle tag assignments and outbox events when repositories are implemented

	return quiz, nil
}

func (s *quizService) GetQuizByID(ctx context.Context, id uuid.UUID) (*models.Quiz, error) {
	return s.quizRepo.GetByID(ctx, id)
}

func (s *quizService) ListQuizzes(ctx context.Context, filter *repository.QuizFilter, sort *repository.SortOption, page, pageSize int) ([]models.Quiz, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.quizRepo.List(ctx, filter, sort, pageSize, offset)
}

func (s *quizService) UpdateQuiz(ctx context.Context, id uuid.UUID, updates *models.Quiz) (*models.Quiz, error) {
	existing, err := s.quizRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if updates.Title != "" {
		existing.Title = updates.Title
	}
	if updates.Description != "" {
		existing.Description = updates.Description
	}
	if updates.LessonID != nil {
		existing.LessonID = updates.LessonID
	}
	if updates.TimeLimitS != 0 {
		existing.TimeLimitS = updates.TimeLimitS
	}
	if updates.TotalPoints != 0 {
		existing.TotalPoints = updates.TotalPoints
	}

	if err := s.quizRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *quizService) DeleteQuiz(ctx context.Context, id uuid.UUID) error {
	return s.quizRepo.Delete(ctx, id)
}

func (s *quizService) AddQuestion(ctx context.Context, quizID uuid.UUID, question *models.QuizQuestion) (*models.QuizQuestion, error) {
	if question == nil {
		return nil, errors.New("quiz question: nil question")
	}

	quiz, err := s.quizRepo.GetByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	question.QuizID = quizID
	if question.ID == uuid.Nil {
		question.ID = uuid.New()
	}
	if question.Metadata == nil {
		question.Metadata = map[string]any{}
	}
	if question.Points == 0 {
		question.Points = 1
	}

	if err := s.questionRepo.Create(ctx, question); err != nil {
		return nil, err
	}

	created, err := s.questionRepo.GetByID(ctx, question.ID)
	if err != nil {
		return nil, err
	}

	quiz.TotalPoints += created.Points
	if err := s.quizRepo.Update(ctx, quiz); err != nil {
		return nil, err
	}

	return created, nil
}

func (s *quizService) UpdateQuestion(ctx context.Context, id uuid.UUID, updates *models.QuizQuestion) (*models.QuizQuestion, error) {
	existing, err := s.questionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if updates.Prompt != "" {
		existing.Prompt = updates.Prompt
	}
	if updates.Type != "" {
		existing.Type = updates.Type
	}
	if updates.Points != 0 {
		delta := updates.Points - existing.Points
		existing.Points = updates.Points

		quiz, err := s.quizRepo.GetByID(ctx, existing.QuizID)
		if err == nil {
			quiz.TotalPoints += delta
			_ = s.quizRepo.Update(ctx, quiz)
		}
	}
	if updates.Metadata != nil {
		existing.Metadata = updates.Metadata
	}
	if updates.PromptMedia != nil {
		existing.PromptMedia = updates.PromptMedia
	}
	if updates.Ord != 0 {
		existing.Ord = updates.Ord
	}

	if err := s.questionRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *quizService) ReorderQuestions(ctx context.Context, quizID uuid.UUID, questionIDs []uuid.UUID) ([]models.QuizQuestion, error) {
	if err := s.questionRepo.Reorder(ctx, quizID, questionIDs); err != nil {
		return nil, err
	}
	questions, _, err := s.questionRepo.ListByQuizID(ctx, quizID, nil, &repository.SortOption{Field: "ord", Direction: repository.SortAscending}, 0, 0)
	return questions, err
}

func (s *quizService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	question, err := s.questionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}
		return err
	}

	if err := s.questionRepo.Delete(ctx, id); err != nil {
		return err
	}

	quiz, err := s.quizRepo.GetByID(ctx, question.QuizID)
	if err == nil {
		quiz.TotalPoints -= question.Points
		if quiz.TotalPoints < 0 {
			quiz.TotalPoints = 0
		}
		_ = s.quizRepo.Update(ctx, quiz)
	}

	return nil
}

func (s *quizService) GetQuizQuestions(ctx context.Context, quizID uuid.UUID) ([]models.QuizQuestion, error) {
	questions, _, err := s.questionRepo.ListByQuizID(ctx, quizID, nil, &repository.SortOption{Field: "ord", Direction: repository.SortAscending}, 0, 0)
	return questions, err
}

func (s *quizService) ListQuizQuestions(ctx context.Context, quizID uuid.UUID, filter *repository.QuizQuestionFilter, sort *repository.SortOption, page, pageSize int) ([]models.QuizQuestion, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.questionRepo.ListByQuizID(ctx, quizID, filter, sort, pageSize, offset)
}

func (s *quizService) AddOption(ctx context.Context, questionID uuid.UUID, option *models.QuestionOption) (*models.QuestionOption, error) {
	return nil, errors.New("question options not implemented")
}

func (s *quizService) UpdateOption(ctx context.Context, id uuid.UUID, updates *models.QuestionOption) (*models.QuestionOption, error) {
	return nil, errors.New("question options not implemented")
}

func (s *quizService) DeleteOption(ctx context.Context, id uuid.UUID) error {
	return errors.New("question options not implemented")
}

func (s *quizService) GetQuestionOptions(ctx context.Context, questionID uuid.UUID) ([]models.QuestionOption, error) {
	return nil, errors.New("question options not implemented")
}
