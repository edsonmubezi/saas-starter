package orgs

import (
	"net/http"

	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/gorilla/mux"
)

// registerEmailConfigRoutes registers email config admin routes.
// Email config is org-wide (used by broadcasts, salary slips, leave, 2FA, etc.).
func registerEmailConfigRoutes(protected *mux.Router) {
	viewPerm := middleware.RequireAnyPermission([]string{"tenant.broadcast.view", "tenant.recruitment.settings.manage"})
	managePerm := middleware.RequireAnyPermission([]string{"tenant.broadcast.create", "tenant.recruitment.settings.manage"})

	protected.Handle("/email-config",
		viewPerm(http.HandlerFunc(handler.GetEmailConfigHandler)),
	).Methods("GET", "OPTIONS")

	protected.Handle("/email-config",
		managePerm(http.HandlerFunc(handler.UpsertEmailConfigHandler)),
	).Methods("PUT", "OPTIONS")
}
