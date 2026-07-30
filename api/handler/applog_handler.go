package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/edsonmubezi/myapp/internal/identity"
	"github.com/edsonmubezi/myapp/internal/platform/applog"
	"github.com/gorilla/mux"
)

// AppLogHandler handles application log API endpoints
type AppLogHandler struct {
	service *applog.Service
}

// NewAppLogHandler creates a new application log handler
func NewAppLogHandler(svc *applog.Service) *AppLogHandler {
	return &AppLogHandler{
		service: svc,
	}
}

// GetApplicationLogs retrieves application logs with filters
// @Summary      Get Application Logs
// @Description  Returns application logs with optional filters and pagination
// @Tags         Logs
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "Page number (default: 1)"
// @Param        page_size query int false "Page size (default: 50, max: 100)"
// @Param        level query string false "Log level filter (DEBUG, INFO, WARN, ERROR, FATAL)"
// @Param        category query string false "Log category filter (http, database, cache, auth, business, scheduler, external, system)"
// @Param        from_date query string false "Start date filter (RFC3339 format)"
// @Param        to_date query string false "End date filter (RFC3339 format)"
// @Param        search query string false "Search in message"
// @Param        trace_id query string false "Filter by trace ID"
// @Param        request_id query string false "Filter by request ID"
// @Success      200 {object} SuccessResponse "Application logs"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/logs/application [get]
func (h *AppLogHandler) GetApplicationLogs(w http.ResponseWriter, r *http.Request) {
	p, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// Parse pagination
	page := 1
	pageSize := 50
	if pg, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && pg > 0 {
		page = pg
	}
	if ps, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	// Build filters
	filters := applog.LogFilters{
		TenantID: &p.OrganizationID,
	}

	if level := r.URL.Query().Get("level"); level != "" {
		lvl := applog.LogLevel(level)
		filters.Level = &lvl
	}
	if category := r.URL.Query().Get("category"); category != "" {
		cat := applog.LogCategory(category)
		filters.Category = &cat
	}
	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = &search
	}
	if traceID := r.URL.Query().Get("trace_id"); traceID != "" {
		filters.TraceID = &traceID
	}
	if requestID := r.URL.Query().Get("request_id"); requestID != "" {
		filters.RequestID = &requestID
	}
	if fromDateStr := r.URL.Query().Get("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse(time.RFC3339, fromDateStr); err == nil {
			filters.FromDate = &fromDate
		}
	}
	if toDateStr := r.URL.Query().Get("to_date"); toDateStr != "" {
		if toDate, err := time.Parse(time.RFC3339, toDateStr); err == nil {
			filters.ToDate = &toDate
		}
	}

	logs, total, err := h.service.GetLogs(r.Context(), filters, page, pageSize)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to retrieve application logs", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Application logs retrieved", map[string]interface{}{
		"logs":        logs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetAccessLogs retrieves HTTP access logs
// @Summary      Get Access Logs
// @Description  Returns HTTP access logs with optional filters and pagination
// @Tags         Logs
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "Page number (default: 1)"
// @Param        page_size query int false "Page size (default: 50, max: 100)"
// @Param        method query string false "HTTP method filter (GET, POST, PUT, DELETE, etc.)"
// @Param        path query string false "Path filter (supports regex)"
// @Param        from_date query string false "Start date filter (RFC3339 format)"
// @Param        to_date query string false "End date filter (RFC3339 format)"
// @Param        request_id query string false "Filter by request ID"
// @Success      200 {object} SuccessResponse "Access logs"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/logs/access [get]
func (h *AppLogHandler) GetAccessLogs(w http.ResponseWriter, r *http.Request) {
	p, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// Parse pagination
	page := 1
	pageSize := 50
	if pg, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && pg > 0 {
		page = pg
	}
	if ps, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	// Build filters
	filters := applog.LogFilters{
		TenantID: &p.OrganizationID,
	}

	if method := r.URL.Query().Get("method"); method != "" {
		filters.Method = &method
	}
	if path := r.URL.Query().Get("path"); path != "" {
		filters.Path = &path
	}
	if requestID := r.URL.Query().Get("request_id"); requestID != "" {
		filters.RequestID = &requestID
	}
	if fromDateStr := r.URL.Query().Get("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse(time.RFC3339, fromDateStr); err == nil {
			filters.FromDate = &fromDate
		}
	}
	if toDateStr := r.URL.Query().Get("to_date"); toDateStr != "" {
		if toDate, err := time.Parse(time.RFC3339, toDateStr); err == nil {
			filters.ToDate = &toDate
		}
	}

	logs, total, err := h.service.GetAccessLogs(r.Context(), filters, page, pageSize)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to retrieve access logs", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Access logs retrieved", map[string]interface{}{
		"logs":        logs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetLogByID retrieves a single application log by ID
// @Summary      Get Log by ID
// @Description  Returns a specific application log entry
// @Tags         Logs
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Log ID"
// @Success      200 {object} SuccessResponse "Log entry"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/logs/application/{id} [get]
func (h *AppLogHandler) GetLogByID(w http.ResponseWriter, r *http.Request) {
	_, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Missing log ID", nil)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid log ID", nil)
		return
	}

	log, err := h.service.GetLogByID(r.Context(), id)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to retrieve log", nil)
		return
	}
	if log == nil {
		SendJSONResponse(w, http.StatusNotFound, "Log not found", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Log retrieved", log)
}

// GetLogStats retrieves aggregated log statistics
// @Summary      Get Log Statistics
// @Description  Returns aggregated statistics for application logs
// @Tags         Logs
// @Security     BearerAuth
// @Produce      json
// @Param        hours query int false "Hours to look back (default: 24, max: 168)"
// @Success      200 {object} SuccessResponse "Log statistics"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/logs/stats [get]
func (h *AppLogHandler) GetLogStats(w http.ResponseWriter, r *http.Request) {
	p, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	hours := 24
	if h, err := strconv.Atoi(r.URL.Query().Get("hours")); err == nil && h > 0 && h <= 168 {
		hours = h
	}

	stats, err := h.service.GetLogStats(r.Context(), &p.OrganizationID, hours)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to retrieve log statistics", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Log statistics retrieved", stats)
}

