package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/edsonmubezi/myapp/internal/identity"
	export "github.com/edsonmubezi/myapp/pkg/exports"
	"github.com/edsonmubezi/myapp/pkg/pagination"
	"github.com/edsonmubezi/myapp/pkg/securejson"

	"github.com/edsonmubezi/myapp/internal/organization"

	"github.com/gorilla/mux"
)

var Organize organization.OrganizationUseCase

func SetOrganizationUseCase(useCase organization.OrganizationUseCase) {
	Organize = useCase
}

// @Summary      Create Organization
// @Description  Registers a new organization in the multi-tenant system with contact details and configuration settings
// @Tags         Organizations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        organization body organization.Organization true "Organization details" example({"name":"Acme Corp","address":"123 Main St","contact_person":"John Doe","phone_number":"+1234567890","email":"contact@acme.com"})
// @Success      201 {object} SuccessResponse{data=organization.Organization} "Organization created successfully"
// @Failure      400 {object} ErrorResponse "Invalid input or validation error"
// @Failure      500 {object} ErrorResponse "Failed to create organization"
// @Router       /api/organizations [post]
func CreateOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	org, validationErrors, err := middleware.ParseAndValidateBody[organization.Organization](r)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Internal error", nil)
		return
	}
	if validationErrors != nil {
		SendValidationErrors(w, validationErrors)
		return
	}

	org.Name = strings.TrimSpace(org.Name)
	org.Address = strings.TrimSpace(org.Address)
	org.ContactPerson = strings.TrimSpace(org.ContactPerson)
	org.PhoneNumber = strings.TrimSpace(org.PhoneNumber)

	now := time.Now()
	org.CreatedAt = now
	org.UpdatedAt = now

	createdOrg, err := Organize.CreateOrganization(r.Context(), org)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to create organization", nil)
		return
	}

	SendJSONResponse(w, http.StatusCreated, "Organization created successfully", createdOrg)
}

// @Summary      Get Organization by ID
// @Description  Retrieves complete organization details including name, contact information, and configuration for a specific organization
// @Tags         Organizations
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "Organization ID" example(1)
// @Success      200 {object} SuccessResponse{data=organization.Organization} "Organization details retrieved successfully"
// @Failure      400 {object} ErrorResponse "Invalid organization ID format"
// @Failure      404 {object} ErrorResponse "Organization not found"
// @Router       /api/organizations/{id} [get]
func GetOrganizationByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	org, err := Organize.GetOrganizationByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(org)
}

// @Summary      Update Organization
// @Description  Updates an existing organization's details including name, address, contact person, and configuration settings
// @Tags         Organizations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "Organization ID" example(1)
// @Param        organization body organization.Organization true "Updated organization details"
// @Success      200 {object} SuccessResponse{data=organization.Organization} "Organization updated successfully"
// @Failure      400 {object} ErrorResponse "Invalid organization ID or request body"
// @Failure      404 {object} ErrorResponse "Organization not found"
// @Failure      500 {object} ErrorResponse "Failed to update organization"
// @Router       /api/organizations/{id} [put]
func UpdateOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	org, validationErrors, err := middleware.ParseAndValidateBody[organization.Organization](r)

	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	if validationErrors != nil {
		SendValidationErrors(w, validationErrors)
		return
	}
	if org == nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid or missing request body", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid organization ID", nil)
		return
	}
	org.ID = id

	org.UpdatedAt = time.Now()

	updated, err := Organize.UpdateOrganization(r.Context(), org)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			SendJSONResponse(w, http.StatusNotFound, "Organization not found", nil)
			return
		}
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to update organization", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Organization updated successfully", updated)
}

