package applog

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ResponseWriter wraps http.ResponseWriter to capture response size and status
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	size         int
	wroteHeader  bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Middleware creates an HTTP middleware for access logging
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate request ID if not present
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			w.Header().Set("X-Request-ID", requestID)
		}

		// Get trace context if available
		traceID := r.Header.Get("X-Trace-ID")
		spanID := r.Header.Get("X-Span-ID")

		// Wrap response writer to capture status and size
		rw := newResponseWriter(w)

		// Process request
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start)

		// Get tenant and user IDs from context (set by auth middleware)
		var tenantID, userID *int64
		if tid := r.Context().Value("tenant_id"); tid != nil {
			if t, ok := tid.(int64); ok {
				tenantID = &t
			}
		}
		if uid := r.Context().Value("user_id"); uid != nil {
			if u, ok := uid.(int64); ok {
				userID = &u
			}
		}

		// Get TLS version if available
		var tlsVersion string
		if r.TLS != nil {
			switch r.TLS.Version {
			case 0x0301:
				tlsVersion = "TLS 1.0"
			case 0x0302:
				tlsVersion = "TLS 1.1"
			case 0x0303:
				tlsVersion = "TLS 1.2"
			case 0x0304:
				tlsVersion = "TLS 1.3"
			}
		}

		// Build access log
		accessLog := &HTTPAccessLog{
			Method:       r.Method,
			Path:         r.URL.Path,
			Query:        r.URL.RawQuery,
			StatusCode:   rw.statusCode,
			Duration:     duration,
			RequestSize:  r.ContentLength,
			ResponseSize: int64(rw.size),
			IPAddress:    getClientIP(r),
			UserAgent:    r.UserAgent(),
			Referer:      r.Referer(),
			TenantID:     tenantID,
			UserID:       userID,
			RequestID:    requestID,
			TraceID:      traceID,
			SpanID:       spanID,
			ContentType:  r.Header.Get("Content-Type"),
			Protocol:     r.Proto,
			TLSVersion:   tlsVersion,
		}

		// Log asynchronously
		s.LogHTTPAccess(accessLog)

		// Also log errors to application logs
		if rw.statusCode >= 500 {
			s.NewLog().
				Error().
				HTTP().
				Message("Server error").
				Request(requestID, r.Method, r.URL.Path).
				Status(rw.statusCode).
				Duration(duration).
				IP(getClientIP(r)).
				Send()
		} else if rw.statusCode >= 400 {
			s.NewLog().
				Warn().
				HTTP().
				Message("Client error").
				Request(requestID, r.Method, r.URL.Path).
				Status(rw.statusCode).
				Duration(duration).
				IP(getClientIP(r)).
				Send()
		}
	})
}

// RecoveryMiddleware creates a recovery middleware that logs panics
func (s *Service) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID
				requestID := r.Header.Get("X-Request-ID")

				// Log the panic
				s.NewLog().
					Fatal().
					Category(CategorySystem).
					Messagef("Panic recovered: %v", err).
					Request(requestID, r.Method, r.URL.Path).
					IP(getClientIP(r)).
					UserAgent(r.UserAgent()).
					Field("panic_value", err).
					Send()

				// Return 500 error
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check common reverse proxy headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

// LogDatabaseQuery logs a database query
func (s *Service) LogDatabaseQuery(query string, duration time.Duration, err error) {
	builder := s.NewLog().
		Database().
		Message("Database query").
		Duration(duration).
		Field("query", truncateQuery(query, 500))

	if err != nil {
		builder.Error().Err(err)
	} else if duration > 100*time.Millisecond {
		// Log slow queries as warnings
		builder.Warn().Field("slow_query", true)
	} else {
		builder.Debug()
	}

	builder.Send()
}

// LogCacheOperation logs a cache operation
func (s *Service) LogCacheOperation(operation, key string, hit bool, duration time.Duration, err error) {
	builder := s.NewLog().
		Category(CategoryCache).
		Messagef("Cache %s", operation).
		Duration(duration).
		Field("key", key).
		Field("hit", hit)

	if err != nil {
		builder.Warn().Err(err)
	} else {
		builder.Debug()
	}

	builder.Send()
}

// LogExternalCall logs an external service call
func (s *Service) LogExternalCall(service, method, url string, statusCode int, duration time.Duration, err error) {
	builder := s.NewLog().
		Category(CategoryExternal).
		Messagef("External call to %s", service).
		Duration(duration).
		Field("service", service).
		Field("method", method).
		Field("url", url).
		Field("status_code", statusCode)

	if err != nil {
		builder.Error().Err(err)
	} else if statusCode >= 400 {
		builder.Warn()
	} else {
		builder.Info()
	}

	builder.Send()
}

// LogScheduledJob logs a scheduled job execution
func (s *Service) LogScheduledJob(jobName string, duration time.Duration, err error) {
	builder := s.NewLog().
		Category(CategoryScheduler).
		Messagef("Scheduled job: %s", jobName).
		Duration(duration).
		Field("job_name", jobName)

	if err != nil {
		builder.Error().Err(err)
	} else {
		builder.Info()
	}

	builder.Send()
}

func truncateQuery(query string, maxLen int) string {
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "..."
}
