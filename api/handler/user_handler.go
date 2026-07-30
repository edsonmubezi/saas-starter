package handler

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/edsonmubezi/myapp/api/middleware"
	internalAuth "github.com/edsonmubezi/myapp/internal/auth"
	"github.com/edsonmubezi/myapp/internal/identity"
	"github.com/edsonmubezi/myapp/internal/organization"
	"github.com/edsonmubezi/myapp/internal/orgsettings"
	"github.com/edsonmubezi/myapp/internal/user"
	"github.com/edsonmubezi/myapp/pkg/auth"
	export "github.com/edsonmubezi/myapp/pkg/exports"
	"github.com/edsonmubezi/myapp/pkg/pagination"
	"github.com/edsonmubezi/myapp/pkg/securejson"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// Package-level variables for password reset
var (
	passwordResetHistoryRepo    *internalAuth.PostgresPasswordResetHistoryRepository
	organizationUseCase         organization.OrganizationUseCase
	orgCoreSettingsUseCaseForMe orgsettings.OrganizationSettingsUseCase
)

// SetPasswordResetHistoryRepo sets the password reset history repository
func SetPasswordResetHistoryRepo(repo *internalAuth.PostgresPasswordResetHistoryRepository) {
	passwordResetHistoryRepo = repo
}

// SetOrganizationUseCaseForReset sets the organization use case for password reset PDF generation
func SetOrganizationUseCaseForReset(uc organization.OrganizationUseCase) {
	organizationUseCase = uc
}

// SetOrgSettingsUseCaseForUserHandler sets the org settings use case for /users/me endpoint
func SetOrgSettingsUseCaseForUserHandler(uc interface{}) {
	if typedUC, ok := uc.(orgsettings.OrganizationSettingsUseCase); ok {
		orgCoreSettingsUseCaseForMe = typedUC
	}
}

func hasAnyChanges(in *user.UpdateUserInput) bool {
	return in.FullName != nil ||
		in.Email != nil ||
		in.Password != nil ||
		in.RoleID != nil ||
		in.ActiveStatus != nil ||
		in.Photo != nil ||
		in.OrganizationID != nil
}

// @Summary      Get User by ID
// @Description  Retrieves complete user details by ID within the authenticated user's organization. Includes role information and enforces tenant isolation.
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "User ID" example(1)
// @Success      200 {object} SuccessResponse{data=user.User} "User retrieved successfully"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Access denied - user belongs to different organization"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /api/users/{id} [get]
func GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user's organization ID
	auth, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	ids, ok := vars["id"]
	if !ok || ids == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Missing user ID in URL", nil)
		return
	}
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	user, err := userUseCase.GetUserByID(r.Context(), id, auth.OrganizationID)
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}

	// Verify organization ownership using standardized identity enforcement
	if err := identity.MustMatchOrg(r.Context(), user.OrganizationID); err != nil {
		log.Printf("Access denied: User %d (org %d) attempted to access user %d (org %d)",
			auth.UserID, auth.OrganizationID, user.ID, user.OrganizationID)
		SendJSONResponse(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	user.Password = ""
	SendJSONResponse(w, http.StatusOK, "User retrieved successfully", user)
}

// @Summary      Get Authenticated User
// @Description  Retrieves the currently authenticated user's details including role and organization. Returns user with encrypted ID for security.
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} SuccessResponse{data=user.User} "Authenticated user fetched successfully"
// @Failure      401 {object} ErrorResponse "Unauthorized - user ID missing"
// @Failure      404 {object} ErrorResponse "User not found"
// @Failure      500 {object} ErrorResponse "Failed to retrieve user"
// @Router       /api/users/me [get]
func GetAuthenticatedUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == 0 {
		// keep your existing simple sender for error paths if you want
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized: user ID missing", nil)
		return
	}

	auth := middleware.GetAuthContextFromContext(r.Context())
	u, err := userUseCase.GetUserByID(r.Context(), userID, auth.OrganizationID)
	if err != nil {
		log.Printf("Error fetching user by ID %d: %v", userID, err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to retrieve user", nil)
		return
	}
	if u == nil {
		log.Printf("User not found in DB for ID: %d", userID)
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}

	u.Password = ""

	// Fetch organization type from org settings
	var organizationType string
	if auth.OrganizationID > 0 && orgCoreSettingsUseCaseForMe != nil {
		if settings, err := orgCoreSettingsUseCaseForMe.GetOrgSettingsByOrganizationID(r.Context(), auth.OrganizationID); err == nil && settings != nil {
			organizationType = settings.OrganizationType
		} else {
			organizationType = orgsettings.OrgTypeSingleCompany // Default
		}
	}

	resp := map[string]any{
		"status":  200,
		"message": "Authenticated user fetched successfully",
		"data": map[string]any{
			"user":              u, // `id` gets encrypted by securejson due to the tag
			"organization_type": organizationType,
		},
	}
	securejson.JSON(w, http.StatusOK, resp)
}

// @Summary      Get User by Email
// @Description  Retrieves user details by email address. Returns user information if found in the system.
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Param        email query string true "User email address" example("user@example.com")
// @Success      200 {object} SuccessResponse{data=user.User} "User fetched successfully"
// @Failure      400 {object} ErrorResponse "Missing email parameter"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /api/users/by-email [get]
func GetUserByEmailHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Missing email", nil)
		return
	}
	user, err := userUseCase.GetUserByEmail(r.Context(), email)
	if err != nil || user == nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}
	user.Password = ""
	SendJSONResponse(w, http.StatusOK, "User fetched successfully", user)
}

// @Summary      Update User Status
// @Description  Toggles a user's active/inactive status. Only affects users within the authenticated user's organization.
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "User ID" example(1)
// @Success      200 {object} SuccessResponse{data=user.User} "User status updated successfully"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      404 {object} ErrorResponse "User not found"
// @Failure      500 {object} ErrorResponse "Failed to update user status"
// @Router       /api/users/{id}/status [put]
func UpdateUserStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Missing user ID in URL path", nil)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	auth := middleware.GetAuthContextFromContext(r.Context())

	// SuperAdmin can toggle status for any user across orgs
	var u *user.User
	if auth.Role == "SuperAdmin" {
		u, err = userUseCase.GetUserByIDNoOrg(r.Context(), id)
	} else {
		u, err = userUseCase.GetUserByID(r.Context(), id, auth.OrganizationID)
	}
	if err != nil || u == nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}
	u.ActiveStatus = !u.ActiveStatus
	if err := userUseCase.UpdateUserStatus(r.Context(), u.ID, u.OrganizationID); err != nil {
		log.Printf("Error updating user status for ID %d: %v", u.ID, err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to update user status", nil)
		return
	}
	u.Password = ""
	SendJSONResponse(w, http.StatusOK, "User status updated successfully", u)
}

