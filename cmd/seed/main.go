// Seed command - run with: go run cmd/seed/main.go [action] [args...]
// Actions: seed (default), cleanup, cleanup-all
package main

import (
	"log"
	"os"

	"github.com/edsonmubezi/myapp/internal/seeder"
	"github.com/edsonmubezi/myapp/pkg/database"
)

func main() {
	database.InitDB()
	defer database.CloseDB()

	seedManager := seeder.NewSeedManager(database.DB)

	// Get action from command line argument
	action := "seed"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	switch action {
	case "cleanup":
		log.Println("Starting database cleanup...")
		if err := seedManager.CleanupSeededData(); err != nil {
			log.Fatalf("Cleanup failed: %v", err)
		}
		log.Println("Database cleanup completed successfully!")

	case "cleanup-all":
		log.Println("⚠️  Starting FULL database cleanup...")
		log.Println("⚠️  This will remove ALL data from seeding tables!")
		if err := seedManager.CleanupAll(); err != nil {
			log.Fatalf("Full cleanup failed: %v", err)
		}
		log.Println("⚠️  FULL database cleanup completed!")

	case "seed":
		fallthrough
	default:
		log.Println("Starting database seeding...")
		if err := seedManager.SeedAll(); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
		log.Println("Database seeding completed successfully!")
	}
}