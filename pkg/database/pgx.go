package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

var DB *pgxpool.Pool
var RecruitmentDB *pgxpool.Pool

// PoolConfig holds database connection pool configuration
type PoolConfig struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}

// DefaultPoolConfig returns production-ready defaults for connection pooling
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxConns:          25,               // Maximum connections in the pool
		MinConns:          5,                // Minimum connections to keep open
		MaxConnLifetime:   time.Hour,        // Recycle connections after 1 hour
		MaxConnIdleTime:   30 * time.Minute, // Close idle connections after 30 minutes
		HealthCheckPeriod: time.Minute,      // Check connection health every minute
		ConnectTimeout:    10 * time.Second, // Timeout for new connections
	}
}

// LoadPoolConfigFromEnv loads pool configuration from environment variables
// Falls back to defaults if not set
func LoadPoolConfigFromEnv() *PoolConfig {
	cfg := DefaultPoolConfig()

	if v := os.Getenv("DB_MAX_CONNS"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			cfg.MaxConns = int32(n)
		}
	}
	if v := os.Getenv("DB_MIN_CONNS"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			cfg.MinConns = int32(n)
		}
	}
	if v := os.Getenv("DB_MAX_CONN_LIFETIME_MINUTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.MaxConnLifetime = time.Duration(n) * time.Minute
		}
	}
	if v := os.Getenv("DB_HEALTH_CHECK_PERIOD_SECONDS"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.HealthCheckPeriod = time.Duration(n) * time.Second
		}
	}

	return cfg
}

func loadEnv() {
	// Try to load .env file, but don't fail if it doesn't exist
	// In production (Coolify), environment variables are injected directly
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found (this is OK in production environments)")
	}
}

// InitDBWithConfig initializes the main database using config
func InitDBWithConfig(host string, port int, user, password, dbName, sslMode string) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user, password, host, port, dbName, sslMode)

	// Load pool configuration
	poolCfg := LoadPoolConfigFromEnv()

	// Parse the connection string to get a config object
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Failed to parse database URL: %v", err)
	}

	initDBWithConfig(config, poolCfg, "Main")
}

// InitDB initializes the main database using DATABASE_URL env var (legacy support)
func InitDB() {
	loadEnv()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set in .env")
	}

	// Load pool configuration
	poolCfg := LoadPoolConfigFromEnv()

	// Parse the connection string to get a config object
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Failed to parse database URL: %v", err)
	}

	initDBWithConfig(config, poolCfg, "Main")
}

func initDBWithConfig(config *pgxpool.Config, poolCfg *PoolConfig, name string) {
	// Apply connection pool settings
	config.MaxConns = poolCfg.MaxConns
	config.MinConns = poolCfg.MinConns
	config.MaxConnLifetime = poolCfg.MaxConnLifetime
	config.MaxConnIdleTime = poolCfg.MaxConnIdleTime
	config.HealthCheckPeriod = poolCfg.HealthCheckPeriod
	config.ConnConfig.ConnectTimeout = poolCfg.ConnectTimeout

	// Add connection lifecycle hooks for monitoring
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		log.Printf("%s database: new connection established", name)
		return nil
	}

	config.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		return conn.Ping(ctx) == nil
	}

	config.AfterRelease = func(conn *pgx.Conn) bool {
		return true
	}

	// Create the connection pool with timeout
	// Use a longer timeout for pool creation since it must establish MinConns connections
	poolCreateTimeout := poolCfg.ConnectTimeout * time.Duration(poolCfg.MinConns+1)
	ctx, cancel := context.WithTimeout(context.Background(), poolCreateTimeout)
	defer cancel()

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create %s connection pool: %v", name, err)
	}

	// Ping to verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		log.Fatalf("%s database ping failed: %v", name, err)
	}

	DB = pool

	log.Printf("✓ %s database pool initialized: MaxConns=%d, MinConns=%d",
		name, poolCfg.MaxConns, poolCfg.MinConns)
}

// GetPoolStats returns current connection pool statistics
func GetPoolStats() *pgxpool.Stat {
	if DB == nil {
		return nil
	}
	return DB.Stat()
}

// LogPoolStats logs the current pool statistics (useful for monitoring)
func LogPoolStats() {
	if DB == nil {
		return
	}
	stats := DB.Stat()
	log.Printf("DB Pool Stats: TotalConns=%d, AcquiredConns=%d, IdleConns=%d, MaxConns=%d",
		stats.TotalConns(), stats.AcquiredConns(), stats.IdleConns(), stats.MaxConns())
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}

// ConnectWithURL creates a new connection pool with the given database URL
// Returns nil if the URL is empty
func ConnectWithURL(dbURL string, name string) *pgxpool.Pool {
	if dbURL == "" {
		return nil
	}

	poolCfg := LoadPoolConfigFromEnv()

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Printf("Failed to parse %s database URL: %v", name, err)
		return nil
	}

	// Apply connection pool settings (smaller pool for secondary DBs)
	config.MaxConns = poolCfg.MaxConns / 2
	config.MinConns = poolCfg.MinConns / 2
	if config.MinConns < 1 {
		config.MinConns = 1
	}
	config.MaxConnLifetime = poolCfg.MaxConnLifetime
	config.MaxConnIdleTime = poolCfg.MaxConnIdleTime
	config.HealthCheckPeriod = poolCfg.HealthCheckPeriod
	config.ConnConfig.ConnectTimeout = poolCfg.ConnectTimeout

	poolCreateTimeout := poolCfg.ConnectTimeout * time.Duration(config.MinConns+1)
	ctx, cancel := context.WithTimeout(context.Background(), poolCreateTimeout)
	defer cancel()

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		log.Printf("Failed to connect to %s database: %v", name, err)
		return nil
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		log.Printf("%s database ping failed: %v", name, err)
		return nil
	}

	log.Printf("✓ %s database pool initialized: MaxConns=%d, MinConns=%d",
		name, config.MaxConns, config.MinConns)
	return pool
}

// InitRecruitmentDBWithConfig initializes the recruitment database using config
func InitRecruitmentDBWithConfig(host string, port int, user, password, dbName, sslMode string) {
	if password == "" {
		log.Fatal("RECRUITMENT_DB_PASSWORD is required")
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user, password, host, port, dbName, sslMode)

	RecruitmentDB = ConnectWithURL(dbURL, "Recruitment")
	if RecruitmentDB == nil {
		log.Fatal("Failed to initialize recruitment database")
	}
}

// InitRecruitmentDB initializes the recruitment database using env vars (legacy support)
func InitRecruitmentDB() {
	dbURL := os.Getenv("RECRUITMENT_DATABASE_URL")
	if dbURL == "" {
		log.Println("RECRUITMENT_DATABASE_URL not set, recruitment will use main database")
		return
	}

	RecruitmentDB = ConnectWithURL(dbURL, "Recruitment")
}

// CloseRecruitmentDB closes the recruitment database connection
func CloseRecruitmentDB() {
	if RecruitmentDB != nil {
		RecruitmentDB.Close()
		log.Println("Recruitment database connection closed")
	}
}