// @Summary      Soft Delete User
// @Description  Soft-deletes a user by marking them as deleted without removing data from database. Enforces organization isolation.
// @Tags         Users
// @Security     BearerAuth
// @Param        id path int true "User ID" example(1)
// @Success      204 "User soft deleted successfully"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      500 {object} ErrorResponse "Failed to delete user"
// @Router       /api/users/{id} [delete]
func SoftDeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Verify organization ownership before allowing deletion
	auth, err := identity.Require(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch existing user to verify organization ownership
	existingUser, err := userUseCase.GetUserByID(r.Context(), id, auth.OrganizationID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Enforce organization isolation
	if err := identity.MustMatchOrg(r.Context(), existingUser.OrganizationID); err != nil {
		log.Printf("Access denied: User %d (org %d) attempted to delete user %d (org %d)",
			auth.UserID, auth.OrganizationID, existingUser.ID, existingUser.OrganizationID)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	err = userUseCase.SoftDeleteUser(r.Context(), id, auth.OrganizationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Update User
// @Description  Updates user details with partial update support. Only provided fields are updated. Enforces organization isolation.
// @Tags         Users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "User ID" example(1)
// @Param        user body user.UpdateUserInput true "Fields to update (all optional)"
// @Success      200 {object} SuccessResponse{data=user.User} "User updated successfully"
// @Failure      400 {object} ErrorResponse "Invalid user ID, JSON, or no fields to update"
// @Failure      404 {object} ErrorResponse "User not found"
// @Failure      500 {object} ErrorResponse "Failed to update user"
// @Router       /api/users/{id} [put]
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// 1) Get ID from path
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Missing user ID in URL path", nil)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// 2) Verify organization ownership before allowing update
	auth, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// Fetch existing user to verify organization ownership
	existingUser, err := userUseCase.GetUserByID(r.Context(), id, auth.OrganizationID)
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}

	// Enforce organization isolation
	if err := identity.MustMatchOrg(r.Context(), existingUser.OrganizationID); err != nil {
		log.Printf("Access denied: User %d (org %d) attempted to update user %d (org %d)",
			auth.UserID, auth.OrganizationID, existingUser.ID, existingUser.OrganizationID)
		SendJSONResponse(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	// 3) Parse + validate body using your generic helper
	payload, validationErrors, err := middleware.ParseAndValidateBody[user.UpdateUserInput](r)
	if err != nil {
		// parse error (malformed JSON etc.)
		SendJSONResponse(w, http.StatusBadRequest, "Invalid JSON", nil)
		return
	}
	if validationErrors != nil {
		// structured validation errors from validator
		SendValidationErrors(w, validationErrors)
		return
	}

	// 4) If body contained an ID and it conflicts, reject
	if payload.ID != 0 && payload.ID != id {
		SendJSONResponse(w, http.StatusBadRequest, "Body ID does not match URL ID", nil)
		return
	}
	// 5) Force the ID from path
	payload.ID = id
	if !hasAnyChanges(payload) {
		SendJSONResponse(w, http.StatusBadRequest, "No fields to update", nil)
		return
	}

	updatedUser, err := userUseCase.UpdateUserPartial(r.Context(), payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
			return
		}
		log.Printf("Error updating user: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to update user", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "User updated successfully", updatedUser)
}

// UpdateUserForHQHandler is the SuperAdmin variant — no org isolation check.
func UpdateUserForHQHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Missing user ID in URL path", nil)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// Verify the user exists (no org filter for HQ)
	_, err = userUseCase.GetUserByIDNoOrg(r.Context(), id)
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}

	payload, validationErrors, err := middleware.ParseAndValidateBody[user.UpdateUserInput](r)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid JSON", nil)
		return
	}
	if validationErrors != nil {
		SendValidationErrors(w, validationErrors)
		return
	}

	if payload.ID != 0 && payload.ID != id {
		SendJSONResponse(w, http.StatusBadRequest, "Body ID does not match URL ID", nil)
		return
	}
	payload.ID = id
	if !hasAnyChanges(payload) {
		SendJSONResponse(w, http.StatusBadRequest, "No fields to update", nil)
		return
	}

	updatedUser, err := userUseCase.UpdateUserPartial(r.Context(), payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
			return
		}
		log.Printf("Error updating user: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to update user", nil)
		return
	}
	SendJSONResponse(w, http.StatusOK, "User updated successfully", updatedUser)
}

