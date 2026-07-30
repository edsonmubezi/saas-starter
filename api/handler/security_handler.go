package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/edsonmubezi/myapp/internal/identity"
	"github.com/edsonmubezi/myapp/internal/platform/security"
	"github.com/gorilla/mux"
)

// SecurityHandler handles security-related API endpoints
type SecurityHandler struct {
	securityService *security.Service
}

// NewSecurityHandler creates a new security handler
func NewSecurityHandler(svc *security.Service) *SecurityHandler {
	return &SecurityHandler{securityService: svc}
}

// GetSecurityDashboard returns security metrics for the organization
// @Summary      Get Security Dashboard
// @Description  Returns security event statistics for the organization dashboard
// @Tags         Security
// @Security     BearerAuth
// @Produce      json
// @Param        hours query int false "Number of hours to look back (default: 24)" example(24)
// @Success      200 {object} SuccessResponse "Security dashboard data"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/security/dashboard [get]
func (h *SecurityHandler) GetSecurityDashboard(w http.ResponseWriter, r *http.Request) {
	if h.securityService == nil {
		SendJSONResponse(w, http.StatusServiceUnavailable, "Security service not configured", nil)
		return
	}

	p, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// Get hours parameter (default 24)
	hoursStr := r.URL.Query().Get("hours")
	hours := 24
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 720 { // max 30 days
			hours = h
		}
	}

	dashboard, err := h.securityService.GetSecurityDashboard(r.Context(), p.OrganizationID, hours)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to retrieve security dashboard", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Security dashboard retrieved", dashboard)
}

// GetSecurityEvents returns security events with filtering and pagination
// @Summary      Get Security Events
// @Description  Returns security events for the organization with optional filters
// @Tags         Security
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "Page number (default: 1)" example(1)
// @Param        page_size query int false "Page size (default: 20, max: 100)" example(20)
// @Param        severity query string false "Filter by severity (INFO, WARNING, HIGH, CRITICAL)" example(HIGH)
// @Param        category query string false "Filter by category (authentication, authorization, data_access, anomaly)" example(authentication)
// @Param        event_type query string false "Filter by event type" example(LOGIN_FAILED)
// @Param        ip_address query string false "Filter by IP address" example(192.168.1.1)
// @Param        actor_id query int false "Filter by actor/user ID" example(1)
// @Param        from query string false "Filter from date (RFC3339)" example(2024-01-01T00:00:00Z)
// @Param        to query string false "Filter to date (RFC3339)" example(2024-12-31T23:59:59Z)
// @Success      200 {object} SuccessResponse "Security events"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/security/events [get]
func (h *SecurityHandler) GetSecurityEvents(w http.ResponseWriter, r *http.Request) {
	if h.securityService == nil {
		SendJSONResponse(w, http.StatusServiceUnavailable, "Security service not configured", nil)
		return
	}

	p, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// Parse pagination
	page := 1
	pageSize := 20
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	// Build filters
	filters := security.SecurityFilters{
		TenantID: &p.OrganizationID,
	}

	if sev := r.URL.Query().Get("severity"); sev != "" {
		severity := security.Severity(sev)
		filters.Severity = &severity
	}
	if cat := r.URL.Query().Get("category"); cat != "" {
		category := security.Category(cat)
		filters.Category = &category
	}
	if eventType := r.URL.Query().Get("event_type"); eventType != "" {
		filters.EventType = &eventType
	}
	if ip := r.URL.Query().Get("ip_address"); ip != "" {
		filters.IPAddress = &ip
	}
	if actorID := r.URL.Query().Get("actor_id"); actorID != "" {
		if id, err := strconv.ParseInt(actorID, 10, 64); err == nil {
			filters.ActorID = &id
		}
	}
	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filters.FromDate = &t
		}
	}
	if to := r.URL.Query().Get("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			filters.ToDate = &t
		}
	}

	events, total, err := h.securityService.GetEvents(r.Context(), filters, page, pageSize)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to retrieve security events", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Security events retrieved", map[string]interface{}{
		"events":     events,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetSecurityEventByID returns a single security event by ID
// @Summary      Get Security Event by ID
// @Description  Returns a single security event by its database ID
// @Tags         Security
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Security event ID"
// @Success      200 {object} SuccessResponse "Security event"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Event not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/security/events/{id} [get]
func (h *SecurityHandler) GetSecurityEventByID(w http.ResponseWriter, r *http.Request) {
	if h.securityService == nil {
		SendJSONResponse(w, http.StatusServiceUnavailable, "Security service not configured", nil)
		return
	}

	_, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Missing event ID", nil)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid event ID", nil)
		return
	}

	event, err := h.securityService.GetByID(r.Context(), id)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to retrieve security event", nil)
		return
	}
	if event == nil {
		SendJSONResponse(w, http.StatusNotFound, "Security event not found", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Security event retrieved", event)
}
