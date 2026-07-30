package super

import (
	"net/http"

	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/edsonmubezi/myapp/internal/organization"
	"github.com/edsonmubezi/myapp/internal/orgsettings"
	"github.com/edsonmubezi/myapp/internal/permission"
	"github.com/edsonmubezi/myapp/internal/role"
	"github.com/edsonmubezi/myapp/internal/user"
	"github.com/gorilla/mux"
)

// wrapWithAuthorize wraps handler with authorization middleware
func wrapWithAuthorize(h http.HandlerFunc, roles []string, permissions []string) http.HandlerFunc {
	mw := middleware.AuthorizationMiddleware(roles, permissions)
	return func(w http.ResponseWriter, r *http.Request) {
		mw(h).ServeHTTP(w, r)
	}
}

// RegisterSuperAdminLevelRoutes registers all routes available only to SuperAdmin
func RegisterSuperAdminLevelRoutes(protected *mux.Router, useCases map[string]interface{}) organization.OrganizationUseCase {
	var orgUC organization.OrganizationUseCase

	// Register SuperAdmin use cases
	for name, useCase := range useCases {
		switch name {
		case "user":
			if uc, ok := useCase.(user.UserUseCase); ok {
				handler.SetUserUseCase(uc)
			}
		case "role":
			if uc, ok := useCase.(role.RoleUseCase); ok {
				handler.SetRoleUseCase(uc)
			}
		case "permission":
			if uc, ok := useCase.(permission.PermissionUseCase); ok {
				handler.SetPermissionUseCase(uc)
			}
		case "organization":
			if uc, ok := useCase.(organization.OrganizationUseCase); ok {
				orgUC = uc
				handler.SetOrganizationUseCase(uc)
			}
		case "orgcoresettings":
			if uc, ok := useCase.(orgsettings.OrganizationSettingsUseCase); ok {
				handler.SetOrgSettingsUseCase(uc)
			}
		}
	}

	// Only register routes if protected router is provided
	if protected != nil {
		registerUserRoutes(protected)
		registerRoleRoutes(protected)
		registerPermissionRoutes(protected)
		registerOrganizationRoutes(protected)
		registerOrgSettingsRoutes(protected)
	}

	return orgUC
}
