package resolver

import (
	"content-services/graph/model"
	"context"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

// CreateLevel is the resolver for the createLevel field.
func (r *mutationResolver) CreateLevel(ctx context.Context, input model.CreateLevelInput) (*model.Level, error) {
	if r.Taxonomy == nil {
		return nil, gqlerror.Errorf("taxonomy store not configured")
	}
	level, err := r.Taxonomy.CreateLevel(ctx, input.Code, input.Name)
	if err != nil {
		return nil, mapTaxonomyError("level", err)
	}
	return mapLevel(level), nil
}

// UpdateLevel is the resolver for the updateLevel field.
func (r *mutationResolver) UpdateLevel(ctx context.Context, id string, input model.UpdateLevelInput) (*model.Level, error) {
	if r.Taxonomy == nil {
		return nil, gqlerror.Errorf("taxonomy store not configured")
	}
	level, err := r.Taxonomy.UpdateLevel(ctx, id, input.Code, input.Name)
	if err != nil {
		return nil, mapTaxonomyError("level", err)
	}
	return mapLevel(level), nil
}

// DeleteLevel is the resolver for the deleteLevel field.
func (r *mutationResolver) DeleteLevel(ctx context.Context, id string) (bool, error) {
	if r.Taxonomy == nil {
		return false, gqlerror.Errorf("taxonomy store not configured")
	}
	if err := r.Taxonomy.DeleteLevel(ctx, id); err != nil {
		return false, mapTaxonomyError("level", err)
	}
	return true, nil
}
