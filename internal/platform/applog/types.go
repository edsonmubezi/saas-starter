package applog

import (
	"time"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

// LogCategory represents the category of log entry
type LogCategory string

const (
	CategoryHTTP      LogCategory = "http"
	CategoryDatabase  LogCategory = "database"
	CategoryCache     LogCategory = "cache"
	CategoryAuth      LogCategory = "auth"
	CategoryBusiness  LogCategory = "business"
	CategoryScheduler LogCategory = "scheduler"
	CategoryExternal  LogCategory = "external"
	CategorySystem    LogCategory = "system"
)

// ApplicationLog represents an application log entry stored in PostgreSQL
type ApplicationLog struct {
	ID         int64                  `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	Level      LogLevel               `json:"level"`
	Category   LogCategory            `json:"category"`
	Message    string                 `json:"message"`
	Service    string                 `json:"service"`
	TraceID    string                 `json:"trace_id,omitempty"`
	SpanID     string                 `json:"span_id,omitempty"`
	TenantID   *int64                 `json:"tenant_id,omitempty"`
	UserID     *int64                 `json:"user_id,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Duration   time.Duration          `json:"duration,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Error      *ErrorDetails          `json:"error,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Stack      string                 `json:"stack,omitempty"`
}

// ErrorDetails contains error-specific information
type ErrorDetails struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	StackTrace string `json:"stack_trace,omitempty"`
	Code       string `json:"code,omitempty"`
}

// HTTPAccessLog represents an HTTP request/response log
type HTTPAccessLog struct {
	ID           int64         `json:"id"`
	Timestamp    time.Time     `json:"timestamp"`
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	Query        string        `json:"query,omitempty"`
	StatusCode   int           `json:"status_code"`
	Duration     time.Duration `json:"duration"`
	RequestSize  int64         `json:"request_size"`
	ResponseSize int64         `json:"response_size"`
	IPAddress    string        `json:"ip_address"`
	UserAgent    string        `json:"user_agent"`
	Referer      string        `json:"referer,omitempty"`
	TenantID     *int64        `json:"tenant_id,omitempty"`
	UserID       *int64        `json:"user_id,omitempty"`
	UserEmail    *string       `json:"user_email,omitempty"`
	RequestID    string        `json:"request_id"`
	TraceID      string        `json:"trace_id,omitempty"`
	SpanID       string        `json:"span_id,omitempty"`
	Error        string        `json:"error,omitempty"`
	ContentType  string        `json:"content_type,omitempty"`
	Protocol     string        `json:"protocol"`
	TLSVersion   string        `json:"tls_version,omitempty"`
	GeoLocation  *GeoLocation  `json:"geo_location,omitempty"`
}

// GeoLocation contains IP geolocation data
type GeoLocation struct {
	Country   string  `json:"country,omitempty"`
	City      string  `json:"city,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

// LogFilters for querying application logs
type LogFilters struct {
	Level     *LogLevel
	Category  *LogCategory
	TenantID  *int64
	UserID    *int64
	TraceID   *string
	RequestID *string
	Method    *string
	Path      *string
	FromDate  *time.Time
	ToDate    *time.Time
	Search    *string // Full-text search in message
}

// AccessLogFilters for querying HTTP access logs
type AccessLogFilters struct {
	TenantID  *int64
	UserID    *int64
	TraceID   *string
	RequestID *string
	Method    *string
	Path      *string
	FromDate  *time.Time
	ToDate    *time.Time
}

// PathErrorStat represents error counts for a specific endpoint path
type PathErrorStat struct {
	Path  string `json:"path"`
	Count int64  `json:"count"`
}

// HourlyErrorStat represents error and warning counts for a single hour
type HourlyErrorStat struct {
	Hour   time.Time `json:"hour"`
	Errors int64     `json:"errors"`
	Warns  int64     `json:"warns"`
}

// LogStatsData contains raw aggregate statistics returned by the repository
type LogStatsData struct {
	ByLevel       map[string]int64
	ByCategory    map[string]int64
	TopErrorPaths []PathErrorStat
	ErrorTrend    []HourlyErrorStat
}

// LogStats represents aggregated log statistics
type LogStats struct {
	TotalLogs     int64             `json:"total_logs"`
	ByLevel       map[string]int64  `json:"by_level"`
	ByCategory    map[string]int64  `json:"by_category"`
	ErrorRate     float64           `json:"error_rate"`
	AvgDuration   float64           `json:"avg_duration_ms"`
	TopEndpoints  []EndpointStats   `json:"top_endpoints"`
	TopErrors     []ErrorStats      `json:"top_errors"`
	TopErrorPaths []PathErrorStat   `json:"top_error_paths"`
	ErrorTrend    []HourlyErrorStat `json:"error_trend"`
}

// EndpointStats represents stats for an endpoint
type EndpointStats struct {
	Method      string  `json:"method"`
	Path        string  `json:"path"`
	Count       int64   `json:"count"`
	AvgDuration float64 `json:"avg_duration_ms"`
	ErrorRate   float64 `json:"error_rate"`
}

// ErrorStats represents stats for errors
type ErrorStats struct {
	Message  string    `json:"message"`
	Count    int64     `json:"count"`
	LastSeen time.Time `json:"last_seen"`
}
