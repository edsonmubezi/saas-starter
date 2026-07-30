package seeder

import (
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

type SeedManager struct {
	db *pgxpool.Pool
}

func NewSeedManager(db *pgxpool.Pool) *SeedManager {
	return &SeedManager{db: db}
}

func (sm *SeedManager) SeedAll() error {
	log.Println("🏢 Creating default organization...")
	organizationID, err := sm.SeedDefaultOrganization()
	if err != nil {
		return err
	}

	log.Println("🌱 Seeding permissions...")
	if err := sm.SeedPermissions(); err != nil {
		return err
	}

	log.Println("👥 Seeding roles...")
	if err := sm.SeedRoles(organizationID); err != nil {
		return err
	}

	log.Println("🔗 Assigning permissions to roles...")
	if err := sm.SeedRolePermissions(organizationID); err != nil {
		return err
	}

	log.Println("👤 Seeding default user...")
	if err := sm.SeedDefaultUser(organizationID); err != nil {
		return err
	}

	log.Println("📊 Generating summary...")
	sm.GetRolePermissionSummary(organizationID)

	return nil
}