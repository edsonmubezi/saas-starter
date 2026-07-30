package database

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
)

// SetupTestDB connects to the test database and returns a pool and cleanup function.
func SetupTestDB() (*pgxpool.Pool, func()) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		log.Fatal("TEST_DATABASE_URL environment variable must be set for testing")
	}

	// Ensure SSL is enabled (unless explicitly disabled for local testing)
	if !strings.Contains(dsn, "sslmode=") {
		if strings.Contains(dsn, "?") {
			dsn += "&sslmode=require"
		} else {
			dsn += "?sslmode=require"
		}
	} else if strings.Contains(dsn, "sslmode=disable") {
		log.Println("WARNING: SSL is disabled for database connection. This should only be used for local testing.")
	}

	db, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to test DB: %v", err)
	}

	//  Cleanup function for teardown
	cleanup := func() {
		// Truncate tables you use in tests
		_, err := db.Exec(context.Background(), `
			TRUNCATE TABLE departments RESTART IDENTITY CASCADE;
		`)
		if err != nil {
			log.Printf("Warning: failed to truncate tables: %v", err)
		}

		db.Close()
	}

	return db, cleanup
}
