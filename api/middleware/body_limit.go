package middleware

import "net/http"

const (
	// DefaultMaxBodySize is the default limit for JSON/form API requests (4 MB).
	DefaultMaxBodySize int64 = 4 << 20 // 4 MB

	// UploadMaxBodySize is the limit for file upload endpoints (32 MB).
	UploadMaxBodySize int64 = 32 << 20 // 32 MB
)

// BodyLimit returns middleware that caps the request body at maxBytes.
// Requests exceeding the limit will receive a 413 Request Entity Too Large
// response from the standard library when the body is read.
//
// Usage:
//
//	r.Use(BodyLimit(middleware.DefaultMaxBodySize))          // global default
//	r.Handle("/upload", BodyLimit(middleware.UploadMaxBodySize)(handler))
func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
