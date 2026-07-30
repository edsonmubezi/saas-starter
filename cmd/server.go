// @title SaaS Starter API
// @version 1.0.0
// @description Multi-tenant SaaS Starter API with organization administration, role-based access control, and platform services.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/edsonmubezi/myapp
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes https http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Provide your JWT token as: Bearer <token>
package main

import (
	"context"
	"log"
	"os"

	"github.com/edsonmubezi/myapp/internal/config"
	"github.com/edsonmubezi/myapp/internal/container"
	"github.com/edsonmubezi/myapp/internal/seeder"
	"github.com/edsonmubezi/myapp/internal/server"
	"github.com/edsonmubezi/myapp/pkg/auth"
	"github.com/edsonmubezi/myapp/pkg/database"
	"github.com/edsonmubezi/myapp/pkg/migrations"
	redisclient "github.com/edsonmubezi/myapp/pkg/redis"
	"github.com/edsonmubezi/myapp/pkg/resilience"
	"github.com/joho/godotenv"
)

// validateEnvironment checks that all required environment variables are set
func validateEnvironment() {
	requiredVars := map[string]func(string) error{
		"DATABASE_URL": func(val string) error {
			if val == "" {
				return nil // Will be caught by database.InitDB()
			}
			return nil
		},
		"SECRET_KEY": func(val string) error {
			if val == "" {
				log.Fatal("SECRET_KEY environment variable is not set")
			}
			if len(val) != 32 {
				log.Fatalf("SECRET_KEY must be exactly 32 bytes long, got %d bytes", len(val))
			}
			return nil
		},
		"JWT_SECRET": func(val string) error {
			if val == "" {
				log.Fatal("JWT_SECRET environment variable is not set")
			}
			return nil
		},
	}

	for varName, validator := range requiredVars {
		val := os.Getenv(varName)
		if err := validator(val); err != nil {
			log.Fatalf("Environment validation failed for %s: %v", varName, err)
		}
	}

	log.Println("✓ Environment validation passed")
}

func main() {
	// Load .env file if it exists (for local development)
	// In production, environment variables are injected directly
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found (this is OK in production environments)")
	}

	// Validate critical environment variables before starting
	validateEnvironment()

	// Initialize circuit breakers for resilience
	log.Println("Initializing circuit breakers...")
	resilience.InitializeCircuitBreakers()
	log.Println("✓ Circuit breakers initialized")

	// Validate JWT secret strength on startup
	log.Println("Validating JWT secret...")
	if err := auth.ValidateJWTSecretOnStartup(); err != nil {
		log.Fatalf("❌ JWT secret validation failed: %v", err)
	}
	log.Println("✓ JWT secret validated")

	// Load configuration
	cfg := config.Load()

	// Initialize Redis for rate limiting and token blacklist
	log.Println("Initializing Redis...")
	if err := redisclient.Initialize(cfg.Redis); err != nil {
		if cfg.Redis.Enabled {
			log.Fatalf("❌ Failed to initialize Redis: %v", err)
		} else {
			log.Println("⚠️  Redis disabled, rate limiting and token blacklist unavailable")
		}
	} else {
		log.Println("✓ Redis initialized successfully")
	}
	defer func() {
		if err := redisclient.Close(); err != nil {
			log.Printf("Warning: Failed to close Redis connection: %v", err)
		}
	}()

	// Initialize main database
	if cfg.Database.Password != "" {
		database.InitDBWithConfig(
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Name,
			cfg.Database.SSLMode,
		)
	} else {
		// Fallback to DATABASE_URL for backward compatibility
		database.InitDB()
	}
	defer database.CloseDB()

	// Run database migrations on server startup
	log.Println("Running database migrations...")
	migrator := migrations.NewMigrator(database.DB, os.DirFS("db/migrations"))
	if err := migrator.Up(context.Background()); err != nil {
		// In production, fail fast on migration errors to prevent running with outdated schema
		runtimeEnv := os.Getenv("RUNTIME_ENV")
		if runtimeEnv == "production" {
			log.Fatalf("❌ Migration failed in production: %v", err)
		}
		log.Printf("⚠️  Migration failed: %v", err)
		log.Println("⚠️  Server will continue (dev mode), but database schema may be outdated...")
	} else {
		log.Println("✓ Database migrations completed successfully")
	}

	// Run automatic seeding on server startup
	seedManager := seeder.NewSeedManager(database.DB)
	if err := seedManager.AutoSeed(); err != nil {
		log.Printf("  Auto-seeding failed: %v", err)
		log.Println(" Server will continue without seeding...")
	}

	// Initialize application container and server
	containerInstance := container.New(database.DB, cfg)
	srv := server.New(cfg, containerInstance)

	// Use graceful shutdown to prevent data loss during restarts/deployments
	if err := srv.StartWithGracefulShutdown(); err != nil {
		log.Fatal(err)
	}
}