// GetLogLevels returns available log levels
// @Summary      Get Log Levels
// @Description  Returns the list of available log levels
// @Tags         Logs
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} SuccessResponse "Log levels"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/logs/levels [get]
func (h *AppLogHandler) GetLogLevels(w http.ResponseWriter, r *http.Request) {
	_, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	levels := []map[string]string{
		{"value": string(applog.LevelDebug), "label": "Debug", "description": "Detailed debugging information"},
		{"value": string(applog.LevelInfo), "label": "Info", "description": "General informational messages"},
		{"value": string(applog.LevelWarn), "label": "Warning", "description": "Warning conditions"},
		{"value": string(applog.LevelError), "label": "Error", "description": "Error conditions"},
		{"value": string(applog.LevelFatal), "label": "Fatal", "description": "Critical errors causing shutdown"},
	}

	SendJSONResponse(w, http.StatusOK, "Log levels retrieved", map[string]interface{}{
		"levels": levels,
	})
}

// GetLogCategories returns available log categories
// @Summary      Get Log Categories
// @Description  Returns the list of available log categories
// @Tags         Logs
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} SuccessResponse "Log categories"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/logs/categories [get]
func (h *AppLogHandler) GetLogCategories(w http.ResponseWriter, r *http.Request) {
	_, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	categories := []map[string]string{
		{"value": string(applog.CategoryHTTP), "label": "HTTP", "description": "HTTP request/response logs"},
		{"value": string(applog.CategoryDatabase), "label": "Database", "description": "Database operation logs"},
		{"value": string(applog.CategoryCache), "label": "Cache", "description": "Cache operation logs"},
		{"value": string(applog.CategoryAuth), "label": "Authentication", "description": "Authentication and authorization logs"},
		{"value": string(applog.CategoryBusiness), "label": "Business", "description": "Business logic logs"},
		{"value": string(applog.CategoryScheduler), "label": "Scheduler", "description": "Scheduled job logs"},
		{"value": string(applog.CategoryExternal), "label": "External", "description": "External service call logs"},
		{"value": string(applog.CategorySystem), "label": "System", "description": "System-level logs"},
	}

	SendJSONResponse(w, http.StatusOK, "Log categories retrieved", map[string]interface{}{
		"categories": categories,
	})
}
