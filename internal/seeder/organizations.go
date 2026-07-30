package seeder

import (
	"context"
	"fmt"
	"log"
)

// SeedDefaultOrganization creates a default organization and returns its ID (only if none exists)
func (sm *SeedManager) SeedDefaultOrganization() (int64, error) {
	// Check if any organization exists
	var count int64
	countQuery := `SELECT COUNT(*) FROM organizations`
	err := sm.db.QueryRow(context.Background(), countQuery).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to check existing organizations: %v", err)
	}

	if count > 0 {
		// Return the first organization found
		var orgID int64
		getQuery := `SELECT id FROM organizations ORDER BY id LIMIT 1`
		err := sm.db.QueryRow(context.Background(), getQuery).Scan(&orgID)
		if err != nil {
			return 0, fmt.Errorf("failed to get existing organization: %v", err)
		}
		log.Printf("  ✓ Organizations already exist, using first one (ID: %d)", orgID)
		return orgID, nil
	}

	orgName := "Demo Organization"

	// Create new organization
	insertQuery := `
		INSERT INTO organizations (
			name,
			address,
			phone_number,
			contact_person,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`

	var orgID int64
	err = sm.db.QueryRow(context.Background(), insertQuery,
		orgName,
		"123 Main Street, Demo City",
		"+1234567890",
		"Demo Admin",
	).Scan(&orgID)

	if err != nil {
		return 0, fmt.Errorf("failed to create default organization: %v", err)
	}

	log.Printf("  ✓ Created default organization: %s (ID: %d)", orgName, orgID)

	// Create default organization settings
	if err := sm.createDefaultOrgSettings(orgID); err != nil {
		log.Printf("  ⚠️  Warning: Failed to create default organization settings: %v", err)
		// Don't fail the entire seeding process if just settings fail
	}

	return orgID, nil
}

// createDefaultOrgSettings creates default organization settings for a new organization
func (sm *SeedManager) createDefaultOrgSettings(orgID int64) error {
	// Check if settings already exist
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM organization_settings WHERE organization_id = $1)`
	err := sm.db.QueryRow(context.Background(), checkQuery, orgID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check existing organization settings: %v", err)
	}

	if exists {
		log.Printf("  ✓ Organization settings already exist for organization %d", orgID)
		return nil
	}

	// Insert default organization settings
	insertQuery := `
		INSERT INTO organization_settings (
			organization_id,
			organization_type,
			created_at,
			updated_at
		)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`

	var settingsID int64
	err = sm.db.QueryRow(context.Background(), insertQuery,
		orgID,
		"single_company", // Default to single company
	).Scan(&settingsID)

	if err != nil {
		return fmt.Errorf("failed to create default organization settings: %v", err)
	}

	log.Printf("  ✓ Created default organization settings (ID: %d) for organization %d", settingsID, orgID)
	return nil
}

// GetAllOrganizationIDs returns all organization IDs from the database
func (sm *SeedManager) GetAllOrganizationIDs() ([]int64, error) {
	query := `SELECT id FROM organizations ORDER BY id`
	rows, err := sm.db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization IDs: %v", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}
