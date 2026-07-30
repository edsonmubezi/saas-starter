package apierror

import (
	"encoding/json"
	"net/http"
)

// Error code constants for consistent client-side handling
const (
	ErrCodeBadRequest   = "BAD_REQUEST"
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeForbidden    = "FORBIDDEN"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeRateLimit    = "RATE_LIMIT_EXCEEDED"
)

// APIError is a structured error returned by all API endpoints.
type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
	TraceID string            `json:"trace_id,omitempty"`
}

// errorEnvelope wraps APIError in the standard response shape: { "error": { ... } }
type errorEnvelope struct {
	Error APIError `json:"error"`
}

// New creates a new APIError with the given code and message.
func New(code, message string) *APIError {
	return &APIError{Code: code, Message: message}
}

// WithDetails attaches field-level validation details (field → reason).
func (e *APIError) WithDetails(details map[string]string) *APIError {
	e.Details = details
	return e
}

// WithTraceID attaches a trace/request ID for log correlation.
func (e *APIError) WithTraceID(traceID string) *APIError {
	e.TraceID = traceID
	return e
}

// Write serialises the error as JSON and writes it to the response.
func Write(w http.ResponseWriter, statusCode int, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorEnvelope{Error: *err})
}

// --- Convenience helpers ---

// BadRequest writes a 400 Bad Request error.
func BadRequest(w http.ResponseWriter, message string) {
	Write(w, http.StatusBadRequest, New(ErrCodeBadRequest, message))
}

// Validation writes a 422 Unprocessable Entity with optional field details.
func Validation(w http.ResponseWriter, message string, details map[string]string) {
	Write(w, http.StatusUnprocessableEntity, New(ErrCodeValidation, message).WithDetails(details))
}

// Unauthorized writes a 401 Unauthorized error.
func Unauthorized(w http.ResponseWriter, message string) {
	Write(w, http.StatusUnauthorized, New(ErrCodeUnauthorized, message))
}

// Forbidden writes a 403 Forbidden error.
func Forbidden(w http.ResponseWriter, message string) {
	Write(w, http.StatusForbidden, New(ErrCodeForbidden, message))
}

// NotFound writes a 404 Not Found error.
func NotFound(w http.ResponseWriter, message string) {
	Write(w, http.StatusNotFound, New(ErrCodeNotFound, message))
}

// Conflict writes a 409 Conflict error.
func Conflict(w http.ResponseWriter, message string) {
	Write(w, http.StatusConflict, New(ErrCodeConflict, message))
}

// Internal writes a 500 Internal Server Error.
func Internal(w http.ResponseWriter, message string) {
	Write(w, http.StatusInternalServerError, New(ErrCodeInternal, message))
}

// RateLimit writes a 429 Too Many Requests error.
func RateLimit(w http.ResponseWriter) {
	Write(w, http.StatusTooManyRequests, New(ErrCodeRateLimit, "too many requests, please slow down"))
}
