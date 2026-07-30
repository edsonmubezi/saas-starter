package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/edsonmubezi/myapp/internal/identity"
	"github.com/edsonmubezi/myapp/internal/platform/audit"
	"github.com/gorilla/mux"
)

// AuditHandler handles audit log API requests
type AuditHandler struct {
	auditService *audit.PostgresService
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *audit.PostgresService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

// AuditEventResponse represents an audit event in API responses
type AuditEventResponse struct {
	ID           string                 `json:"id"`
	AuditID      string                 `json:"audit_id"`
	Timestamp    time.Time              `json:"timestamp"`
	ActorType    string                 `json:"actor_type"`
	ActorID      *int64                 `json:"actor_id,omitempty"`
	ActorEmail   string                 `json:"actor_email,omitempty"`
	ActorName    string                 `json:"actor_name,omitempty"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *int64                 `json:"resource_id,omitempty"`
	ResourceName string                 `json:"resource_name,omitempty"`
	Severity     string                 `json:"severity"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	BeforeState  map[string]interface{} `json:"before_state,omitempty"`
	AfterState   map[string]interface{} `json:"after_state,omitempty"`
	Changes      []audit.FieldChange    `json:"changes,omitempty"`
}

// AuditListResponse represents paginated audit events
type AuditListResponse struct {
	Events     []AuditEventResponse `json:"events"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// GetAuditEvents handles GET /api/audit/events
// @Summary      List Audit Events
// @Description  Retrieves paginated audit events with optional filters
// @Tags         Audit
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        page_size query int false "Page size" default(20)
// @Param        resource_type query string false "Filter by resource type"
// @Param        resource_id query int false "Filter by resource ID"
// @Param        actor_id query int false "Filter by actor ID"
// @Param        action query string false "Filter by action"
// @Param        severity query string false "Filter by severity"
// @Param        from query string false "Filter from date (RFC3339)"
// @Param        to query string false "Filter to date (RFC3339)"
// @Success      200 {object} AuditListResponse
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/audit/events [get]
func (h *AuditHandler) GetAuditEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	auth, err := identity.Require(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := audit.AuditFilters{
		TenantID: auth.OrganizationID,
	}

	if rt := r.URL.Query().Get("resource_type"); rt != "" {
		filters.ResourceType = &rt
	}
	if rid := r.URL.Query().Get("resource_id"); rid != "" {
		if id, err := strconv.ParseInt(rid, 10, 64); err == nil {
			filters.ResourceID = &id
		}
	}
	if aid := r.URL.Query().Get("actor_id"); aid != "" {
		if id, err := strconv.ParseInt(aid, 10, 64); err == nil {
			filters.ActorID = &id
		}
	}
	if action := r.URL.Query().Get("action"); action != "" {
		filters.Action = &action
	}
	if sev := r.URL.Query().Get("severity"); sev != "" {
		severity := audit.Severity(sev)
		filters.Severity = &severity
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

	events, total, err := h.auditService.GetEvents(ctx, filters, page, pageSize)
	if err != nil {
		http.Error(w, "Failed to fetch audit events", http.StatusInternalServerError)
		return
	}

	// Log audit access (audit-the-audit)
	go func() {
		_ = h.auditService.LogAuditAccess(ctx, auth.UserID, auth.OrganizationID, filters, len(events), identity.IPAddress(ctx))
	}()

	responseEvents := make([]AuditEventResponse, len(events))
	for i, e := range events {
		responseEvents[i] = AuditEventResponse{
			ID:           e.AuditID,
			AuditID:      e.AuditID,
			Timestamp:    e.Timestamp,
			ActorType:    string(e.ActorType),
			ActorID:      e.ActorID,
			ActorEmail:   e.ActorEmail,
			ActorName:    e.ActorName,
			Action:       e.Action,
			ResourceType: e.ResourceType,
			ResourceID:   e.ResourceID,
			ResourceName: e.ResourceName,
			Severity:     string(e.Severity),
			IPAddress:    e.IPAddress,
			UserAgent:    e.UserAgent,
			BeforeState:  e.BeforeState,
			AfterState:   e.AfterState,
			Changes:      e.Changes,
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	response := AuditListResponse{
		Events:     responseEvents,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAuditEventByID handles GET /api/audit/events/{id}
// @Summary      Get Audit Event by ID
// @Description  Retrieves a single audit event by its ID
// @Tags         Audit
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Audit event ID"
// @Success      200 {object} AuditEventResponse
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/audit/events/{id} [get]
func (h *AuditHandler) GetAuditEventByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	auth, err := identity.Require(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	event, err := h.auditService.GetByID(ctx, auth.OrganizationID, id)
	if err != nil {
		http.Error(w, "Failed to fetch audit event", http.StatusInternalServerError)
		return
	}
	if event == nil {
		http.Error(w, "Audit event not found", http.StatusNotFound)
		return
	}

	response := AuditEventResponse{
		ID:           event.AuditID,
		AuditID:      event.AuditID,
		Timestamp:    event.Timestamp,
		ActorType:    string(event.ActorType),
		ActorID:      event.ActorID,
		ActorEmail:   event.ActorEmail,
		ActorName:    event.ActorName,
		Action:       event.Action,
		ResourceType: event.ResourceType,
		ResourceID:   event.ResourceID,
		ResourceName: event.ResourceName,
		Severity:     string(event.Severity),
		IPAddress:    event.IPAddress,
		UserAgent:    event.UserAgent,
		BeforeState:  event.BeforeState,
		AfterState:   event.AfterState,
		Changes:      event.Changes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetResourceHistory handles GET /api/audit/resources/{type}/{id}/history
// @Summary      Get Resource History
// @Description  Retrieves the audit history for a specific resource
// @Tags         Audit
// @Security     BearerAuth
// @Produce      json
// @Param        type path string true "Resource type (e.g., User, Role)"
// @Param        id path int true "Resource ID"
// @Success      200 {array} AuditEventResponse
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/audit/resources/{type}/{id}/history [get]
func (h *AuditHandler) GetResourceHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	auth, err := identity.Require(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	resourceType := vars["type"]
	resourceID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid resource ID", http.StatusBadRequest)
		return
	}

	events, err := h.auditService.GetResourceHistory(ctx, auth.OrganizationID, resourceType, resourceID)
	if err != nil {
		http.Error(w, "Failed to fetch resource history", http.StatusInternalServerError)
		return
	}

	responseEvents := make([]AuditEventResponse, len(events))
	for i, e := range events {
		responseEvents[i] = AuditEventResponse{
			ID:           e.AuditID,
			AuditID:      e.AuditID,
			Timestamp:    e.Timestamp,
			ActorType:    string(e.ActorType),
			ActorID:      e.ActorID,
			ActorEmail:   e.ActorEmail,
			ActorName:    e.ActorName,
			Action:       e.Action,
			ResourceType: e.ResourceType,
			ResourceID:   e.ResourceID,
			ResourceName: e.ResourceName,
			Severity:     string(e.Severity),
			IPAddress:    e.IPAddress,
			UserAgent:    e.UserAgent,
			BeforeState:  e.BeforeState,
			AfterState:   e.AfterState,
			Changes:      e.Changes,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseEvents)
}

// GetUserActivity handles GET /api/audit/users/{id}/activity
// @Summary      Get User Activity
// @Description  Retrieves audit events for a specific user's actions
// @Tags         Audit
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "User ID"
// @Param        from query string false "From date (RFC3339)" default(7 days ago)
// @Param        to query string false "To date (RFC3339)" default(now)
// @Success      200 {array} AuditEventResponse
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/audit/users/{id}/activity [get]
func (h *AuditHandler) GetUserActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	auth, err := identity.Require(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	userID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	to := time.Now().UTC()
	from := to.AddDate(0, 0, -7)

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = t
		}
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = t
		}
	}

	events, err := h.auditService.GetUserActivity(ctx, auth.OrganizationID, userID, from, to)
	if err != nil {
		http.Error(w, "Failed to fetch user activity", http.StatusInternalServerError)
		return
	}

	responseEvents := make([]AuditEventResponse, len(events))
	for i, e := range events {
		responseEvents[i] = AuditEventResponse{
			ID:           e.AuditID,
			AuditID:      e.AuditID,
			Timestamp:    e.Timestamp,
			ActorType:    string(e.ActorType),
			ActorID:      e.ActorID,
			ActorEmail:   e.ActorEmail,
			ActorName:    e.ActorName,
			Action:       e.Action,
			ResourceType: e.ResourceType,
			ResourceID:   e.ResourceID,
			ResourceName: e.ResourceName,
			Severity:     string(e.Severity),
			IPAddress:    e.IPAddress,
			UserAgent:    e.UserAgent,
			BeforeState:  e.BeforeState,
			AfterState:   e.AfterState,
			Changes:      e.Changes,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseEvents)
}

// GetAuditActions returns available audit actions
// @Summary      Get Audit Actions
// @Description  Returns the list of available audit action types
// @Tags         Audit
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Router       /api/audit/metadata/actions [get]
func (h *AuditHandler) GetAuditActions(w http.ResponseWriter, r *http.Request) {
	actions := []map[string]interface{}{
		{"value": audit.ActionLoginSuccess, "label": "Login Success", "color": "text-cyan-400"},
		{"value": audit.ActionLoginFailed, "label": "Login Failed", "color": "text-rose-400"},
		{"value": audit.ActionLogout, "label": "Logout", "color": "text-white/60"},
		{"value": audit.ActionPasswordReset, "label": "Password Reset", "color": "text-amber-400"},
		{"value": audit.ActionMFAEnabled, "label": "MFA Enabled", "color": "text-emerald-400"},
		{"value": audit.ActionMFADisabled, "label": "MFA Disabled", "color": "text-rose-400"},
		{"value": audit.ActionCreated, "label": "Created", "color": "text-emerald-400"},
		{"value": audit.ActionUpdated, "label": "Updated", "color": "text-blue-400"},
		{"value": audit.ActionDeleted, "label": "Deleted", "color": "text-rose-400"},
		{"value": audit.ActionViewed, "label": "Viewed", "color": "text-white/60"},
		{"value": audit.ActionRoleGranted, "label": "Role Granted", "color": "text-emerald-400"},
		{"value": audit.ActionRoleRevoked, "label": "Role Revoked", "color": "text-rose-400"},
		{"value": audit.ActionPermGranted, "label": "Permission Granted", "color": "text-emerald-400"},
		{"value": audit.ActionPermRevoked, "label": "Permission Revoked", "color": "text-rose-400"},
		{"value": audit.ActionExported, "label": "Exported", "color": "text-amber-400"},
		{"value": audit.ActionImported, "label": "Imported", "color": "text-blue-400"},
		{"value": audit.ActionArchived, "label": "Archived", "color": "text-white/60"},
		{"value": audit.ActionRestored, "label": "Restored", "color": "text-emerald-400"},
		{"value": audit.ActionApproved, "label": "Approved", "color": "text-emerald-400"},
		{"value": audit.ActionRejected, "label": "Rejected", "color": "text-rose-400"},
		{"value": audit.ActionCompleted, "label": "Completed", "color": "text-emerald-400"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"actions": actions,
	})
}

// GetAuditSeverities returns available audit severity levels
// @Summary      Get Audit Severities
// @Description  Returns the list of available audit severity levels
// @Tags         Audit
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Router       /api/audit/metadata/severities [get]
func (h *AuditHandler) GetAuditSeverities(w http.ResponseWriter, r *http.Request) {
	severities := []map[string]interface{}{
		{"value": string(audit.SeverityLow), "label": "Low", "color": "text-blue-400"},
		{"value": string(audit.SeverityMedium), "label": "Medium", "color": "text-amber-400"},
		{"value": string(audit.SeverityHigh), "label": "High", "color": "text-orange-400"},
		{"value": string(audit.SeverityCritical), "label": "Critical", "color": "text-rose-400"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"severities": severities,
	})
}

// GetAuditResourceTypes returns available audit resource types
// @Summary      Get Audit Resource Types
// @Description  Returns the list of available resource types for audit events
// @Tags         Audit
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Router       /api/audit/metadata/resource-types [get]
func (h *AuditHandler) GetAuditResourceTypes(w http.ResponseWriter, r *http.Request) {
	resourceTypes := []map[string]interface{}{
		{"value": audit.ResourceUser, "label": "User"},
		{"value": audit.ResourceRole, "label": "Role"},
		{"value": audit.ResourcePermission, "label": "Permission"},
		{"value": audit.ResourceOrganization, "label": "Organization"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"resource_types": resourceTypes,
	})
}
