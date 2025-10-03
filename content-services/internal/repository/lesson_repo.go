package repository

import (
	"content-services/internal/models"
	"content-services/internal/types"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LessonFilter struct {
	TopicID     *uuid.UUID
	LevelID     *uuid.UUID
	IsPublished *bool
	Search      string
}

type LessonRepository interface {
	Create(ctx context.Context, lesson *models.Lesson) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Lesson, error)
	GetByCode(ctx context.Context, code string) (*models.Lesson, error)
	List(ctx context.Context, filter *LessonFilter, limit, offset int) ([]models.Lesson, int64, error)
	Update(ctx context.Context, lesson *models.Lesson) error
	Publish(ctx context.Context, id uuid.UUID) error
	Unpublish(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type lessonRepository struct {
	collection *mongo.Collection
}

func NewLessonRepository(db *mongo.Database) LessonRepository {
	return &lessonRepository{
		collection: db.Collection("lessons"),
	}
}

// lessonDoc represents the MongoDB document structure
type lessonDoc struct {
	ID          string     `bson:"_id"`
	Code        *string    `bson:"code,omitempty"`
	Title       string     `bson:"title"`
	Description string     `bson:"description"`
	TopicID     *string    `bson:"topic_id,omitempty"`
	LevelID     *string    `bson:"level_id,omitempty"`
	IsPublished bool       `bson:"is_published"`
	Version     int        `bson:"version"`
	CreatedBy   *string    `bson:"created_by,omitempty"`
	CreatedAt   time.Time  `bson:"created_at"`
	UpdatedAt   time.Time  `bson:"updated_at"`
	PublishedAt *time.Time `bson:"published_at,omitempty"`
}

// toModel converts lessonDoc to models.Lesson
func (d *lessonDoc) toModel() *models.Lesson {
	lesson := &models.Lesson{
		ID:          uuid.MustParse(d.ID),
		Code:        strPtrOrEmpty(d.Code),
		Title:       d.Title,
		Description: d.Description,
		TopicID:     uuidPtrFromStr(d.TopicID),
		LevelID:     uuidPtrFromStr(d.LevelID),
		IsPublished: d.IsPublished,
		Version:     d.Version,
		CreatedBy:   uuidPtrFromStr(d.CreatedBy),
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}

	if d.PublishedAt != nil {
		lesson.PublishedAt = sql.NullTime{
			Time:  *d.PublishedAt,
			Valid: true,
		}
	}

	return lesson
}

// fromModel converts models.Lesson to lessonDoc
func fromModel(lesson *models.Lesson) *lessonDoc {
	doc := &lessonDoc{
		ID:          lesson.ID.String(),
		Title:       lesson.Title,
		Description: lesson.Description,
		IsPublished: lesson.IsPublished,
		Version:     lesson.Version,
		CreatedAt:   lesson.CreatedAt,
		UpdatedAt:   lesson.UpdatedAt,
	}

	if lesson.Code != "" {
		doc.Code = &lesson.Code
	}
	if lesson.TopicID != nil {
		s := lesson.TopicID.String()
		doc.TopicID = &s
	}
	if lesson.LevelID != nil {
		s := lesson.LevelID.String()
		doc.LevelID = &s
	}
	if lesson.CreatedBy != nil {
		s := lesson.CreatedBy.String()
		doc.CreatedBy = &s
	}
	if lesson.PublishedAt.Valid {
		doc.PublishedAt = &lesson.PublishedAt.Time
	}

	return doc
}

func strPtrOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func uuidPtrFromStr(s *string) *uuid.UUID {
	if s == nil || *s == "" {
		return nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return nil
	}
	return &id
}

func (r *lessonRepository) Create(ctx context.Context, lesson *models.Lesson) error {
	doc := fromModel(lesson)

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return types.ErrDuplicateCode
		}
		return err
	}

	return nil
}

func (r *lessonRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Lesson, error) {
	var doc lessonDoc
	err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, types.ErrLessonNotFound
		}
		return nil, err
	}

	return doc.toModel(), nil
}

func (r *lessonRepository) GetByCode(ctx context.Context, code string) (*models.Lesson, error) {
	var doc lessonDoc
	err := r.collection.FindOne(ctx, bson.M{"code": code}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, types.ErrLessonNotFound
		}
		return nil, err
	}

	return doc.toModel(), nil
}

func (r *lessonRepository) List(ctx context.Context, filter *LessonFilter, limit, offset int) ([]models.Lesson, int64, error) {
	// Build filter
	filterDoc := bson.M{}

	if filter != nil {
		if filter.TopicID != nil {
			filterDoc["topic_id"] = filter.TopicID.String()
		}
		if filter.LevelID != nil {
			filterDoc["level_id"] = filter.LevelID.String()
		}
		if filter.IsPublished != nil {
			filterDoc["is_published"] = *filter.IsPublished
		}
		if filter.Search != "" {
			filterDoc["$or"] = []bson.M{
				{"title": bson.M{"$regex": filter.Search, "$options": "i"}},
				{"description": bson.M{"$regex": filter.Search, "$options": "i"}},
				{"code": bson.M{"$regex": filter.Search, "$options": "i"}},
			}
		}
	}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filterDoc)
	if err != nil {
		return nil, 0, err
	}

	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filterDoc, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var docs []lessonDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, err
	}

	lessons := make([]models.Lesson, len(docs))
	for i, doc := range docs {
		lessons[i] = *doc.toModel()
	}

	return lessons, total, nil
}

func (r *lessonRepository) Update(ctx context.Context, lesson *models.Lesson) error {
	doc := fromModel(lesson)
	doc.UpdatedAt = time.Now().UTC()

	update := bson.M{
		"$set": doc,
		"$inc": bson.M{"version": 1},
	}

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": lesson.ID.String()},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	if err := result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return types.ErrLessonNotFound
		}
		return err
	}

	return nil
}

func (r *lessonRepository) Publish(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	update := bson.M{
		"$set": bson.M{
			"is_published": true,
			"published_at": now,
			"updated_at":   now,
		},
		"$inc": bson.M{"version": 1},
	}

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id.String(), "is_published": false},
		update,
	)

	if err := result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Check if lesson exists but is already published
			var doc lessonDoc
			checkErr := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
			if checkErr == nil && doc.IsPublished {
				return types.ErrAlreadyPublished
			}
			return types.ErrLessonNotFound
		}
		return err
	}

	return nil
}

func (r *lessonRepository) Unpublish(ctx context.Context, id uuid.UUID) error {
	update := bson.M{
		"$set": bson.M{
			"is_published": false,
			"updated_at":   time.Now().UTC(),
		},
		"$inc": bson.M{"version": 1},
	}

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id.String()},
		update,
	)

	if err := result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return types.ErrLessonNotFound
		}
		return err
	}

	return nil
}

func (r *lessonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return types.ErrLessonNotFound
	}

	return nil
}
