package orgs

import (
	"net/http"

	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/gorilla/mux"
)

// registerImportLogRoutes registers import log routes
func registerImportLogRoutes(protected *mux.Router) {
	anyImportPerm := middleware.RequireAnyPermission([]string{"tenant.employee.import_export", "tenant.payroll.view"})

	protected.Handle("/import-logs",
		anyImportPerm(http.HandlerFunc(handler.GetImportSessionsHandler)),
	).Methods("GET")

	protected.Handle("/import-logs/{id}/records",
		anyImportPerm(http.HandlerFunc(handler.GetImportFailedRecordsHandler)),
	).Methods("GET")

	protected.Handle("/import-logs/{id}/duplicates",
		anyImportPerm(http.HandlerFunc(handler.GetDuplicateRecordsHandler)),
	).Methods("GET")
}
