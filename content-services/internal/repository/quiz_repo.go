package repository

import (
	"content-services/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuizRepository interface {
	Create(ctx context.Context, quiz *models.Quiz) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Quiz, error)
	List(ctx context.Context, lessonID *uuid.UUID, limit, offset int) ([]models.Quiz, int64, error)
	Update(ctx context.Context, quiz *models.Quiz) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type QuizQuestionRepository interface {
	Create(ctx context.Context, question *models.QuizQuestion) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.QuizQuestion, error)
	GetByQuizID(ctx context.Context, quizID uuid.UUID) ([]models.QuizQuestion, error)
	Update(ctx context.Context, question *models.QuizQuestion) error
	Reorder(ctx context.Context, quizID uuid.UUID, questionIDs []uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type QuestionOptionRepository interface {
	Create(ctx context.Context, option *models.QuestionOption) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.QuestionOption, error)
	GetByQuestionID(ctx context.Context, questionID uuid.UUID) ([]models.QuestionOption, error)
	Update(ctx context.Context, option *models.QuestionOption) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type quizRepository struct {
	db *gorm.DB
}

type quizQuestionRepository struct {
	db *gorm.DB
}

type questionOptionRepository struct {
	db *gorm.DB
}

func NewQuizRepository(db *gorm.DB) QuizRepository {
	return &quizRepository{db: db}
}

func NewQuizQuestionRepository(db *gorm.DB) QuizQuestionRepository {
	return &quizQuestionRepository{db: db}
}

func NewQuestionOptionRepository(db *gorm.DB) QuestionOptionRepository {
	return &questionOptionRepository{db: db}
}

// Quiz implementations
func (r *quizRepository) Create(ctx context.Context, quiz *models.Quiz) error {
	// TODO: implement
	return nil
}

func (r *quizRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Quiz, error) {
	// TODO: implement
	return nil, nil
}

func (r *quizRepository) List(ctx context.Context, lessonID *uuid.UUID, limit, offset int) ([]models.Quiz, int64, error) {
	// TODO: implement with filters
	return nil, 0, nil
}

func (r *quizRepository) Update(ctx context.Context, quiz *models.Quiz) error {
	// TODO: implement
	return nil
}

func (r *quizRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}

// QuizQuestion implementations
func (r *quizQuestionRepository) Create(ctx context.Context, question *models.QuizQuestion) error {
	// TODO: implement - auto-increment ord
	return nil
}

func (r *quizQuestionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.QuizQuestion, error) {
	// TODO: implement
	return nil, nil
}

func (r *quizQuestionRepository) GetByQuizID(ctx context.Context, quizID uuid.UUID) ([]models.QuizQuestion, error) {
	// TODO: implement - order by ord
	return nil, nil
}

func (r *quizQuestionRepository) Update(ctx context.Context, question *models.QuizQuestion) error {
	// TODO: implement
	return nil
}

func (r *quizQuestionRepository) Reorder(ctx context.Context, quizID uuid.UUID, questionIDs []uuid.UUID) error {
	// TODO: implement
	return nil
}

func (r *quizQuestionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}

// QuestionOption implementations
func (r *questionOptionRepository) Create(ctx context.Context, option *models.QuestionOption) error {
	// TODO: implement - auto-increment ord
	return nil
}

func (r *questionOptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.QuestionOption, error) {
	// TODO: implement
	return nil, nil
}

func (r *questionOptionRepository) GetByQuestionID(ctx context.Context, questionID uuid.UUID) ([]models.QuestionOption, error) {
	// TODO: implement - order by ord
	return nil, nil
}

func (r *questionOptionRepository) Update(ctx context.Context, option *models.QuestionOption) error {
	// TODO: implement
	return nil
}

func (r *questionOptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: implement
	return nil
}
