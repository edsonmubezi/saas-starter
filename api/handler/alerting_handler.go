package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/edsonmubezi/myapp/internal/identity"
	"github.com/edsonmubezi/myapp/internal/platform/alerting"
	"github.com/gorilla/mux"
)

// AlertingHandler handles alerting configuration API endpoints
type AlertingHandler struct {
	repo    *alerting.Repository
	service *alerting.Service
}

// NewAlertingHandler creates a new alerting handler
func NewAlertingHandler(repo *alerting.Repository, svc *alerting.Service) *AlertingHandler {
	return &AlertingHandler{
		repo:    repo,
		service: svc,
	}
}

// CreateAlertConfig creates a new alert configuration
// @Summary      Create Alert Configuration
// @Description  Creates a new alert configuration for the organization
// @Tags         Alerting
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        config body alerting.CreateAlertConfigInput true "Alert configuration"
// @Success      201 {object} SuccessResponse "Alert configuration created"
// @Failure      400 {object} ErrorResponse "Invalid input"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/alerting/configs [post]
func (h *AlertingHandler) CreateAlertConfig(w http.ResponseWriter, r *http.Request) {
	p, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var input alerting.CreateAlertConfigInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Validate required fields
	if input.Name == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Name is required", nil)
		return
	}
	if len(input.Severities) == 0 {
		SendJSONResponse(w, http.StatusBadRequest, "At least one severity level is required", nil)
		return
	}

	config, err := h.repo.CreateConfig(r.Context(), p.OrganizationID, p.UserID, &input)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to create alert configuration", nil)
		return
	}

	SendJSONResponse(w, http.StatusCreated, "Alert configuration created", config)
}

// GetAlertConfigs retrieves all alert configurations for the organization
// @Summary      Get Alert Configurations
// @Description  Returns all alert configurations for the organization
// @Tags         Alerting
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "Page number (default: 1)"
// @Param        page_size query int false "Page size (default: 20, max: 100)"
// @Success      200 {object} SuccessResponse "Alert configurations"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/alerting/configs [get]
func (h *AlertingHandler) GetAlertConfigs(w http.ResponseWriter, r *http.Request) {
	p, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// Parse pagination
	page := 1
	pageSize := 20
	if pg, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && pg > 0 {
		page = pg
	}
	if ps, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	configs, total, err := h.repo.GetConfigsByTenant(r.Context(), p.OrganizationID, page, pageSize)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to retrieve alert configurations", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Alert configurations retrieved", map[string]interface{}{
		"configs":     configs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetAlertConfigByID retrieves a specific alert configuration
// @Summary      Get Alert Configuration by ID
// @Description  Returns a specific alert configuration
// @Tags         Alerting
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Alert configuration ID"
// @Success      200 {object} SuccessResponse "Alert configuration"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/alerting/configs/{id} [get]
func (h *AlertingHandler) GetAlertConfigByID(w http.ResponseWriter, r *http.Request) {
	p, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Missing configuration ID", nil)
		return
	}

	config, err := h.repo.GetConfigByID(r.Context(), id, p.OrganizationID)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to retrieve alert configuration", nil)
		return
	}
	if config == nil {
		SendJSONResponse(w, http.StatusNotFound, "Alert configuration not found", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Alert configuration retrieved", config)
}

// UpdateAlertConfig updates an alert configuration
// @Summary      Update Alert Configuration
// @Description  Updates an existing alert configuration
// @Tags         Alerting
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Alert configuration ID"
// @Param        config body alerting.UpdateAlertConfigInput true "Alert configuration update"
// @Success      200 {object} SuccessResponse "Alert configuration updated"
// @Failure      400 {object} ErrorResponse "Invalid input"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/alerting/configs/{id} [put]
func (h *AlertingHandler) UpdateAlertConfig(w http.ResponseWriter, r *http.Request) {
	p, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Missing configuration ID", nil)
		return
	}

	var input alerting.UpdateAlertConfigInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	config, err := h.repo.UpdateConfig(r.Context(), id, p.OrganizationID, &input)
	if err != nil {
		if err.Error() == "alert config not found" {
			SendJSONResponse(w, http.StatusNotFound, "Alert configuration not found", nil)
			return
		}
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to update alert configuration", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Alert configuration updated", config)
}

// DeleteAlertConfig deletes an alert configuration
// @Summary      Delete Alert Configuration
// @Description  Deletes an alert configuration
// @Tags         Alerting
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Alert configuration ID"
// @Success      200 {object} SuccessResponse "Alert configuration deleted"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/alerting/configs/{id} [delete]
func (h *AlertingHandler) DeleteAlertConfig(w http.ResponseWriter, r *http.Request) {
	p, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Missing configuration ID", nil)
		return
	}

	if err := h.repo.DeleteConfig(r.Context(), id, p.OrganizationID); err != nil {
		if err.Error() == "alert config not found" {
			SendJSONResponse(w, http.StatusNotFound, "Alert configuration not found", nil)
			return
		}
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to delete alert configuration", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Alert configuration deleted", nil)
}

