package handler

import (
	"encoding/json"
	"net/http"

	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/edsonmubezi/myapp/internal/emailconfig"
)

var emailConfigUC emailconfig.UseCase

func SetEmailConfigUseCase(uc emailconfig.UseCase) {
	emailConfigUC = uc
}

// GetEmailConfigHandler returns the email config for the authenticated user's organization.
// Password is masked in the response for security.
func GetEmailConfigHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	if auth.OrganizationID == 0 {
		SendJSONResponse(w, http.StatusUnauthorized, "Organization not found", nil)
		return
	}

	cfg, err := emailConfigUC.GetEmailConfig(r.Context(), auth.OrganizationID)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to get email config: "+err.Error(), nil)
		return
	}

	if cfg == nil {
		SendJSONResponse(w, http.StatusOK, "No email config found", nil)
		return
	}

	// Mask password for admin display
	cfg.SMTPPassword = "********"

	SendJSONResponse(w, http.StatusOK, "Email config retrieved", cfg)
}

// UpsertEmailConfigHandler creates or updates email config for the authenticated user's organization.
func UpsertEmailConfigHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	if auth.OrganizationID == 0 {
		SendJSONResponse(w, http.StatusUnauthorized, "Organization not found", nil)
		return
	}

	var req emailconfig.UpsertEmailConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	// If password is masked, keep the existing password
	if req.SMTPPassword == "********" {
		existing, err := emailConfigUC.GetEmailConfig(r.Context(), auth.OrganizationID)
		if err != nil {
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to get existing config: "+err.Error(), nil)
			return
		}
		if existing != nil {
			req.SMTPPassword = existing.SMTPPassword
		}
	}

	cfg, err := emailConfigUC.UpsertEmailConfig(r.Context(), auth.OrganizationID, &req)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Mask password in response
	cfg.SMTPPassword = "********"

	SendJSONResponse(w, http.StatusOK, "Email config saved", cfg)
}
