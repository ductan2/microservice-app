package repository

import (
	"content-services/internal/models"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// ErrMediaNotFound is returned when no media asset matches the lookup criteria.
	ErrMediaNotFound = errors.New("media: not found")
)

type MediaRepository interface {
	Create(ctx context.Context, media *models.MediaAsset) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.MediaAsset, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.MediaAsset, error)
	GetBySHA256(ctx context.Context, sha256 string) (*models.MediaAsset, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type mediaRepository struct {
	collection *mongo.Collection
}

func NewMediaRepository(db *mongo.Database) MediaRepository {
	return &mediaRepository{
		collection: db.Collection("media_assets"),
	}
}

func (r *mediaRepository) Create(ctx context.Context, media *models.MediaAsset) error {
	if media == nil {
		return errors.New("media: nil asset")
	}
	if media.ID == uuid.Nil {
		media.ID = uuid.New()
	}
	if media.CreatedAt.IsZero() {
		media.CreatedAt = time.Now().UTC()
	}
	_, err := r.collection.InsertOne(ctx, media)
	return err
}

func (r *mediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.MediaAsset, error) {
	var media models.MediaAsset
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&media)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrMediaNotFound
		}
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.MediaAsset, error) {
	if len(ids) == 0 {
		return []models.MediaAsset{}, nil
	}
	cursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var assets []models.MediaAsset
	for cursor.Next(ctx) {
		var media models.MediaAsset
		if err := cursor.Decode(&media); err != nil {
			return nil, err
		}
		assets = append(assets, media)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return assets, nil
}

func (r *mediaRepository) GetBySHA256(ctx context.Context, sha256 string) (*models.MediaAsset, error) {
	var media models.MediaAsset
	err := r.collection.FindOne(ctx, bson.M{"sha256": sha256}).Decode(&media)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrMediaNotFound
		}
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrMediaNotFound
	}
	return nil
}