// @Summary      List All Organizations (Unpaginated)
// @Description  Retrieves all organizations without pagination. For paginated results, use /api/all-organizations endpoint instead.
// @Tags         Organizations
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} SuccessResponse{data=[]organization.Organization} "All organizations retrieved successfully"
// @Failure      500 {object} ErrorResponse "Failed to fetch organizations"
// @Router       /api/organizations [get]
func GetAllOrganizationsHandler(w http.ResponseWriter, r *http.Request) {

	result, err := Organize.GetAllOrganizations(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

// @Summary      List Organizations (Paginated)
// @Description  Retrieves paginated list of organizations with support for searching and sorting by various fields
// @Tags         Organizations
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "Page number" default(1) example(1)
// @Param        page_size query int false "Items per page" default(10) example(10)
// @Param        sort_by query string false "Sort field" Enums(id, name, created_at) default(id)
// @Param        sort_order query string false "Sort order" Enums(asc, desc) default(asc)
// @Param        q query string false "Search by organization name" example("Acme")
// @Success      200 {object} PaginatedResponse{data=[]organization.Organization} "Organizations retrieved successfully"
// @Failure      500 {object} ErrorResponse "Failed to fetch organizations"
// @Router       /api/all-organizations [get]
func GetOrganizationsHandler(w http.ResponseWriter, r *http.Request) {
	pag := pagination.Parse(r)
	if v := r.URL.Query().Get("sort_by"); v != "" {
		pag.SortBy = v
	}
	if v := r.URL.Query().Get("sort_order"); v != "" {
		pag.SortOrder = v
	}
	// Search
	if q := r.URL.Query().Get("q"); q != "" {
		pag.Search = q
	}

	result, err := Organize.GetOrganizations(r.Context(), pag)

	if err != nil {
		securejson.JSON(w, http.StatusInternalServerError, map[string]any{
			"status":  http.StatusInternalServerError,
			"message": "Failed to fetch organisation",
		})
		return
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  http.StatusOK,
		"message": "Organisation fetched successfully",
		"data":    result, // fields tagged secure:"encrypt_id" will be encrypted
	})

}

// @Summary      Update Organization with Logo
// @Description  Updates organization details including logo upload. Supports multipart/form-data for file upload.
// @Tags         Organizations
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id path int true "Organization ID" example(1)
// @Param        name formData string false "Organization name"
// @Param        phone_number formData string false "Phone number"
// @Param        address formData string false "Address"
// @Param        contact_person formData string false "Contact person"
// @Param        email formData string false "Email address"
// @Param        tin formData string false "Tax Identification Number"
// @Param        registration_number formData string false "Registration number"
// @Param        logo formData file false "Logo image file (JPG, PNG, GIF, WebP - max 5MB)"
// @Success      200 {object} SuccessResponse{data=organization.Organization} "Organization updated successfully"
// @Failure      400 {object} ErrorResponse "Invalid request or file upload error"
// @Failure      404 {object} ErrorResponse "Organization not found"
// @Failure      500 {object} ErrorResponse "Failed to update organization"
// @Router       /api/organizations/{id}/update-with-logo [put]
func UpdateOrganizationWithLogoHandler(w http.ResponseWriter, r *http.Request) {
	// Get organization ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid organization ID", nil)
		return
	}

	// Get existing organization
	existingOrg, err := Organize.GetOrganizationByID(r.Context(), orgID)
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "Organization not found", nil)
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Failed to parse form data", nil)
		return
	}

	// Update fields from form data (only if provided)
	if name := r.FormValue("name"); name != "" {
		existingOrg.Name = strings.TrimSpace(name)
	}
	if phoneNumber := r.FormValue("phone_number"); phoneNumber != "" {
		existingOrg.PhoneNumber = strings.TrimSpace(phoneNumber)
	}
	if address := r.FormValue("address"); address != "" {
		existingOrg.Address = strings.TrimSpace(address)
	}
	if contactPerson := r.FormValue("contact_person"); contactPerson != "" {
		existingOrg.ContactPerson = strings.TrimSpace(contactPerson)
	}

	// Optional fields
	if email := r.FormValue("email"); email != "" {
		trimmedEmail := strings.TrimSpace(email)
		existingOrg.Email = &trimmedEmail
	}
	if tin := r.FormValue("tin"); tin != "" {
		trimmedTIN := strings.TrimSpace(tin)
		existingOrg.TIN = &trimmedTIN
	}
	if regNum := r.FormValue("registration_number"); regNum != "" {
		trimmedRegNum := strings.TrimSpace(regNum)
		existingOrg.RegistrationNumber = &trimmedRegNum
	}

	// Handle logo upload (optional)
	_, _, err = r.FormFile("logo")
	if err == nil {
		// Logo file provided, upload it
		logoPath, err := UploadImageFile(r, "logo", orgID, "org-branding")
		if err != nil {
			SendJSONResponse(w, http.StatusBadRequest, fmt.Sprintf("Logo upload failed: %v", err), nil)
			return
		}
		existingOrg.LogoURL = &logoPath
	}
	// If err != nil, no logo was uploaded, which is fine - we keep the existing one

	// Update organization
	existingOrg.UpdatedAt = time.Now()
	updatedOrg, err := Organize.UpdateOrganization(r.Context(), existingOrg)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to update organization", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Organization updated successfully", updatedOrg)
}

