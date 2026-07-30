package seeder

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
)

// OrgDefaultsSeeder interface for seeding organization defaults
// This interface can be implemented by SeedManager and injected into other packages
type OrgDefaultsSeeder interface {
	SeedOrgDefaults(ctx context.Context, tx pgx.Tx, orgID int64) error
}

// OrgDefaultsData contains all default data to seed for a new organization
type OrgDefaultsData struct{}

// SeedOrgDefaults seeds all default configuration data for a new organization
// This should be called within a transaction during organization creation
func (sm *SeedManager) SeedOrgDefaults(ctx context.Context, tx pgx.Tx, orgID int64) error {
	now := time.Now()

	log.Printf("Seeding defaults for organization %d...", orgID)

	// 1. Seed Roles for this organization
	type seedRole struct {
		Name         string
		IsAssignable bool
	}
	seedRoles := []seedRole{
		{"SuperAdmin", false},
		{"OrgsAdmin", false},
	}
	roleIDs := make(map[string]int64)
	for _, sr := range seedRoles {
		var roleID int64
		err := tx.QueryRow(ctx, `
			INSERT INTO roles (name, organization_id, is_assignable)
			VALUES ($1, $2, $3)
			ON CONFLICT (name, organization_id) DO UPDATE SET is_assignable = EXCLUDED.is_assignable
			RETURNING id
		`, sr.Name, orgID, sr.IsAssignable).Scan(&roleID)
		if err != nil {
			return fmt.Errorf("failed to seed role %s: %w", sr.Name, err)
		}
		roleIDs[sr.Name] = roleID
		log.Printf("  Role: %s (ID: %d, assignable: %v)", sr.Name, roleID, sr.IsAssignable)
	}

	// 2. Seed Role Permissions
	// Get all permissions
	permRows, err := tx.Query(ctx, `SELECT id, name FROM permissions`)
	if err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}
	defer permRows.Close()

	type permInfo struct {
		ID   int64
		Name string
	}
	var allPerms []permInfo
	for permRows.Next() {
		var p permInfo
		if err := permRows.Scan(&p.ID, &p.Name); err != nil {
			continue
		}
		allPerms = append(allPerms, p)
	}

	// Assign permissions to roles based on prefix
	for _, perm := range allPerms {
		var targetRoles []string

		switch {
		case strings.HasPrefix(perm.Name, "admin."):
			targetRoles = []string{"SuperAdmin"}
		case strings.HasPrefix(perm.Name, "tenant."):
			targetRoles = []string{"OrgsAdmin"}
		}

		for _, roleName := range targetRoles {
			roleID, ok := roleIDs[roleName]
			if !ok {
				continue
			}
			_, err := tx.Exec(ctx, `
				INSERT INTO role_permissions (role_id, permission_id, created_at)
				VALUES ($1, $2, $3)
				ON CONFLICT (role_id, permission_id) DO NOTHING
			`, roleID, perm.ID, now)
			if err != nil {
				log.Printf("  Warning: failed to assign permission %s to role %s: %v", perm.Name, roleName, err)
			}
		}
	}
	log.Printf("  Role permissions assigned")

	log.Printf("Successfully seeded all defaults for organization %d", orgID)
	return nil
}

// SeedOrgDefaultsNonTx seeds all default configuration data for a new organization
// without a transaction (for standalone usage)
func (sm *SeedManager) SeedOrgDefaultsNonTx(ctx context.Context, orgID int64) error {
	tx, err := sm.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := sm.SeedOrgDefaults(ctx, tx, orgID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