// SoftDeleteUserForHQHandler is the SuperAdmin variant — no org isolation check.
func SoftDeleteUserForHQHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Verify the user exists (no org filter for HQ)
	existingUser, err := userUseCase.GetUserByIDNoOrg(r.Context(), id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	err = userUseCase.SoftDeleteUser(r.Context(), id, existingUser.OrganizationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Get My Token
// @Description  Generates new access and refresh tokens for the currently authenticated user. Returns tokens with user information from context or database.
// @Tags         Authentication
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} AuthSuccessResponse "Tokens issued successfully"
// @Failure      401 {object} ErrorResponse "Unauthorized - user not found"
// @Failure      500 {object} ErrorResponse "Failed to generate tokens"
// @Router       /api/users/my-token [get]
func GetMyTokenHandler(w http.ResponseWriter, r *http.Request) {
	ac := middleware.GetAuthContextFromContext(r.Context())
	if ac.UserID == 0 {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userEmail := ac.Email

	roleName := ac.Role
	orgID := ac.OrganizationID

	needEmail := strings.TrimSpace(userEmail) == ""
	needRole := strings.TrimSpace(roleName) == ""
	needOrg := orgID == 0

	fullName := identity.FullName(r.Context())

	if needEmail || needRole || needOrg || fullName == "" {
		u, err := userUseCase.GetUserByID(r.Context(), ac.UserID, ac.OrganizationID)
		if err != nil || u == nil {
			SendJSONResponse(w, http.StatusUnauthorized, "User not found", nil)
			return
		}

		if needEmail {
			userEmail = u.Email
		}
		if needRole && u.Role != nil {
			roleName = u.Role.Name
		}
		if needOrg {
			orgID = u.OrganizationID
		}
		if fullName == "" {
			fullName = u.FullName
		}
	}

	// Avoid caching token response
	w.Header().Set("Cache-Control", "no-store")

	access, err := auth.GenerateAccessToken(ac.UserID, userEmail, orgID, roleName, fullName)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate access token", nil)
		return
	}
	refresh, err := auth.GenerateRefreshToken(ac.UserID, orgID)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate refresh token", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "Tokens issued", LoginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	})
}

// =============================================================================
// SUPERADMIN USER LISTING HANDLER
// =============================================================================

// @Summary      Get All Users (SuperAdmin)
// @Description  Retrieves all users across all organizations for SuperAdmin. Supports optional organization filter via query param.
// @Tags         Users (SuperAdmin)
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        page_size query int false "Page size" default(20)
// @Param        search query string false "Search by email or name"
// @Param        organization_id query int false "Filter by organization ID (0 = all organizations)"
// @Success      200 {object} SuccessResponse{data=pagination.Result[user.UserListItem]} "Users retrieved successfully"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Failed to fetch users"
// @Router       /api/users [get]
func GetAllUsersForHQHandler(w http.ResponseWriter, r *http.Request) {
	pg := pagination.Parse(r)

	// Optional organization filter from query param (SuperAdmin can filter by org)
	orgID := int64(0)
	if v := r.URL.Query().Get("organization_id"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed >= 0 {
			orgID = parsed
		}
	}

	userType := r.URL.Query().Get("user_type")

	result, err := userUseCase.GetAllUsersForAdmin(r.Context(), pg, orgID, userType)
	if err != nil {
		log.Printf("Error fetching users for admin: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to fetch users", nil)
		return
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": "Users retrieved successfully",
		"data":    result,
	})
}

func GetUserCountsHandler(w http.ResponseWriter, r *http.Request) {
	counts, err := userUseCase.GetUserCounts(r.Context())
	if err != nil {
		log.Printf("Error fetching user counts: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to fetch user counts", nil)
		return
	}
	SendJSONResponse(w, http.StatusOK, "User counts retrieved", counts)
}

func LockUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// Verify user exists — SuperAdmin can lock any user, others only their org
	var u *user.User
	if authUser.Role == "SuperAdmin" {
		u, err = userUseCase.GetUserByIDNoOrg(r.Context(), id)
	} else {
		u, err = userUseCase.GetUserByID(r.Context(), id, authUser.OrganizationID)
	}
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}

	if authUser.Role != "SuperAdmin" && u.OrganizationID != authUser.OrganizationID {
		SendJSONResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}

	if userRepo != nil {
		if err := userRepo.LockAccount(r.Context(), id, "Locked by administrator"); err != nil {
			log.Printf("Error locking user account %d: %v", id, err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to lock account", nil)
			return
		}
	}

	log.Printf("Account locked: User %d locked user %d", authUser.UserID, id)
	SendJSONResponse(w, http.StatusOK, "Account locked successfully", nil)
}

// =============================================================================
// ORGANIZATION-SCOPED USER HANDLERS
// =============================================================================

// @Summary      Get Organization Users
// @Description  Retrieves users within the caller's organization only. Organization is extracted from JWT - no org filter accepted from request.
// @Tags         Users (Organization)
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        page_size query int false "Page size" default(20)
// @Param        search query string false "Search by email or name"
// @Success      200 {object} SuccessResponse{data=pagination.Result[user.UserListItem]} "Organization users retrieved successfully"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Failed to fetch organization users"
// @Router       /api/org-users [get]
func GetOrganizationUsersHandler(w http.ResponseWriter, r *http.Request) {
	pg := pagination.Parse(r)
	userType := r.URL.Query().Get("user_type") // "employee", "admin", or "" (all)

	// No organization filter from request - org ID is extracted from JWT context in usecase
	result, err := userUseCase.GetUsersForOrganization(r.Context(), pg, userType)
	if err != nil {
		log.Printf("Error fetching organization users: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to fetch organization users", nil)
		return
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": "Organization users retrieved successfully",
		"data":    result,
	})
}

// @Summary      Create Organization User
// @Description  Creates a new user within the caller's organization. Organization ID is enforced from JWT context.
// @Tags         Users (Organization)
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        user body user.User true "User details"
// @Success      201 {object} SuccessResponse{data=user.User} "User created successfully"
// @Failure      400 {object} ErrorResponse "Invalid input or validation error"
// @Failure      409 {object} ErrorResponse "Email already exists"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/org-users/create [post]
func CreateOrgUserHandler(w http.ResponseWriter, r *http.Request) {
	// This handler reuses CreateUserHandler logic but the usecase enforces org from context
	CreateUserHandler(w, r)
}

// @Summary      Get Organization User by ID
// @Description  Retrieves user details by ID within the caller's organization only.
// @Tags         Users (Organization)
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "User ID"
// @Success      200 {object} SuccessResponse{data=user.User} "User retrieved successfully"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      403 {object} ErrorResponse "Access denied - user belongs to different organization"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /api/org-users/{id} [get]
func GetOrgUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user's organization ID from context
	auth, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Missing user ID in URL", nil)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// Fetch user with org filter enforced
	u, err := userUseCase.GetUserByID(r.Context(), id, auth.OrganizationID)
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}

	// Double-check organization ownership
	if u.OrganizationID != auth.OrganizationID {
		log.Printf("Access denied: User %d (org %d) attempted to access user %d (org %d)",
			auth.UserID, auth.OrganizationID, u.ID, u.OrganizationID)
		SendJSONResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}

	u.Password = ""
	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  http.StatusOK,
		"message": "User retrieved successfully",
		"data":    u,
	})
}

