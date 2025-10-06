package resolver

import (
	"content-services/graph/generated"
	"content-services/graph/model"
	"content-services/internal/models"
	"content-services/internal/service"
	"context"

	"github.com/google/uuid"
)

// Parent is the resolver for the parent field.
func (r *folderResolver) Parent(ctx context.Context, obj *model.Folder) (*model.Folder, error) {
	if obj.ParentID == nil {
		return nil, nil
	}

	if r.FolderService == nil {
		return nil, nil
	}

	parentID, err := uuid.Parse(*obj.ParentID)
	if err != nil {
		return nil, err
	}

	parent, err := r.FolderService.GetFolder(ctx, parentID)
	if err != nil {
		return nil, err
	}

	return mapFolder(parent), nil
}

// Children is the resolver for the children field.
func (r *folderResolver) Children(ctx context.Context, obj *model.Folder) ([]*model.Folder, error) {
	if r.FolderService == nil {
		return []*model.Folder{}, nil
	}

	// Don't fetch children if at max depth
	if obj.Depth >= 3 {
		return []*model.Folder{}, nil
	}

	folderID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, err
	}

	children, err := r.FolderService.GetSubfolders(ctx, folderID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Folder, 0, len(children))
	for i := range children {
		result = append(result, mapFolder(&children[i]))
	}

	return result, nil
}

// ChildrenCount is the resolver for the childrenCount field.
func (r *folderResolver) ChildrenCount(ctx context.Context, obj *model.Folder) (int, error) {
	if r.FolderService == nil {
		return 0, nil
	}

	folderID, err := uuid.Parse(obj.ID)
	if err != nil {
		return 0, err
	}

	count, err := r.FolderService.CountChildren(ctx, folderID)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// MediaCount is the resolver for the mediaCount field.
func (r *folderResolver) MediaCount(ctx context.Context, obj *model.Folder) (int, error) {
	if r.FolderService == nil {
		return 0, nil
	}

	folderID, err := uuid.Parse(obj.ID)
	if err != nil {
		return 0, err
	}

	count, err := r.FolderService.CountMediaAssets(ctx, folderID)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// Folder returns FolderResolver implementation.
func (r *Resolver) Folder() generated.FolderResolver { return &folderResolver{r} }

type folderResolver struct{ *Resolver }

// Helper mappers

func mapFolder(folder *models.Folder) *model.Folder {
	if folder == nil {
		return nil
	}

	var parentID *string
	if folder.ParentID != nil {
		pid := folder.ParentID.String()
		parentID = &pid
	}

	return &model.Folder{
		ID:        folder.ID.String(),
		Name:      folder.Name,
		ParentID:  parentID,
		Depth:     folder.Depth,
		CreatedAt: folder.CreatedAt,
		UpdatedAt: folder.UpdatedAt,
	}
}

func mapFolderTree(tree *service.FolderTree) *model.FolderTree {
	if tree == nil {
		return nil
	}

	children := make([]*model.FolderTree, 0, len(tree.Children))
	for i := range tree.Children {
		children = append(children, mapFolderTree(&tree.Children[i]))
	}

	return &model.FolderTree{
		Folder:   mapFolder(&tree.Folder),
		Children: children,
	}
}
