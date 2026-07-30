package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/edsonmubezi/myapp/internal/role"
	"github.com/edsonmubezi/myapp/pkg/pagination"
	"github.com/edsonmubezi/myapp/pkg/securejson"
	"github.com/gorilla/mux"
)

// ===== Use case wiring (roles) =====
var roleUseCase role.RoleUseCase

func SetRoleUseCase(useCase role.RoleUseCase) {
	roleUseCase = useCase
}

// ===== Role-specific DTOs =====
type RoleRequest struct {
	Name string `json:"name"`
}

// used by AssignPermissionsToRoleHandler (remains under role, since the target aggregate is a role)
type AssignPermissionsRequest struct {
	RoleID        int64   `json:"role_id"`
	PermissionIDs []int64 `json:"permission_ids"`
}

type assignPermissionsResponse struct {
	RoleID        int64   `json:"role_id"`
	PermissionIDs []int64 `json:"permission_ids"`
	Message       string  `json:"message"`
}

// ===== Handlers: Roles =====

func CreateRoleHandler(w http.ResponseWriter, r *http.Request) {
	var req RoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	created, err := roleUseCase.CreateRole(r.Context(), req.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating role: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(created)
}

func GetRoleByIDHandler(w http.ResponseWriter, r *http.Request) {
	roleIDStr := mux.Vars(r)["id"]
	roleID, err := strconv.ParseInt(roleIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}
	rl, err := roleUseCase.GetRoleByID(r.Context(), roleID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving role: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(rl)
}

func GetAllRolesHandler(w http.ResponseWriter, r *http.Request) {
	roles, err := roleUseCase.GetAllRoles(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving roles: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(roles)
}

func GetAllRolesAdminHandler(w http.ResponseWriter, r *http.Request) {
	roles, err := roleUseCase.GetAllRolesAdmin(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving roles: %v", err), http.StatusInternalServerError)
		return
	}
	securejson.JSON(w, http.StatusOK, roles)
}

func GetAllRolesAdminPaginatedHandler(w http.ResponseWriter, r *http.Request) {
	pag := pagination.Parse(r)
	result, err := roleUseCase.GetAllRolesAdminPaginated(r.Context(), pag)
	if err != nil {
		securejson.JSON(w, http.StatusInternalServerError, map[string]any{
			"status":  http.StatusInternalServerError,
			"message": "Failed to fetch roles",
		})
		return
	}
	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  http.StatusOK,
		"message": "roles fetched successfully",
		"data":    result,
	})
}

func GetSystemRolesHandler(w http.ResponseWriter, r *http.Request) {
	roles, err := roleUseCase.GetSystemRoles(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving system roles: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(roles)
}

func GetAllEncyrptRolesHandler(w http.ResponseWriter, r *http.Request) {
	roles, err := roleUseCase.GetAllRoles(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving roles: %v", err), http.StatusInternalServerError)
		return
	}
	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  http.StatusOK,
		"message": "Users fetched successfully",
		"data":    roles, // fields with secure:"encrypt_id" will be encrypted
	})
}

func GetAllRolesPaginatedHandler(w http.ResponseWriter, r *http.Request) {
	pag := pagination.Parse(r)

	result, err := roleUseCase.GetAllRolesPaginated(r.Context(), pag)

	if err != nil {
		securejson.JSON(w, http.StatusInternalServerError, map[string]any{
			"status":  http.StatusInternalServerError,
			"message": "Failed to fetch roles",
		})
		return
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  http.StatusOK,
		"message": "roles fetched successfully",
		"data":    result,
	})
}

func UpdateRoleHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid role ID", nil)
		return
	}
	var payload struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Name == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid or missing name in request body", nil)
		return
	}
	updated, err := roleUseCase.UpdateRole(r.Context(), id, payload.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			SendJSONResponse(w, http.StatusNotFound, "Role not found", nil)
			return
		}
		log.Printf("Error updating role for ID %d: %v", id, err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to update role", nil)
		return
	}
	SendJSONResponse(w, http.StatusOK, "Role updated successfully", updated)
}

func DeleteRoleHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid role ID", nil)
		return
	}
	if err := roleUseCase.DeleteRole(r.Context(), id); err != nil {
		log.Printf("Error deleting role for ID %d: %v", id, err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to delete role", nil)
		return
	}
	SendJSONResponse(w, http.StatusOK, "Role deleted successfully", nil)
}

// Assign/replace the permission set for a role: PUT /roles/{id}/permissions
func AssignPermissionsToRoleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	roleIDStr := mux.Vars(r)["id"]
	roleID, err := strconv.ParseInt(roleIDStr, 10, 64)
	if err != nil || roleID <= 0 {
		http.Error(w, `{"error":"invalid role id"}`, http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var payload AssignPermissionsRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid JSON body: `+escapeForJSON(err.Error())+`"}`, http.StatusBadRequest)
		return
	}

	// Optional: allow empty to clear all
	if len(payload.PermissionIDs) > 10_000 {
		http.Error(w, `{"error":"too many permission_ids"}`, http.StatusBadRequest)
		return
	}

	seen := make(map[int64]struct{}, len(payload.PermissionIDs))
	clean := make([]int64, 0, len(payload.PermissionIDs))
	for _, id := range payload.PermissionIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		clean = append(clean, id)
	}

	if err := roleUseCase.AssignPermissionsToRole(r.Context(), roleID, clean); err != nil {

		http.Error(
			w,
			`{"error":"failed to assign permissions: `+escapeForJSON(err.Error())+`"}`,
			http.StatusInternalServerError,
		)
		return
	}

	_ = json.NewEncoder(w).Encode(assignPermissionsResponse{
		RoleID:        roleID,
		PermissionIDs: clean,
		Message:       "permissions assigned successfully",
	})
}

func GetRoleWithPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	roleIDStr := mux.Vars(r)["id"]
	roleID, err := strconv.ParseInt(roleIDStr, 10, 64)
	if err != nil || roleID <= 0 {
		http.Error(w, `{"error":"invalid role id"}`, http.StatusBadRequest)
		return
	}

	data, err := roleUseCase.GetRoleWithPermissions(r.Context(), roleID)
	if err != nil {
		http.Error(w, `{"error":"`+escapeForJSON(err.Error())+`"}`, http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(data)
}
