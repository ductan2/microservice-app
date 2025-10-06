package resolver

import (
	"content-services/graph/model"
	"content-services/internal/repository"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Folder is the resolver for the folder field.
func (r *queryResolver) Folder(ctx context.Context, id string) (*model.Folder, error) {
	if r.FolderService == nil {
		return nil, gqlerror.Errorf("folder service not configured")
	}

	folderID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid folder id: %v", err)
	}

	folder, err := r.FolderService.GetFolder(ctx, folderID)
	if err != nil {
		if errors.Is(err, repository.ErrFolderNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mapFolder(folder), nil
}

// Folders is the resolver for the folders field.
func (r *queryResolver) Folders(ctx context.Context, filter *model.FolderFilterInput, page, pageSize *int, orderBy *model.FolderOrderInput) (*model.FolderCollection, error) {
	if r.FolderService == nil {
		return nil, gqlerror.Errorf("folder service not configured")
	}

	repoFilter, err := buildFolderFilter(filter)
	if err != nil {
		return nil, err
	}

	pageNum := 1
	if page != nil && *page > 0 {
		pageNum = *page
	}
	pageSizeNum := 100
	if pageSize != nil && *pageSize > 0 {
		pageSizeNum = *pageSize
	}

	sortOption := buildFolderOrder(orderBy)
	folders, total, err := r.FolderService.ListFolders(ctx, repoFilter, sortOption, pageNum, pageSizeNum)
	if err != nil {
		return nil, err
	}

	items := make([]*model.Folder, 0, len(folders))
	for i := range folders {
		items = append(items, mapFolder(&folders[i]))
	}

	return &model.FolderCollection{
		Items:      items,
		TotalCount: int(total),
		Page:       pageNum,
		PageSize:   pageSizeNum,
	}, nil
}

// RootFolders is the resolver for the rootFolders field.
func (r *queryResolver) RootFolders(ctx context.Context) ([]*model.Folder, error) {
	if r.FolderService == nil {
		return nil, gqlerror.Errorf("folder service not configured")
	}

	folders, err := r.FolderService.GetRootFolders(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Folder, 0, len(folders))
	for i := range folders {
		result = append(result, mapFolder(&folders[i]))
	}

	return result, nil
}

// FolderTree is the resolver for the folderTree field.
func (r *queryResolver) FolderTree(ctx context.Context, id string) (*model.FolderTree, error) {
	if r.FolderService == nil {
		return nil, gqlerror.Errorf("folder service not configured")
	}

	folderID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid folder id: %v", err)
	}

	tree, err := r.FolderService.GetFolderWithChildren(ctx, folderID)
	if err != nil {
		if errors.Is(err, repository.ErrFolderNotFound) {
			return nil, gqlerror.Errorf("folder not found")
		}
		return nil, err
	}

	return mapFolderTree(tree), nil
}

// FolderPath is the resolver for the folderPath field.
func (r *queryResolver) FolderPath(ctx context.Context, id string) ([]*model.Folder, error) {
	if r.FolderService == nil {
		return nil, gqlerror.Errorf("folder service not configured")
	}

	folderID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid folder id: %v", err)
	}

	path, err := r.FolderService.GetFolderPath(ctx, folderID)
	if err != nil {
		if errors.Is(err, repository.ErrFolderNotFound) {
			return nil, gqlerror.Errorf("folder not found")
		}
		return nil, err
	}

	result := make([]*model.Folder, 0, len(path))
	for i := range path {
		result = append(result, mapFolder(&path[i]))
	}

	return result, nil
}

// Helper functions

func buildFolderFilter(input *model.FolderFilterInput) (*repository.FolderFilter, error) {
	if input == nil {
		return nil, nil
	}

	filter := &repository.FolderFilter{}

	if input.ParentID != nil && *input.ParentID != "" {
		parentID, err := uuid.Parse(*input.ParentID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid parentId: %v", err)
		}
		filter.ParentID = &parentID
	}

	if input.Depth != nil {
		filter.Depth = input.Depth
	}

	if input.Search != nil {
		filter.Search = strings.TrimSpace(*input.Search)
	}

	return filter, nil
}

func buildFolderOrder(input *model.FolderOrderInput) *repository.SortOption {
	if input == nil {
		return nil
	}
	option := &repository.SortOption{Direction: mapOrderDirection(input.Direction)}
	switch input.Field {
	case model.FolderOrderFieldCreatedAt:
		option.Field = "created_at"
	default:
		option.Field = "name"
	}
	return option
}
