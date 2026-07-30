package seeder

import (
	"log"
)

// AutoSeed runs intelligent seeding on server startup
func (sm *SeedManager) AutoSeed() error {
	log.Println("Starting automatic database seeding...")

	// 1. Ensure default organization exists (only if none exist)
	orgID, err := sm.SeedDefaultOrganization()
	if err != nil {
		return err
	}

	// 2. Smart permission management (add new, update existing, remove obsolete)
	if err := sm.SeedPermissions(); err != nil {
		return err
	}

	// 3. Ensure roles and permissions are up to date for ALL organizations
	allOrgIDs, err := sm.GetAllOrganizationIDs()
	if err != nil {
		return err
	}

	for _, oid := range allOrgIDs {
		log.Printf("  Syncing roles and permissions for organization %d...", oid)
		if err := sm.SeedRoles(oid); err != nil {
			return err
		}
		if err := sm.SeedRolePermissions(oid); err != nil {
			return err
		}
	}

	// 4. Ensure default user exists (only if no users exist)
	if err := sm.SeedDefaultUser(orgID); err != nil {
		return err
	}

	log.Println("Automatic database seeding completed successfully!")
	return nil
}
