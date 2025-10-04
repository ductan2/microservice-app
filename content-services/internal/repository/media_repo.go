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
	// ErrMediaNotFound is returned when no media asset matches the lookup criteria.
	ErrMediaNotFound = errors.New("media: not found")
)

type MediaRepository interface {
	Create(ctx context.Context, media *models.MediaAsset) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.MediaAsset, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.MediaAsset, error)
	GetBySHA256(ctx context.Context, sha256 string) (*models.MediaAsset, error)
	List(ctx context.Context, filter *MediaFilter, sort *SortOption, limit, offset int) ([]models.MediaAsset, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type MediaFilter struct {
	FolderID   *uuid.UUID
	Kind       string
	UploadedBy *uuid.UUID
	SHA256     string
	Search     string
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

func (r *mediaRepository) List(ctx context.Context, filter *MediaFilter, sort *SortOption, limit, offset int) ([]models.MediaAsset, int64, error) {
	filterDoc := bson.M{}
	if filter != nil {
		if filter.FolderID != nil {
			filterDoc["folder_id"] = *filter.FolderID
		}
		if filter.Kind != "" {
			filterDoc["kind"] = filter.Kind
		}
		if filter.UploadedBy != nil {
			filterDoc["uploaded_by"] = *filter.UploadedBy
		}
		if filter.SHA256 != "" {
			filterDoc["sha256"] = filter.SHA256
		}
		if filter.Search != "" {
			filterDoc["original_name"] = bson.M{"$regex": filter.Search, "$options": "i"}
		}
	}

	opts := options.Find()
	sortField, sortDir := "created_at", SortDescending
	if sort != nil {
		sortField, sortDir = sort.apply(sortField, sortDir)
	}
	if sortField != "created_at" && sortField != "bytes" {
		sortField = "created_at"
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

	var assets []models.MediaAsset
	for cursor.Next(ctx) {
		var asset models.MediaAsset
		if err := cursor.Decode(&asset); err != nil {
			return nil, 0, err
		}
		assets = append(assets, asset)
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	total, err := r.collection.CountDocuments(ctx, filterDoc)
	if err != nil {
		return nil, 0, err
	}

	if assets == nil {
		assets = []models.MediaAsset{}
	}

	return assets, total, nil
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
