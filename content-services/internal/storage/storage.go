package storage

import (
	"context"
	"io"
	"time"
)

// ObjectStorage captures the operations required by the media service for storing
// and retrieving binary objects.
type ObjectStorage interface {
	PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	DeleteObject(ctx context.Context, key string) error
	GeneratePresignedURL(ctx context.Context, key string, expiresIn time.Duration) (string, error)
}
