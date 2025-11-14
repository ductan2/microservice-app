package resolver

import (
	"content-services/graph/model"
	"content-services/internal/repository"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// UploadMedia is the resolver for the uploadMedia field.
func (r *mutationResolver) UploadMedia(ctx context.Context, input model.UploadMediaInput) (*model.MediaAsset, error) {
	return r.handleUploadInput(ctx, input)
}

// UploadMediaBatch is the resolver for the uploadMediaBatch field.
func (r *mutationResolver) UploadMediaBatch(ctx context.Context, inputs []*model.UploadMediaInput) ([]*model.MediaAsset, error) {
	if len(inputs) == 0 {
		return nil, gqlerror.Errorf("at least one input is required")
	}

	results := make([]*model.MediaAsset, 0, len(inputs))
	for idx, input := range inputs {
		if input == nil {
			return nil, gqlerror.Errorf("input %d is nil", idx)
		}
		asset, err := r.handleUploadInput(ctx, *input)
		if err != nil {
			return nil, gqlerror.Errorf("upload %d failed: %v", idx, err)
		}
		results = append(results, asset)
	}

	return results, nil
}

// DeleteMedia is the resolver for the deleteMedia field.
func (r *mutationResolver) DeleteMedia(ctx context.Context, id string) (bool, error) {
	if r.Media == nil {
		return false, gqlerror.Errorf("media service not configured")
	}
	mediaID, err := uuid.Parse(id)
	if err != nil {
		return false, gqlerror.Errorf("invalid media id: %v", err)
	}
	if err := r.Media.DeleteMedia(ctx, mediaID); err != nil {
		if errors.Is(err, repository.ErrMediaNotFound) {
			return false, gqlerror.Errorf("media asset not found")
		}
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) handleUploadInput(ctx context.Context, input model.UploadMediaInput) (*model.MediaAsset, error) {
	if r.Media == nil {
		return nil, gqlerror.Errorf("media service not configured")
	}

	upload := input.File
	defer func() {
		if closer, ok := upload.File.(io.ReadCloser); ok {
			closer.Close()
		}
	}()

	filename := upload.Filename
	if input.Filename != nil && *input.Filename != "" {
		filename = *input.Filename
	}
	if filename == "" {
		filename = "upload"
	}

	kind := strings.ToLower(input.Kind.String())
	var userID uuid.UUID
	if input.UploadedBy != nil && *input.UploadedBy != "" {
		parsed, err := uuid.Parse(*input.UploadedBy)
		if err != nil {
			return nil, gqlerror.Errorf("invalid uploadedBy: %v", err)
		}
		userID = parsed
	}

	var folderID *uuid.UUID
	if input.FolderID != nil && *input.FolderID != "" {
		parsed, err := uuid.Parse(*input.FolderID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid folderId: %v", err)
		}
		folderID = &parsed
	}

	media, err := r.Media.UploadMedia(ctx, upload.File, filename, input.MimeType, kind, userID, folderID)
	if err != nil {
		return nil, err
	}

	return r.mapMediaAsset(ctx, media)
}
