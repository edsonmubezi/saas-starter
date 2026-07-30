// internal/permission/repository.go
package permission

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/edsonmubezi/myapp/pkg/pagination"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// View models for grouped listing
type PermissionItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"` // Computed: "ADMIN_ONLY" or "ALL"
}

type PermissionResourceRow struct {
	ResourceID  int64            `json:"resource_id"`
	Resource    string           `json:"resource"`
	Permissions []PermissionItem `json:"permissions"`
}

// PermissionRepository defines the methods for interacting with permissions in the database.
// NOTE: Permissions are seeded elsewhere; no Create/Delete here.
type PermissionRepository interface {
	// Single item
	GetPermissionByID(ctx context.Context, id int64) (*Permission, error)

	// Update (description only)
	UpdatePermissionDescription(ctx context.Context, id int64, description string) (*Permission, error)

	// List permissions grouped by resource
	// scope: "admin" (admin only), "tenant" (non-admin), "" (all)
	ListPermissionResourcesPaginated(
		ctx context.Context,
		pag pagination.Pagination,
		scope string,
	) (pagination.Result[PermissionResourceRow], error)

	ListPermissionResources(ctx context.Context, scope string) ([]PermissionResourceRow, error)
}

type PostgresPermissionRepository struct {
	db *pgxpool.Pool
}

func NewPostgresPermissionRepository(db *pgxpool.Pool) *PostgresPermissionRepository {
	return &PostgresPermissionRepository{db: db}
}

// ----- Single item -----

func (r *PostgresPermissionRepository) GetPermissionByID(ctx context.Context, id int64) (*Permission, error) {
	const sql = `
		SELECT id, name, COALESCE(description, '')
		FROM permissions
		WHERE id = $1
	`
	row := r.db.QueryRow(ctx, sql, id)

	var p Permission
	if err := row.Scan(&p.ID, &p.Name, &p.Description); err != nil {
		return nil, fmt.Errorf("could not find permission: %v", err)
	}
	return &p, nil
}

// ----- Update (description only) -----

func (r *PostgresPermissionRepository) UpdatePermissionDescription(ctx context.Context, id int64, description string) (*Permission, error) {
	const sql = `
		UPDATE permissions
		SET description = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, name, COALESCE(description, '')
	`
	row := r.db.QueryRow(ctx, sql, description, id)

	var p Permission
	if err := row.Scan(&p.ID, &p.Name, &p.Description); err != nil {
		return nil, fmt.Errorf("could not update permission description: %v", err)
	}
	return &p, nil
}

// buildScopePredicate returns a SQL WHERE clause for filtering by scope prefix
// scope: "admin" (admin.* only), "tenant" (tenant.* only), "" (all)
func buildScopePredicate(scope string) string {
	switch scope {
	case ScopeAdmin:
		// Only admin.* permissions
		return "p.name LIKE 'admin.%'"
	case ScopeTenant:
		// Only tenant.* permissions (excludes admin.*)
		return "p.name LIKE 'tenant.%'"
	default:
		// All permissions
		return "1=1"
	}
}

// computeVisibility returns the visibility string based on permission name prefix
func computeVisibility(name string) string {
	p := &Permission{Name: name}
	return p.GetVisibility()
}

