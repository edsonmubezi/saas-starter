package seeder

import (
	"context"
	"fmt"
	"log"
)

// Permission struct for seeding - visibility is now derived from name prefix
type Permission struct {
	Name        string
	Description string
}

// GetAllPermissions returns all permissions organized by scope prefix
// Naming convention: [scope].[resource].[action]
// Scopes: admin (SuperAdmin/HQ), tenant (OrgAdmin), self (User)
func (sm *SeedManager) GetAllPermissions() []Permission {
	return []Permission{
		// ===== ADMIN PERMISSIONS (SuperAdmin/HQ only) =====

		// User Management (HQ)
		{Name: "admin.user.create", Description: "Create new users"},
		{Name: "admin.user.view", Description: "View all users"},
		{Name: "admin.user.edit", Description: "Edit user details"},
		{Name: "admin.user.delete", Description: "Delete users"},
		{Name: "admin.user.edit_status", Description: "Edit user status"},

		// Role Management (HQ)
		{Name: "admin.role.create", Description: "Create new roles"},
		{Name: "admin.role.view", Description: "View all roles"},
		{Name: "admin.role.edit", Description: "Edit role details"},
		{Name: "admin.role.delete", Description: "Delete roles"},
		{Name: "admin.role.assign", Description: "Assign roles to users"},

		// Permission Management (HQ)
		{Name: "admin.permission.view", Description: "View all permissions"},
		{Name: "admin.permission.edit", Description: "Edit permissions"},

		// Organization Management (HQ)
		{Name: "admin.organization.create", Description: "Create new organizations"},
		{Name: "admin.organization.view", Description: "View all organizations"},
		{Name: "admin.organization.edit", Description: "Edit organization details"},
		{Name: "admin.organization.delete", Description: "Delete organizations"},

		// Organization Settings (HQ Core Settings)
		{Name: "admin.org_settings.create", Description: "Create organization core settings"},
		{Name: "admin.org_settings.view", Description: "View organization core settings"},
		{Name: "admin.org_settings.edit", Description: "Edit organization core settings"},

		// Dashboard (HQ)
		{Name: "admin.dashboard.view", Description: "View admin dashboard"},

		// ===== TENANT PERMISSIONS (OrgAdmin level) =====

		// Permission Management (Tenant)
		{Name: "tenant.permission.view", Description: "View limited permissions"},

		// User Management (Tenant)
		{Name: "tenant.user.create", Description: "Create users within organization"},
		{Name: "tenant.user.view", Description: "View users within organization"},
		{Name: "tenant.user.edit", Description: "Edit users within organization"},
		{Name: "tenant.user.delete", Description: "Delete users within organization"},
		{Name: "tenant.user.edit_status", Description: "Edit user status within organization"},

		// Role Management (Tenant)
		{Name: "tenant.role.create", Description: "Create roles within organization"},
		{Name: "tenant.role.view", Description: "View roles within organization"},
		{Name: "tenant.role.assign", Description: "Assign roles to users within organization"},

		// Dashboard (Tenant)
		{Name: "tenant.dashboard.view", Description: "View organization dashboard"},

		// ===== SECURITY PERMISSIONS =====

		// User Security Management (Admin)
		{Name: "admin.user.unlock", Description: "Unlock locked user accounts"},
		{Name: "admin.user.reset_password", Description: "Reset user passwords"},
		{Name: "admin.user.view_security", Description: "View user security settings and login history"},
		{Name: "admin.user.view_reset_history", Description: "View user password reset history"},
		{Name: "admin.user.reset_2fa", Description: "Reset user's two-factor authentication"},

		// User Security Management (Tenant)
		{Name: "tenant.user.unlock", Description: "Unlock locked user accounts within organization"},
		{Name: "tenant.user.reset_password", Description: "Reset user passwords within organization"},
		{Name: "tenant.user.view_security", Description: "View user security settings within organization"},
		{Name: "tenant.user.view_reset_history", Description: "View user password reset history within organization"},

		// Role 2FA Management
		{Name: "admin.role.set_2fa", Description: "Configure role 2FA requirements"},

		// ===== LOGGING & AUDIT PERMISSIONS =====

		// Alerting (Admin - HQ level)
		{Name: "admin.alerting.view", Description: "View alert configurations and history"},
		{Name: "admin.alerting.manage", Description: "Create and manage alert configurations"},

		// Alerting (Tenant - Organization level)
		{Name: "tenant.alerting.view", Description: "View organization alert configurations and history"},
		{Name: "tenant.alerting.manage", Description: "Manage organization alert configurations"},

		// Audit Logs (Admin - HQ level)
		{Name: "admin.audit.view", Description: "View audit logs across all organizations"},

		// Audit Logs (Tenant - Organization level)
		{Name: "tenant.audit.view", Description: "View audit logs within organization"},

		// Security Events (Admin - HQ level)
		{Name: "admin.security.view", Description: "View security dashboard and events across all organizations"},

		// Security Events (Tenant - Organization level)
		{Name: "tenant.security.view", Description: "View security dashboard and events within organization"},

		// Application Logs (Admin - HQ level)
		{Name: "admin.logs.view", Description: "View application and access logs across all organizations"},

		// Application Logs (Tenant - Organization level)
		{Name: "tenant.logs.view", Description: "View application and access logs within organization"},

		// ===== BRANDING PERMISSIONS =====

		// Document Branding (Tenant)
		{Name: "tenant.document_branding.manage", Description: "Manage document branding settings (PDF/Excel headers, watermarks, colors)"},

		// Email Branding (Tenant)
		{Name: "tenant.email_branding.manage", Description: "Manage email branding settings (colors, fonts, footer text)"},

		// ===== COMMUNICATION PERMISSIONS =====

		// Email Broadcast (Tenant)
		{Name: "tenant.broadcast.view", Description: "View email broadcasts"},
		{Name: "tenant.broadcast.create", Description: "Create and edit email broadcasts"},
		{Name: "tenant.broadcast.approve", Description: "Approve or reject email broadcasts"},
		{Name: "tenant.broadcast.send", Description: "Send approved email broadcasts"},
		{Name: "tenant.broadcast.delete", Description: "Delete email broadcast drafts"},

		// ===== DATA IMPORT PERMISSIONS =====

		// Data Import (Tenant)
		{Name: "tenant.dataimport.view", Description: "View import mappings and drafts"},
		{Name: "tenant.dataimport.manage", Description: "Create/edit mappings, upload and edit drafts"},
		{Name: "tenant.dataimport.confirm", Description: "Confirm and commit import drafts"},

		// ===== AI ASSISTANT PERMISSIONS =====

		// AI Chat Assistant (Tenant)
		{Name: "tenant.chat.view", Description: "Use the AI help assistant chatbot"},
		{Name: "tenant.chat.manage", Description: "Manage AI assistant settings (API key configuration)"},

		// Knowledge Base
		{Name: "tenant.knowledge_base.view", Description: "View knowledge base articles"},
		{Name: "tenant.knowledge_base.manage", Description: "Create, edit, and delete knowledge base articles"},
	}
}

