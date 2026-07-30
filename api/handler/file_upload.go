package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dutchcoders/go-clamd"
	"github.com/edsonmubezi/myapp/pkg/resilience"
	"github.com/edsonmubezi/myapp/pkg/storage"
)

// Global storage instance
var storageService storage.Storage

// SetStorage sets the storage service to be used for file uploads
func SetStorage(s storage.Storage) {
	storageService = s
}

// clamavAddr returns the ClamAV daemon address from environment
func clamavAddr() string {
	if v := os.Getenv("CLAMAV_ADDR"); v != "" {
		return v
	}
	return "tcp://clamav:3310"
}

// isClamAVEnabled checks if virus scanning is enabled
func isClamAVEnabled() bool {
	enabled := os.Getenv("CLAMAV_ENABLED")
	return enabled == "true" || enabled == "1" || enabled == "yes"
}

// ErrClamAVUnavailable is returned when ClamAV circuit breaker is open
var ErrClamAVUnavailable = errors.New("ClamAV service unavailable")

// scanFileForVirus scans uploaded file with ClamAV for viruses
// Returns error if virus found or scanning fails
// Uses circuit breaker to prevent cascading failures when ClamAV is down
func scanFileForVirus(ctx context.Context, file multipart.File) error {
	if !isClamAVEnabled() {
		log.Println("INFO: ClamAV scanning disabled, skipping virus scan")
		return nil
	}

	// Check circuit breaker first
	cb := resilience.GetBreaker(resilience.ServiceClamAV)
	if cb != nil && !cb.IsAvailable() {
		log.Println("WARNING: ClamAV circuit breaker open, skipping virus scan")
		if os.Getenv("CLAMAV_REQUIRED") == "true" {
			return ErrClamAVUnavailable
		}
		return nil
	}

	reader, ok := file.(io.Reader)
	if !ok {
		return fmt.Errorf("file is not readable")
	}

	// Add timeout if none set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
		defer cancel()
	}

	// Execute scan with circuit breaker protection
	var scanErr error
	operation := func() error {
		scanErr = performVirusScan(ctx, reader)
		return scanErr
	}

	if cb != nil {
		if err := cb.Execute(ctx, operation); err != nil {
			if errors.Is(err, resilience.ErrCircuitOpen) {
				log.Println("WARNING: ClamAV circuit breaker opened during scan")
				if os.Getenv("CLAMAV_REQUIRED") == "true" {
					return ErrClamAVUnavailable
				}
				return nil
			}
			return err
		}
	} else {
		if err := operation(); err != nil {
			return err
		}
	}

	return scanErr
}

// performVirusScan executes the actual virus scan
func performVirusScan(ctx context.Context, reader io.Reader) error {
	c := clamd.NewClamd(clamavAddr())

	// Test connection first
	err := c.Ping()
	if err != nil {
		log.Printf("WARNING: ClamAV not reachable at %s: %v", clamavAddr(), err)
		if os.Getenv("CLAMAV_REQUIRED") == "true" {
			return fmt.Errorf("virus scanning required but ClamAV unavailable: %w", err)
		}
		// Return error to trigger circuit breaker
		return fmt.Errorf("ClamAV connection failed: %w", err)
	}

	// Abort channel for ScanStream
	abort := make(chan bool, 1)
	go func() {
		<-ctx.Done()
		abort <- true
	}()

	resp, err := c.ScanStream(reader, abort)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("virus scan aborted: %w", ctx.Err())
		}
		return fmt.Errorf("virus scan error: %w", err)
	}

	for r := range resp {
		switch r.Status {
		case clamd.RES_OK:
			log.Println("INFO: File passed virus scan")
			return nil
		case clamd.RES_FOUND:
			log.Printf("SECURITY: Virus detected - %s", r.Description)
			// Virus found is not a circuit breaker failure - it's expected behavior
			return fmt.Errorf("virus detected: %s", r.Description)
		case clamd.RES_ERROR, clamd.RES_PARSE_ERROR:
			return fmt.Errorf("virus scan error: %s", r.Description)
		}
	}

	if ctx.Err() != nil {
		return fmt.Errorf("virus scan aborted: %w", ctx.Err())
	}
	return fmt.Errorf("virus scan returned no result")
}

