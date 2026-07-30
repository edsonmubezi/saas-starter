package super

import (
	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/gorilla/mux"
)

// Permission constants for SuperAdmin organization management
const (
	PermOrgCreate = "admin.organization.create"
	PermOrgView   = "admin.organization.view"
	PermOrgEdit   = "admin.organization.edit"
	PermOrgDelete = "admin.organization.delete"
)

func registerOrganizationRoutes(r *mux.Router) {

	r.Handle("/organizations",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.CreateOrganizationHandler,
				nil,
				[]string{PermOrgCreate},
			),
		),
	).Methods("POST", "OPTIONS")

	r.Handle("/encrpted-organizations",
		wrapWithAuthorize(
			handler.GetAllOrgsEncrptyedHandler,
			nil,
			[]string{PermOrgView},
		),
	).Methods("GET", "OPTIONS")

	r.Handle("/organizations",
		wrapWithAuthorize(
			handler.GetAllOrganizationsHandler,
			nil,
			[]string{PermOrgView},
		),
	).Methods("GET", "OPTIONS")

	r.Handle("/all-organizations",
		wrapWithAuthorize(
			handler.GetOrganizationsHandler,
			nil,
			[]string{PermOrgView},
		),
	).Methods("GET", "OPTIONS")

	// Route for /organizations/all - must be before /{id} to prevent "all" being treated as an ID
	r.Handle("/organizations/all",
		wrapWithAuthorize(
			handler.GetOrganizationsHandler,
			nil,
			[]string{PermOrgView},
		),
	).Methods("GET", "OPTIONS")

	r.Handle("/organizations/{id}",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.GetOrganizationByIDHandler,
				nil,
				[]string{PermOrgView},
			),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("GET", "OPTIONS")

	r.Handle("/organizations/{id}",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.UpdateOrganizationHandler,
				nil,
				[]string{PermOrgEdit},
			),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("PUT", "OPTIONS")

	// Update organization with logo upload (supports multipart/form-data)
	r.Handle("/organizations/{id}/update-with-logo",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.UpdateOrganizationWithLogoHandler,
				nil,
				[]string{PermOrgEdit},
			),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("PUT", "OPTIONS")

	r.Handle("/organizations/{id}",
		middleware.ChainMiddleware(
			wrapWithAuthorize(
				handler.SoftDeleteOrganizationHandler,
				nil,
				[]string{PermOrgDelete},
			),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("DELETE", "OPTIONS")

}
