package mappers

import (
	"context"
	"strings"

	"content-services/graph/model"
	"content-services/internal/models"
	"content-services/internal/utils"
)

// MediaAssetToGraphQL converts models.MediaAsset to model.MediaAsset with presigned URL
func MediaAssetToGraphQL(ctx context.Context, media *models.MediaAsset, mediaService interface {
	GetPresignedURL(ctx context.Context, id interface{}) (string, error)
}) (*model.MediaAsset, error) {
	if media == nil {
		return nil, nil
	}

	var uploadedBy *string
	if media.UploadedBy != nil {
		id := media.UploadedBy.String()
		uploadedBy = &id
	}

	var folderID *string
	if media.FolderID != nil {
		id := media.FolderID.String()
		folderID = &id
	}

	var thumbnailURL *string
	if media.ThumbnailURL != "" {
		url := media.ThumbnailURL
		thumbnailURL = &url
	}

	duration := utils.ToIntPtr(media.DurationMs)

	downloadURL := ""
	if mediaService != nil {
		url, err := mediaService.GetPresignedURL(ctx, media.ID)
		if err != nil {
			return nil, err
		}
		downloadURL = url
	}

	return &model.MediaAsset{
		ID:           media.ID.String(),
		StorageKey:   media.StorageKey,
		Kind:         MapMediaKind(media.Kind),
		MimeType:     media.MimeType,
		FolderID:     folderID,
		OriginalName: media.OriginalName,
		ThumbnailURL: thumbnailURL,
		Bytes:        media.Bytes,
		DurationMs:   duration,
		Sha256:       media.SHA256,
		CreatedAt:    media.CreatedAt,
		UploadedBy:   uploadedBy,
		DownloadURL:  downloadURL,
	}, nil
}

// MapMediaKind converts string to model.MediaKind enum
func MapMediaKind(kind string) model.MediaKind {
	switch strings.ToLower(kind) {
	case "image":
		return model.MediaKindImage
	case "audio":
		return model.MediaKindAudio
	default:
		return model.MediaKind(strings.ToUpper(kind))
	}
}

// RepositoryTagsToGraphQL converts repository tag models to GraphQL models
func RepositoryTagsToGraphQL(tags []models.Tag) []*model.Tag {
	result := make([]*model.Tag, 0, len(tags))
	for i := range tags {
		tag := tags[i]
		result = append(result, &model.Tag{
			ID:   tag.ID.String(),
			Slug: tag.Slug,
			Name: tag.Name,
		})
	}
	return result
}

// ContentTagKindToModel converts GraphQL enum to repository kind string
func ContentTagKindToModel(kind model.ContentTagKind) string {
	switch kind {
	case model.ContentTagKindLesson:
		return "lesson"
	case model.ContentTagKindQuiz:
		return "quiz"
	case model.ContentTagKindFlashcardSet:
		return "flashcard_set"
	default:
		return strings.ToLower(string(kind))
	}
}