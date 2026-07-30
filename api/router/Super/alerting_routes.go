package super

import (
	"log"
	"net/http"

	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/edsonmubezi/myapp/internal/platform/alerting"
	"github.com/gorilla/mux"
)

// Permission constants for admin alerting management
const (
	PermAdminAlertingView   = "admin.alerting.view"
	PermAdminAlertingManage = "admin.alerting.manage"
)

// RegisterAdminAlertingRoutes registers alerting routes for admin (HQ) users
// These routes are accessed via /api/admin/alerting/*
func RegisterAdminAlertingRoutes(r *mux.Router, alertingService *alerting.Service) {
	// Create repository for alert configs
	alertRepo := alerting.NewRepository()
	alertingHandler := handler.NewAlertingHandler(alertRepo, alertingService)

	// Create subrouter for alerting
	alertingRouter := r.PathPrefix("/alerting").Subrouter()

	// View routes - require view permission
	alertingRouter.Handle("/configs",
		middleware.RequireAnyPermission([]string{PermAdminAlertingView})(
			http.HandlerFunc(alertingHandler.GetAlertConfigs),
		),
	).Methods("GET", "OPTIONS")

	alertingRouter.Handle("/configs/{id}",
		middleware.RequireAnyPermission([]string{PermAdminAlertingView})(
			http.HandlerFunc(alertingHandler.GetAlertConfigByID),
		),
	).Methods("GET", "OPTIONS")

	alertingRouter.Handle("/history",
		middleware.RequireAnyPermission([]string{PermAdminAlertingView})(
			http.HandlerFunc(alertingHandler.GetAlertHistory),
		),
	).Methods("GET", "OPTIONS")

	alertingRouter.Handle("/channels",
		middleware.RequireAnyPermission([]string{PermAdminAlertingView})(
			http.HandlerFunc(alertingHandler.GetConfiguredChannels),
		),
	).Methods("GET", "OPTIONS")

	// Management routes - require manage permission
	alertingRouter.Handle("/configs",
		middleware.RequireAnyPermission([]string{PermAdminAlertingManage})(
			http.HandlerFunc(alertingHandler.CreateAlertConfig),
		),
	).Methods("POST", "OPTIONS")

	alertingRouter.Handle("/configs/{id}",
		middleware.RequireAnyPermission([]string{PermAdminAlertingManage})(
			http.HandlerFunc(alertingHandler.UpdateAlertConfig),
		),
	).Methods("PUT", "OPTIONS")

	alertingRouter.Handle("/configs/{id}",
		middleware.RequireAnyPermission([]string{PermAdminAlertingManage})(
			http.HandlerFunc(alertingHandler.DeleteAlertConfig),
		),
	).Methods("DELETE", "OPTIONS")

	alertingRouter.Handle("/test/{channel}",
		middleware.RequireAnyPermission([]string{PermAdminAlertingManage})(
			http.HandlerFunc(alertingHandler.TestAlertChannel),
		),
	).Methods("POST", "OPTIONS")

	log.Println("  Admin alerting routes registered at /admin/alerting/*")
}
