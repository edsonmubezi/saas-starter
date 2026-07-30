package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/edsonmubezi/myapp/pkg/database"
	"github.com/edsonmubezi/myapp/pkg/migrations"
)

func main() {
	var (
		command = flag.String("command", "", "Migration command: up, down, status, create")
		steps   = flag.Int("steps", 0, "Number of migrations to rollback (for down command)")
		name    = flag.String("name", "", "Migration name (for create command)")
	)
	flag.Parse()

	if *command == "" {
		printUsage()
		os.Exit(1)
	}

	// Initialize database for all commands except create
	var migrator *migrations.Migrator
	if *command != "create" {
		initDatabase()
		defer database.CloseDB()
		migrator = migrations.NewMigrator(database.DB, os.DirFS("db/migrations"))
	} else {
		migrator = migrations.NewMigrator(nil, os.DirFS("db/migrations"))
	}
	ctx := context.Background()

	switch *command {
	case "up":
		if err := migrator.Up(ctx); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("All migrations applied successfully")

	case "down":
		if err := migrator.Down(ctx, *steps); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Printf("Rolled back %d migration(s) successfully\n", *steps)

	case "status":
		if err := migrator.Status(ctx); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}

	case "create":
		if *name == "" {
			fmt.Println("Error: Migration name is required for create command")
			fmt.Println("Usage: go run cmd/migrate/main.go -command=create -name=\"your_migration_name\"")
			os.Exit(1)
		}
		if err := migrator.Create(*name); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", *command)
		printUsage()
		os.Exit(1)
	}
}

// initDatabase connects using DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME
// or falls back to DATABASE_URL for backward compatibility
func initDatabase() {
	host := os.Getenv("DB_HOST")
	password := os.Getenv("DB_PASSWORD")
	if host != "" && password != "" {
		port, _ := strconv.Atoi(os.Getenv("DB_PORT"))
		if port == 0 {
			port = 5432
		}
		user := os.Getenv("DB_USER")
		if user == "" {
			user = "microfinance"
		}
		dbName := os.Getenv("DB_NAME")
		if dbName == "" {
			dbName = "microfinance"
		}
		sslMode := os.Getenv("DB_SSLMODE")
		if sslMode == "" {
			sslMode = "disable"
		}
		database.InitDBWithConfig(host, port, user, password, dbName, sslMode)
		return
	}
	// Fallback to DATABASE_URL
	database.InitDB()
}

func printUsage() {
	fmt.Println("Database Migration Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/migrate/main.go -command=<command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up              Apply all pending migrations")
	fmt.Println("  down            Rollback migrations")
	fmt.Println("  status          Show migration status")
	fmt.Println("  create          Create new migration files")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -steps=N        Number of migrations to rollback (down command)")
	fmt.Println("  -name=NAME      Migration name (create command)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/migrate/main.go -command=up")
	fmt.Println("  go run cmd/migrate/main.go -command=down -steps=1")
	fmt.Println("  go run cmd/migrate/main.go -command=status")
	fmt.Println("  go run cmd/migrate/main.go -command=create -name=\"add_users_table\"")
}