// @Summary      Update Organization User
// @Description  Updates user details within the caller's organization only.
// @Tags         Users (Organization)
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "User ID"
// @Param        user body user.UpdateUserInput true "Fields to update"
// @Success      200 {object} SuccessResponse{data=user.User} "User updated successfully"
// @Failure      400 {object} ErrorResponse "Invalid user ID or input"
// @Failure      403 {object} ErrorResponse "Access denied"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /api/org-users/{id} [put]
func UpdateOrgUserHandler(w http.ResponseWriter, r *http.Request) {
	auth, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Missing user ID in URL path", nil)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// Verify user exists and belongs to caller's organization
	existingUser, err := userUseCase.GetUserByID(r.Context(), id, auth.OrganizationID)
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}
	if existingUser.OrganizationID != auth.OrganizationID {
		SendJSONResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}

	payload, validationErrors, err := middleware.ParseAndValidateBody[user.UpdateUserInput](r)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid JSON", nil)
		return
	}
	if validationErrors != nil {
		SendValidationErrors(w, validationErrors)
		return
	}

	if payload.ID != 0 && payload.ID != id {
		SendJSONResponse(w, http.StatusBadRequest, "Body ID does not match URL ID", nil)
		return
	}
	payload.ID = id

	if !hasAnyChanges(payload) {
		SendJSONResponse(w, http.StatusBadRequest, "No fields to update", nil)
		return
	}

	updatedUser, err := userUseCase.UpdateUserPartial(r.Context(), payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
			return
		}
		log.Printf("Error updating user: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to update user", nil)
		return
	}

	SendJSONResponse(w, http.StatusOK, "User updated successfully", updatedUser)
}

// @Summary      Delete Organization User
// @Description  Soft-deletes a user within the caller's organization only.
// @Tags         Users (Organization)
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      204 "User deleted successfully"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      403 {object} ErrorResponse "Access denied"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /api/org-users/{id} [delete]
func DeleteOrgUserHandler(w http.ResponseWriter, r *http.Request) {
	auth, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// Verify user exists and belongs to caller's organization
	existingUser, err := userUseCase.GetUserByID(r.Context(), id, auth.OrganizationID)
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}
	if existingUser.OrganizationID != auth.OrganizationID {
		SendJSONResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}

	err = userUseCase.SoftDeleteUser(r.Context(), id, auth.OrganizationID)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to delete user", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Update Organization User Status
// @Description  Toggles user active/inactive status within the caller's organization only.
// @Tags         Users (Organization)
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200 {object} SuccessResponse{data=user.User} "User status updated successfully"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      403 {object} ErrorResponse "Access denied"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /api/org-users/{id}/status [put]
func UpdateOrgUserStatusHandler(w http.ResponseWriter, r *http.Request) {
	auth, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Missing user ID in URL path", nil)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// Verify user exists and belongs to caller's organization
	u, err := userUseCase.GetUserByID(r.Context(), id, auth.OrganizationID)
	if err != nil || u == nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}
	if u.OrganizationID != auth.OrganizationID {
		SendJSONResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}

	u.ActiveStatus = !u.ActiveStatus
	if err := userUseCase.UpdateUserStatus(r.Context(), u.ID, auth.OrganizationID); err != nil {
		log.Printf("Error updating user status for ID %d: %v", u.ID, err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to update user status", nil)
		return
	}
	u.Password = ""
	SendJSONResponse(w, http.StatusOK, "User status updated successfully", u)
}

// @Summary      Send Organization Password Reset
// @Description  Advanced password reset for a user within the caller's organization. Supports email, form, or both methods.
// @Tags         Users (Organization)
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "User ID"
// @Param        request body AdminResetPasswordRequest true "Reset options"
// @Success      200 {object} AdminResetPasswordResponse "Password reset initiated"
// @Failure      400 {object} ErrorResponse "Invalid user ID or request"
// @Failure      403 {object} ErrorResponse "Access denied"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /api/org-users/{id}/reset-password [post]
func SendOrgPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// Parse request body
	var req AdminResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Default method is email
	if req.Method == "" {
		req.Method = "email"
	}

	// Validate method
	if req.Method != "email" && req.Method != "form" && req.Method != "both" {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid method. Must be 'email', 'form', or 'both'", nil)
		return
	}

	// Verify user exists and belongs to caller's organization
	targetUser, err := userUseCase.GetUserByID(r.Context(), id, authUser.OrganizationID)
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}
	if targetUser.OrganizationID != authUser.OrganizationID {
		SendJSONResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}

	response := AdminResetPasswordResponse{
		Success: true,
	}

	// Generate form reference
	formReference := generateFormReference()
	response.FormReference = formReference

	// Generate temporary password if requested
	var tempPassword string
	if req.GeneratePassword {
		tempPassword, err = generateTemporaryPassword(14)
		if err != nil {
			log.Printf("Error generating temporary password: %v", err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate password", nil)
			return
		}
		response.TemporaryPassword = tempPassword

		// Hash the password
		hashBytes, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to process password", nil)
			return
		}
		hashedPassword := string(hashBytes)

		// Update user password
		if err := userUseCase.UpdatePassword(r.Context(), id, hashedPassword); err != nil {
			log.Printf("Error updating password: %v", err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to update password", nil)
			return
		}

		// Set must_change_password flag
		if userRepo != nil {
			if err := userRepo.SetMustChangePassword(r.Context(), id, true); err != nil {
				log.Printf("Error setting must_change_password flag: %v", err)
			}
		}
	}

	// Determine expiry time
	expiryHours := 24
	if req.ExpiryHours > 0 && req.ExpiryHours <= 72 {
		expiryHours = req.ExpiryHours
	}
	expiresAt := time.Now().Add(time.Duration(expiryHours) * time.Hour)

	// Get IP address
	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Create password reset history record
	var resetID int64
	if passwordResetHistoryRepo != nil {
		resetMethod := internalAuth.ResetMethodAdmin
		if req.Method == "email" {
			resetMethod = internalAuth.ResetMethodEmail
		} else if req.Method == "form" {
			resetMethod = internalAuth.ResetMethodForm
		}

		input := internalAuth.PasswordResetHistoryInput{
			UserID:         id,
			InitiatedByID:  &authUser.UserID,
			ResetMethod:    resetMethod,
			IPAddress:      ipAddress,
			UserAgent:      userAgent,
			FormAttached:   req.Method == "form" || req.Method == "both",
			Notes:          &req.Notes,
			TokenExpiresAt: &expiresAt,
			OrganizationID: &targetUser.OrganizationID,
		}

		resetID, err = passwordResetHistoryRepo.Create(r.Context(), input)
		if err != nil {
			log.Printf("Error creating password reset history: %v", err)
		}
		response.ResetID = resetID
	}

	// Send email if method is email or both
	if req.Method == "email" || req.Method == "both" {
		if emailResolver != nil && req.GeneratePassword {
			err = sendPasswordResetEmail(r.Context(), targetUser, tempPassword, expiresAt)
			if err != nil {
				log.Printf("Error sending password reset email: %v", err)
			} else {
				response.EmailSent = true
			}
		} else if req.Method == "email" && !req.GeneratePassword {
			if resetTokenService != nil {
				token, err := resetTokenService.CreateResetToken(
					r.Context(),
					id,
					ipAddress,
					userAgent,
					time.Duration(expiryHours)*time.Hour,
				)
				if err == nil {
					baseURL := getBaseURL()
					resetLink := fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)
					err = sendPasswordResetLinkEmail(r.Context(), targetUser, resetLink, expiresAt)
					if err != nil {
						log.Printf("Error sending password reset link email: %v", err)
					} else {
						response.EmailSent = true
					}
				}
			}
		}
	}

	// Generate PDF form if method is form or both
	if req.Method == "form" || req.Method == "both" {
		pdfData := export.PasswordResetFormData{
			UserFullName:      targetUser.FullName,
			UserEmail:         targetUser.Email,
			UserID:            idStr,
			RequestDate:       time.Now(),
			RequestedBy:       authUser.Email,
			RequestedByRole:   authUser.Role,
			TemporaryPassword: tempPassword,
			ExpiresAt:         expiresAt,
			Notes:             req.Notes,
			FormReference:     formReference,
		}

		var pdfBytes []byte
		if organizationUseCase != nil {
			pdfBytes, err = export.GeneratePasswordResetForm(r.Context(), organizationUseCase, targetUser.OrganizationID, pdfData)
			if err != nil {
				log.Printf("Error generating PDF with org details: %v", err)
				companyName := "Microfinance"
				if targetUser.Organization != nil {
					companyName = targetUser.Organization.Name
				}
				pdfBytes, err = export.GenerateSimplePasswordResetForm(pdfData, companyName)
			}
		} else {
			companyName := "Microfinance"
			if targetUser.Organization != nil {
				companyName = targetUser.Organization.Name
			}
			pdfBytes, err = export.GenerateSimplePasswordResetForm(pdfData, companyName)
		}

		if err != nil {
			log.Printf("Error generating PDF: %v", err)
		} else if len(pdfBytes) > 0 {
			response.PDFGenerated = true

			if storageService != nil {
				pdfPath := fmt.Sprintf("password-reset-forms/%d/%s.pdf", targetUser.OrganizationID, formReference)
				reader := bytes.NewReader(pdfBytes)
				err := storageService.Upload(r.Context(), pdfPath, reader, "application/pdf")
				if err != nil {
					log.Printf("Error saving PDF to storage: %v", err)
				} else {
					pdfURL, err := storageService.GetURL(r.Context(), pdfPath, 24*time.Hour)
					if err != nil {
						log.Printf("Error getting PDF URL: %v", err)
						response.PDFURL = &pdfPath
					} else {
						response.PDFURL = &pdfURL
					}
				}
			}
		}
	}

	// Build response message
	var messageParts []string
	if response.EmailSent {
		messageParts = append(messageParts, "email sent")
	}
	if response.PDFGenerated {
		messageParts = append(messageParts, "PDF form generated")
	}
	if req.GeneratePassword {
		messageParts = append(messageParts, "temporary password created")
	}

	if len(messageParts) > 0 {
		response.Message = fmt.Sprintf("Password reset initiated: %s", strings.Join(messageParts, ", "))
	} else {
		response.Message = "Password reset initiated"
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": response.Message,
		"data":    response,
	})
}

