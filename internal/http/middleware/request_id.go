package middleware

import (
	"context"
	"net/http"

	"github.com/edsonmubezi/myapp/internal/platform/logging"
	"github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

// RequestID middleware generates or propagates X-Request-ID
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get or generate request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// Add to logger context
		ctx = logging.WithRequestID(ctx, requestID)

		// Return in response headers
		w.Header().Set("X-Request-ID", requestID)

		// Also support traceparent header for W3C trace context
		traceparent := r.Header.Get("traceparent")
		if traceparent != "" {
			ctx = context.WithValue(ctx, "traceparent", traceparent)
		}

		// Continue with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}
