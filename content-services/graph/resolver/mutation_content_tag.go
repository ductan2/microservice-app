package resolver

import (
	"content-services/graph/model"
	"content-services/internal/taxonomy"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// AddContentTag is the resolver for the addContentTag field.
func (r *mutationResolver) AddContentTag(ctx context.Context, input model.ContentTagInput) (*model.ContentTag, error) {
	if r.TagRepo == nil {
		return nil, gqlerror.Errorf("tag repository not configured")
	}

	tagID, err := uuid.Parse(input.TagID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid tag ID: %v", err)
	}

	objectID, err := uuid.Parse(input.ObjectID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid object ID: %v", err)
	}

	kind := contentTagKindToModel(input.Kind)
	if err := r.TagRepo.AddTagToContent(ctx, tagID, kind, objectID); err != nil {
		return nil, err
	}

	var tag *model.Tag
	if r.Taxonomy != nil {
		taxonomyTag, err := r.Taxonomy.GetTagByID(ctx, input.TagID)
		if err != nil {
			if !errors.Is(err, taxonomy.ErrNotFound) {
				return nil, err
			}
		} else {
			tag = mapTag(taxonomyTag)
		}
	}

	return &model.ContentTag{
		TagID:    input.TagID,
		Kind:     input.Kind,
		ObjectID: input.ObjectID,
		Tag:      tag,
	}, nil
}

// RemoveContentTag is the resolver for the removeContentTag field.
func (r *mutationResolver) RemoveContentTag(ctx context.Context, input model.ContentTagInput) (bool, error) {
	if r.TagRepo == nil {
		return false, gqlerror.Errorf("tag repository not configured")
	}

	tagID, err := uuid.Parse(input.TagID)
	if err != nil {
		return false, gqlerror.Errorf("invalid tag ID: %v", err)
	}

	objectID, err := uuid.Parse(input.ObjectID)
	if err != nil {
		return false, gqlerror.Errorf("invalid object ID: %v", err)
	}

	kind := contentTagKindToModel(input.Kind)
	if err := r.TagRepo.RemoveTagFromContent(ctx, tagID, kind, objectID); err != nil {
		return false, err
	}

	return true, nil
}
