package super

import (
	"net/http"

	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/gorilla/mux"
)

// Permission constants for SuperAdmin user management
const (
	PermUserCreate         = "admin.user.create"
	PermUserView           = "admin.user.view"
	PermUserEdit           = "admin.user.edit"
	PermUserDelete         = "admin.user.delete"
	PermUserEditStatus     = "admin.user.edit_status"
	PermUserViewEmail      = "admin.user.view_email"
	PermUserUnlock         = "admin.user.unlock"
	PermUserViewSecurity   = "admin.user.view_security"
	PermUserResetPassword  = "admin.user.reset_password"
	PermUserViewResetHist  = "admin.user.view_reset_history"
)

func registerUserRoutes(r *mux.Router) {
	r.HandleFunc("/users/me", handler.GetAuthenticatedUserHandler).
		Methods("GET", "OPTIONS")

	r.Handle("/users/token", middleware.ChainMiddleware(
		http.HandlerFunc(handler.GetMyTokenHandler),
	)).Methods("GET", "OPTIONS")

	r.Handle("/logout", middleware.ChainMiddleware(
		http.HandlerFunc(handler.LogoutHandler),
	)).Methods("POST", "OPTIONS")

	// Create user
	r.Handle("/users/create", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.CreateUserHandler, nil, []string{PermUserCreate}),
	)).Methods("POST", "OPTIONS")

	// List users (HQ/SuperAdmin - all organizations)
	r.Handle("/users", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.GetAllUsersForHQHandler, nil, []string{PermUserView}),
	)).Methods("GET", "OPTIONS")

	// User counts for tab badges
	r.Handle("/users/counts", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.GetUserCountsHandler, nil, []string{PermUserView}),
	)).Methods("GET", "OPTIONS")

	// Get user by ID
	r.Handle("/users/{id}", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.GetUserByIDHandler, nil, []string{PermUserView}),
		middleware.DecryptMiddleware("id"),
	)).Methods("GET", "OPTIONS")

	// Update user status
	r.Handle("/users/{id}/status", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.UpdateUserStatusHandler, nil, []string{PermUserEditStatus}),
		middleware.DecryptMiddleware("id"),
	)).Methods("PUT", "OPTIONS")

	// Get user by email
	r.Handle("/users/email", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.GetUserByEmailHandler, nil, []string{PermUserViewEmail}),
	)).Methods("GET", "OPTIONS")

	// Update user details (HQ — cross-org)
	r.Handle("/users/{id}", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.UpdateUserForHQHandler, nil, []string{PermUserEdit}),
		middleware.DecryptMiddleware("id"),
	)).Methods("PUT", "OPTIONS")

	// Delete user (HQ — cross-org)
	r.Handle("/users/{id}", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.SoftDeleteUserForHQHandler, nil, []string{PermUserDelete}),
		middleware.DecryptMiddleware("id"),
	)).Methods("DELETE", "OPTIONS")

	// Lock user account (admin only)
	r.Handle("/users/{id}/lock", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.LockUserAccountHandler, nil, []string{PermUserUnlock}),
		middleware.DecryptMiddleware("id"),
	)).Methods("POST", "OPTIONS")

	// Unlock user account (admin only)
	r.Handle("/users/{id}/unlock", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.UnlockUserAccountHandler, nil, []string{PermUserUnlock}),
		middleware.DecryptMiddleware("id"),
	)).Methods("POST", "OPTIONS")

	// Get user security info (admin only)
	r.Handle("/users/{id}/security", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.GetUserSecurityInfoHandler, nil, []string{PermUserViewSecurity}),
		middleware.DecryptMiddleware("id"),
	)).Methods("GET", "OPTIONS")

	// Admin password reset
	r.Handle("/users/{id}/admin-reset-password", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.AdminResetPasswordHandler, nil, []string{PermUserResetPassword}),
		middleware.DecryptMiddleware("id"),
	)).Methods("POST", "OPTIONS")

	// Get password reset history for a user
	r.Handle("/users/{id}/password-reset-history", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.GetUserPasswordResetHistoryHandler, nil, []string{PermUserViewResetHist}),
		middleware.DecryptMiddleware("id"),
	)).Methods("GET", "OPTIONS")

	// Download password reset form PDF
	r.Handle("/password-reset-forms/{id}/download", middleware.ChainMiddleware(
		wrapWithAuthorize(handler.DownloadPasswordResetFormHandler, nil, []string{PermUserViewResetHist}),
	)).Methods("GET", "OPTIONS")
}