// @Summary      Delete Organization
// @Description  Soft-deletes an organization by ID. Organization data is retained in database but marked as deleted and excluded from queries.
// @Tags         Organizations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "Organization ID" example(1)
// @Success      200 {object} MessageResponse "Organization deleted successfully"
// @Failure      400 {object} ErrorResponse "Invalid organization ID"
// @Failure      500 {object} ErrorResponse "Failed to delete organization"
// @Router       /api/organizations/{id} [delete]
func SoftDeleteOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid organization ID", nil)
		return
	}
	err = Organize.SoftDeleteOrganization(r.Context(), id)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to delete organization", nil)
		return
	}
	SendJSONResponse(w, http.StatusOK, "Organization deleted successfully", nil)
}

// @Summary      List Organizations with Encrypted IDs
// @Description  Retrieves all organizations with ID encryption enabled. Organization IDs are encrypted in the response for additional security.
// @Tags         Organizations
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} SuccessResponse{data=[]organization.Organization} "Organizations retrieved successfully with encrypted IDs"
// @Failure      500 {object} ErrorResponse "Failed to fetch organizations"
// @Router       /api/organizations/encrypted [get]
func GetAllOrgsEncrptyedHandler(w http.ResponseWriter, r *http.Request) {

	result, err := Organize.GetAllOrganizations(r.Context())
	if err != nil {
		securejson.JSON(w, http.StatusInternalServerError, map[string]any{
			"status":  http.StatusInternalServerError,
			"message": "Failed to fetch Organization",
		})
		return
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  http.StatusOK,
		"message": "Organization fetched successfully",
		"data":    result,
	})
}

func GetDocumentBrandingHandler(w http.ResponseWriter, r *http.Request) {
	orgID, err := identity.OrgID(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	settings, err := Organize.GetDocumentBranding(r.Context(), orgID)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to fetch document branding settings", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Document branding settings fetched", settings)
}

func UploadWatermarkImageHandler(w http.ResponseWriter, r *http.Request) {
	orgID, err := identity.OrgID(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Failed to parse form data", nil)
		return
	}

	_, _, err = r.FormFile("watermark")
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "No watermark image provided", nil)
		return
	}

	imagePath, err := UploadImageFile(r, "watermark", orgID, "org-branding")
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, fmt.Sprintf("Upload failed: %v", err), nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Watermark image uploaded", map[string]string{"path": imagePath})
}

func SaveDocumentBrandingHandler(w http.ResponseWriter, r *http.Request) {
	orgID, err := identity.OrgID(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var input organization.DocumentBrandingSettings
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Validate font family against available fonts (built-in + custom TTF)
	fontValid := false
	for _, f := range export.AvailableFonts() {
		if f == input.FontFamily {
			fontValid = true
			break
		}
	}
	if !fontValid {
		input.FontFamily = "Arial"
	}

	// Validate colors (basic hex check)
	if len(input.PrimaryColor) != 7 || input.PrimaryColor[0] != '#' {
		input.PrimaryColor = "#1a365d"
	}
	if len(input.HeaderTextColor) != 7 || input.HeaderTextColor[0] != '#' {
		input.HeaderTextColor = "#FFFFFF"
	}

	input.OrganizationID = orgID

	if err := Organize.SaveDocumentBranding(r.Context(), &input); err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to save document branding settings", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Document branding settings saved", nil)
}

func GetAvailableFontsHandler(w http.ResponseWriter, r *http.Request) {
	SendJSONResponse(w, http.StatusOK, "Available fonts", export.AvailableFonts())
}

func GetEmailBrandingHandler(w http.ResponseWriter, r *http.Request) {
	orgID, err := identity.OrgID(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	settings, err := Organize.GetEmailBranding(r.Context(), orgID)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to fetch email branding settings", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Email branding settings fetched", settings)
}

func SaveEmailBrandingHandler(w http.ResponseWriter, r *http.Request) {
	orgID, err := identity.OrgID(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var input organization.EmailBrandingSettings
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Validate colors (basic hex check)
	validateHexColor := func(color *string, fallback string) {
		if len(*color) != 7 || (*color)[0] != '#' {
			*color = fallback
		}
	}
	validateHexColor(&input.PrimaryColor, "#4F46E5")
	validateHexColor(&input.HeaderTextColor, "#FFFFFF")
	validateHexColor(&input.AccentColor, "#4F46E5")

	input.OrganizationID = orgID

	if err := Organize.SaveEmailBranding(r.Context(), &input); err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to save email branding settings", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Email branding settings saved", nil)
}