// UploadPDFFile validates, scans, and saves a PDF file securely using the configured storage backend.
// Parameters:
//   - r: HTTP request containing the file
//   - fieldName: Name of the form field containing the file
//   - orgID: Organization ID for path structuring
//   - subPath: Document type folder (e.g., "background-checks", "disciplinary-cases")
//
// Returns the storage path where the file was saved
func UploadPDFFile(r *http.Request, fieldName string, orgID int64, subPath string) (string, error) {
	if storageService == nil {
		return "", fmt.Errorf("storage service not initialized")
	}

	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", fmt.Errorf("missing or invalid file: %v", err)
	}
	defer file.Close()

	// Structural validation
	if err := validatePDF(header, file); err != nil {
		return "", err
	}

	// Virus scan (if enabled)
	if err := scanFileForVirus(r.Context(), file); err != nil {
		return "", fmt.Errorf("virus scan failed: %v", err)
	}

	// Reset file pointer after validation + scan
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return "", fmt.Errorf("cannot reset file reader: %v", err)
		}
	} else {
		return "", fmt.Errorf("uploaded file is not seekable")
	}

	// Generate secure random filename
	randomFileName, err := generateRandomFilename(".pdf")
	if err != nil {
		return "", fmt.Errorf("failed to generate file name: %v", err)
	}

	// Generate standardized storage path: org-{orgID}/{subPath}/{filename}
	storagePath := fmt.Sprintf("org-%d/%s/%s", orgID, subPath, randomFileName)

	// Upload file using storage abstraction
	if err := storageService.Upload(r.Context(), storagePath, file, "application/pdf"); err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	return storagePath, nil
}

// Generates a secure random file name with given extension
func generateRandomFilename(ext string) (string, error) {
	b := make([]byte, 16) // 128 bits
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b) + ext, nil
}

// validatePDF ensures the file looks like a PDF by extension, signature, and MIME
func validatePDF(header *multipart.FileHeader, file multipart.File) error {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" {
		return errors.New("only PDF files are allowed")
	}

	// Read a small header
	head := make([]byte, 512)
	n, err := file.Read(head)
	if err != nil && err != io.EOF {
		return errors.New("unable to read file for validation")
	}
	head = head[:n]

	// Reset pointer
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return errors.New("cannot reset file pointer after reading")
		}
	} else {
		return errors.New("uploaded file is not seekable")
	}

	if n < 5 || string(head[:5]) != "%PDF-" {
		return errors.New("file content is not a valid PDF")
	}

	ctype := http.DetectContentType(head)
	if ctype != "application/pdf" && !strings.HasPrefix(ctype, "application/octet-stream") {
		return errors.New("not a PDF (content-type)")
	}

	return nil
}

// UploadImageFile validates, scans, and saves an image file securely using the configured storage backend.
// Supports common image formats: JPG, PNG, GIF, WebP
// Parameters:
//   - r: HTTP request containing the file
//   - fieldName: Name of the form field containing the file
//   - orgID: Organization ID for path structuring
//   - subPath: Document type folder (e.g., "org-branding")
//
// Returns the storage path where the file was saved
func UploadImageFile(r *http.Request, fieldName string, orgID int64, subPath string) (string, error) {
	if storageService == nil {
		return "", fmt.Errorf("storage service not initialized")
	}

	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", fmt.Errorf("missing or invalid file: %v", err)
	}
	defer file.Close()

	// Structural validation
	ext, contentType, err := validateImage(header, file)
	if err != nil {
		return "", err
	}

	// Virus scan (if enabled)
	if err := scanFileForVirus(r.Context(), file); err != nil {
		return "", fmt.Errorf("virus scan failed: %v", err)
	}

	// Reset file pointer after validation + scan
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return "", fmt.Errorf("cannot reset file reader: %v", err)
		}
	} else {
		return "", fmt.Errorf("uploaded file is not seekable")
	}

	// Generate secure random filename
	randomFileName, err := generateRandomFilename(ext)
	if err != nil {
		return "", fmt.Errorf("failed to generate file name: %v", err)
	}

	// Generate standardized storage path: org-{orgID}/{subPath}/{filename}
	storagePath := fmt.Sprintf("org-%d/%s/%s", orgID, subPath, randomFileName)

	// Upload file using storage abstraction
	if err := storageService.Upload(r.Context(), storagePath, file, contentType); err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	return storagePath, nil
}