// =============================================================================
// ACCOUNT SECURITY HANDLERS
// =============================================================================

// @Summary      Unlock User Account
// @Description  Unlocks a locked user account. Resets failed login attempts and clears lock status.
// @Tags         Users (Admin)
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "User ID to unlock"
// @Success      200 {object} SuccessResponse "Account unlocked successfully"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Access denied"
// @Failure      404 {object} ErrorResponse "User not found"
// @Failure      500 {object} ErrorResponse "Failed to unlock account"
// @Router       /api/admin/users/{id}/unlock [post]
func UnlockUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// Verify user exists — SuperAdmin can unlock any user across orgs
	var u *user.User
	if authUser.Role == "SuperAdmin" {
		u, err = userUseCase.GetUserByIDNoOrg(r.Context(), id)
	} else {
		u, err = userUseCase.GetUserByID(r.Context(), id, authUser.OrganizationID)
	}
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}

	// OrgAdmin can only unlock users in their own org
	if authUser.Role != "SuperAdmin" && u.OrganizationID != authUser.OrganizationID {
		SendJSONResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}

	// Unlock the account using the repository
	if userRepo != nil {
		if err := userRepo.UnlockAccount(r.Context(), id); err != nil {
			log.Printf("Error unlocking account for user %d: %v", id, err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to unlock account", nil)
			return
		}
	} else {
		SendJSONResponse(w, http.StatusInternalServerError, "User repository not configured", nil)
		return
	}

	// Log the action
	log.Printf("Account unlocked: User %d unlocked user %d", authUser.UserID, id)

	SendJSONResponse(w, http.StatusOK, "Account unlocked successfully", map[string]any{
		"user_id":   id,
		"full_name": u.FullName,
		"email":     u.Email,
	})
}

// @Summary      Get User Security Info
// @Description  Retrieves security-related information for a user (locked status, failed attempts, last login, etc.)
// @Tags         Users (Admin)
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "User ID"
// @Success      200 {object} SuccessResponse "Security info retrieved successfully"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Access denied"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /api/admin/users/{id}/security [get]
func GetUserSecurityInfoHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// Get user — SuperAdmin can view any user across orgs
	var u *user.User
	if authUser.Role == "SuperAdmin" {
		u, err = userUseCase.GetUserByIDNoOrg(r.Context(), id)
	} else {
		u, err = userUseCase.GetUserByID(r.Context(), id, authUser.OrganizationID)
	}
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}

	// OrgAdmin can only view users in their own org
	if authUser.Role != "SuperAdmin" && u.OrganizationID != authUser.OrganizationID {
		SendJSONResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}

	// Return security info
	securityInfo := map[string]any{
		"user_id":               u.ID,
		"full_name":             u.FullName,
		"email":                 u.Email,
		"active_status":         u.ActiveStatus,
		"is_locked":             u.IsLocked(),
		"locked_at":             u.LockedAt,
		"lock_reason":           u.LockReason,
		"failed_login_attempts": u.FailedLoginAttempts,
		"last_login_at":         u.LastLoginAt,
		"last_login_ip":         u.LastLoginIP,
		"two_factor_enabled":    u.TwoFactorEnabled,
		"two_factor_method":     u.TwoFactorMethod,
		"must_change_password":  u.MustChangePassword,
	}

	SendJSONResponse(w, http.StatusOK, "Security info retrieved successfully", securityInfo)
}