// Paginated: group by resource and nest permissions
// scope: "admin" (admin only), "tenant" (non-admin), "" (all)
func (r *PostgresPermissionRepository) ListPermissionResourcesPaginated(
	ctx context.Context,
	pag pagination.Pagination,
	scope string,
) (pagination.Result[PermissionResourceRow], error) {

	scopePredicate := buildScopePredicate(scope)

	// Note: visibility is computed from the name prefix using CASE WHEN
	table := fmt.Sprintf(`(
		WITH visible AS (
			SELECT p.id, p.name, COALESCE(p.description,'') AS description,
			       CASE WHEN p.name LIKE 'admin.%%' THEN 'ADMIN_ONLY' ELSE 'ALL' END AS visibility
			FROM permissions p
			WHERE %s
		),
		resources AS (
			SELECT DISTINCT split_part(name, '.', 2) AS resource
			FROM visible
		),
		numbered AS (
			SELECT resource,
			       DENSE_RANK() OVER (ORDER BY resource) AS resource_id
			FROM resources
		),
		aggregated AS (
			SELECT
				n.resource_id,
				n.resource,
				JSONB_AGG(
					JSONB_BUILD_OBJECT(
						'id', v.id,
						'name', v.name,
						'description', v.description,
						'visibility', v.visibility
					)
					ORDER BY v.name
				) AS permissions
			FROM numbered n
			JOIN visible v
			  ON split_part(v.name, '.', 2) = n.resource
			GROUP BY n.resource_id, n.resource
		)
		SELECT resource_id AS id, resource, permissions
		FROM aggregated
	) AS perm_resources_view`, scopePredicate)

	opts := pagination.Options[PermissionResourceRow]{
		Table:      table,
		Fields:     []string{"id", "resource", "permissions"},
		SearchCols: []string{"resource"},
		Pagination: pag,
		SortBy:     "resource",
		ScanFunc: func(row pgx.Row) (PermissionResourceRow, error) {
			var out PermissionResourceRow
			var tmpID int64
			var permsJSON []byte
			if err := row.Scan(&tmpID, &out.Resource, &permsJSON); err != nil {
				return out, err
			}
			out.ResourceID = tmpID
			if len(permsJSON) > 0 {
				if err := json.Unmarshal(permsJSON, &out.Permissions); err != nil {
					return out, fmt.Errorf("decode permissions json: %w", err)
				}
			} else {
				out.Permissions = []PermissionItem{}
			}
			return out, nil
		},
	}

	return pagination.QueryWithPagination(ctx, r.db, opts)
}

// ListPermissionResources returns all permissions grouped by resource (non-paginated)
// scope: "admin" (admin only), "tenant" (non-admin), "" (all)
func (r *PostgresPermissionRepository) ListPermissionResources(
	ctx context.Context,
	scope string,
) ([]PermissionResourceRow, error) {

	scopePredicate := buildScopePredicate(scope)

	// Note: visibility is computed from the name prefix using CASE WHEN
	sql := fmt.Sprintf(`
		WITH visible AS (
			SELECT p.id, p.name, COALESCE(p.description,'') AS description,
			       CASE WHEN p.name LIKE 'admin.%%' THEN 'ADMIN_ONLY' ELSE 'ALL' END AS visibility
			FROM permissions p
			WHERE %s
		),
		resources AS (
			SELECT DISTINCT LOWER(split_part(name, '.', 2)) AS resource
			FROM visible
		),
		numbered AS (
			SELECT resource,
			       DENSE_RANK() OVER (ORDER BY resource) AS resource_id
			FROM resources
		),
		aggregated AS (
			SELECT
				n.resource_id,
				n.resource,
				JSONB_AGG(
					JSONB_BUILD_OBJECT(
						'id', v.id,
						'name', v.name,
						'description', v.description,
						'visibility', v.visibility
					)
					ORDER BY v.name
				) AS permissions
			FROM numbered n
			JOIN visible v
			  ON LOWER(split_part(v.name, '.', 2)) = n.resource
			GROUP BY n.resource_id, n.resource
		)
		SELECT resource_id, resource, permissions
		FROM aggregated
		ORDER BY resource_id;
	`, scopePredicate)

	rows, err := r.db.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve permission resources: %w", err)
	}
	defer rows.Close()

	var out []PermissionResourceRow
	for rows.Next() {
		var row PermissionResourceRow
		var permsJSON []byte
		if err := rows.Scan(&row.ResourceID, &row.Resource, &permsJSON); err != nil {
			return nil, err
		}
		if len(permsJSON) > 0 {
			if err := json.Unmarshal(permsJSON, &row.Permissions); err != nil {
				return nil, fmt.Errorf("decode permissions json: %w", err)
			}
		} else {
			row.Permissions = []PermissionItem{}
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
