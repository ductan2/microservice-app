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

type LessonSectionFilter struct {
	Type *string
}

type LessonSectionRepository interface {
	Create(ctx context.Context, section *models.LessonSection) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.LessonSection, error)
	ListByLessonID(ctx context.Context, lessonID uuid.UUID, filter *LessonSectionFilter, sort *SortOption, limit, offset int) ([]models.LessonSection, int64, error)
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

func (r *lessonSectionRepository) ListByLessonID(ctx context.Context, lessonID uuid.UUID, filter *LessonSectionFilter, sort *SortOption, limit, offset int) ([]models.LessonSection, int64, error) {
	filterDoc := bson.M{"lesson_id": lessonID.String()}
	if filter != nil && filter.Type != nil && *filter.Type != "" {
		filterDoc["type"] = *filter.Type
	}

	opts := options.Find()
	sortField, sortDir := "ord", SortAscending
	if sort != nil {
		sortField, sortDir = sort.apply(sortField, sortDir)
	}
	opts.SetSort(bson.D{{Key: sortField, Value: int(sortDir)}})
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := r.collection().Find(ctx, filterDoc, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var sections []models.LessonSection
	for cursor.Next(ctx) {
		var doc lessonSectionDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, err
		}
		sections = append(sections, *doc.toModel())
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	total, err := r.collection().CountDocuments(ctx, filterDoc)
	if err != nil {
		return nil, 0, err
	}

	if sections == nil {
		sections = []models.LessonSection{}
	}

	return sections, total, nil
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
