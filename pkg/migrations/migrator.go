package migrations

import (
	"context"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Migration struct {
	Version string
	Name    string
	UpSQL   string
	DownSQL string
}

type Migrator struct {
	db           *pgxpool.Pool
	migrations   []Migration
	migrationsFS fs.FS
}

func NewMigrator(db *pgxpool.Pool, migrationsFS fs.FS) *Migrator {
	return &Migrator{
		db:           db,
		migrationsFS: migrationsFS,
	}
}

func (m *Migrator) LoadMigrations() error {
	migrations := make(map[string]*Migration)
	versionPattern := regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)

	err := fs.WalkDir(m.migrationsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		matches := versionPattern.FindStringSubmatch(d.Name())
		if len(matches) != 4 {
			return nil
		}

		version := matches[1]
		name := matches[2]
		direction := matches[3]

		content, err := fs.ReadFile(m.migrationsFS, path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		if migrations[version] == nil {
			migrations[version] = &Migration{
				Version: version,
				Name:    name,
			}
		}

		switch direction {
		case "up":
			migrations[version].UpSQL = string(content)
		case "down":
			migrations[version].DownSQL = string(content)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk migrations directory: %w", err)
	}

	var migrationsList []Migration
	for _, migration := range migrations {
		if migration.UpSQL != "" {
			migrationsList = append(migrationsList, *migration)
		}
	}

	sort.Slice(migrationsList, func(i, j int) bool {
		vi, _ := strconv.Atoi(migrationsList[i].Version)
		vj, _ := strconv.Atoi(migrationsList[j].Version)
		return vi < vj
	})

	m.migrations = migrationsList
	return nil
}

func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`

	_, err := m.db.Exec(ctx, query)
	return err
}

func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	if err := m.createMigrationsTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to create migrations table: %w", err)
	}

	rows, err := m.db.Query(ctx, "SELECT version FROM schema_migrations WHERE dirty = FALSE")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func (m *Migrator) Up(ctx context.Context) error {
	if err := m.LoadMigrations(); err != nil {
		return err
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	for _, migration := range m.migrations {
		if applied[migration.Version] {
			fmt.Printf("Migration %s already applied, skipping\n", migration.Version)
			continue
		}

		fmt.Printf("Applying migration %s: %s\n", migration.Version, migration.Name)

		tx, err := m.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Mark migration as dirty
		_, err = tx.Exec(ctx, "INSERT INTO schema_migrations (version, dirty) VALUES ($1, TRUE) ON CONFLICT (version) DO UPDATE SET dirty = TRUE", migration.Version)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to mark migration as dirty: %w", err)
		}

		// Execute migration
		_, err = tx.Exec(ctx, migration.UpSQL)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
		}

		// Mark migration as clean
		_, err = tx.Exec(ctx, "UPDATE schema_migrations SET dirty = FALSE, applied_at = $1 WHERE version = $2", time.Now(), migration.Version)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to mark migration as clean: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration transaction: %w", err)
		}

		fmt.Printf("Migration %s applied successfully\n", migration.Version)
	}

	return nil
}

func (m *Migrator) Down(ctx context.Context, steps int) error {
	if err := m.LoadMigrations(); err != nil {
		return err
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	var appliedMigrations []Migration
	for i := len(m.migrations) - 1; i >= 0; i-- {
		migration := m.migrations[i]
		if applied[migration.Version] {
			appliedMigrations = append(appliedMigrations, migration)
		}
	}

	if steps <= 0 || steps > len(appliedMigrations) {
		steps = len(appliedMigrations)
	}

	for i := 0; i < steps; i++ {
		migration := appliedMigrations[i]

		if migration.DownSQL == "" {
			return fmt.Errorf("migration %s has no down migration", migration.Version)
		}

		fmt.Printf("Rolling back migration %s: %s\n", migration.Version, migration.Name)

		tx, err := m.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Mark migration as dirty
		_, err = tx.Exec(ctx, "UPDATE schema_migrations SET dirty = TRUE WHERE version = $1", migration.Version)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to mark migration as dirty: %w", err)
		}

		// Execute rollback
		_, err = tx.Exec(ctx, migration.DownSQL)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to execute rollback for migration %s: %w", migration.Version, err)
		}

		// Remove migration record
		_, err = tx.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", migration.Version)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to remove migration record: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit rollback transaction: %w", err)
		}

		fmt.Printf("Migration %s rolled back successfully\n", migration.Version)
	}

	return nil
}

func (m *Migrator) Status(ctx context.Context) error {
	if err := m.LoadMigrations(); err != nil {
		return err
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("%-10s %-20s %s\n", "VERSION", "STATUS", "NAME")
	fmt.Println(strings.Repeat("-", 50))

	for _, migration := range m.migrations {
		status := "pending"
		if applied[migration.Version] {
			status = "applied"
		}
		fmt.Printf("%-10s %-20s %s\n", migration.Version, status, migration.Name)
	}

	return nil
}

func (m *Migrator) Create(name string) error {
	if name == "" {
		return fmt.Errorf("migration name cannot be empty")
	}

	// Load migrations to get the next version number
	if err := m.LoadMigrations(); err != nil {
		return fmt.Errorf("failed to load existing migrations: %w", err)
	}

	// Find next version number
	var maxVersion int
	for _, migration := range m.migrations {
		if version, err := strconv.Atoi(migration.Version); err == nil && version > maxVersion {
			maxVersion = version
		}
	}

	nextVersion := fmt.Sprintf("%06d", maxVersion+1)
	cleanName := strings.ReplaceAll(strings.ToLower(name), " ", "_")

	upFile := fmt.Sprintf("%s_%s.up.sql", nextVersion, cleanName)
	downFile := fmt.Sprintf("%s_%s.down.sql", nextVersion, cleanName)

	fmt.Printf("Creating migration files:\n")
	fmt.Printf("  %s\n", upFile)
	fmt.Printf("  %s\n", downFile)
	fmt.Printf("\nPlease add your SQL statements to these files.\n")

	return nil
}
