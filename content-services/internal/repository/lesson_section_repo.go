package repository

import (
	"content-services/internal/models"
	"content-services/internal/types"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LessonSectionRepository interface {
	Create(ctx context.Context, section *models.LessonSection) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.LessonSection, error)
	GetByLessonID(ctx context.Context, lessonID uuid.UUID) ([]models.LessonSection, error)
	Update(ctx context.Context, section *models.LessonSection) error
	Reorder(ctx context.Context, lessonID uuid.UUID, sectionIDs []uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type lessonSectionRepository struct {
	db *mongo.Database
}

func NewLessonSectionRepository(db *mongo.Database) LessonSectionRepository {
	return &lessonSectionRepository{db: db}
}

func (r *lessonSectionRepository) Create(ctx context.Context, section *models.LessonSection) error {
	if section == nil {
		return errors.New("lesson section is nil")
	}

	doc := sectionDocFromModel(section)
	_, err := r.collection().InsertOne(ctx, doc)
	return err
}

func (r *lessonSectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.LessonSection, error) {
	var doc lessonSectionDoc
	err := r.collection().FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, types.ErrLessonSectionNotFound
		}
		return nil, err
	}

	return doc.toModel(), nil
}

func (r *lessonSectionRepository) GetByLessonID(ctx context.Context, lessonID uuid.UUID) ([]models.LessonSection, error) {
	cursor, err := r.collection().Find(
		ctx,
		bson.M{"lesson_id": lessonID.String()},
		options.Find().SetSort(bson.D{{Key: "ord", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sections []models.LessonSection
	for cursor.Next(ctx) {
		var doc lessonSectionDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		sections = append(sections, *doc.toModel())
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if sections == nil {
		sections = []models.LessonSection{}
	}

	return sections, nil
}

func (r *lessonSectionRepository) Update(ctx context.Context, section *models.LessonSection) error {
	if section == nil {
		return errors.New("lesson section is nil")
	}

	updates := bson.M{}
	if section.Type != "" {
		updates["type"] = section.Type
	}
	if section.Body != nil {
		updates["body"] = section.Body
	}
	if section.Ord != 0 {
		updates["ord"] = section.Ord
	}
	if len(updates) == 0 {
		return nil
	}

	result, err := r.collection().UpdateOne(
		ctx,
		bson.M{"_id": section.ID.String()},
		bson.M{"$set": updates},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return types.ErrLessonSectionNotFound
	}

	return nil
}

func (r *lessonSectionRepository) Reorder(ctx context.Context, lessonID uuid.UUID, sectionIDs []uuid.UUID) error {
	for idx, sectionID := range sectionIDs {
		result, err := r.collection().UpdateOne(
			ctx,
			bson.M{"_id": sectionID.String(), "lesson_id": lessonID.String()},
			bson.M{"$set": bson.M{"ord": idx + 1}},
		)
		if err != nil {
			return err
		}
		if result.MatchedCount == 0 {
			return types.ErrLessonSectionNotFound
		}
	}
	return nil
}

func (r *lessonSectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection().DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return types.ErrLessonSectionNotFound
	}
	return nil
}

func (r *lessonSectionRepository) collection() *mongo.Collection {
	return r.db.Collection("lesson_sections")
}

type lessonSectionDoc struct {
	ID        string         `bson:"_id"`
	LessonID  string         `bson:"lesson_id"`
	Ord       int            `bson:"ord"`
	Type      string         `bson:"type"`
	Body      map[string]any `bson:"body"`
	CreatedAt time.Time      `bson:"created_at"`
}

func sectionDocFromModel(section *models.LessonSection) *lessonSectionDoc {
	doc := &lessonSectionDoc{
		ID:        section.ID.String(),
		LessonID:  section.LessonID.String(),
		Ord:       section.Ord,
		Type:      section.Type,
		Body:      section.Body,
		CreatedAt: section.CreatedAt,
	}

	if doc.Body == nil {
		doc.Body = map[string]any{}
	}

	return doc
}

func (d *lessonSectionDoc) toModel() *models.LessonSection {
	if d == nil {
		return nil
	}

	section := &models.LessonSection{
		ID:        uuid.MustParse(d.ID),
		LessonID:  uuid.MustParse(d.LessonID),
		Ord:       d.Ord,
		Type:      d.Type,
		Body:      d.Body,
		CreatedAt: d.CreatedAt,
	}

	if section.Body == nil {
		section.Body = map[string]any{}
	}

	return section
}
