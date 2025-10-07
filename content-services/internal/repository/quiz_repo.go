package repository

import (
	"content-services/internal/models"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type QuizRepository interface {
	Create(ctx context.Context, quiz *models.Quiz) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Quiz, error)
	List(ctx context.Context, filter *QuizFilter, sort *SortOption, limit, offset int) ([]models.Quiz, int64, error)
	Update(ctx context.Context, quiz *models.Quiz) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type QuizQuestionRepository interface {
	Create(ctx context.Context, question *models.QuizQuestion) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.QuizQuestion, error)
	ListByQuizID(ctx context.Context, quizID uuid.UUID, filter *QuizQuestionFilter, sort *SortOption, limit, offset int) ([]models.QuizQuestion, int64, error)
	Update(ctx context.Context, question *models.QuizQuestion) error
	Reorder(ctx context.Context, quizID uuid.UUID, questionIDs []uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type QuizFilter struct {
	LessonID *uuid.UUID
	TopicID  *uuid.UUID
	LevelID  *uuid.UUID
	Search   string
}

type QuizQuestionFilter struct {
	Type *string
}

type QuestionOptionRepository interface {
	Create(ctx context.Context, option *models.QuestionOption) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.QuestionOption, error)
	GetByQuestionID(ctx context.Context, questionID uuid.UUID) ([]models.QuestionOption, error)
	Update(ctx context.Context, option *models.QuestionOption) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type quizRepository struct {
	collection *mongo.Collection
}

type quizQuestionRepository struct {
	collection *mongo.Collection
}

type questionOptionRepository struct {
	collection *mongo.Collection
}

func NewQuizRepository(db *mongo.Database) QuizRepository {
	return &quizRepository{collection: db.Collection("quizzes")}
}

func NewQuizQuestionRepository(db *mongo.Database) QuizQuestionRepository {
	return &quizQuestionRepository{collection: db.Collection("quiz_questions")}
}

func NewQuestionOptionRepository(db *mongo.Database) QuestionOptionRepository {
	return &questionOptionRepository{collection: db.Collection("question_options")}
}

type quizDoc struct {
	ID          string    `bson:"_id"`
	LessonID    *string   `bson:"lesson_id,omitempty"`
	Title       string    `bson:"title"`
	Description string    `bson:"description"`
	TotalPoints int       `bson:"total_points"`
	TimeLimitS  int       `bson:"time_limit_s,omitempty"`
	CreatedAt   time.Time `bson:"created_at"`
	TopicID     *string   `bson:"topic_id,omitempty"`
	LevelID     *string   `bson:"level_id,omitempty"`
}

func quizDocFromModel(quiz *models.Quiz) *quizDoc {
	doc := &quizDoc{
		ID:          quiz.ID.String(),
		Title:       quiz.Title,
		Description: quiz.Description,
		TotalPoints: quiz.TotalPoints,
		TimeLimitS:  quiz.TimeLimitS,
		CreatedAt:   quiz.CreatedAt,
	}

	if quiz.LessonID != nil {
		lessonID := quiz.LessonID.String()
		doc.LessonID = &lessonID
	}

	if quiz.TopicID != nil {
		topicID := quiz.TopicID.String()
		doc.TopicID = &topicID
	}

	if quiz.LevelID != nil {
		levelID := quiz.LevelID.String()
		doc.LevelID = &levelID
	}

	return doc
}

func (d *quizDoc) toModel() *models.Quiz {
	quiz := &models.Quiz{
		ID:          uuid.MustParse(d.ID),
		Title:       d.Title,
		Description: d.Description,
		TotalPoints: d.TotalPoints,
		TimeLimitS:  d.TimeLimitS,
		CreatedAt:   d.CreatedAt,
	}

	if d.LessonID != nil && *d.LessonID != "" {
		id, err := uuid.Parse(*d.LessonID)
		if err == nil {
			quiz.LessonID = &id
		}
	}

	if d.TopicID != nil && *d.TopicID != "" {
		id, err := uuid.Parse(*d.TopicID)
		if err == nil {
			quiz.TopicID = &id
		}
	}

	if d.LevelID != nil && *d.LevelID != "" {
		id, err := uuid.Parse(*d.LevelID)
		if err == nil {
			quiz.LevelID = &id
		}
	}

	return quiz
}

type quizQuestionDoc struct {
	ID          string         `bson:"_id"`
	QuizID      string         `bson:"quiz_id"`
	Ord         int            `bson:"ord"`
	Type        string         `bson:"type"`
	Prompt      string         `bson:"prompt"`
	PromptMedia *string        `bson:"prompt_media,omitempty"`
	Points      int            `bson:"points"`
	Metadata    map[string]any `bson:"metadata"`
}

func quizQuestionDocFromModel(question *models.QuizQuestion) *quizQuestionDoc {
	doc := &quizQuestionDoc{
		ID:       question.ID.String(),
		QuizID:   question.QuizID.String(),
		Ord:      question.Ord,
		Type:     question.Type,
		Prompt:   question.Prompt,
		Points:   question.Points,
		Metadata: question.Metadata,
	}

	if question.PromptMedia != nil {
		mediaID := question.PromptMedia.String()
		doc.PromptMedia = &mediaID
	}

	return doc
}

func (d *quizQuestionDoc) toModel() *models.QuizQuestion {
	question := &models.QuizQuestion{
		ID:       uuid.MustParse(d.ID),
		QuizID:   uuid.MustParse(d.QuizID),
		Ord:      d.Ord,
		Type:     d.Type,
		Prompt:   d.Prompt,
		Points:   d.Points,
		Metadata: d.Metadata,
	}

	if d.PromptMedia != nil && *d.PromptMedia != "" {
		id, err := uuid.Parse(*d.PromptMedia)
		if err == nil {
			question.PromptMedia = &id
		}
	}

	return question
}

// Quiz implementations
func (r *quizRepository) Create(ctx context.Context, quiz *models.Quiz) error {
	if quiz == nil {
		return errors.New("quiz: nil quiz")
	}
	if quiz.ID == uuid.Nil {
		quiz.ID = uuid.New()
	}
	if quiz.CreatedAt.IsZero() {
		quiz.CreatedAt = time.Now().UTC()
	}

	_, err := r.collection.InsertOne(ctx, quizDocFromModel(quiz))
	return err
}

func (r *quizRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Quiz, error) {
	var doc quizDoc
	err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}
	return doc.toModel(), nil
}

func (r *quizRepository) List(ctx context.Context, filter *QuizFilter, sort *SortOption, limit, offset int) ([]models.Quiz, int64, error) {
	filterDoc := bson.M{}
	if filter != nil {
		if filter.LessonID != nil {
			filterDoc["lesson_id"] = filter.LessonID.String()
		}
		if filter.TopicID != nil {
			filterDoc["topic_id"] = filter.TopicID.String()
		}
		if filter.LevelID != nil {
			filterDoc["level_id"] = filter.LevelID.String()
		}
		if filter.Search != "" {
			regex := bson.M{"$regex": filter.Search, "$options": "i"}
			filterDoc["$or"] = []bson.M{
				{"title": regex},
				{"description": regex},
			}
		}
	}

	findOptions := options.Find()
	sortField, sortDir := "created_at", SortDescending
	if sort != nil {
		sortField, sortDir = sort.apply(sortField, sortDir)
	}
	if sortField != "created_at" && sortField != "total_points" {
		sortField = "created_at"
	}
	findOptions.SetSort(bson.D{{Key: sortField, Value: int(sortDir)}})
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}
	if offset > 0 {
		findOptions.SetSkip(int64(offset))
	}

	cursor, err := r.collection.Find(ctx, filterDoc, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var quizzes []models.Quiz
	for cursor.Next(ctx) {
		var doc quizDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, err
		}
		quizzes = append(quizzes, *doc.toModel())
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	total, err := r.collection.CountDocuments(ctx, filterDoc)
	if err != nil {
		return nil, 0, err
	}

	return quizzes, total, nil
}

func (r *quizRepository) Update(ctx context.Context, quiz *models.Quiz) error {
	if quiz == nil {
		return errors.New("quiz: nil quiz")
	}
	update := bson.M{
		"title":        quiz.Title,
		"description":  quiz.Description,
		"total_points": quiz.TotalPoints,
		"time_limit_s": quiz.TimeLimitS,
	}
	if quiz.LessonID != nil {
		update["lesson_id"] = quiz.LessonID.String()
	} else {
		update["lesson_id"] = nil
	}
	if quiz.TopicID != nil {
		update["topic_id"] = quiz.TopicID.String()
	} else {
		update["topic_id"] = nil
	}
	if quiz.LevelID != nil {
		update["level_id"] = quiz.LevelID.String()
	} else {
		update["level_id"] = nil
	}

	_, err := r.collection.UpdateByID(ctx, quiz.ID.String(), bson.M{"$set": update})
	return err
}

func (r *quizRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// QuizQuestion implementations
func (r *quizQuestionRepository) Create(ctx context.Context, question *models.QuizQuestion) error {
	if question == nil {
		return errors.New("quiz question: nil question")
	}
	if question.ID == uuid.Nil {
		question.ID = uuid.New()
	}
	if question.Metadata == nil {
		question.Metadata = map[string]any{}
	}

	if question.Ord == 0 {
		opts := options.FindOne().SetSort(bson.D{{Key: "ord", Value: -1}})
		var last quizQuestionDoc
		err := r.collection.FindOne(ctx, bson.M{"quiz_id": question.QuizID.String()}, opts).Decode(&last)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				question.Ord = 1
			} else {
				return err
			}
		} else {
			question.Ord = last.Ord + 1
		}
	}

	_, err := r.collection.InsertOne(ctx, quizQuestionDocFromModel(question))
	return err
}

