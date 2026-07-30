package orgs

import (
	"log"

	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/edsonmubezi/myapp/internal/platform/audit"
	"github.com/gorilla/mux"
)

// Permission constants for tenant audit management
const (
	PermTenantAuditView = "tenant.audit.view"
)

// RegisterTenantAuditRoutes registers audit routes for tenant (organization) users
// These routes are accessed via /api/org/audit/*
func RegisterTenantAuditRoutes(r *mux.Router, auditService *audit.PostgresService) {
	if auditService == nil {
		log.Println("  Audit service not configured, skipping tenant audit routes")
		return
	}

	auditHandler := handler.NewAuditHandler(auditService)

	// Create subrouter with tenant audit permission
	auditRouter := r.PathPrefix("/audit").Subrouter()
	auditRouter.Use(middleware.RequireAnyPermission([]string{PermTenantAuditView}))

	// Audit event routes - tenant users see only their organization's events
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

	log.Println("  Tenant audit routes registered at /org/audit/*")
}
