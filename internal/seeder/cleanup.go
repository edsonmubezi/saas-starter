package seeder

import (
	"context"
	"fmt"
	"log"
)

// CleanupSeededData removes all seeded data from the database
func (sm *SeedManager) CleanupSeededData() error {
	log.Println("🧹 Starting database cleanup...")

	// Define cleanup queries in dependency order (child tables first)
	cleanupQueries := []struct {
		name  string
		query string
	}{
		{
			name:  "role_permissions",
			query: `DELETE FROM role_permissions`,
		},
		{
			name:  "users",
			query: `DELETE FROM users WHERE email = 'edsonmubezi@gmail.com'`,
		},
		{
			name:  "roles",
			query: `DELETE FROM roles WHERE name IN ('SuperAdmin', 'OrgsAdmin', 'Employee')`,
		},
		{
			name:  "permissions",
			query: `DELETE FROM permissions WHERE name LIKE 'hq-only.%' OR name LIKE 'orgs.%' OR name LIKE 'emp.%'`,
		},
		{
			name:  "organizations",
			query: `DELETE FROM organizations WHERE name = 'Microfinance Demo Organization'`,
		},
	}

	// Execute cleanup queries
	for _, cleanup := range cleanupQueries {
		log.Printf("  🗑️  Cleaning up %s...", cleanup.name)

		result, err := sm.db.Exec(context.Background(), cleanup.query)
		if err != nil {
			log.Printf("  ❌ Failed to cleanup %s: %v", cleanup.name, err)
			continue
		}

		rowsAffected := result.RowsAffected()
		if rowsAffected > 0 {
			log.Printf("  ✓ Cleaned up %d rows from %s", rowsAffected, cleanup.name)
		} else {
			log.Printf("  ✓ No data to cleanup in %s", cleanup.name)
		}
	}

	log.Println("🧹 Database cleanup completed!")
	return nil
}

// CleanupAll removes ALL data from seeding tables (use with caution!)
func (sm *SeedManager) CleanupAll() error {
	log.Println("⚠️  Starting FULL database cleanup (all data will be removed)...")

	// Define cleanup queries for all data
	cleanupQueries := []struct {
		name  string
		query string
	}{
		{
			name:  "role_permissions",
			query: `TRUNCATE TABLE role_permissions RESTART IDENTITY CASCADE`,
		},
		{
			name:  "users",
			query: `TRUNCATE TABLE users RESTART IDENTITY CASCADE`,
		},
		{
			name:  "roles",
			query: `TRUNCATE TABLE roles RESTART IDENTITY CASCADE`,
		},
		{
			name:  "permissions",
			query: `TRUNCATE TABLE permissions RESTART IDENTITY CASCADE`,
		},
		{
			name:  "organizations",
			query: `TRUNCATE TABLE organizations RESTART IDENTITY CASCADE`,
		},
	}

	// Execute cleanup queries
	for _, cleanup := range cleanupQueries {
		log.Printf("  🗑️  Truncating %s...", cleanup.name)

		_, err := sm.db.Exec(context.Background(), cleanup.query)
		if err != nil {
			return fmt.Errorf("failed to truncate %s: %v", cleanup.name, err)
		}

		log.Printf("  ✓ Truncated %s", cleanup.name)
	}

	log.Println("⚠️  FULL database cleanup completed!")
	return nil
}