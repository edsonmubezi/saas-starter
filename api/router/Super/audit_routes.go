package super

import (
	"log"
	"net/http"

	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/edsonmubezi/myapp/internal/platform/audit"
	"github.com/gorilla/mux"
)

// Permission constants for admin audit management
const (
	PermAdminAuditView = "admin.audit.view"
)

// RegisterAdminAuditRoutes registers audit routes for admin (HQ) users
// These routes are accessed via /api/admin/audit/*
func RegisterAdminAuditRoutes(r *mux.Router, auditService *audit.PostgresService) {
	if auditService == nil {
		log.Println("  Audit service not configured, skipping admin audit routes")
		return
	}

	auditHandler := handler.NewAuditHandler(auditService)

	// Create subrouter with admin audit permission
	auditRouter := r.PathPrefix("/audit").Subrouter()
	auditRouter.Use(middleware.RequireAnyPermission([]string{PermAdminAuditView}))

	// Audit event routes
	auditRouter.HandleFunc("/events", auditHandler.GetAuditEvents).Methods("GET", "OPTIONS")
	auditRouter.HandleFunc("/events/{id}", auditHandler.GetAuditEventByID).Methods("GET", "OPTIONS")

	// Resource history
	auditRouter.HandleFunc("/resources/{type}/{id}/history", auditHandler.GetResourceHistory).Methods("GET", "OPTIONS")

	// User activity
	auditRouter.HandleFunc("/users/{id}/activity", auditHandler.GetUserActivity).Methods("GET", "OPTIONS")

	// Metadata endpoints
	auditRouter.HandleFunc("/metadata/actions", auditHandler.GetAuditActions).Methods("GET", "OPTIONS")
	auditRouter.HandleFunc("/metadata/severities", auditHandler.GetAuditSeverities).Methods("GET", "OPTIONS")
	auditRouter.HandleFunc("/metadata/resource-types", auditHandler.GetAuditResourceTypes).Methods("GET", "OPTIONS")

	log.Println("  Admin audit routes registered at /admin/audit/*")
}

// RegisterAdminSecurityRoutes registers security routes for admin (HQ) users
// These routes are accessed via /api/admin/security/*
func RegisterAdminSecurityRoutes(r *mux.Router, securityService interface {
	GetDashboard(ctx interface{}, tenantID int64) (interface{}, error)
}) {
	if securityService == nil {
		log.Println("  Security service not configured, skipping admin security routes")
		return
	}

	// Security routes are admin-only, no need for tenant version
	securityRouter := r.PathPrefix("/security").Subrouter()
	securityRouter.Use(middleware.RequireAnyPermission([]string{"admin.security.view"}))

	// Note: Security handler needs to be set up with the service
	// This will be wired up in router.go

	log.Println("  Admin security routes registered at /admin/security/*")
}

// RegisterAdminLogsRoutes registers application log routes for admin (HQ) users
// These routes are accessed via /api/admin/logs/*
func RegisterAdminLogsRoutes(r *mux.Router, appLogHandler http.Handler) {
	if appLogHandler == nil {
		log.Println("  App log handler not configured, skipping admin logs routes")
		return
	}

	// Logs routes are admin-only
	logsRouter := r.PathPrefix("/logs").Subrouter()
	logsRouter.Use(middleware.RequireAnyPermission([]string{"admin.logs.view"}))

	log.Println("  Admin logs routes registered at /admin/logs/*")
}
