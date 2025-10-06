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
	// ErrFolderNotFound is returned when no folder matches the lookup criteria.
	ErrFolderNotFound = errors.New("folder: not found")
	// ErrMaxDepthExceeded is returned when trying to create folder deeper than max depth
	ErrMaxDepthExceeded = errors.New("folder: maximum depth of 3 exceeded")
	// ErrFolderHasChildren is returned when trying to delete folder with children
	ErrFolderHasChildren = errors.New("folder: cannot delete folder with children")
	// ErrFolderHasMedia is returned when trying to delete folder with media assets
	ErrFolderHasMedia = errors.New("folder: cannot delete folder with media assets")
)

type FolderRepository interface {
	Create(ctx context.Context, folder *models.Folder) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Folder, error)
	GetByParentID(ctx context.Context, parentID *uuid.UUID) ([]models.Folder, error)
	GetAncestors(ctx context.Context, folderID uuid.UUID) ([]models.Folder, error)
	List(ctx context.Context, filter *FolderFilter, sort *SortOption, limit, offset int) ([]models.Folder, int64, error)
	Update(ctx context.Context, folder *models.Folder) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountChildren(ctx context.Context, folderID uuid.UUID) (int64, error)
	CountMediaAssets(ctx context.Context, folderID uuid.UUID) (int64, error)
}

type FolderFilter struct {
	ParentID *uuid.UUID
	Depth    *int
	Search   string
}

type folderRepository struct {
	collection      *mongo.Collection
	mediaCollection *mongo.Collection
}

func NewFolderRepository(db *mongo.Database) FolderRepository {
	return &folderRepository{
		collection:      db.Collection("folders"),
		mediaCollection: db.Collection("media_assets"),
	}
}

func (r *folderRepository) Create(ctx context.Context, folder *models.Folder) error {
	if folder == nil {
		return errors.New("folder: nil folder")
	}
	if folder.ID == uuid.Nil {
		folder.ID = uuid.New()
	}
	now := time.Now().UTC()
	if folder.CreatedAt.IsZero() {
		folder.CreatedAt = now
	}
	if folder.UpdatedAt.IsZero() {
		folder.UpdatedAt = now
	}

	// Validate depth
	if folder.Depth < 1 || folder.Depth > 3 {
		return ErrMaxDepthExceeded
	}

	// If has parent, verify parent exists and depth is valid
	if folder.ParentID != nil {
		parent, err := r.GetByID(ctx, *folder.ParentID)
		if err != nil {
			return err
		}
		// Parent depth + 1 should equal this folder's depth
		if parent.Depth+1 != folder.Depth {
			return errors.New("folder: invalid depth for parent")
		}
		if folder.Depth > 3 {
			return ErrMaxDepthExceeded
		}
	} else {
		// Root folder must have depth 1
		if folder.Depth != 1 {
			folder.Depth = 1
		}
	}

	_, err := r.collection.InsertOne(ctx, folder)
	return err
}

func (r *folderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Folder, error) {
	var folder models.Folder
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&folder)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrFolderNotFound
		}
		return nil, err
	}
	return &folder, nil
}

func (r *folderRepository) GetByParentID(ctx context.Context, parentID *uuid.UUID) ([]models.Folder, error) {
	filter := bson.M{}
	if parentID == nil {
		filter["parent_id"] = bson.M{"$exists": false}
	} else {
		filter["parent_id"] = *parentID
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var folders []models.Folder
	if err := cursor.All(ctx, &folders); err != nil {
		return nil, err
	}

	if folders == nil {
		folders = []models.Folder{}
	}
	return folders, nil
}

func (r *folderRepository) GetAncestors(ctx context.Context, folderID uuid.UUID) ([]models.Folder, error) {
	folder, err := r.GetByID(ctx, folderID)
	if err != nil {
		return nil, err
	}

	var ancestors []models.Folder
	current := folder

	// Traverse up the tree
	for current.ParentID != nil {
		parent, err := r.GetByID(ctx, *current.ParentID)
		if err != nil {
			return nil, err
		}
		ancestors = append([]models.Folder{*parent}, ancestors...) // Prepend to maintain order
		current = parent
	}

	return ancestors, nil
}

func (r *folderRepository) List(ctx context.Context, filter *FolderFilter, sort *SortOption, limit, offset int) ([]models.Folder, int64, error) {
	filterDoc := bson.M{}
	if filter != nil {
		if filter.ParentID != nil {
			filterDoc["parent_id"] = *filter.ParentID
		}
		if filter.Depth != nil {
			filterDoc["depth"] = *filter.Depth
		}
		if filter.Search != "" {
			filterDoc["name"] = bson.M{"$regex": filter.Search, "$options": "i"}
		}
	}

	opts := options.Find()
	sortField, sortDir := "name", SortAscending
	if sort != nil {
		sortField, sortDir = sort.apply(sortField, sortDir)
	}
	if sortField != "name" && sortField != "created_at" {
		sortField = "name"
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

	var folders []models.Folder
	if err := cursor.All(ctx, &folders); err != nil {
		return nil, 0, err
	}

	total, err := r.collection.CountDocuments(ctx, filterDoc)
	if err != nil {
		return nil, 0, err
	}

	if folders == nil {
		folders = []models.Folder{}
	}

	return folders, total, nil
}

func (r *folderRepository) Update(ctx context.Context, folder *models.Folder) error {
	if folder == nil || folder.ID == uuid.Nil {
		return errors.New("folder: invalid folder for update")
	}

	folder.UpdatedAt = time.Now().UTC()

	update := bson.M{
		"$set": bson.M{
			"name":       folder.Name,
			"updated_at": folder.UpdatedAt,
		},
	}

	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": folder.ID}, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrFolderNotFound
	}
	return nil
}

func (r *folderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if folder has children
	childrenCount, err := r.CountChildren(ctx, id)
	if err != nil {
		return err
	}
	if childrenCount > 0 {
		return ErrFolderHasChildren
	}

	// Check if folder has media assets
	mediaCount, err := r.CountMediaAssets(ctx, id)
	if err != nil {
		return err
	}
	if mediaCount > 0 {
		return ErrFolderHasMedia
	}

	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrFolderNotFound
	}
	return nil
}

func (r *folderRepository) CountChildren(ctx context.Context, folderID uuid.UUID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"parent_id": folderID})
}

func (r *folderRepository) CountMediaAssets(ctx context.Context, folderID uuid.UUID) (int64, error) {
	return r.mediaCollection.CountDocuments(ctx, bson.M{"folder_id": folderID})
}
