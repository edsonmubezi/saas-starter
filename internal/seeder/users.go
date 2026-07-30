package seeder

import (
	"context"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// SeedDefaultUser creates the default SuperAdmin user (only if no users exist)
func (sm *SeedManager) SeedDefaultUser(organizationID int64) error {
	// Check if any users exist
	var count int64
	countQuery := `SELECT COUNT(*) FROM users`
	err := sm.db.QueryRow(context.Background(), countQuery).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing users: %v", err)
	}

	if count > 0 {
		log.Printf("  ✓ Users already exist, skipping user seeding")
		return nil
	}

	email := "admin@example.com"
	password := "Admin@1234"
	fullName := "System Administrator"

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Get SuperAdmin role ID for this organization
	var roleID int64
	roleQuery := `SELECT id FROM roles WHERE name = 'SuperAdmin' AND organization_id = $1 LIMIT 1`
	err = sm.db.QueryRow(context.Background(), roleQuery, organizationID).Scan(&roleID)
	if err != nil {
		return fmt.Errorf("failed to get SuperAdmin role ID: %v", err)
	}

	// Create the user
	userQuery := `
		INSERT INTO users (
			full_name,
			email,
			password,
			role_id,
			active_status,
			must_change_password,
			organization_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id
	`

	var userID int64
	err = sm.db.QueryRow(context.Background(), userQuery,
		fullName,
		email,
		string(hashedPassword),
		roleID,
		true,
		false, // must_change_password = false for default admin
		organizationID,
	).Scan(&userID)

	if err != nil {
		// Handle duplicate key error gracefully
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			log.Printf("  ✓ User already exists: %s", email)
			log.Printf("  🔑 Login credentials: %s / %s", email, password)
			return nil
		}
		return fmt.Errorf("failed to create default user: %v", err)
	}

	log.Printf("  ✓ Created default user: %s (ID: %d) with SuperAdmin role", email, userID)
	log.Printf("  🔑 Login credentials: %s / %s", email, password)

	return nil
}