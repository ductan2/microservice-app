package resolver

import (
	"content-services/graph/model"
	"context"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

// CreateTopic is the resolver for the createTopic field.
func (r *mutationResolver) CreateTopic(ctx context.Context, input model.CreateTopicInput) (*model.Topic, error) {
	if r.Taxonomy == nil {
		return nil, gqlerror.Errorf("taxonomy store not configured")
	}
	topic, err := r.Taxonomy.CreateTopic(ctx, input.Slug, input.Name)
	if err != nil {
		return nil, mapTaxonomyError("topic", err)
	}
	return mapTopic(topic), nil
}

// UpdateTopic is the resolver for the updateTopic field.
func (r *mutationResolver) UpdateTopic(ctx context.Context, id string, input model.UpdateTopicInput) (*model.Topic, error) {
	if r.Taxonomy == nil {
		return nil, gqlerror.Errorf("taxonomy store not configured")
	}
	topic, err := r.Taxonomy.UpdateTopic(ctx, id, input.Slug, input.Name)
	if err != nil {
		return nil, mapTaxonomyError("topic", err)
	}
	return mapTopic(topic), nil
}

// DeleteTopic is the resolver for the deleteTopic field.
func (r *mutationResolver) DeleteTopic(ctx context.Context, id string) (bool, error) {
	if r.Taxonomy == nil {
		return false, gqlerror.Errorf("taxonomy store not configured")
	}
	if err := r.Taxonomy.DeleteTopic(ctx, id); err != nil {
		return false, mapTaxonomyError("topic", err)
	}
	return true, nil
}
