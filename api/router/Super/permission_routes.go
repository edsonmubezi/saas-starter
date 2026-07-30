package super

import (
	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"

	"github.com/gorilla/mux"
)

// Permission constants for SuperAdmin permission management
// Admin-only permissions - tenant routes moved to Orgs/permission_routes.go
const (
	PermissionView = "admin.permission.view"
	PermissionEdit = "admin.permission.edit"
)

func registerPermissionRoutes(r *mux.Router) {
	// ===== View single permission (admin only)
	r.Handle("/permissions/{id}",
		middleware.ChainMiddleware(
			wrapWithAuthorize(handler.GetPermissionByIDHandler, nil, []string{PermissionView}),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("GET", "OPTIONS")

	// ===== Grouped, paginated list for HQ (ALL permissions including admin.*)
	r.Handle("/permissions",
		middleware.ChainMiddleware(
			wrapWithAuthorize(handler.AdminGetPermissionResourcesPaginatedHandler, nil, []string{PermissionView}),
		),
	).Methods("GET", "OPTIONS")

	// ===== Grouped, non-paginated list (exports/dropdowns) - all scopes
	r.Handle("/permissions-all",
		middleware.ChainMiddleware(
			wrapWithAuthorize(handler.GetPermissionResourcesAdminHandler, nil, []string{PermissionView}),
		),
	).Methods("GET", "OPTIONS")

	// ===== Update permission description (admin only)
	r.Handle("/permissions/{id}/description",
		middleware.ChainMiddleware(
			wrapWithAuthorize(handler.UpdatePermissionDescriptionHandler, nil, []string{PermissionEdit}),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("PATCH", "OPTIONS")
}
