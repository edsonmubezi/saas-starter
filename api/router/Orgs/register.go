package orgs

import (
	"net/http"

	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/edsonmubezi/myapp/internal/chat"
	"github.com/edsonmubezi/myapp/internal/emailconfig"
	"github.com/edsonmubezi/myapp/internal/importlog"
	"github.com/edsonmubezi/myapp/internal/organization"
	"github.com/edsonmubezi/myapp/internal/orgsettings"
	"github.com/gorilla/mux"
)

// wrapWithAuthorize wraps handler with authorization middleware
func wrapWithAuthorize(h http.HandlerFunc, roles []string, permissions []string) http.HandlerFunc {
	mw := middleware.AuthorizationMiddleware(roles, permissions)
	return func(w http.ResponseWriter, r *http.Request) {
		mw(h).ServeHTTP(w, r)
	}
}

// RegisterOrganizationLevelRoutes registers all routes available to Organization admins and above
func RegisterOrganizationLevelRoutes(protected *mux.Router, useCases map[string]interface{}) {
	// Register Organization use cases
	for name, useCase := range useCases {
		switch name {
		case "organization":
			if uc, ok := useCase.(organization.OrganizationUseCase); ok {
				handler.SetOrganizationUseCase(uc)
			}
		case "orgcoresettings":
			if uc, ok := useCase.(orgsettings.OrganizationSettingsUseCase); ok {
				handler.SetOrgSettingsUseCase(uc)
			}
		case "emailconfig":
			if uc, ok := useCase.(emailconfig.UseCase); ok {
				handler.SetEmailConfigUseCase(uc)
			}
		case "importlog":
			if uc, ok := useCase.(importlog.UseCase); ok {
				handler.SetImportLogUseCase(uc)
			}
		case "chat":
			if uc, ok := useCase.(chat.UseCase); ok {
				handler.SetChatUseCase(uc)
			}
		case "knowledgerepo":
			if repo, ok := useCase.(chat.KnowledgeRepository); ok {
				handler.SetKnowledgeRepository(repo)
			}
		}
	}

	registerUserRoutes(protected)
	registerRoleRoutes(protected)
	registerPermissionRoutes(protected)
	registerEmailConfigRoutes(protected)
	registerDocumentBrandingRoutes(protected)
	registerEmailBrandingRoutes(protected)
	registerChatRoutes(protected)
	registerImportLogRoutes(protected)
}
