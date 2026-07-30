package orgs

import (
	"log"
	"net/http"

	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/edsonmubezi/myapp/internal/platform/alerting"
	"github.com/gorilla/mux"
)

// Permission constants for tenant alerting management
const (
	PermTenantAlertingView   = "tenant.alerting.view"
	PermTenantAlertingManage = "tenant.alerting.manage"
	// Admin permissions (for SuperAdmin accessing org routes)
	PermAdminAlertingView   = "admin.alerting.view"
	PermAdminAlertingManage = "admin.alerting.manage"
)

// RegisterTenantAlertingRoutes registers alerting routes for tenant (organization) users
// These routes are accessed via /api/org/alerting/*
func RegisterTenantAlertingRoutes(r *mux.Router, alertingService *alerting.Service) {
	// Create repository for alert configs
	alertRepo := alerting.NewRepository()
	alertingHandler := handler.NewAlertingHandler(alertRepo, alertingService)

	// Create subrouter for alerting
	alertingRouter := r.PathPrefix("/alerting").Subrouter()

	// View routes - require view permission (tenant OR admin)
	alertingRouter.Handle("/configs",
		middleware.RequireAnyPermission([]string{PermTenantAlertingView, PermAdminAlertingView})(
			http.HandlerFunc(alertingHandler.GetAlertConfigs),
		),
	).Methods("GET", "OPTIONS")

	alertingRouter.Handle("/configs/{id}",
		middleware.RequireAnyPermission([]string{PermTenantAlertingView, PermAdminAlertingView})(
			http.HandlerFunc(alertingHandler.GetAlertConfigByID),
		),
	).Methods("GET", "OPTIONS")

	alertingRouter.Handle("/history",
		middleware.RequireAnyPermission([]string{PermTenantAlertingView, PermAdminAlertingView})(
			http.HandlerFunc(alertingHandler.GetAlertHistory),
		),
	).Methods("GET", "OPTIONS")

	alertingRouter.Handle("/channels",
		middleware.RequireAnyPermission([]string{PermTenantAlertingView, PermAdminAlertingView})(
			http.HandlerFunc(alertingHandler.GetConfiguredChannels),
		),
	).Methods("GET", "OPTIONS")

	// Management routes - require manage permission (tenant OR admin)
	alertingRouter.Handle("/configs",
		middleware.RequireAnyPermission([]string{PermTenantAlertingManage, PermAdminAlertingManage})(
			http.HandlerFunc(alertingHandler.CreateAlertConfig),
		),
	).Methods("POST", "OPTIONS")

	alertingRouter.Handle("/configs/{id}",
		middleware.RequireAnyPermission([]string{PermTenantAlertingManage, PermAdminAlertingManage})(
			http.HandlerFunc(alertingHandler.UpdateAlertConfig),
		),
	).Methods("PUT", "OPTIONS")

	alertingRouter.Handle("/configs/{id}",
		middleware.RequireAnyPermission([]string{PermTenantAlertingManage, PermAdminAlertingManage})(
			http.HandlerFunc(alertingHandler.DeleteAlertConfig),
		),
	).Methods("DELETE", "OPTIONS")

	alertingRouter.Handle("/test/{channel}",
		middleware.RequireAnyPermission([]string{PermTenantAlertingManage, PermAdminAlertingManage})(
			http.HandlerFunc(alertingHandler.TestAlertChannel),
		),
	).Methods("POST", "OPTIONS")

	log.Println("  Tenant alerting routes registered at /org/alerting/*")
}
