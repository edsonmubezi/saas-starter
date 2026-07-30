package orgs

import (
	"net/http"

	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/gorilla/mux"
)

const permDocBrandingManage = "tenant.document_branding.manage"

func registerDocumentBrandingRoutes(protected *mux.Router) {
	perm := middleware.RequireAnyPermission([]string{permDocBrandingManage})

	protected.Handle("/organization/document-branding",
		perm(http.HandlerFunc(handler.GetDocumentBrandingHandler)),
	).Methods("GET", "OPTIONS")

	protected.Handle("/organization/document-branding",
		perm(http.HandlerFunc(handler.SaveDocumentBrandingHandler)),
	).Methods("PUT", "OPTIONS")

	protected.Handle("/organization/document-branding/watermark",
		perm(http.HandlerFunc(handler.UploadWatermarkImageHandler)),
	).Methods("POST", "OPTIONS")

	// Available fonts list (no special permission needed beyond auth)
	protected.Handle("/organization/document-branding/fonts",
		http.HandlerFunc(handler.GetAvailableFontsHandler),
	).Methods("GET", "OPTIONS")
}