// SeedPermissions smartly manages permissions (add new, update existing, remove obsolete)
// Note: visibility is now derived from the name prefix, not stored in DB
func (sm *SeedManager) SeedPermissions() error {
	permissions := sm.GetAllPermissions()

	// Create a map of current permissions for quick lookup
	currentPerms := make(map[string]Permission)
	for _, perm := range permissions {
		currentPerms[perm.Name] = perm
	}

	// Get existing permissions from database
	existingPerms := make(map[string]bool)
	rows, err := sm.db.Query(context.Background(), `SELECT name FROM permissions`)
	if err != nil {
		return fmt.Errorf("failed to get existing permissions: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		existingPerms[name] = true
	}

	// Add new permissions or update existing ones
	for _, perm := range permissions {
		query := `
			INSERT INTO permissions (name, description, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			ON CONFLICT (name) DO UPDATE SET
				description = EXCLUDED.description,
				updated_at = NOW()
		`

		_, err := sm.db.Exec(context.Background(), query, perm.Name, perm.Description)
		if err != nil {
			return fmt.Errorf("failed to seed permission %s: %v", perm.Name, err)
		}

		if existingPerms[perm.Name] {
			log.Printf("   Updated permission: %s", perm.Name)
		} else {
			log.Printf("   Added new permission: %s", perm.Name)
		}
	}

	// Remove obsolete permissions (ones in DB but not in current list)
	for existingPerm := range existingPerms {
		if _, exists := currentPerms[existingPerm]; !exists {
			// Delete from role_permissions first (foreign key constraint)
			_, err := sm.db.Exec(context.Background(), `DELETE FROM role_permissions WHERE permission_id = (SELECT id FROM permissions WHERE name = $1)`, existingPerm)
			if err != nil {
				log.Printf("    Failed to remove role_permissions for obsolete permission %s: %v", existingPerm, err)
				continue
			}

			// Delete the permission
			_, err = sm.db.Exec(context.Background(), `DELETE FROM permissions WHERE name = $1`, existingPerm)
			if err != nil {
				log.Printf("    Failed to remove obsolete permission %s: %v", existingPerm, err)
				continue
			}
			log.Printf("   Removed obsolete permission: %s", existingPerm)
		}
	}

	log.Printf(" Successfully managed permissions: %d current, %d existing", len(permissions), len(existingPerms))
	return nil
}
