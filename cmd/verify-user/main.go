// Verify user command - run with: go run cmd/verify-user/main.go
package main

import (
	"context"
	"log"

	"github.com/edsonmubezi/myapp/pkg/database"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	database.InitDB()
	defer database.CloseDB()

	email := "edsonmubezi@gmail.com"
	password := "edson2525"

	log.Printf("🔍 Verifying user: %s", email)

	// Check if user exists
	query := `
		SELECT
			u.id,
			u.full_name,
			u.email,
			u.password,
			u.active_status,
			u.organization_id,
			r.name as role_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1
	`

	var userID, orgID int64
	var fullName, dbEmail, hashedPassword, roleName string
	var activeStatus bool

	err := database.DB.QueryRow(context.Background(), query, email).Scan(
		&userID, &fullName, &dbEmail, &hashedPassword, &activeStatus, &orgID, &roleName,
	)

	if err != nil {
		log.Printf("❌ User not found: %v", err)
		log.Println("🔧 Try running the seeder: go run cmd/seed/main.go")
		return
	}

	log.Printf("✅ User found:")
	log.Printf("   ID: %d", userID)
	log.Printf("   Name: %s", fullName)
	log.Printf("   Email: %s", dbEmail)
	log.Printf("   Active: %v", activeStatus)
	log.Printf("   Organization ID: %d", orgID)
	log.Printf("   Role: %s", roleName)

	// Test password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Printf("❌ Password mismatch: %v", err)
		log.Printf("🔧 The stored password hash doesn't match '%s'", password)
	} else {
		log.Printf("✅ Password matches!")
	}

	if !activeStatus {
		log.Printf("❌ Account is not active")
		log.Printf("🔧 You need to activate the account")
	}

	if activeStatus && err == nil {
		log.Printf("🎉 User should be able to login successfully!")
	}
}