package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// TimeoutConfig holds configuration for request timeouts
type TimeoutConfig struct {
	// DefaultTimeout is the default timeout for all requests
	DefaultTimeout time.Duration
	// RouteTimeouts allows custom timeouts for specific route prefixes
	RouteTimeouts map[string]time.Duration
	// ExcludedRoutes are routes that should not have timeout applied
	ExcludedRoutes []string
}

// DefaultTimeoutConfig returns production-ready timeout defaults
func DefaultTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		DefaultTimeout: 30 * time.Second,
		RouteTimeouts: map[string]time.Duration{
			// Health checks should be fast
			"/health": 5 * time.Second,

			// Authentication endpoints
			"/api/login":  15 * time.Second,
			"/api/logout": 10 * time.Second,
			"/api/auth/":  15 * time.Second,

			// Heavy operations need more time
			"/api/payroll/generate":    120 * time.Second,
			"/api/payroll/reports":     60 * time.Second,
			"/api/employees/import":    300 * time.Second,
			"/api/employees/export":    120 * time.Second,
			"/api/org/data-migration":  600 * time.Second,
			"/api/org/data-import/drafts": 600 * time.Second, // bulk import confirm can be slow
			"/api/loan-reports":        60 * time.Second,
			"/api/salary-slip":         60 * time.Second,
			"/api/self/salary-slip":    60 * time.Second,

			// File operations
			"/api/employee-docs":       60 * time.Second,
			"/uploads/":                60 * time.Second,

			// Standard CRUD operations
			"/api/employees":           30 * time.Second,
			"/api/organizations":       30 * time.Second,
			"/api/departments":         20 * time.Second,
			"/api/users":               20 * time.Second,
			"/api/roles":               20 * time.Second,
			"/api/leave":               20 * time.Second,
			"/api/loans":               30 * time.Second,
		},
		ExcludedRoutes: []string{
			"/swagger/",       // Swagger UI needs no timeout
			"/static/",        // Static files
			"/api/org/chat/",  // SSE streaming for AI chat
			"/api/ws",         // WebSocket connections are long-lived
		},
	}
}

// TimeoutMiddleware returns a middleware that enforces request timeouts
func TimeoutMiddleware(cfg *TimeoutConfig) func(http.Handler) http.Handler {
	if cfg == nil {
		cfg = DefaultTimeoutConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if route is excluded
			for _, excluded := range cfg.ExcludedRoutes {
				if strings.HasPrefix(r.URL.Path, excluded) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Determine timeout for this route
			timeout := getTimeoutForRoute(r.URL.Path, cfg)

			// Create context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a channel to signal completion
			done := make(chan struct{})

			// Wrap ResponseWriter to track if response was written
			tw := &timeoutResponseWriter{
				ResponseWriter: w,
				wroteHeader:    false,
			}

			// Run handler in goroutine
			go func() {
				defer func() {
					if err := recover(); err != nil {
						log.Printf("PANIC in handler (recovered): %v", err)
						tw.mu.Lock()
						if !tw.wroteHeader {
							tw.WriteHeader(http.StatusInternalServerError)
							json.NewEncoder(tw).Encode(map[string]string{
								"error": "Internal server error",
							})
						}
						tw.mu.Unlock()
					}
					close(done)
				}()

				next.ServeHTTP(tw, r.WithContext(ctx))
			}()

			// Wait for either completion or timeout
			select {
			case <-done:
				// Handler completed normally
				return

			case <-ctx.Done():
				// Timeout occurred
				tw.mu.Lock()
				defer tw.mu.Unlock()

				if !tw.wroteHeader {
					log.Printf("REQUEST TIMEOUT: %s %s (timeout: %v)", r.Method, r.URL.Path, timeout)

					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Connection", "close")
					w.WriteHeader(http.StatusServiceUnavailable)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status":  503,
						"message": "Request timeout",
						"error":   "The request took too long to process. Please try again.",
					})
				}
				return
			}
		})
	}
}

// getTimeoutForRoute returns the appropriate timeout for a given route
func getTimeoutForRoute(path string, cfg *TimeoutConfig) time.Duration {
	// Check for exact match first
	if timeout, ok := cfg.RouteTimeouts[path]; ok {
		return timeout
	}

	// Check for prefix match (longest prefix wins)
	var bestMatch string
	var bestTimeout time.Duration = cfg.DefaultTimeout

	for prefix, timeout := range cfg.RouteTimeouts {
		if strings.HasPrefix(path, prefix) && len(prefix) > len(bestMatch) {
			bestMatch = prefix
			bestTimeout = timeout
		}
	}

	return bestTimeout
}

// timeoutResponseWriter wraps http.ResponseWriter to track if headers were written
type timeoutResponseWriter struct {
	http.ResponseWriter
	mu          sync.Mutex
	wroteHeader bool
}

func (tw *timeoutResponseWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.wroteHeader {
		return
	}
	tw.wroteHeader = true
	tw.ResponseWriter.WriteHeader(code)
}

func (tw *timeoutResponseWriter) Write(b []byte) (int, error) {
	tw.mu.Lock()
	if !tw.wroteHeader {
		tw.wroteHeader = true
		tw.ResponseWriter.WriteHeader(http.StatusOK)
	}
	tw.mu.Unlock()

	return tw.ResponseWriter.Write(b)
}

// Unwrap returns the underlying ResponseWriter (for http.Flusher, etc.)
func (tw *timeoutResponseWriter) Unwrap() http.ResponseWriter {
	return tw.ResponseWriter
}

// TimeoutMiddlewareSimple returns a simple timeout middleware with a single timeout value
func TimeoutMiddlewareSimple(timeout time.Duration) func(http.Handler) http.Handler {
	return TimeoutMiddleware(&TimeoutConfig{
		DefaultTimeout: timeout,
		RouteTimeouts:  make(map[string]time.Duration),
		ExcludedRoutes: []string{"/swagger/", "/static/"},
	})
}

// WithTimeout wraps a handler with a specific timeout
// Useful for individual handlers that need custom timeouts
func WithTimeout(timeout time.Duration, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		done := make(chan struct{})

		go func() {
			defer close(done)
			handler(w, r.WithContext(ctx))
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			log.Printf("HANDLER TIMEOUT: %s %s (timeout: %v)", r.Method, r.URL.Path, timeout)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  503,
				"message": "Request timeout",
			})
		}
	}
}

// GetRequestTimeout extracts the timeout from the request context
// Returns the remaining time until deadline, or 0 if no deadline set
func GetRequestTimeout(r *http.Request) time.Duration {
	deadline, ok := r.Context().Deadline()
	if !ok {
		return 0
	}
	return time.Until(deadline)
}

// HasRequestDeadline checks if the request has a timeout deadline
func HasRequestDeadline(r *http.Request) bool {
	_, ok := r.Context().Deadline()
	return ok
}
