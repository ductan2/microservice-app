package resolver

import (
	"content-services/graph/model"
	"content-services/internal/models"
	"content-services/internal/taxonomy"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func mapTaxonomyError(resource string, err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, taxonomy.ErrDuplicate):
		return gqlerror.Errorf("%s already exists", resource)
	case errors.Is(err, taxonomy.ErrNotFound):
		return gqlerror.Errorf("%s not found", resource)
	default:
		return err
	}
}

func mapTopic(topic *taxonomy.Topic) *model.Topic {
	if topic == nil {
		return nil
	}
	return &model.Topic{
		ID:        topic.ID,
		Slug:      topic.Slug,
		Name:      topic.Name,
		CreatedAt: topic.CreatedAt,
	}
}

func mapLevel(level *taxonomy.Level) *model.Level {
	if level == nil {
		return nil
	}
	return &model.Level{
		ID:   level.ID,
		Code: level.Code,
		Name: level.Name,
	}
}

func mapTag(tag *taxonomy.Tag) *model.Tag {
	if tag == nil {
		return nil
	}
	return &model.Tag{
		ID:   tag.ID,
		Slug: tag.Slug,
		Name: tag.Name,
	}
}

func (r *Resolver) mapMediaAsset(ctx context.Context, media *models.MediaAsset) (*model.MediaAsset, error) {
	if media == nil {
		return nil, nil
	}
	var uploadedBy *string
	if media.UploadedBy != nil {
		id := media.UploadedBy.String()
		uploadedBy = &id
	}
	var duration *int
	if media.DurationMs > 0 {
		duration = &media.DurationMs
	}
	if media.ID == uuid.Nil {
		media.ID = uuid.New()
	}
	downloadURL := ""
	if r.Media != nil {
		url, err := r.Media.GetPresignedURL(ctx, media.ID)
		if err != nil {
			return nil, err
		}
		downloadURL = url
	}
	return &model.MediaAsset{
		ID:          media.ID.String(),
		StorageKey:  media.StorageKey,
		Kind:        mapMediaKind(media.Kind),
		MimeType:    media.MimeType,
		Bytes:       media.Bytes,
		DurationMs:  duration,
		Sha256:      media.SHA256,
		CreatedAt:   media.CreatedAt,
		UploadedBy:  uploadedBy,
		DownloadURL: downloadURL,
	}, nil
}

func mapMediaKind(kind string) model.MediaKind {
	switch strings.ToLower(kind) {
	case "image":
		return model.MediaKindImage
	case "audio":
		return model.MediaKindAudio
	default:
		return model.MediaKind(strings.ToUpper(kind))
	}
}
