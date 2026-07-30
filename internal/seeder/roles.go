package seeder

import (
	"context"
	"fmt"
	"log"
)

type Role struct {
	Name           string
	OrganizationID int64
	IsAssignable   bool
}

// GetAllRoles returns all roles to be created for an organization
func (sm *SeedManager) GetAllRoles(organizationID int64) []Role {
	return []Role{
		{Name: "SuperAdmin", OrganizationID: organizationID, IsAssignable: false},
		{Name: "OrgsAdmin", OrganizationID: organizationID, IsAssignable: false},
	}
}

// SeedRoles creates roles if they don't exist for the organization (smart - adds missing ones)
func (sm *SeedManager) SeedRoles(organizationID int64) error {
	// Get existing roles for this organization
	existingRoles := make(map[string]bool)
	query := `SELECT name FROM roles WHERE organization_id = $1`
	rows, err := sm.db.Query(context.Background(), query, organizationID)
	if err != nil {
		return fmt.Errorf("failed to get existing roles: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		existingRoles[name] = true
	}

	roles := sm.GetAllRoles(organizationID)
	addedCount := 0

	for _, role := range roles {
		if existingRoles[role.Name] {
			continue // Role already exists, skip
		}

		insertQuery := `
			INSERT INTO roles (name, organization_id, is_assignable, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())
		`

		_, err := sm.db.Exec(context.Background(), insertQuery, role.Name, role.OrganizationID, role.IsAssignable)
		if err != nil {
			return fmt.Errorf("failed to seed role %s: %v", role.Name, err)
		}

		log.Printf("  ✓ Seeded role: %s", role.Name)
		addedCount++
	}

	if addedCount == 0 {
		log.Printf("  ✓ All roles already exist for organization %d (%d roles)", organizationID, len(existingRoles))
	} else {
		log.Printf("👥 Successfully added %d new roles (total: %d)", addedCount, len(existingRoles)+addedCount)
	}

	return nil
}