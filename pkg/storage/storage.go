package storage

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Storage defines the interface for file storage operations
type Storage interface {
	// Upload uploads a file and returns the storage path/key
	Upload(ctx context.Context, path string, reader io.Reader, contentType string) error

	// Download retrieves a file
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes a file
	Delete(ctx context.Context, path string) error

	// GetURL returns a publicly accessible URL for the file
	// For local storage, this might be a relative path
	// For S3, this could be a presigned URL
	GetURL(ctx context.Context, path string, expiry time.Duration) (string, error)

	// Exists checks if a file exists at the given path
	Exists(ctx context.Context, path string) (bool, error)
}

// Config holds storage configuration
type Config struct {
	Type      string // "local" or "s3"
	LocalPath string // Base path for local storage

	// S3 Configuration
	S3Region    string
	S3Bucket    string
	S3Endpoint  string // Optional: for custom endpoints
	S3AccessKey string
	S3SecretKey string
}

// New creates a storage instance based on configuration
func New(cfg Config) (Storage, error) {
	switch cfg.Type {
	case "s3":
		return NewS3Storage(cfg)
	case "local", "":
		return NewLocalStorage(cfg)
	default:
		return NewLocalStorage(cfg) // Default to local
	}
}

// GeneratePath creates a standardized storage path
// Format: org-{orgID}/employee-{empID}/{filename}
func GeneratePath(orgID, employeeID int64, filename string) string {
	return fmt.Sprintf("org-%d/employee-%d/%s", orgID, employeeID, filename)
}