func (r *quizQuestionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.QuizQuestion, error) {
	var doc quizQuestionDoc
	err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}
	return doc.toModel(), nil
}

func (r *quizQuestionRepository) ListByQuizID(ctx context.Context, quizID uuid.UUID, filter *QuizQuestionFilter, sort *SortOption, limit, offset int) ([]models.QuizQuestion, int64, error) {
	filterDoc := bson.M{"quiz_id": quizID.String()}
	if filter != nil && filter.Type != nil && *filter.Type != "" {
		filterDoc["type"] = *filter.Type
	}

	opts := options.Find()
	sortField, sortDir := "ord", SortAscending
	if sort != nil {
		sortField, sortDir = sort.apply(sortField, sortDir)
	}
	if sortField != "ord" && sortField != "points" {
		sortField = "ord"
	}
	opts.SetSort(bson.D{{Key: sortField, Value: int(sortDir)}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}

	cursor, err := r.collection.Find(ctx, filterDoc, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var questions []models.QuizQuestion
	for cursor.Next(ctx) {
		var doc quizQuestionDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, err
		}
		questions = append(questions, *doc.toModel())
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	total, err := r.collection.CountDocuments(ctx, filterDoc)
	if err != nil {
		return nil, 0, err
	}

	if questions == nil {
		questions = []models.QuizQuestion{}
	}

	return questions, total, nil
}

func (r *quizQuestionRepository) Update(ctx context.Context, question *models.QuizQuestion) error {
	if question == nil {
		return errors.New("quiz question: nil question")
	}

	update := bson.M{
		"ord":      question.Ord,
		"type":     question.Type,
		"prompt":   question.Prompt,
		"points":   question.Points,
		"metadata": question.Metadata,
	}
	if question.PromptMedia != nil {
		update["prompt_media"] = question.PromptMedia.String()
	} else {
		update["prompt_media"] = nil
	}

	_, err := r.collection.UpdateByID(ctx, question.ID.String(), bson.M{"$set": update})
	return err
}

func (r *quizQuestionRepository) Reorder(ctx context.Context, quizID uuid.UUID, questionIDs []uuid.UUID) error {
	for idx, id := range questionIDs {
		ord := idx + 1
		_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id.String(), "quiz_id": quizID.String()}, bson.M{"$set": bson.M{"ord": ord}})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *quizQuestionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// QuestionOption implementations
func (r *questionOptionRepository) Create(ctx context.Context, option *models.QuestionOption) error {
	return errors.New("question options not implemented")
}

func (r *questionOptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.QuestionOption, error) {
	return nil, errors.New("question options not implemented")
}

func (r *questionOptionRepository) GetByQuestionID(ctx context.Context, questionID uuid.UUID) ([]models.QuestionOption, error) {
	return nil, errors.New("question options not implemented")
}

func (r *questionOptionRepository) Update(ctx context.Context, option *models.QuestionOption) error {
	return errors.New("question options not implemented")
}

func (r *questionOptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return errors.New("question options not implemented")
}