// =============================================================================
// ADMIN PASSWORD RESET HANDLERS
// =============================================================================

// AdminResetPasswordRequest represents the request body for admin-initiated password reset
type AdminResetPasswordRequest struct {
	Method           string `json:"method"` // "email", "form", or "both"
	Notes            string `json:"notes,omitempty"`
	GeneratePassword bool   `json:"generate_password"` // Generate a temporary password
	ExpiryHours      int    `json:"expiry_hours,omitempty"`
}

// AdminResetPasswordResponse represents the response for admin password reset
type AdminResetPasswordResponse struct {
	Success           bool    `json:"success"`
	Message           string  `json:"message"`
	ResetID           int64   `json:"reset_id"`
	FormReference     string  `json:"form_reference,omitempty"`
	TemporaryPassword string  `json:"temporary_password,omitempty"` // Only returned if generate_password is true
	EmailSent         bool    `json:"email_sent"`
	PDFGenerated      bool    `json:"pdf_generated"`
	PDFURL            *string `json:"pdf_url,omitempty"`
}

// generateTemporaryPassword generates a secure temporary password
func generateTemporaryPassword(length int) (string, error) {
	if length < 12 {
		length = 12
	}

	// Character sets
	const (
		lowercase = "abcdefghijkmnopqrstuvwxyz"    // Removed l (looks like 1)
		uppercase = "ABCDEFGHJKLMNPQRSTUVWXYZ"     // Removed I, O (look like 1, 0)
		digits    = "23456789"                      // Removed 0, 1 (look like O, l)
		special   = "!@#$%^&*"
	)

	all := lowercase + uppercase + digits + special

	// Ensure at least one from each category
	password := make([]byte, length)

	// First 4 characters: one from each category
	password[0] = lowercase[randomInt(len(lowercase))]
	password[1] = uppercase[randomInt(len(uppercase))]
	password[2] = digits[randomInt(len(digits))]
	password[3] = special[randomInt(len(special))]

	// Rest: random from all
	for i := 4; i < length; i++ {
		password[i] = all[randomInt(len(all))]
	}

	// Shuffle the password
	for i := len(password) - 1; i > 0; i-- {
		j := randomInt(i + 1)
		password[i], password[j] = password[j], password[i]
	}

	return string(password), nil
}

// randomInt returns a random integer between 0 and max-1
func randomInt(max int) int {
	b := make([]byte, 1)
	_, _ = rand.Read(b)
	return int(b[0]) % max
}

// generateFormReference generates a unique form reference number
func generateFormReference() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	ref := base64.URLEncoding.EncodeToString(b)
	return fmt.Sprintf("PWR-%s-%s", time.Now().Format("20060102"), ref[:8])
}

