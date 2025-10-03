package taxonomy

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// ErrNotFound is returned when a taxonomy document cannot be located.
	ErrNotFound = errors.New("taxonomy: not found")
	// ErrDuplicate is returned when attempting to create or update a document with a duplicate unique field.
	ErrDuplicate = errors.New("taxonomy: duplicate value")
)

// Store wraps Mongo collections that hold taxonomy data.
type Store struct {
	topics  *mongo.Collection
	levels  *mongo.Collection
	tags    *mongo.Collection
	lessons *mongo.Collection
}

// NewStore prepares a taxonomy store backed by Mongo collections and ensures indexes.
func NewStore(ctx context.Context, db *mongo.Database) (*Store, error) {
	store := &Store{
		topics:  db.Collection("topics"),
		levels:  db.Collection("levels"),
		tags:    db.Collection("tags"),
		lessons: db.Collection("lessons"),
	}

	if err := store.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) ensureIndexes(ctx context.Context) error {
	models := []struct {
		collection *mongo.Collection
		indexes    []mongo.IndexModel
	}{
		{
			collection: s.topics,
			indexes: []mongo.IndexModel{
				{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "created_at", Value: -1}}},
			},
		},
		{
			collection: s.levels,
			indexes: []mongo.IndexModel{
				{Keys: bson.D{{Key: "code", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "name", Value: 1}}},
			},
		},
		{
			collection: s.tags,
			indexes: []mongo.IndexModel{
				{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "name", Value: 1}}},
			},
		},
		{
			collection: s.lessons,
			indexes: []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "code", Value: 1}},
					Options: options.Index().SetUnique(true).SetSparse(true),
				},
				{
					Keys: bson.D{
						{Key: "topic_id", Value: 1},
						{Key: "level_id", Value: 1},
						{Key: "is_published", Value: 1},
					},
				},
				{Keys: bson.D{{Key: "created_at", Value: -1}}},
				{Keys: bson.D{{Key: "updated_at", Value: -1}}},
			},
		},
	}

	for _, model := range models {
		if len(model.indexes) == 0 {
			continue
		}
		if _, err := model.collection.Indexes().CreateMany(ctx, model.indexes); err != nil {
			if mongo.IsDuplicateKeyError(err) {
				// Index already exists, ignore.
				continue
			}
			return err
		}
	}
	return nil
}

// Topic represents a topic document stored in Mongo.
type Topic struct {
	ID        string    `bson:"_id"`
	Slug      string    `bson:"slug"`
	Name      string    `bson:"name"`
	CreatedAt time.Time `bson:"created_at"`
}

// Level represents a CEFR level document.
type Level struct {
	ID   string `bson:"_id"`
	Code string `bson:"code"`
	Name string `bson:"name"`
}

// Tag represents a flexible content tag document.
type Tag struct {
	ID   string `bson:"_id"`
	Slug string `bson:"slug"`
	Name string `bson:"name"`
}