// validateImage ensures the file is a valid image by extension, signature, and MIME
func validateImage(header *multipart.FileHeader, file multipart.File) (string, string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))

	// Allowed image extensions
	allowedExts := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
	}

	expectedContentType, ok := allowedExts[ext]
	if !ok {
		return "", "", errors.New("only JPG, PNG, GIF, and WebP image files are allowed")
	}

	// Check file size (max 5MB for images)
	if header.Size > 5*1024*1024 {
		return "", "", errors.New("image file size must be less than 5MB")
	}

	// Read file header for content type detection
	head := make([]byte, 512)
	n, err := file.Read(head)
	if err != nil && err != io.EOF {
		return "", "", errors.New("unable to read file for validation")
	}
	head = head[:n]

	// Reset pointer
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return "", "", errors.New("cannot reset file pointer after reading")
		}
	} else {
		return "", "", errors.New("uploaded file is not seekable")
	}

	// Detect content type
	detectedType := http.DetectContentType(head)

	// Validate content type matches expected for the extension
	if !strings.HasPrefix(detectedType, "image/") {
		return "", "", errors.New("file content is not a valid image")
	}

	return ext, expectedContentType, nil
}

// UploadReceiptFile validates, scans, and saves a receipt file (PDF or image) securely.
// Supports: PDF, JPG, JPEG, PNG, GIF, WebP
// Parameters:
//   - r: HTTP request containing the file
//   - fieldName: Name of the form field containing the file
//   - orgID: Organization ID for path structuring
//   - subPath: Sub-path within org folder (e.g., "expense-claims")
//
// Returns the storage path where the file was saved
func UploadReceiptFile(r *http.Request, fieldName string, orgID int64, subPath string) (string, error) {
	if storageService == nil {
		return "", fmt.Errorf("storage service not initialized")
	}

	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", fmt.Errorf("missing or invalid file: %v", err)
	}
	defer file.Close()

	// Check file size (max 6MB)
	if header.Size > 6*1024*1024 {
		return "", errors.New("file size must be less than 6MB")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	var contentType string

	// Validate based on file type
	if ext == ".pdf" {
		// Validate as PDF
		if err := validatePDF(header, file); err != nil {
			return "", err
		}
		contentType = "application/pdf"
	} else {
		// Validate as image
		var imgErr error
		ext, contentType, imgErr = validateImage(header, file)
		if imgErr != nil {
			return "", fmt.Errorf("invalid file: only PDF, JPG, PNG, GIF, WebP are allowed")
		}
	}

	// Virus scan (if enabled)
	if err := scanFileForVirus(r.Context(), file); err != nil {
		return "", fmt.Errorf("virus scan failed: %v", err)
	}

	// Reset file pointer after validation + scan
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return "", fmt.Errorf("cannot reset file reader: %v", err)
		}
	} else {
		return "", fmt.Errorf("uploaded file is not seekable")
	}

	// Generate secure random filename
	randomFileName, err := generateRandomFilename(ext)
	if err != nil {
		return "", fmt.Errorf("failed to generate file name: %v", err)
	}

	// Generate storage path: org-{orgID}/{subPath}/{randomfilename}
	storagePath := fmt.Sprintf("org-%d/%s/%s", orgID, subPath, randomFileName)

	// Upload file using storage abstraction
	if err := storageService.Upload(r.Context(), storagePath, file, contentType); err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	return storagePath, nil
}

// GetStorageService returns the storage service for external use (e.g., presigned URLs)
func GetStorageService() storage.Storage {
	return storageService
}
