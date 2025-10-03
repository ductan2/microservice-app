package service

import (
	"bytes"
	"content-services/internal/models"
	"content-services/internal/repository"
	"content-services/internal/storage"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

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
	mediaRepo  repository.MediaRepository
	storage    storage.ObjectStorage
	presignTTL time.Duration
}

func NewMediaService(mediaRepo repository.MediaRepository, storage storage.ObjectStorage, presignTTL time.Duration) MediaService {
	return &mediaService{
		mediaRepo:  mediaRepo,
		storage:    storage,
		presignTTL: presignTTL,
	}
}

func (s *mediaService) UploadMedia(ctx context.Context, file io.Reader, filename, mimeType, kind string, userID uuid.UUID) (*models.MediaAsset, error) {
	if file == nil {
		return nil, errors.New("media: file reader is required")
	}
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind != "image" && kind != "audio" {
		return nil, fmt.Errorf("media: unsupported kind %q", kind)
	}
	var buffer bytes.Buffer
	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(&buffer, hasher), file)
	if err != nil {
		return nil, fmt.Errorf("media: copy failed: %w", err)
	}
	if written == 0 {
		return nil, errors.New("media: empty file")
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))
	if existing, err := s.mediaRepo.GetBySHA256(ctx, checksum); err == nil && existing != nil {
		return existing, nil
	} else if err != nil && !errors.Is(err, repository.ErrMediaNotFound) {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(filename))
	key := fmt.Sprintf("media/%s/%s", kind, checksum)
	if ext != "" {
		key = fmt.Sprintf("%s%s", key, ext)
	}

	if s.storage == nil {
		return nil, errors.New("media: storage client not configured")
	}
	reader := bytes.NewReader(buffer.Bytes())
	if err := s.storage.PutObject(ctx, key, reader, written, mimeType); err != nil {
		return nil, err
	}

	media := &models.MediaAsset{
		StorageKey: key,
		Kind:       kind,
		MimeType:   mimeType,
		Bytes:      int(written),
		SHA256:     checksum,
		CreatedAt:  time.Now().UTC(),
	}
	if userID != uuid.Nil {
		uid := userID
		media.UploadedBy = &uid
	}

	if err := s.mediaRepo.Create(ctx, media); err != nil {
		_ = s.storage.DeleteObject(ctx, key)
		return nil, err
	}
	return media, nil
}

func (s *mediaService) GetMediaByID(ctx context.Context, id uuid.UUID) (*models.MediaAsset, error) {
	return s.mediaRepo.GetByID(ctx, id)
}

func (s *mediaService) GetMediaByIDs(ctx context.Context, ids []uuid.UUID) ([]models.MediaAsset, error) {
	return s.mediaRepo.GetByIDs(ctx, ids)
}

func (s *mediaService) GetPresignedURL(ctx context.Context, id uuid.UUID) (string, error) {
	media, err := s.mediaRepo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	if s.storage == nil {
		return "", errors.New("media: storage client not configured")
	}
	return s.storage.GeneratePresignedURL(ctx, media.StorageKey, s.presignTTL)
}

func (s *mediaService) DeleteMedia(ctx context.Context, id uuid.UUID) error {
	media, err := s.mediaRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if s.storage != nil {
		if err := s.storage.DeleteObject(ctx, media.StorageKey); err != nil {
			return err
		}
	}
	return s.mediaRepo.Delete(ctx, id)
}