// GetAlertHistory retrieves alert history for the organization
// @Summary      Get Alert History
// @Description  Returns alert history for the organization
// @Tags         Alerting
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "Page number (default: 1)"
// @Param        page_size query int false "Page size (default: 20, max: 100)"
// @Success      200 {object} SuccessResponse "Alert history"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/alerting/history [get]
func (h *AlertingHandler) GetAlertHistory(w http.ResponseWriter, r *http.Request) {
	p, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// Parse pagination
	page := 1
	pageSize := 20
	if pg, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && pg > 0 {
		page = pg
	}
	if ps, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	history, total, err := h.repo.GetAlertHistory(r.Context(), p.OrganizationID, page, pageSize)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to retrieve alert history", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Alert history retrieved", map[string]interface{}{
		"history":     history,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetConfiguredChannels returns the list of configured alert channels
// @Summary      Get Configured Channels
// @Description  Returns the list of alert channels that are configured and available
// @Tags         Alerting
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} SuccessResponse "Configured channels"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/alerting/channels [get]
func (h *AlertingHandler) GetConfiguredChannels(w http.ResponseWriter, r *http.Request) {
	_, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	channels := h.service.GetConfiguredChannels()
	channelNames := make([]string, len(channels))
	for i, ch := range channels {
		channelNames[i] = string(ch)
	}

	SendJSONResponse(w, http.StatusOK, "Configured channels retrieved", map[string]interface{}{
		"channels": channelNames,
		"enabled":  h.service.IsEnabled(),
	})
}

// TestAlertChannel sends a test alert through a specific channel
// @Summary      Test Alert Channel
// @Description  Sends a test alert through a specific channel
// @Tags         Alerting
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        channel path string true "Channel type (email, slack, sms, teams, webhook)"
// @Success      200 {object} SuccessResponse "Test alert sent"
// @Failure      400 {object} ErrorResponse "Invalid channel"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Failed to send test alert"
// @Router       /api/alerting/test/{channel} [post]
func (h *AlertingHandler) TestAlertChannel(w http.ResponseWriter, r *http.Request) {
	p, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	channelStr := vars["channel"]

	channelType := alerting.ChannelType(channelStr)

	// Validate channel type
	validChannels := map[alerting.ChannelType]bool{
		alerting.ChannelEmail:   true,
		alerting.ChannelSlack:   true,
		alerting.ChannelSMS:     true,
		alerting.ChannelTeams:   true,
		alerting.ChannelWebhook: true,
	}

	if !validChannels[channelType] {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid channel type", nil)
		return
	}

	// Check if channel is configured
	if _, exists := h.service.GetChannel(channelType); !exists {
		SendJSONResponse(w, http.StatusBadRequest, "Channel not configured", nil)
		return
	}

	// Create test alert
	testAlert := alerting.CreateSecurityAlert(
		"TEST_ALERT",
		"INFO",
		p.OrganizationID,
		map[string]interface{}{
			"message": "This is a test alert from Microfinance",
			"user_id": p.UserID,
		},
	)

	// Send test alert
	results := h.service.Send(r.Context(), testAlert, []alerting.ChannelType{channelType})

	if len(results) > 0 && results[0].Success {
		SendJSONResponse(w, http.StatusOK, "Test alert sent successfully", results[0])
	} else {
		errMsg := "Unknown error"
		if len(results) > 0 {
			errMsg = results[0].Error
		}
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to send test alert: "+errMsg, nil)
	}
}
