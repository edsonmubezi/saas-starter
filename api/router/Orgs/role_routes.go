package orgs

import (
	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/gorilla/mux"
)

// Permission constants for tenant role management
const (
	PermTenantRoleCreate = "tenant.role.create"
	PermTenantRoleView   = "tenant.role.view"
	PermTenantRoleEdit   = "tenant.role.edit"
	PermTenantRoleDelete = "tenant.role.delete"
	PermTenantRoleAssign = "tenant.role.assign"
)

func registerRoleRoutes(r *mux.Router) {
	// Create role (tenant)
	r.Handle("/roles",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.CreateRoleHandler,
				nil,
				[]string{PermTenantRoleCreate},
			),
		),
	).Methods("POST", "OPTIONS")

	// Assign permissions to role (tenant)
	r.Handle("/roles/{id}/permissions",
		middleware.ChainMiddleware(
			wrapWithAuthorize(handler.AssignPermissionsToRoleHandler, nil, []string{PermTenantRoleAssign}),
		),
	).Methods("PUT", "OPTIONS")

	// Get role by ID (tenant)
	r.Handle("/roles/{id}",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.GetRoleByIDHandler,
				nil,
				[]string{PermTenantRoleView},
			),
		),
	).Methods("GET", "OPTIONS")

	// Get all roles (tenant)
	r.Handle("/roles",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.GetAllRolesHandler,
				nil,
				[]string{PermTenantRoleView},
			),
		),
	).Methods("GET", "OPTIONS")

	// Get all roles paginated (tenant)
	r.Handle("/roles-paginated",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.GetAllRolesPaginatedHandler,
				nil,
				[]string{PermTenantRoleView},
			),
		),
	).Methods("GET", "OPTIONS")

	// Update role (tenant)
	r.Handle("/roles/{id}",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.UpdateRoleHandler,
				nil,
				[]string{PermTenantRoleEdit},
			),
		),
	).Methods("PUT", "OPTIONS")

	// Delete role (tenant)
	r.Handle("/roles/{id}",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.DeleteRoleHandler,
				nil,
				[]string{PermTenantRoleDelete},
			),
		),
	).Methods("DELETE", "OPTIONS")

	// Get role with permissions (tenant)
	r.Handle("/roles/{id}/details",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.GetRoleWithPermissionsHandler,
				nil,
				[]string{PermTenantRoleView},
			),
		),
	).Methods("GET", "OPTIONS")
}
