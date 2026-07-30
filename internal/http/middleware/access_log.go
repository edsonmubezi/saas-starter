package middleware

import (
	"net/http"
	"time"

	"github.com/edsonmubezi/myapp/internal/platform/logging"
	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	status        int
	bytesWritten  int64
	wroteHeader   bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.ResponseWriter.WriteHeader(code)
		rw.wroteHeader = true
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// AccessLog middleware logs HTTP requests with structured fields
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		rw := &responseWriter{
			ResponseWriter: w,
			status:         200,
			wroteHeader:    false,
		}

		// Get logger from context
		logger := logging.FromContext(r.Context())

		// Add request fields
		logger = logger.With(
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", getRemoteIP(r)),
			zap.String("user_agent", r.UserAgent()),
		)

		// Only log content length if present
		if r.ContentLength > 0 {
			logger = logger.With(zap.Int64("bytes_in", r.ContentLength))
		}

		// Process request
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start)

		// Log after completion
		fields := []zap.Field{
			zap.Int("status", rw.status),
			zap.Duration("latency", duration),
			zap.Float64("latency_ms", float64(duration.Milliseconds())),
		}

		// Only include bytes_out if we wrote something
		if rw.bytesWritten > 0 {
			fields = append(fields, zap.Int64("bytes_out", rw.bytesWritten))
		}

		// Log at appropriate level
		if rw.status >= 500 {
			logger.Error("HTTP request", fields...)
		} else if rw.status >= 400 {
			logger.Warn("HTTP request", fields...)
		} else {
			logger.Info("HTTP request", fields...)
		}
	})
}

// getRemoteIP extracts the real client IP considering proxies
func getRemoteIP(r *http.Request) string {
	// Check X-Forwarded-For header (set by proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
