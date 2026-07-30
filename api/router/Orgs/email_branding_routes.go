package orgs

import (
	"net/http"

	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/gorilla/mux"
)

const permEmailBrandingManage = "tenant.email_branding.manage"

func registerEmailBrandingRoutes(protected *mux.Router) {
	perm := middleware.RequireAnyPermission([]string{permEmailBrandingManage})

	protected.Handle("/organization/email-branding",
		perm(http.HandlerFunc(handler.GetEmailBrandingHandler)),
	).Methods("GET", "OPTIONS")

	protected.Handle("/organization/email-branding",
		perm(http.HandlerFunc(handler.SaveEmailBrandingHandler)),
	).Methods("PUT", "OPTIONS")
}
