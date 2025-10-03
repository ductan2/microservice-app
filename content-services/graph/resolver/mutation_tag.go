package resolver

import (
	"content-services/graph/model"
	"context"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

// CreateTag is the resolver for the createTag field.
func (r *mutationResolver) CreateTag(ctx context.Context, input model.CreateTagInput) (*model.Tag, error) {
	if r.Taxonomy == nil {
		return nil, gqlerror.Errorf("taxonomy store not configured")
	}
	tag, err := r.Taxonomy.CreateTag(ctx, input.Slug, input.Name)
	if err != nil {
		return nil, mapTaxonomyError("tag", err)
	}
	return mapTag(tag), nil
}

// UpdateTag is the resolver for the updateTag field.
func (r *mutationResolver) UpdateTag(ctx context.Context, id string, input model.UpdateTagInput) (*model.Tag, error) {
	if r.Taxonomy == nil {
		return nil, gqlerror.Errorf("taxonomy store not configured")
	}
	tag, err := r.Taxonomy.UpdateTag(ctx, id, input.Slug, input.Name)
	if err != nil {
		return nil, mapTaxonomyError("tag", err)
	}
	return mapTag(tag), nil
}

// DeleteTag is the resolver for the deleteTag field.
func (r *mutationResolver) DeleteTag(ctx context.Context, id string) (bool, error) {
	if r.Taxonomy == nil {
		return false, gqlerror.Errorf("taxonomy store not configured")
	}
	if err := r.Taxonomy.DeleteTag(ctx, id); err != nil {
		return false, mapTaxonomyError("tag", err)
	}
	return true, nil
}
