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

var (
	// ErrFlashcardSetNotFound indicates a flashcard set lookup failed.
	ErrFlashcardSetNotFound = errors.New("flashcard set not found")
	// ErrFlashcardNotFound indicates a flashcard lookup failed.
	ErrFlashcardNotFound = errors.New("flashcard not found")
)

type FlashcardSetRepository interface {
	Create(ctx context.Context, set *models.FlashcardSet) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.FlashcardSet, error)
	List(ctx context.Context, filter *FlashcardSetFilter, sort *SortOption, limit, offset int) ([]models.FlashcardSet, int64, error)
	Update(ctx context.Context, set *models.FlashcardSet) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type FlashcardRepository interface {
	Create(ctx context.Context, card *models.Flashcard) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Flashcard, error)
	ListBySetID(ctx context.Context, setID uuid.UUID, filter *FlashcardFilter, sort *SortOption, limit, offset int) ([]models.Flashcard, int64, error)
	Update(ctx context.Context, card *models.Flashcard) error
	Reorder(ctx context.Context, setID uuid.UUID, cardIDs []uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type FlashcardSetFilter struct {
	TopicID   *uuid.UUID
	LevelID   *uuid.UUID
	CreatedBy *uuid.UUID
	Search    string
}

type FlashcardFilter struct {
	HasMedia *bool
}

type flashcardSetRepository struct {
	collection *mongo.Collection
}

type flashcardRepository struct {
	collection *mongo.Collection
}

func NewFlashcardSetRepository(db *mongo.Database) FlashcardSetRepository {
	return &flashcardSetRepository{collection: db.Collection("flashcard_sets")}
}

func NewFlashcardRepository(db *mongo.Database) FlashcardRepository {
	return &flashcardRepository{collection: db.Collection("flashcards")}
}

func (r *flashcardSetRepository) Create(ctx context.Context, set *models.FlashcardSet) error {
	if set.ID == uuid.Nil {
		set.ID = uuid.New()
	}
	if set.CreatedAt.IsZero() {
		set.CreatedAt = time.Now().UTC()
	}
	_, err := r.collection.InsertOne(ctx, set)
	return err
}

func (r *flashcardSetRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.FlashcardSet, error) {
	var set models.FlashcardSet
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&set)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrFlashcardSetNotFound
		}
		return nil, err
	}
	return &set, nil
}

func (r *flashcardSetRepository) List(ctx context.Context, filter *FlashcardSetFilter, sort *SortOption, limit, offset int) ([]models.FlashcardSet, int64, error) {
	match := bson.D{}
	if filter != nil {
		if filter.TopicID != nil {
			match = append(match, bson.E{Key: "topic_id", Value: *filter.TopicID})
		}
		if filter.LevelID != nil {
			match = append(match, bson.E{Key: "level_id", Value: *filter.LevelID})
		}
		if filter.CreatedBy != nil {
			match = append(match, bson.E{Key: "created_by", Value: *filter.CreatedBy})
		}
		if filter.Search != "" {
			regex := bson.D{{Key: "$regex", Value: filter.Search}, {Key: "$options", Value: "i"}}
			match = append(match, bson.E{Key: "$or", Value: bson.A{
				bson.D{{Key: "title", Value: regex}},
				bson.D{{Key: "description", Value: regex}},
			}})
		}
	}

	totalFilter := bson.D{}
	if len(match) > 0 {
		totalFilter = append(totalFilter, match...)
	}

	total, err := r.collection.CountDocuments(ctx, totalFilter)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []models.FlashcardSet{}, 0, nil
	}

	pipeline := mongo.Pipeline{}
	if len(match) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}

	lookupStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "flashcards"},
		{Key: "let", Value: bson.D{{Key: "setId", Value: "$_id"}}},
		{Key: "pipeline", Value: bson.A{
			bson.D{{Key: "$match", Value: bson.D{{Key: "$expr", Value: bson.D{{Key: "$eq", Value: bson.A{"$set_id", "$$setId"}}}}}}},
			bson.D{{Key: "$count", Value: "count"}},
		}},
		{Key: "as", Value: "card_counts"},
	}}}
	addFieldsStage := bson.D{{Key: "$addFields", Value: bson.D{{Key: "card_count", Value: bson.D{
		{Key: "$ifNull", Value: bson.A{
			bson.D{{Key: "$arrayElemAt", Value: bson.A{"$card_counts.count", 0}}},
			0,
		}},
	}}}}}
	projectStage := bson.D{{Key: "$project", Value: bson.D{{Key: "card_counts", Value: 0}}}}

	pipeline = append(pipeline, lookupStage, addFieldsStage, projectStage)

	sortField, sortDir := "created_at", SortDescending
	if sort != nil {
		sortField, sortDir = sort.apply(sortField, sortDir)
	}
	if sortField != "created_at" && sortField != "card_count" {
		sortField = "created_at"
	}
	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.D{{Key: sortField, Value: int(sortDir)}}}})
	if offset > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$skip", Value: int64(offset)}})
	}
	if limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: int64(limit)}})
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var sets []models.FlashcardSet
	for cursor.Next(ctx) {
		var set models.FlashcardSet
		if err := cursor.Decode(&set); err != nil {
			return nil, 0, err
		}
		sets = append(sets, set)
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	if sets == nil {
		sets = []models.FlashcardSet{}
	}

	return sets, total, nil
}

