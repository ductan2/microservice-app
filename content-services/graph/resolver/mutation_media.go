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

	media, err := r.Media.UploadMedia(ctx, upload.File, filename, input.MimeType, kind, userID)
	if err != nil {
		return nil, err
	}

	return r.mapMediaAsset(ctx, media)
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
