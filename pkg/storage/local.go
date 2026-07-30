package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStorage implements Storage interface for local filesystem
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(cfg Config) (*LocalStorage, error) {
	basePath := cfg.LocalPath
	if basePath == "" {
		basePath = "./uploads"
	}

	// Try to create the directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0o750); err != nil {
		// Check if directory already exists (maybe created by Docker volume mount)
		if stat, statErr := os.Stat(basePath); statErr == nil && stat.IsDir() {
			// Directory exists, we can proceed
			return &LocalStorage{basePath: basePath}, nil
		}

		// If we can't create the directory and it doesn't exist, fail with helpful error
		return nil, fmt.Errorf("failed to create base storage directory '%s': %w. "+
			"In production, consider using S3 storage by setting STORAGE_TYPE=s3 "+
			"or ensure the directory exists and has proper permissions", basePath, err)
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

// validatePath checks for path traversal attempts and returns the safe full path
func (l *LocalStorage) validatePath(path string) (string, error) {
	// Clean the path to remove any .. or . components
	cleanPath := filepath.Clean(filepath.FromSlash(path))

	// Check for path traversal attempts
	if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, string(filepath.Separator)+"..") {
		return "", fmt.Errorf("invalid path: path traversal not allowed")
	}

	// Build the full path
	fullPath := filepath.Join(l.basePath, cleanPath)

	// Verify the result is still within basePath (defense in depth)
	absBase, err := filepath.Abs(l.basePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base path: %w", err)
	}
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}
	if !strings.HasPrefix(absPath, absBase) {
		return "", fmt.Errorf("invalid path: outside storage directory")
	}

	return fullPath, nil
}

// Upload saves a file to the local filesystem
func (l *LocalStorage) Upload(ctx context.Context, path string, reader io.Reader, contentType string) error {
	fullPath, err := l.validatePath(path)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create file with secure permissions (fail if exists)
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o640)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("file already exists: %s", path)
		}
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data to file
	if _, err := io.Copy(file, reader); err != nil {
		// Clean up partial file on error
		os.Remove(fullPath)
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Download reads a file from the local filesystem
func (l *LocalStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath, err := l.validatePath(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete removes a file from the local filesystem
func (l *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath, err := l.validatePath(path)
	if err != nil {
		return err
	}

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted, not an error
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetURL returns a relative path for local files
// For local storage, we return the path relative to the base
// The expiry parameter is ignored for local storage
func (l *LocalStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	fullPath, err := l.validatePath(path)
	if err != nil {
		return "", err
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", path)
		}
		return "", fmt.Errorf("failed to check file: %w", err)
	}

	// Return the path prefixed with /api/uploads/ so it routes through the API proxy
	// (nginx proxies /api/* to the backend; plain /uploads/ goes to the frontend SPA)
	return "/api/uploads/" + filepath.ToSlash(path), nil
}

// Exists checks if a file exists at the given path
func (l *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath, err := l.validatePath(path)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}