func (r *flashcardSetRepository) Update(ctx context.Context, set *models.FlashcardSet) error {
	update := bson.M{
		"$set": bson.M{
			"title":       set.Title,
			"description": set.Description,
			"topic_id":    set.TopicID,
			"level_id":    set.LevelID,
			"created_by":  set.CreatedBy,
		},
	}

	res, err := r.collection.UpdateByID(ctx, set.ID, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrFlashcardSetNotFound
	}
	return nil
}

func (r *flashcardSetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrFlashcardSetNotFound
	}
	return nil
}

func (r *flashcardRepository) Create(ctx context.Context, card *models.Flashcard) error {
	if card.ID == uuid.Nil {
		card.ID = uuid.New()
	}
	if card.CreatedAt.IsZero() {
		card.CreatedAt = time.Now().UTC()
	}
	if card.Ord == 0 {
		var last models.Flashcard
		err := r.collection.FindOne(ctx, bson.M{"set_id": card.SetID}, options.FindOne().SetSort(bson.D{{Key: "ord", Value: -1}})).Decode(&last)
		switch {
		case err == nil:
			card.Ord = last.Ord + 1
		case errors.Is(err, mongo.ErrNoDocuments):
			card.Ord = 1
		default:
			return err
		}
	}

	_, err := r.collection.InsertOne(ctx, card)
	return err
}

func (r *flashcardRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Flashcard, error) {
	var card models.Flashcard
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&card)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrFlashcardNotFound
		}
		return nil, err
	}
	return &card, nil
}

func (r *flashcardRepository) ListBySetID(ctx context.Context, setID uuid.UUID, filter *FlashcardFilter, sort *SortOption, limit, offset int) ([]models.Flashcard, int64, error) {
	filterDoc := bson.M{"set_id": setID}
	if filter != nil && filter.HasMedia != nil {
		hasMedia := *filter.HasMedia
		if hasMedia {
			filterDoc["$or"] = []bson.M{
				{"front_media_id": bson.M{"$exists": true}},
				{"back_media_id": bson.M{"$exists": true}},
			}
		} else {
			filterDoc["$and"] = []bson.M{
				{"$or": []bson.M{{"front_media_id": bson.M{"$exists": false}}, {"front_media_id": nil}}},
				{"$or": []bson.M{{"back_media_id": bson.M{"$exists": false}}, {"back_media_id": nil}}},
			}
		}
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

	cursor, err := r.collection.Find(ctx, filterDoc, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var cards []models.Flashcard
	for cursor.Next(ctx) {
		var card models.Flashcard
		if err := cursor.Decode(&card); err != nil {
			return nil, 0, err
		}
		cards = append(cards, card)
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	total, err := r.collection.CountDocuments(ctx, filterDoc)
	if err != nil {
		return nil, 0, err
	}

	if cards == nil {
		cards = []models.Flashcard{}
	}

	return cards, total, nil
}

func (r *flashcardRepository) Update(ctx context.Context, card *models.Flashcard) error {
	update := bson.M{
		"$set": bson.M{
			"front_text":     card.FrontText,
			"back_text":      card.BackText,
			"front_media_id": card.FrontMediaID,
			"back_media_id":  card.BackMediaID,
			"ord":            card.Ord,
			"hints":          card.Hints,
		},
	}

	res, err := r.collection.UpdateByID(ctx, card.ID, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrFlashcardNotFound
	}
	return nil
}

func (r *flashcardRepository) Reorder(ctx context.Context, setID uuid.UUID, cardIDs []uuid.UUID) error {
	for index, cardID := range cardIDs {
		res, err := r.collection.UpdateOne(ctx, bson.M{"_id": cardID, "set_id": setID}, bson.M{"$set": bson.M{"ord": index + 1}})
		if err != nil {
			return err
		}
		if res.MatchedCount == 0 {
			return ErrFlashcardNotFound
		}
	}
	return nil
}

func (r *flashcardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrFlashcardNotFound
	}
	return nil
}