// Lesson represents a lesson document.
type Lesson struct {
	ID          string    `bson:"_id"`
	Code        string    `bson:"code"`
	Title       string    `bson:"title"`
	Description string    `bson:"description"`
	TopicID     string    `bson:"topic_id"`
	LevelID     string    `bson:"level_id"`
	IsPublished bool      `bson:"is_published"`
	Version     int       `bson:"version"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}
type CreateLessonInput struct {
	Code        *string
	Title       string
	Description string
	TopicID     *string
	LevelID     *string
	CreatedBy   *string
}

// CreateTopic inserts a new topic document.
func (s *Store) CreateTopic(ctx context.Context, slug, name string) (*Topic, error) {
	doc := Topic{
		ID:        uuid.NewString(),
		Slug:      slug,
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}
	if _, err := s.topics.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return &doc, nil
}

// UpdateTopic updates an existing topic document and returns the new value.
func (s *Store) UpdateTopic(ctx context.Context, id string, slug, name *string) (*Topic, error) {
	update := bson.M{}
	if slug != nil {
		update["slug"] = *slug
	}
	if name != nil {
		update["name"] = *name
	}
	if len(update) == 0 {
		// Nothing to update, return current document if it exists.
		return s.GetTopicByID(ctx, id)
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	res := s.topics.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": update}, opts)
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	var doc Topic
	if err := res.Decode(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetTopicByID fetches a topic by its identifier.
func (s *Store) GetTopicByID(ctx context.Context, id string) (*Topic, error) {
	var doc Topic
	err := s.topics.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &doc, nil
}

// GetTopicBySlug fetches a topic by its slug.
func (s *Store) GetTopicBySlug(ctx context.Context, slug string) (*Topic, error) {
	var doc Topic
	err := s.topics.FindOne(ctx, bson.M{"slug": slug}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &doc, nil
}

// ListTopics returns all topics sorted by creation date (newest first).
func (s *Store) ListTopics(ctx context.Context) ([]Topic, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := s.topics.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var topics []Topic
	for cursor.Next(ctx) {
		var doc Topic
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		topics = append(topics, doc)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return topics, nil
}

// DeleteTopic removes a topic by its identifier.
func (s *Store) DeleteTopic(ctx context.Context, id string) error {
	res, err := s.topics.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateLevel inserts a new level document.
func (s *Store) CreateLevel(ctx context.Context, code, name string) (*Level, error) {
	doc := Level{
		ID:   uuid.NewString(),
		Code: code,
		Name: name,
	}
	if _, err := s.levels.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return &doc, nil
}

// UpdateLevel updates an existing level document.
func (s *Store) UpdateLevel(ctx context.Context, id string, code, name *string) (*Level, error) {
	update := bson.M{}
	if code != nil {
		update["code"] = *code
	}
	if name != nil {
		update["name"] = *name
	}
	if len(update) == 0 {
		return s.GetLevelByID(ctx, id)
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	res := s.levels.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": update}, opts)
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	var doc Level
	if err := res.Decode(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetLevelByID fetches a level document by ID.
func (s *Store) GetLevelByID(ctx context.Context, id string) (*Level, error) {
	var doc Level
	err := s.levels.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &doc, nil
}

// GetLevelByCode fetches a level by its code.
func (s *Store) GetLevelByCode(ctx context.Context, code string) (*Level, error) {
	var doc Level
	err := s.levels.FindOne(ctx, bson.M{"code": code}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &doc, nil
}

// ListLevels returns all levels sorted alphabetically by code.
func (s *Store) ListLevels(ctx context.Context) ([]Level, error) {
	opts := options.Find().SetSort(bson.D{{Key: "code", Value: 1}})
	cursor, err := s.levels.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var levels []Level
	for cursor.Next(ctx) {
		var doc Level
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		levels = append(levels, doc)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return levels, nil
}

// DeleteLevel removes a level by its identifier.
func (s *Store) DeleteLevel(ctx context.Context, id string) error {
	res, err := s.levels.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateTag inserts a new tag document.
func (s *Store) CreateTag(ctx context.Context, slug, name string) (*Tag, error) {
	doc := Tag{
		ID:   uuid.NewString(),
		Slug: slug,
		Name: name,
	}
	if _, err := s.tags.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return &doc, nil
}

// UpdateTag updates an existing tag document.
func (s *Store) UpdateTag(ctx context.Context, id string, slug, name *string) (*Tag, error) {
	update := bson.M{}
	if slug != nil {
		update["slug"] = *slug
	}
	if name != nil {
		update["name"] = *name
	}
	if len(update) == 0 {
		return s.GetTagByID(ctx, id)
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	res := s.tags.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": update}, opts)
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	var doc Tag
	if err := res.Decode(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetTagByID fetches a tag by ID.
func (s *Store) GetTagByID(ctx context.Context, id string) (*Tag, error) {
	var doc Tag
	err := s.tags.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &doc, nil
}

// GetTagBySlug fetches a tag by slug.
func (s *Store) GetTagBySlug(ctx context.Context, slug string) (*Tag, error) {
	var doc Tag
	err := s.tags.FindOne(ctx, bson.M{"slug": slug}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &doc, nil
}

// ListTags returns all tags sorted alphabetically by name.
func (s *Store) ListTags(ctx context.Context) ([]Tag, error) {
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := s.tags.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tags []Tag
	for cursor.Next(ctx) {
		var doc Tag
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		tags = append(tags, doc)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return tags, nil
}

// DeleteTag removes a tag by its identifier.
func (s *Store) DeleteTag(ctx context.Context, id string) error {
	res, err := s.tags.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
