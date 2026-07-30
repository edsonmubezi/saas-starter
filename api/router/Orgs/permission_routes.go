package orgs

import (
	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"

	"github.com/gorilla/mux"
)

// Permission constants for tenant permission management
const (
	PermTenantPermissionView = "tenant.permission.view"
)

func registerPermissionRoutes(r *mux.Router) {
	// ===== View single permission (tenant)
	r.Handle("/permissions/{id}",
		middleware.ChainMiddleware(
			wrapWithAuthorize(handler.GetPermissionByIDHandler, nil, []string{PermTenantPermissionView}),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("GET", "OPTIONS")

	// ===== Grouped, paginated list for tenant (tenant.* only, excludes admin.*)
	r.Handle("/permissions",
		middleware.ChainMiddleware(
			wrapWithAuthorize(handler.GetPermissionRrscForNonHqHandler, nil, []string{PermTenantPermissionView}),
		),
	).Methods("GET", "OPTIONS")

	// ===== Grouped, non-paginated list (exports/dropdowns) - tenant scope
	r.Handle("/permissions-all",
		middleware.ChainMiddleware(
			wrapWithAuthorize(handler.GetPermissionResourcesHandler, nil, []string{PermTenantPermissionView}),
		),
	).Methods("GET", "OPTIONS")
}
