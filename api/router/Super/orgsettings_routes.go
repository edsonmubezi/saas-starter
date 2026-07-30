package super

import (
	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/gorilla/mux"
)

// Permission constants for SuperAdmin org settings
const (
	// Organization settings permissions
	OrgSettingCreate = "admin.org_settings.create"
	OrgSettingView   = "admin.org_settings.view"
	OrgSettingEdit   = "admin.org_settings.edit"
)

func registerOrgSettingsRoutes(r *mux.Router) {

	// Create organization settings
	r.Handle("/org-settings",
		middleware.ChainMiddleware(
			wrapWithAuthorize(handler.CreateOrgSettingsHandler, nil, []string{OrgSettingCreate}),
		),
	).Methods("POST", "OPTIONS")

	// Get organization settings by ID
	r.Handle("/org-settings/{id}",
		middleware.ChainMiddleware(
			wrapWithAuthorize(handler.GetOrgSettingsByIDHandler, nil, []string{OrgSettingView}),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("GET", "OPTIONS")

	// Get organization settings by organization ID
	r.Handle("/org-settings/organization/{organizationId}",
		middleware.ChainMiddleware(
			wrapWithAuthorize(handler.GetOrgSettingsByOrgIDHandler, nil, []string{OrgSettingView}),
			middleware.DecryptMiddleware("organizationId"),
		),
	).Methods("GET", "OPTIONS")

	// Create organization settings by organization ID (with encrypted ID in URL)
	r.Handle("/org-settings/organization/{organizationId}",
		middleware.ChainMiddleware(
			wrapWithAuthorize(handler.CreateOrgSettingsByOrgIDHandler, nil, []string{OrgSettingCreate}),
			middleware.DecryptMiddleware("organizationId"),
		),
	).Methods("POST", "OPTIONS")

	// Update organization settings
	r.Handle("/org-settings",
		middleware.ChainMiddleware(
			wrapWithAuthorize(handler.UpdateOrgSettingsHandler, nil, []string{OrgSettingEdit}),
		),
	).Methods("PUT", "OPTIONS")
}
