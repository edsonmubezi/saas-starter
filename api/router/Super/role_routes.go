package super

import (
	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/gorilla/mux"
)

// Permission constants for SuperAdmin role management
// Admin-only permissions - tenant routes moved to Orgs/role_routes.go
const (
	PermRoleCreate = "admin.role.create"
	PermRoleView   = "admin.role.view"
	PermRoleEdit   = "admin.role.edit"
	PermRoleDelete = "admin.role.delete"
	PermRoleAssign = "admin.role.assign"
)

func registerRoleRoutes(r *mux.Router) {
	// Create role (admin only)
	r.Handle("/roles",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.CreateRoleHandler,
				nil,
				[]string{PermRoleCreate},
			),
		),
	).Methods("POST", "OPTIONS")

	// Get system roles only (SuperAdmin, OrgsAdmin - non-assignable)
	// Must be registered before /roles/{id} to avoid {id} matching "system"
	r.Handle("/roles/system",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.GetSystemRolesHandler,
				nil,
				[]string{PermRoleView},
			),
		),
	).Methods("GET", "OPTIONS")

	// Assign permissions to role (admin only)
	r.Handle("/roles/{id}/permissions",
		middleware.ChainMiddleware(
			wrapWithAuthorize(handler.AssignPermissionsToRoleHandler, nil, []string{PermRoleAssign}),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("PUT", "OPTIONS")

	// Get role by ID (admin only)
	r.Handle("/roles/{id}",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.GetRoleByIDHandler,
				nil,
				[]string{PermRoleView},
			),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("GET", "OPTIONS")

	// Get all roles including system roles (admin only)
	r.Handle("/roles",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.GetAllRolesAdminHandler,
				nil,
				[]string{PermRoleView},
			),
		),
	).Methods("GET", "OPTIONS")

	// Get all roles paginated including system roles (admin only)
	r.Handle("/roles-paginated",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.GetAllRolesAdminPaginatedHandler,
				nil,
				[]string{PermRoleView},
			),
		),
	).Methods("GET", "OPTIONS")

	// Update role (admin only)
	r.Handle("/roles/{id}",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.UpdateRoleHandler,
				nil,
				[]string{PermRoleEdit},
			),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("PUT", "OPTIONS")

	// Delete role (admin only)
	r.Handle("/roles/{id}",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.DeleteRoleHandler,
				nil,
				[]string{PermRoleDelete},
			),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("DELETE", "OPTIONS")

	// Get role with permissions (admin only)
	r.Handle("/roles/{id}/details",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.GetRoleWithPermissionsHandler,
				nil,
				[]string{PermRoleView},
			),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("GET", "OPTIONS")
}