// @Summary      Admin Reset User Password
// @Description  Admin-initiated password reset for a user. Can send email, generate form, or both. Only available for non-Employee roles (Employee can use self-service).
// @Tags         Users (Admin)
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "User ID to reset password for"
// @Param        request body AdminResetPasswordRequest true "Reset options"
// @Success      200 {object} AdminResetPasswordResponse "Password reset initiated successfully"
// @Failure      400 {object} ErrorResponse "Invalid request or user ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Access denied or user role not allowed"
// @Failure      404 {object} ErrorResponse "User not found"
// @Failure      500 {object} ErrorResponse "Failed to reset password"
// @Router       /api/users/{id}/admin-reset-password [post]
func AdminResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// Parse request body
	var req AdminResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Default method is email
	if req.Method == "" {
		req.Method = "email"
	}

	// Validate method
	if req.Method != "email" && req.Method != "form" && req.Method != "both" {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid method. Must be 'email', 'form', or 'both'", nil)
		return
	}

	// Get target user — SuperAdmin can reset any user across orgs
	var targetUser *user.User
	if authUser.Role == "SuperAdmin" {
		targetUser, err = userUseCase.GetUserByIDNoOrg(r.Context(), id)
	} else {
		targetUser, err = userUseCase.GetUserByID(r.Context(), id, authUser.OrganizationID)
	}
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}

	// OrgAdmin can only reset users in their own org
	if authUser.Role != "SuperAdmin" && targetUser.OrganizationID != authUser.OrganizationID {
		SendJSONResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}

	// Check if target user's role is Employee - they should use self-service
	if targetUser.Role != nil && strings.ToLower(targetUser.Role.Name) == "employee" {
		SendJSONResponse(w, http.StatusForbidden, "Employee accounts should use self-service password reset", nil)
		return
	}

	response := AdminResetPasswordResponse{
		Success: true,
	}

	// Generate form reference
	formReference := generateFormReference()
	response.FormReference = formReference

	// Generate temporary password if requested
	var tempPassword string
	var hashedPassword string
	if req.GeneratePassword {
		tempPassword, err = generateTemporaryPassword(14)
		if err != nil {
			log.Printf("Error generating temporary password: %v", err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate password", nil)
			return
		}
		response.TemporaryPassword = tempPassword

		// Hash the password
		hashBytes, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to process password", nil)
			return
		}
		hashedPassword = string(hashBytes)

		// Update user password and set must_change_password flag
		if err := userUseCase.UpdatePassword(r.Context(), id, hashedPassword); err != nil {
			log.Printf("Error updating password: %v", err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to update password", nil)
			return
		}

		// Set must_change_password flag
		if userRepo != nil {
			if err := userRepo.SetMustChangePassword(r.Context(), id, true); err != nil {
				log.Printf("Error setting must_change_password flag: %v", err)
			}
		}
	}

	// Determine expiry time
	expiryHours := 24
	if req.ExpiryHours > 0 && req.ExpiryHours <= 72 {
		expiryHours = req.ExpiryHours
	}
	expiresAt := time.Now().Add(time.Duration(expiryHours) * time.Hour)

	// Get IP address
	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Create password reset history record
	var resetID int64
	if passwordResetHistoryRepo != nil {
		resetMethod := internalAuth.ResetMethodAdmin
		if req.Method == "email" {
			resetMethod = internalAuth.ResetMethodEmail
		} else if req.Method == "form" {
			resetMethod = internalAuth.ResetMethodForm
		}

		input := internalAuth.PasswordResetHistoryInput{
			UserID:         id,
			InitiatedByID:  &authUser.UserID,
			ResetMethod:    resetMethod,
			IPAddress:      ipAddress,
			UserAgent:      userAgent,
			FormAttached:   req.Method == "form" || req.Method == "both",
			Notes:          &req.Notes,
			TokenExpiresAt: &expiresAt,
			OrganizationID: &targetUser.OrganizationID,
		}

		resetID, err = passwordResetHistoryRepo.Create(r.Context(), input)
		if err != nil {
			log.Printf("Error creating password reset history: %v", err)
		}
		response.ResetID = resetID
	}

	// Send email if method is email or both
	if req.Method == "email" || req.Method == "both" {
		if emailResolver != nil && req.GeneratePassword {
			// Send email with temporary password
			err = sendPasswordResetEmail(r.Context(), targetUser, tempPassword, expiresAt)
			if err != nil {
				log.Printf("Error sending password reset email: %v", err)
			} else {
				response.EmailSent = true
			}
		} else if req.Method == "email" && !req.GeneratePassword {
			// Generate reset token and send email
			if resetTokenService != nil {
				token, err := resetTokenService.CreateResetToken(
					r.Context(),
					id,
					ipAddress,
					userAgent,
					time.Duration(expiryHours)*time.Hour,
				)
				if err == nil {
					baseURL := getBaseURL()
					resetLink := fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)
					err = sendPasswordResetLinkEmail(r.Context(), targetUser, resetLink, expiresAt)
					if err != nil {
						log.Printf("Error sending password reset link email: %v", err)
					} else {
						response.EmailSent = true
					}
				}
			}
		}
	}

	// Generate PDF form if method is form or both
	if req.Method == "form" || req.Method == "both" {
		pdfData := export.PasswordResetFormData{
			UserFullName:      targetUser.FullName,
			UserEmail:         targetUser.Email,
			UserID:            idStr,
			RequestDate:       time.Now(),
			RequestedBy:       authUser.Email,
			RequestedByRole:   authUser.Role,
			TemporaryPassword: tempPassword,
			ExpiresAt:         expiresAt,
			Notes:             req.Notes,
			FormReference:     formReference,
		}

		// Try to generate PDF with organization details
		var pdfBytes []byte
		if organizationUseCase != nil {
			pdfBytes, err = export.GeneratePasswordResetForm(r.Context(), organizationUseCase, targetUser.OrganizationID, pdfData)
			if err != nil {
				log.Printf("Error generating PDF with org details: %v", err)
				// Fall back to simple PDF
				companyName := "Microfinance"
				if targetUser.Organization != nil {
					companyName = targetUser.Organization.Name
				}
				pdfBytes, err = export.GenerateSimplePasswordResetForm(pdfData, companyName)
			}
		} else {
			// Generate simple PDF without organization details
			companyName := "Microfinance"
			if targetUser.Organization != nil {
				companyName = targetUser.Organization.Name
			}
			pdfBytes, err = export.GenerateSimplePasswordResetForm(pdfData, companyName)
		}

		if err != nil {
			log.Printf("Error generating PDF: %v", err)
		} else if len(pdfBytes) > 0 {
			response.PDFGenerated = true

			// Save PDF to storage if available
			if storageService != nil {
				pdfPath := fmt.Sprintf("password-reset-forms/%d/%s.pdf", targetUser.OrganizationID, formReference)
				reader := bytes.NewReader(pdfBytes)
				err := storageService.Upload(r.Context(), pdfPath, reader, "application/pdf")
				if err != nil {
					log.Printf("Error saving PDF to storage: %v", err)
				} else {
					// Get URL for the uploaded file
					pdfURL, err := storageService.GetURL(r.Context(), pdfPath, 24*time.Hour)
					if err != nil {
						log.Printf("Error getting PDF URL: %v", err)
						response.PDFURL = &pdfPath // Fallback to path
					} else {
						response.PDFURL = &pdfURL
					}
				}
			}
		}
	}

	// Build response message
	var messageParts []string
	if response.EmailSent {
		messageParts = append(messageParts, "email sent")
	}
	if response.PDFGenerated {
		messageParts = append(messageParts, "PDF form generated")
	}
	if req.GeneratePassword {
		messageParts = append(messageParts, "temporary password created")
	}

	if len(messageParts) > 0 {
		response.Message = fmt.Sprintf("Password reset initiated: %s", strings.Join(messageParts, ", "))
	} else {
		response.Message = "Password reset initiated"
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": response.Message,
		"data":    response,
	})
}

// sendPasswordResetEmail sends an email with temporary password
func sendPasswordResetEmail(ctx context.Context, u *user.User, tempPassword string, expiresAt time.Time) error {
	if emailResolver == nil {
		return fmt.Errorf("email service not configured")
	}

	loginURL := getBaseURL() + "/login"
	expiresIn := formatDuration(time.Until(expiresAt))

	// Build email content
	subject := "Your Password Has Been Reset"
	content := fmt.Sprintf(`
Hello %s,

Your password has been reset by an administrator. Here are your new login credentials:

Email: %s
Temporary Password: %s

IMPORTANT: This password will expire in %s (%s).

You can log in here: %s

You will be required to change your password upon first login.

For security reasons:
- Do not share this password with anyone
- Change your password immediately after logging in
- Use a strong, unique password that you don't use elsewhere

If you did not expect this password reset, please contact your administrator immediately.

Best regards,
Microfinance Team
`, u.FullName, u.Email, tempPassword, expiresIn, expiresAt.Format("02 Jan 2006 15:04 MST"), loginURL)

	emailSvc := emailResolver.ForOrg(ctx, u.OrganizationID)
	return emailSvc.SendEmail(ctx, []string{u.Email}, subject, content, false)
}

// sendPasswordResetLinkEmail sends an email with reset link
func sendPasswordResetLinkEmail(ctx context.Context, u *user.User, resetLink string, expiresAt time.Time) error {
	if emailResolver == nil {
		return fmt.Errorf("email service not configured")
	}

	expiresIn := formatDuration(time.Until(expiresAt))

	subject := "Password Reset Request"
	content := fmt.Sprintf(`
Hello %s,

A password reset has been initiated for your account by an administrator.

Click the link below to reset your password:
%s

This link will expire in %s (%s).

If you did not expect this reset, please contact your administrator immediately.

Best regards,
Microfinance Team
`, u.FullName, resetLink, expiresIn, expiresAt.Format("02 Jan 2006 15:04 MST"))

	emailSvc := emailResolver.ForOrg(ctx, u.OrganizationID)
	return emailSvc.SendEmail(ctx, []string{u.Email}, subject, content, false)
}

