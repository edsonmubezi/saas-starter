package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

// ServeExpenseClaimFileHandler serves expense claim receipt files securely
// GET /api/files/expense-claims/{orgId}/{filename}
func ServeExpenseClaimFileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgId := vars["orgId"]
	filename := vars["filename"]

	// Validate parameters to prevent path traversal
	if !isValidPathSegment(orgId) || !isValidPathSegment(filename) {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	// Get uploads directory from environment
	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}

	// Construct file path: uploads/org-{orgId}/expense-claims/{filename}
	filePath := filepath.Join(uploadsDir, orgId, "expense-claims", filename)

	// Serve the file with proper headers
	serveFileWithHeaders(w, r, filePath, filename)
}

// ServeEmployeeDocumentFileHandler serves employee document files securely
// GET /api/files/employees/{orgId}/{employeeId}/{filename}
func ServeEmployeeDocumentFileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgId := vars["orgId"]
	employeeId := vars["employeeId"]
	filename := vars["filename"]

	// Validate parameters to prevent path traversal
	if !isValidPathSegment(orgId) || !isValidPathSegment(employeeId) || !isValidPathSegment(filename) {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	// Get uploads directory from environment
	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}

	// Construct file path: uploads/org-{orgId}/employee-{employeeId}/{filename}
	filePath := filepath.Join(uploadsDir, orgId, employeeId, filename)

	// Serve the file with proper headers
	serveFileWithHeaders(w, r, filePath, filename)
}

// isValidPathSegment validates a path segment to prevent path traversal
func isValidPathSegment(segment string) bool {
	if segment == "" {
		return false
	}
	// Reject path traversal attempts
	if strings.Contains(segment, "..") || strings.Contains(segment, "/") || strings.Contains(segment, "\\") {
		return false
	}
	return true
}

// serveFileWithHeaders serves a file with appropriate Content-Type and Content-Disposition headers
func serveFileWithHeaders(w http.ResponseWriter, r *http.Request, filePath, filename string) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Determine content type based on file extension
	ext := strings.ToLower(filepath.Ext(filename))
	contentType := getContentType(ext)

	// Set headers for proper browser handling
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s", filename))
	// Allow embedding in iframes
	w.Header().Del("X-Frame-Options")

	// Serve file
	http.ServeFile(w, r, filePath)
}

// getContentType returns the MIME type for a given file extension
func getContentType(ext string) string {
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}
