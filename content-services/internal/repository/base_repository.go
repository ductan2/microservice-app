package repository

import (
	"context"

	"github.com/google/uuid"
)

// BaseRepository defines common CRUD operations that all repositories should implement
type BaseRepository[T any] interface {
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id uuid.UUID) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ListRepository extends BaseRepository with list/filter operations
type ListRepository[T any] interface {
	BaseRepository[T]
	List(ctx context.Context, filter any, sort *SortOption, limit, offset int) ([]T, int64, error)
}

// SoftDeleteRepository extends BaseRepository with soft delete functionality
type SoftDeleteRepository[T any] interface {
	BaseRepository[T]
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error
}

// FilterableRepository defines common filter types
type FilterableRepository interface {
	GetSortOption(sort interface{}) *SortOption
	GetFilter(filter interface{}) (interface{}, error)
}

// RepositoryOption defines common repository configuration options
type RepositoryOption struct {
	CollectionName string
	DatabaseName   string
	MaxConnections int
	Timeout        int
}

// DefaultRepositoryOptions returns sensible defaults for repository configuration
func DefaultRepositoryOptions() *RepositoryOption {
	return &RepositoryOption{
		CollectionName: "items",
		DatabaseName:   "content_services",
		MaxConnections: 10,
		Timeout:        30,
	}
}