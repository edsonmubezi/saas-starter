package orgs

import (
	"net/http"

	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/gorilla/mux"
)

// Permission constants for tenant user management
const (
	PermOrgUserCreate     = "tenant.user.create"
	PermOrgUserView       = "tenant.user.view"
	PermOrgUserEdit       = "tenant.user.edit"
	PermOrgUserDelete     = "tenant.user.delete"
	PermOrgUserEditStatus = "tenant.user.edit_status"
	PermOrgUserViewEmail  = "tenant.user.view_email"
	PermOrgUserUnlock     = "tenant.user.unlock"
	PermOrgRoleAssign     = "tenant.role.assign"
)

func registerUserRoutes(r *mux.Router) {
	// ==========================================================================
	// ORG-SCOPED USER ROUTES
	// All routes use dedicated org-scoped handlers that enforce data isolation.
	// Organization context is extracted from JWT and enforced at handler level.
	// ==========================================================================

	// Create user (organization level) - uses org-scoped handler
	r.Handle("/org-users/create", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.CreateOrgUserHandler, nil, []string{PermOrgUserCreate}),
	)).Methods("POST", "OPTIONS")

	// List users (organization level - restricted to caller's organization)
	// Uses OR logic so features needing user selection (discipline committee, supervisor assignment, etc.) can access
	r.Handle("/org-users", middleware.ChainMiddleware(
		http.HandlerFunc(handler.GetOrganizationUsersHandler),
		middleware.RequireAnyPermission([]string{
			PermOrgUserView,
			PermOrgUserCreate,
			PermOrgUserEdit,
			"tenant.employee.view",
			"tenant.employee.edit",
			"tenant.disciplinary_case.create",
			"tenant.disciplinary_case.committee",
			"tenant.leave.approve",
			"tenant.recruitment.manage",
		}),
	)).Methods("GET", "OPTIONS")

	// Get user by ID (organization level) - uses org-scoped handler
	r.Handle("/org-users/{id}", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.GetOrgUserByIDHandler, nil, []string{PermOrgUserView}),
		middleware.DecryptMiddleware("id"),
	)).Methods("GET", "OPTIONS")

	// Update user details (organization level) - uses org-scoped handler
	r.Handle("/org-users/{id}", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.UpdateOrgUserHandler, nil, []string{PermOrgUserEdit}),
		middleware.DecryptMiddleware("id"),
	)).Methods("PUT", "OPTIONS")

	// Delete user (organization level - soft delete) - uses org-scoped handler
	r.Handle("/org-users/{id}", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.DeleteOrgUserHandler, nil, []string{PermOrgUserDelete}),
		middleware.DecryptMiddleware("id"),
	)).Methods("DELETE", "OPTIONS")

	// Update user status (organization level) - uses org-scoped handler
	r.Handle("/org-users/{id}/status", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.UpdateOrgUserStatusHandler, nil, []string{PermOrgUserEditStatus}),
		middleware.DecryptMiddleware("id"),
	)).Methods("PUT", "OPTIONS")

	// Lock user account (organization level)
	r.Handle("/org-users/{id}/lock", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.LockUserAccountHandler, nil, []string{PermOrgUserUnlock}),
		middleware.DecryptMiddleware("id"),
	)).Methods("POST", "OPTIONS")

	// Unlock user account (organization level)
	r.Handle("/org-users/{id}/unlock", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.UnlockUserAccountHandler, nil, []string{PermOrgUserUnlock}),
		middleware.DecryptMiddleware("id"),
	)).Methods("POST", "OPTIONS")

	// Password reset (organization level) - uses org-scoped handler
	r.Handle("/org-users/{id}/reset-password", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.SendOrgPasswordResetHandler, nil, []string{PermOrgUserEdit}),
		middleware.DecryptMiddleware("id"),
	)).Methods("POST", "OPTIONS")

	// Password reset history (organization level) - handler enforces org isolation
	r.Handle("/org-users/{id}/password-reset-history", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.GetUserPasswordResetHistoryHandler, nil, []string{PermOrgUserEdit}),
		middleware.DecryptMiddleware("id"),
	)).Methods("GET", "OPTIONS")

	// Download password reset form PDF (organization level) - handler enforces org isolation
	r.Handle("/password-reset-forms/{id}/download", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.DownloadPasswordResetFormHandler, nil, []string{PermOrgUserEdit}),
	)).Methods("GET", "OPTIONS")
}
