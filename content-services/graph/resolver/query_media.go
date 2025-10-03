package resolver

import (
	"content-services/graph/model"
	"content-services/internal/models"
	"content-services/internal/repository"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// MediaAsset is the resolver for the mediaAsset field.
func (r *queryResolver) MediaAsset(ctx context.Context, id string) (*model.MediaAsset, error) {
	if r.Media == nil {
		return nil, gqlerror.Errorf("media service not configured")
	}
	mediaID, err := uuid.Parse(id)
	if err != nil {
		return nil, gqlerror.Errorf("invalid media id: %v", err)
	}
	media, err := r.Media.GetMediaByID(ctx, mediaID)
	if err != nil {
		if errors.Is(err, repository.ErrMediaNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.mapMediaAsset(ctx, media)
}

// MediaAssets is the resolver for the mediaAssets field.
func (r *queryResolver) MediaAssets(ctx context.Context, ids []string) ([]*model.MediaAsset, error) {
	if r.Media == nil {
		return nil, gqlerror.Errorf("media service not configured")
	}
	if len(ids) == 0 {
		return []*model.MediaAsset{}, nil
	}
	uuids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, gqlerror.Errorf("invalid media id: %v", err)
		}
		uuids[i] = parsed
	}
	assets, err := r.Media.GetMediaByIDs(ctx, uuids)
	if err != nil {
		return nil, err
	}
	assetMap := make(map[uuid.UUID]models.MediaAsset, len(assets))
	for i := range assets {
		asset := assets[i]
		assetMap[asset.ID] = asset
	}
	result := make([]*model.MediaAsset, 0, len(ids))
	for _, id := range uuids {
		asset, ok := assetMap[id]
		if !ok {
			continue
		}
		mapped, err := r.mapMediaAsset(ctx, &asset)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}
	return result, nil
}
