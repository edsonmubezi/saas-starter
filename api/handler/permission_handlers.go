package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/edsonmubezi/myapp/internal/permission"
	"github.com/edsonmubezi/myapp/pkg/pagination"
	"github.com/gorilla/mux"
)

// ===== Use case wiring (permissions) =====
var permissionUseCase permission.PermissionUseCase

func SetPermissionUseCase(useCase permission.PermissionUseCase) {
	permissionUseCase = useCase
}

// ===== Permission-specific DTOs =====
type UpdatePermissionDescriptionRequest struct {
	Description string `json:"description"`
}

// ===== Handlers: Permissions =====

// GetPermissionByIDHandler returns a single permission by ID
// Admin-only permissions (admin.*) are only visible to HQ callers
func GetPermissionByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid permission ID", http.StatusBadRequest)
		return
	}

	p, err := permissionUseCase.GetPermissionByID(r.Context(), id)
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "Permission not found", nil)
		return
	}

	// Return with computed visibility for backward compatibility
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(p.ToResponse())
}

// GetPermissionResourcesPaginatedHandler returns tenant permissions (non-admin) grouped by resource
// For OrgAdmin users - excludes admin.* permissions
func GetPermissionResourcesPaginatedHandler(w http.ResponseWriter, r *http.Request) {
	pag := pagination.Parse(r)
	// Scope "tenant" filters out admin.* permissions
	scope := permission.ScopeTenant
	res, err := permissionUseCase.ListPermissionResourcesPaginated(r.Context(), pag, scope)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving permissions: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(res)
}

// GetPermissionRrscForNonHqHandler returns tenant permissions (non-admin) for non-HQ users
// Alias for GetPermissionResourcesPaginatedHandler
func GetPermissionRrscForNonHqHandler(w http.ResponseWriter, r *http.Request) {
	pag := pagination.Parse(r)
	scope := permission.ScopeTenant
	res, err := permissionUseCase.ListPermissionResourcesPaginated(r.Context(), pag, scope)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving permissions: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(res)
}

// AdminGetPermissionResourcesPaginatedHandler returns ALL permissions grouped by resource
// For SuperAdmin/HQ users - includes both admin.* and tenant.* permissions
func AdminGetPermissionResourcesPaginatedHandler(w http.ResponseWriter, r *http.Request) {
	pag := pagination.Parse(r)
	// Empty scope returns all permissions
	scope := ""
	res, err := permissionUseCase.ListPermissionResourcesPaginated(r.Context(), pag, scope)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving permissions: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(res)
}

// AdminGetPermissionRrscForNonHqHandler returns ALL permissions for admin view
// Alias for AdminGetPermissionResourcesPaginatedHandler
func AdminGetPermissionRrscForNonHqHandler(w http.ResponseWriter, r *http.Request) {
	pag := pagination.Parse(r)
	scope := ""
	res, err := permissionUseCase.ListPermissionResourcesPaginated(r.Context(), pag, scope)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving permissions: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(res)
}

// GetPermissionResourcesHandler returns tenant permissions (non-paginated)
// For OrgAdmin users - excludes admin.* permissions
func GetPermissionResourcesHandler(w http.ResponseWriter, r *http.Request) {
	scope := permission.ScopeTenant
	rows, err := permissionUseCase.ListPermissionResources(r.Context(), scope)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving permissions: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(rows)
}

// GetPermissionResourcesAdminHandler returns ALL permissions (non-paginated)
// For SuperAdmin/HQ users - includes both admin.* and tenant.* permissions
func GetPermissionResourcesAdminHandler(w http.ResponseWriter, r *http.Request) {
	scope := ""
	rows, err := permissionUseCase.ListPermissionResources(r.Context(), scope)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving permissions: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(rows)
}

// UpdatePermissionDescriptionHandler updates a permission's description
// Only HQ users can update permission descriptions
func UpdatePermissionDescriptionHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid permission ID", nil)
		return
	}

	var payload UpdatePermissionDescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || strings.TrimSpace(payload.Description) == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid or missing description in request body", nil)
		return
	}

	updated, err := permissionUseCase.UpdatePermissionDescription(r.Context(), id, payload.Description)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			SendJSONResponse(w, http.StatusForbidden, "Only HQ users can update permissions", nil)
			return
		}
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to update permission", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Permission updated successfully", updated.ToResponse())
}
