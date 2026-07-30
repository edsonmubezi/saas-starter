package shared

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
)

// DB is an interface matching pgx.Conn or pgxpool.Pool
type DB interface {
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
}

// NameExistenceChecker defines a generic interface for checking
// if a record with a given name exists in an organization (soft-deleted aware).
type NameExistenceChecker interface {
	ExistsByNameInOrg(ctx context.Context, table, nameCol, name string, orgID int64, excludeID ...int64) (bool, error)
}

// PostgresExistenceChecker implements NameExistenceChecker using a pgx DB connection.
type PostgresExistenceChecker struct {
	DB DB
}

// ExistsByNameInOrg checks if a record with the same name (case-insensitive)
// exists in the specified table for the given organization.
// Optionally excludes a record by ID (used during update).
func (c *PostgresExistenceChecker) ExistsByNameInOrg(
	ctx context.Context,
	table string,
	nameCol string,
	name string,
	orgID int64,
	excludeID ...int64,
) (bool, error) {

	// Basic protection against SQL injection via table or column name
	if !isSafeIdentifier(table) || !isSafeIdentifier(nameCol) {
		return false, fmt.Errorf("unsafe table or column name: %s.%s", table, nameCol)
	}

	// Build the query conditionally
	query := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1 FROM %s
			WHERE LOWER(%s) = LOWER($1)
			AND organization_id = $2
			AND delete_status = 0
			%s
		)
	`, table, nameCol, func() string {
		if len(excludeID) > 0 {
			return "AND id != $3"
		}
		return ""
	}())

	var exists bool
	var err error

	// Execute with or without excludeID
	if len(excludeID) > 0 {
		err = c.DB.QueryRow(ctx, query, name, orgID, excludeID[0]).Scan(&exists)
	} else {
		err = c.DB.QueryRow(ctx, query, name, orgID).Scan(&exists)
	}

	return exists, err
}

// isSafeIdentifier checks if a string is safe to use as a SQL identifier (no injection)
func isSafeIdentifier(s string) bool {
	return s != "" && !strings.ContainsAny(s, " ;'\"-")
}