// formatDuration returns a human-friendly duration string like "24 hours" or "30 minutes".
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "less than a minute"
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours >= 24 {
		days := hours / 24
		if days == 1 {
			return "24 hours"
		}
		return fmt.Sprintf("%d days", days)
	}
	if hours > 0 && minutes > 0 {
		return fmt.Sprintf("%d hours %d minutes", hours, minutes)
	}
	if hours > 0 {
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	if minutes == 1 {
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", minutes)
}

// @Summary      Get Password Reset History
// @Description  Retrieves password reset history for a specific user
// @Tags         Users (Admin)
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "User ID"
// @Param        limit query int false "Number of records to return" default(10)
// @Success      200 {object} SuccessResponse "Password reset history retrieved"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Access denied"
// @Failure      500 {object} ErrorResponse "Failed to retrieve history"
// @Router       /api/users/{id}/password-reset-history [get]
func GetUserPasswordResetHistoryHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// Get limit from query
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Verify user exists — SuperAdmin can view any user across orgs
	var targetUser *user.User
	if authUser.Role == "SuperAdmin" {
		targetUser, err = userUseCase.GetUserByIDNoOrg(r.Context(), id)
	} else {
		targetUser, err = userUseCase.GetUserByID(r.Context(), id, authUser.OrganizationID)
	}
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "User not found", nil)
		return
	}

	// OrgAdmin can only view their own org's history
	if authUser.Role != "SuperAdmin" && targetUser.OrganizationID != authUser.OrganizationID {
		SendJSONResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}

	if passwordResetHistoryRepo == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Password reset history not configured", nil)
		return
	}

	history, err := passwordResetHistoryRepo.GetByUserID(r.Context(), id, limit)
	if err != nil {
		log.Printf("Error getting password reset history: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to retrieve history", nil)
		return
	}

	// Transform to match frontend expectations
	type PasswordResetHistoryDTO struct {
		ID                         int64  `json:"id"`
		UserID                     string `json:"user_id"`
		ResetByUserID              string `json:"reset_by_user_id"`
		ResetByName                string `json:"reset_by_name"`
		Method                     string `json:"method"`
		FormReference              string `json:"form_reference,omitempty"`
		Notes                      string `json:"notes,omitempty"`
		TemporaryPasswordGenerated bool   `json:"temporary_password_generated"`
		ResetAt                    string `json:"reset_at"`
		ExpiryTime                 string `json:"expiry_time,omitempty"`
		PDFGenerated               bool   `json:"pdf_generated"`
		EmailSent                  bool   `json:"email_sent"`
	}

	dtoList := make([]PasswordResetHistoryDTO, 0, len(history))
	for _, h := range history {
		// Map reset method to frontend expected format
		method := string(h.ResetMethod)
		// Map backend method types to frontend expectations ('email' | 'form' | 'both')
		switch h.ResetMethod {
		case internalAuth.ResetMethodEmail:
			method = "email"
		case internalAuth.ResetMethodForm:
			method = "form"
		case internalAuth.ResetMethodAdmin:
			method = "both" // Admin reset can be both email and form
		default:
			method = "email"
		}

		dto := PasswordResetHistoryDTO{
			ID:        h.ID,
			UserID:    fmt.Sprintf("%d", h.UserID),
			Method:    method,
			ResetAt:   h.CreatedAt.Format(time.RFC3339),
			EmailSent: false,
		}

		// Use database values directly
		if h.InitiatedByID != nil {
			dto.ResetByUserID = fmt.Sprintf("%d", *h.InitiatedByID)
		}
		if h.InitiatedByName != nil {
			dto.ResetByName = *h.InitiatedByName
		} else {
			dto.ResetByName = "System"
		}
		if h.FormURL != nil {
			dto.FormReference = *h.FormURL
		}
		if h.Notes != nil {
			dto.Notes = *h.Notes
		}
		if h.TokenExpiresAt != nil {
			dto.ExpiryTime = h.TokenExpiresAt.Format(time.RFC3339)
		}

		// Use tracked values from database
		dto.PDFGenerated = h.FormAttached
		dto.TemporaryPasswordGenerated = false // Will be populated once DB is updated
		dto.EmailSent = h.ResetMethod == internalAuth.ResetMethodEmail || h.ResetMethod == internalAuth.ResetMethodAdmin

		dtoList = append(dtoList, dto)
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": "Password reset history retrieved",
		"data":    dtoList,
	})
}

// @Summary      Download Password Reset Form
// @Description  Downloads the PDF password reset form for a specific reset request
// @Tags         Users (Admin)
// @Security     BearerAuth
// @Produce      application/pdf
// @Param        id path int true "Reset History ID"
// @Success      200 {file} application/pdf "PDF file"
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Form not found"
// @Router       /api/password-reset-forms/{id}/download [get]
func DownloadPasswordResetFormHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid ID", nil)
		return
	}

	if passwordResetHistoryRepo == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Password reset history not configured", nil)
		return
	}

	// Get the reset record
	record, err := passwordResetHistoryRepo.GetByID(r.Context(), id)
	if err != nil {
		SendJSONResponse(w, http.StatusNotFound, "Reset record not found", nil)
		return
	}

	// Check authorization
	if authUser.Role != "SuperAdmin" && record.OrganizationID != nil && *record.OrganizationID != authUser.OrganizationID {
		SendJSONResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}

	// Check if form was generated
	if !record.FormAttached || record.FormURL == nil || *record.FormURL == "" {
		SendJSONResponse(w, http.StatusNotFound, "No PDF form available for this reset request", nil)
		return
	}

	// Download from storage
	if storageService == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Storage service not configured", nil)
		return
	}

	readCloser, err := storageService.Download(r.Context(), *record.FormURL)
	if err != nil {
		log.Printf("Error retrieving PDF from storage: %v", err)
		SendJSONResponse(w, http.StatusNotFound, "PDF form not found", nil)
		return
	}
	defer readCloser.Close()

	// Read all bytes from the reader
	pdfBytes, err := io.ReadAll(readCloser)
	if err != nil {
		log.Printf("Error reading PDF: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to read PDF", nil)
		return
	}

	// Set headers and send file
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"password-reset-%d.pdf\"", id))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	w.Write(pdfBytes)
}
