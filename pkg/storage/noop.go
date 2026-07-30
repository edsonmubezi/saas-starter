package storage

import (
	"context"
	"fmt"
	"io"
	"time"
)

// NoOpStorage is a dummy storage implementation that returns errors
// Used when file upload feature is disabled
type NoOpStorage struct{}

// NewNoOpStorage creates a new no-op storage instance
func NewNoOpStorage() *NoOpStorage {
	return &NoOpStorage{}
}

// Upload returns an error indicating file upload is disabled
func (s *NoOpStorage) Upload(ctx context.Context, path string, reader io.Reader, contentType string) error {
	return fmt.Errorf("file upload is disabled - set ENABLE_FILE_UPLOAD=true in .env to enable")
}

// Download returns an error indicating file upload is disabled
func (s *NoOpStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("file upload is disabled - set ENABLE_FILE_UPLOAD=true in .env to enable")
}

// Delete returns an error indicating file upload is disabled
func (s *NoOpStorage) Delete(ctx context.Context, path string) error {
	return fmt.Errorf("file upload is disabled - set ENABLE_FILE_UPLOAD=true in .env to enable")
}

// GetURL returns an error indicating file upload is disabled
func (s *NoOpStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	return "", fmt.Errorf("file upload is disabled - set ENABLE_FILE_UPLOAD=true in .env to enable")
}

// Exists returns false and an error indicating file upload is disabled
func (s *NoOpStorage) Exists(ctx context.Context, path string) (bool, error) {
	return false, fmt.Errorf("file upload is disabled - set ENABLE_FILE_UPLOAD=true in .env to enable")
}
