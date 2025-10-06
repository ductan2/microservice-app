package resolver

import (
	"content-services/graph/model"
	"content-services/internal/repository"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// CreateFolder is the resolver for the createFolder field.
func (r *mutationResolver) CreateFolder(ctx context.Context, input model.CreateFolderInput) (*model.Folder, error) {
	if r.FolderService == nil {
		return nil, gqlerror.Errorf("folder service not configured")
	}

	var parentID *uuid.UUID
	if input.ParentID != nil && *input.ParentID != "" {
		parsed, err := uuid.Parse(*input.ParentID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid parentId: %v", err)
		}
		parentID = &parsed
	}

	folder, err := r.FolderService.CreateFolder(ctx, input.Name, parentID)
	if err != nil {
		if errors.Is(err, repository.ErrMaxDepthExceeded) {
			return nil, gqlerror.Errorf("cannot create folder: maximum depth of 3 exceeded")
		}
		if errors.Is(err, repository.ErrFolderNotFound) {
			return nil, gqlerror.Errorf("parent folder not found")
		}
		return nil, err
	}

	return mapFolder(folder), nil
}

// UpdateFolder is the resolver for the updateFolder field.
func (r *mutationResolver) UpdateFolder(ctx context.Context, id string, input model.UpdateFolderInput) (*model.Folder, error) {
	if r.FolderService == nil {
		return nil, gqlerror.Errorf("folder service not configured")
	}

	folderID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid folder id: %v", err)
	}

	folder, err := r.FolderService.UpdateFolder(ctx, folderID, input.Name)
	if err != nil {
		if errors.Is(err, repository.ErrFolderNotFound) {
			return nil, gqlerror.Errorf("folder not found")
		}
		return nil, err
	}

	return mapFolder(folder), nil
}

// DeleteFolder is the resolver for the deleteFolder field.
func (r *mutationResolver) DeleteFolder(ctx context.Context, id string) (bool, error) {
	if r.FolderService == nil {
		return false, gqlerror.Errorf("folder service not configured")
	}

	folderID, err := uuid.Parse(id)
	if err != nil {
		return false, gqlerror.Errorf("invalid folder id: %v", err)
	}

	err = r.FolderService.DeleteFolder(ctx, folderID)
	if err != nil {
		if errors.Is(err, repository.ErrFolderNotFound) {
			return false, gqlerror.Errorf("folder not found")
		}
		if errors.Is(err, repository.ErrFolderHasChildren) {
			return false, gqlerror.Errorf("cannot delete folder: it contains subfolders")
		}
		if errors.Is(err, repository.ErrFolderHasMedia) {
			return false, gqlerror.Errorf("cannot delete folder: it contains media assets")
		}
		return false, err
	}

	return true, nil
}
