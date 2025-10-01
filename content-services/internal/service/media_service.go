package service

import (
	"content-services/internal/models"
	"content-services/internal/repository"
	"context"
	"io"

	"github.com/google/uuid"
)

type MediaService interface {
	UploadMedia(ctx context.Context, file io.Reader, filename, mimeType, kind string, userID uuid.UUID) (*models.MediaAsset, error)
	GetMediaByID(ctx context.Context, id uuid.UUID) (*models.MediaAsset, error)
	GetMediaByIDs(ctx context.Context, ids []uuid.UUID) ([]models.MediaAsset, error)
	GetPresignedURL(ctx context.Context, id uuid.UUID) (string, error)
	DeleteMedia(ctx context.Context, id uuid.UUID) error
}

type mediaService struct {
	mediaRepo repository.MediaRepository
	// storageClient (S3/MinIO) will be injected here
}

func NewMediaService(
	mediaRepo repository.MediaRepository,
) MediaService {
	return &mediaService{
		mediaRepo: mediaRepo,
	}
}

func (s *mediaService) UploadMedia(ctx context.Context, file io.Reader, filename, mimeType, kind string, userID uuid.UUID) (*models.MediaAsset, error) {
	// TODO: implement
	// 1. Calculate SHA256 hash
	// 2. Check if file already exists (dedup)
	// 3. Upload to S3/MinIO
	// 4. Save metadata to DB
	// 5. For audio, extract duration
	return nil, nil
}

func (s *mediaService) GetMediaByID(ctx context.Context, id uuid.UUID) (*models.MediaAsset, error) {
	// TODO: implement
	return nil, nil
}

func (s *mediaService) GetMediaByIDs(ctx context.Context, ids []uuid.UUID) ([]models.MediaAsset, error) {
	// TODO: implement - for DataLoader pattern
	return nil, nil
}

func (s *mediaService) GetPresignedURL(ctx context.Context, id uuid.UUID) (string, error) {
	// TODO: implement - generate presigned URL for media access
	return "", nil
}

func (s *mediaService) DeleteMedia(ctx context.Context, id uuid.UUID) error {
	// TODO: implement - delete from S3 and DB
	return nil
}
