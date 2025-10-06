package service

import (
	"content-services/internal/models"
	"content-services/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	MaxFolderDepth = 3
)

type FolderService interface {
	CreateFolder(ctx context.Context, name string, parentID *uuid.UUID) (*models.Folder, error)
	GetFolder(ctx context.Context, id uuid.UUID) (*models.Folder, error)
	GetFolderWithChildren(ctx context.Context, id uuid.UUID) (*FolderTree, error)
	GetRootFolders(ctx context.Context) ([]models.Folder, error)
	GetSubfolders(ctx context.Context, parentID uuid.UUID) ([]models.Folder, error)
	GetFolderPath(ctx context.Context, folderID uuid.UUID) ([]models.Folder, error)
	ListFolders(ctx context.Context, filter *repository.FolderFilter, sort *repository.SortOption, page, pageSize int) ([]models.Folder, int64, error)
	UpdateFolder(ctx context.Context, id uuid.UUID, name string) (*models.Folder, error)
	DeleteFolder(ctx context.Context, id uuid.UUID) error
	ValidateDepth(ctx context.Context, parentID *uuid.UUID) error
	CountChildren(ctx context.Context, folderID uuid.UUID) (int64, error)
	CountMediaAssets(ctx context.Context, folderID uuid.UUID) (int64, error)
}

type FolderTree struct {
	Folder   models.Folder
	Children []FolderTree
}

type folderService struct {
	folderRepo repository.FolderRepository
}

func NewFolderService(folderRepo repository.FolderRepository) FolderService {
	return &folderService{
		folderRepo: folderRepo,
	}
}

func (s *folderService) CreateFolder(ctx context.Context, name string, parentID *uuid.UUID) (*models.Folder, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("folder: name is required")
	}

	depth := 1
	if parentID != nil {
		// Verify parent exists
		parent, err := s.folderRepo.GetByID(ctx, *parentID)
		if err != nil {
			if errors.Is(err, repository.ErrFolderNotFound) {
				return nil, errors.New("folder: parent folder not found")
			}
			return nil, err
		}

		// Calculate and validate depth
		depth = parent.Depth + 1
		if depth > MaxFolderDepth {
			return nil, fmt.Errorf("folder: cannot create folder at depth %d (max depth is %d)", depth, MaxFolderDepth)
		}
	}

	folder := &models.Folder{
		Name:     name,
		ParentID: parentID,
		Depth:    depth,
	}

	if err := s.folderRepo.Create(ctx, folder); err != nil {
		return nil, err
	}

	return folder, nil
}

func (s *folderService) GetFolder(ctx context.Context, id uuid.UUID) (*models.Folder, error) {
	return s.folderRepo.GetByID(ctx, id)
}

func (s *folderService) GetFolderWithChildren(ctx context.Context, id uuid.UUID) (*FolderTree, error) {
	folder, err := s.folderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tree := &FolderTree{
		Folder:   *folder,
		Children: []FolderTree{},
	}

	// Recursively get children only if not at max depth
	if folder.Depth < MaxFolderDepth {
		children, err := s.folderRepo.GetByParentID(ctx, &id)
		if err != nil {
			return nil, err
		}

		for _, child := range children {
			childTree, err := s.GetFolderWithChildren(ctx, child.ID)
			if err != nil {
				return nil, err
			}
			tree.Children = append(tree.Children, *childTree)
		}
	}

	return tree, nil
}

func (s *folderService) GetRootFolders(ctx context.Context) ([]models.Folder, error) {
	return s.folderRepo.GetByParentID(ctx, nil)
}

func (s *folderService) GetSubfolders(ctx context.Context, parentID uuid.UUID) ([]models.Folder, error) {
	return s.folderRepo.GetByParentID(ctx, &parentID)
}

func (s *folderService) GetFolderPath(ctx context.Context, folderID uuid.UUID) ([]models.Folder, error) {
	// Get the folder itself
	folder, err := s.folderRepo.GetByID(ctx, folderID)
	if err != nil {
		return nil, err
	}

	// Get ancestors
	ancestors, err := s.folderRepo.GetAncestors(ctx, folderID)
	if err != nil {
		return nil, err
	}

	// Combine ancestors + current folder to form the full path
	path := append(ancestors, *folder)
	return path, nil
}

func (s *folderService) ListFolders(ctx context.Context, filter *repository.FolderFilter, sort *repository.SortOption, page, pageSize int) ([]models.Folder, int64, error) {
	if page < 1 {
		page = 1
	}
	limit := pageSize
	if pageSize < 1 {
		limit = 20
	} else if pageSize > 200 {
		limit = 200
	}
	offset := (page - 1) * limit
	return s.folderRepo.List(ctx, filter, sort, limit, offset)
}

func (s *folderService) UpdateFolder(ctx context.Context, id uuid.UUID, name string) (*models.Folder, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("folder: name is required")
	}

	folder, err := s.folderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	folder.Name = name
	if err := s.folderRepo.Update(ctx, folder); err != nil {
		return nil, err
	}

	return folder, nil
}

func (s *folderService) DeleteFolder(ctx context.Context, id uuid.UUID) error {
	return s.folderRepo.Delete(ctx, id)
}

func (s *folderService) ValidateDepth(ctx context.Context, parentID *uuid.UUID) error {
	if parentID == nil {
		return nil // Root folder is always valid
	}

	parent, err := s.folderRepo.GetByID(ctx, *parentID)
	if err != nil {
		return err
	}

	if parent.Depth >= MaxFolderDepth {
		return fmt.Errorf("folder: cannot create subfolder (parent is at maximum depth %d)", MaxFolderDepth)
	}

	return nil
}

func (s *folderService) CountChildren(ctx context.Context, folderID uuid.UUID) (int64, error) {
	return s.folderRepo.CountChildren(ctx, folderID)
}

func (s *folderService) CountMediaAssets(ctx context.Context, folderID uuid.UUID) (int64, error) {
	return s.folderRepo.CountMediaAssets(ctx, folderID)
}
