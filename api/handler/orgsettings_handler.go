package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/edsonmubezi/myapp/internal/orgsettings"
)

// Declare global usecase var (set from container)
var OrgSettingsUseCase orgsettings.OrganizationSettingsUseCase

func SetOrgSettingsUseCase(useCase orgsettings.OrganizationSettingsUseCase) {
	OrgSettingsUseCase = useCase
}

// @Summary      Create Organization Settings
// @Description  Creates core settings for an organization (Super Admin only)
// @Tags         OrganizationSettings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        orgSettings body orgsettings.OrganizationSettingsCreateInput true "Organization settings payload"
// @Success      201 {object} orgsettings.OrganizationSettings
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/v1/super/org-settings [post]
func CreateOrgSettingsHandler(w http.ResponseWriter, r *http.Request) {
	input := &orgsettings.OrganizationSettingsCreateInput{}
	if err := json.NewDecoder(r.Body).Decode(input); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	created, err := OrgSettingsUseCase.CreateOrgSettings(r.Context(), input)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	SendJSONResponse(w, http.StatusCreated, "Organization settings created successfully", created)
}

// @Summary      Get Organization Settings By ID
// @Description  Retrieves organization settings by its ID (Super Admin only)
// @Tags         OrganizationSettings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "Organization Settings ID"
// @Success      200 {object} orgsettings.OrganizationSettings
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/v1/super/org-settings/{id} [get]
func GetOrgSettingsByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	settings, err := OrgSettingsUseCase.GetOrgSettingsByID(r.Context(), id)
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Organization settings retrieved successfully", settings)
}

// @Summary      Get Organization Settings By Organization ID
// @Description  Retrieves organization settings by organization ID (Super Admin only)
// @Tags         OrganizationSettings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        organizationId path int true "Organization ID"
// @Success      200 {object} orgsettings.OrganizationSettings
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/v1/super/org-settings/organization/{organizationId} [get]
func GetOrgSettingsByOrgIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgIDStr := vars["organizationId"]
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid organization ID format", nil)
		return
	}

	settings, err := OrgSettingsUseCase.GetOrgSettingsByOrganizationID(r.Context(), orgID)
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Organization settings retrieved successfully", settings)
}

// @Summary      Create Organization Settings By Organization ID
// @Description  Creates organization settings for a specific organization (Super Admin only)
// @Tags         OrganizationSettings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        organizationId path int true "Organization ID"
// @Param        orgSettings body orgsettings.OrganizationSettingsCreateInput true "Organization settings payload (without organizationId)"
// @Success      201 {object} orgsettings.OrganizationSettings
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/v1/super/org-settings/organization/{organizationId} [post]
func CreateOrgSettingsByOrgIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgIDStr := vars["organizationId"]
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid organization ID format", nil)
		return
	}

	input := &orgsettings.OrganizationSettingsCreateInput{}
	if err := json.NewDecoder(r.Body).Decode(input); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	// Override organizationId from URL (already decrypted by middleware)
	input.OrganizationID = orgID

	created, err := OrgSettingsUseCase.CreateOrgSettings(r.Context(), input)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	SendJSONResponse(w, http.StatusCreated, "Organization settings created successfully", created)
}

// @Summary      Update Organization Settings
// @Description  Updates organization settings (Super Admin only)
// @Tags         OrganizationSettings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        orgSettings body orgsettings.OrganizationSettingsUpdateInput true "Organization settings update payload"
// @Success      200 {object} orgsettings.OrganizationSettings
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/v1/super/org-settings [put]
func UpdateOrgSettingsHandler(w http.ResponseWriter, r *http.Request) {
	input := &orgsettings.OrganizationSettingsUpdateInput{}
	if err := json.NewDecoder(r.Body).Decode(input); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	updated, err := OrgSettingsUseCase.UpdateOrgSettings(r.Context(), input)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Organization settings updated successfully", updated)
}
