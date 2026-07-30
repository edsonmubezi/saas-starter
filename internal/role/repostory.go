package role

import (
	"context"
	"fmt"

	"github.com/edsonmubezi/myapp/pkg/pagination"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// RoleRepository defines the methods for interacting with roles in the database
type RoleRepository interface {
	CreateRole(ctx context.Context, orgID int64, name string) (*Role, error)
	GetRoleByID(ctx context.Context, id int64) (*Role, error)

	// ---- Non-paginated lists ----
	GetAllRoles(ctx context.Context) ([]Role, error)
	GetRolesByOrganizationID(ctx context.Context, orgID int64) ([]Role, error)
	GetSystemRolesByOrganizationID(ctx context.Context, orgID int64) ([]Role, error)
	GetAllOrgRoles(ctx context.Context, orgID int64) ([]Role, error) // all roles, no is_assignable filter

	// ---- Paginated lists ----
	GetAllRolesPaginated(ctx context.Context, pag pagination.Pagination) (pagination.Result[Role], error)
	GetRolesByOrganizationIDPaginated(ctx context.Context, orgID int64, pag pagination.Pagination) (pagination.Result[Role], error)
	GetAllOrgRolesPaginated(ctx context.Context, orgID int64, pag pagination.Pagination) (pagination.Result[Role], error)

	AssignPermissionsToRole(ctx context.Context, roleID int64, permissionIDs []int64) error

	DeleteRole(ctx context.Context, id int64) error
	UpdateRole(ctx context.Context, id int64, name string) (*Role, error)
	GetRoleWithPermissions(ctx context.Context, roleID int64) (*RoleWithPermissions, error)
	GetDefaultEmployeeRole(ctx context.Context, orgID int64) (*Role, error)
}

type PostgresRoleRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRoleRepository(db *pgxpool.Pool) *PostgresRoleRepository {
	return &PostgresRoleRepository{db: db}
}

func (r *PostgresRoleRepository) CreateRole(ctx context.Context, orgID int64, name string) (*Role, error) {
	sql := `
		INSERT INTO roles (name, organization_id, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, name, organization_id
	`
	row := r.db.QueryRow(ctx, sql, name, orgID)

	var role Role
	if err := row.Scan(&role.ID, &role.Name, &role.OrganizationID); err != nil {
		return nil, fmt.Errorf("could not create role: %v", err)
	}
	return &role, nil
}

func (r *PostgresRoleRepository) GetRoleByID(ctx context.Context, id int64) (*Role, error) {
	sql := `
		SELECT id, name, organization_id
		FROM roles
		WHERE id = $1
	`
	row := r.db.QueryRow(ctx, sql, id)

	var role Role
	if err := row.Scan(&role.ID, &role.Name, &role.OrganizationID); err != nil {
		return nil, fmt.Errorf("could not find role: %v", err)
	}
	return &role, nil
}

// ---------- Non-paginated ----------

func (r *PostgresRoleRepository) GetAllRoles(ctx context.Context) ([]Role, error) {
	sql := `
		SELECT id, name, organization_id
		FROM roles
		WHERE 1=1
		ORDER BY id ASC
	`
	rows, err := r.db.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve roles: %v", err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.OrganizationID); err != nil {
			return nil, fmt.Errorf("could not scan role: %v", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *PostgresRoleRepository) GetRolesByOrganizationID(ctx context.Context, orgID int64) ([]Role, error) {
	sql := `
		SELECT id, name, organization_id
		FROM roles
		WHERE organization_id = $1
		  AND COALESCE(is_assignable, true) = true
		ORDER BY id ASC
	`
	rows, err := r.db.Query(ctx, sql, orgID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve roles for org %d: %v", orgID, err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.OrganizationID); err != nil {
			return nil, fmt.Errorf("could not scan role: %v", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *PostgresRoleRepository) GetSystemRolesByOrganizationID(ctx context.Context, orgID int64) ([]Role, error) {
	sql := `
		SELECT id, name, organization_id
		FROM roles
		WHERE organization_id = $1
		  AND COALESCE(is_assignable, true) = false
		ORDER BY id ASC
	`
	rows, err := r.db.Query(ctx, sql, orgID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve system roles for org %d: %v", orgID, err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.OrganizationID); err != nil {
			return nil, fmt.Errorf("could not scan role: %v", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *PostgresRoleRepository) GetAllOrgRoles(ctx context.Context, orgID int64) ([]Role, error) {
	sql := `
		SELECT id, name, organization_id
		FROM roles
		WHERE organization_id = $1
		ORDER BY id ASC
	`
	rows, err := r.db.Query(ctx, sql, orgID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve all roles for org %d: %v", orgID, err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.OrganizationID); err != nil {
			return nil, fmt.Errorf("could not scan role: %v", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// ---------- Paginated ----------

func (r *PostgresRoleRepository) GetAllRolesPaginated(
	ctx context.Context,
	pag pagination.Pagination,
) (pagination.Result[Role], error) {

	opts := pagination.Options[Role]{
		Table: `(
			SELECT id, name, organization_id
			FROM roles
			WHERE 1=1
		) AS roles_view`,

		Fields:     []string{"id", "name", "organization_id"},
		SearchCols: []string{"name"},
		Pagination: pag,
		SortBy:     "id",

		ScanFunc: func(row pgx.Row) (Role, error) {
			var r Role
			err := row.Scan(&r.ID, &r.Name, &r.OrganizationID)
			return r, err
		},
	}

	return pagination.QueryWithPagination(ctx, r.db, opts)
}

func (r *PostgresRoleRepository) GetRolesByOrganizationIDPaginated(
	ctx context.Context,
	orgID int64,
	pag pagination.Pagination,
) (pagination.Result[Role], error) {

	opts := pagination.Options[Role]{
		Table: fmt.Sprintf(`(
			SELECT id, name, organization_id
			FROM roles
			WHERE organization_id = %d
			  AND COALESCE(is_assignable, true) = true
		) AS roles_view`, orgID),

		Fields:     []string{"id", "name", "organization_id"},
		SearchCols: []string{"name"},
		Pagination: pag,
		SortBy:     "id",

		ScanFunc: func(row pgx.Row) (Role, error) {
			var r Role
			err := row.Scan(&r.ID, &r.Name, &r.OrganizationID)
			return r, err
		},
	}

	return pagination.QueryWithPagination(ctx, r.db, opts)
}

func (r *PostgresRoleRepository) GetAllOrgRolesPaginated(
	ctx context.Context,
	orgID int64,
	pag pagination.Pagination,
) (pagination.Result[Role], error) {

	opts := pagination.Options[Role]{
		Table: fmt.Sprintf(`(
			SELECT id, name, organization_id
			FROM roles
			WHERE organization_id = %d
		) AS roles_view`, orgID),

		Fields:     []string{"id", "name", "organization_id"},
		SearchCols: []string{"name"},
		Pagination: pag,
		SortBy:     "id",

		ScanFunc: func(row pgx.Row) (Role, error) {
			var r Role
			err := row.Scan(&r.ID, &r.Name, &r.OrganizationID)
			return r, err
		},
	}

	return pagination.QueryWithPagination(ctx, r.db, opts)
}

func (r *PostgresRoleRepository) UpdateRole(ctx context.Context, id int64, name string) (*Role, error) {
	sql := `
		UPDATE roles
		SET name = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, name, organization_id
	`
	row := r.db.QueryRow(ctx, sql, name, id)

	var role Role
	if err := row.Scan(&role.ID, &role.Name, &role.OrganizationID); err != nil {
		return nil, fmt.Errorf("could not update role: %v", err)
	}
	return &role, nil
}

func (r *PostgresRoleRepository) DeleteRole(ctx context.Context, id int64) error {
	sql := `
		DELETE FROM roles
		WHERE id = $1
	`
	if _, err := r.db.Exec(ctx, sql, id); err != nil {
		return fmt.Errorf("could not delete role: %v", err)
	}
	return nil
}

func (r *PostgresRoleRepository) AssignPermissionsToRole(
	ctx context.Context,
	roleID int64,
	permissionIDs []int64,
) error {
	// Sanitize + dedupe
	seen := make(map[int64]struct{}, len(permissionIDs))
	out := make([]int64, 0, len(permissionIDs))
	for _, id := range permissionIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // safe if already committed

	// Empty set => clear all for this role
	if len(out) == 0 {
		if _, err := tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID); err != nil {
			return fmt.Errorf("clear role permissions: %w", err)
		}
		return tx.Commit(ctx)
	}

	// 1) DELETE extras (those not in the new set)
	if _, err := tx.Exec(ctx, `
		DELETE FROM role_permissions rp
		WHERE rp.role_id = $1
		  AND NOT EXISTS (
		    SELECT 1
		    FROM unnest($2::bigint[]) AS u(id)
		    WHERE u.id = rp.permission_id
		  )
	`, roleID, out); err != nil {
		return fmt.Errorf("delete extras: %w", err)
	}

	// 2) INSERT missing (ignore existing)
	if _, err := tx.Exec(ctx, `
		INSERT INTO role_permissions (role_id, permission_id, created_at)
		SELECT $1, u.id, NOW()
		FROM unnest($2::bigint[]) AS u(id)
		ON CONFLICT (role_id, permission_id)
		DO NOTHING
	`, roleID, out); err != nil {
		return fmt.Errorf("upsert current set: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PostgresRoleRepository) GetRoleWithPermissions(ctx context.Context, roleID int64) (*RoleWithPermissions, error) {
	// 1) fetch the role itself (visibility-aware via delete_status)
	const roleSQL = `
		SELECT id, name, organization_id
		FROM roles
		WHERE id = $1
	`
	var out RoleWithPermissions
	if err := r.db.QueryRow(ctx, roleSQL, roleID).Scan(&out.ID, &out.Name, &out.OrganizationID); err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}

	// 2) fetch assigned permissions with name + description
	const permsSQL = `
		SELECT p.id, p.name, COALESCE(p.description,'') AS description
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.name ASC
	`
	rows, err := r.db.Query(ctx, permsSQL, roleID)
	if err != nil {
		return nil, fmt.Errorf("get role permissions: %w", err)
	}
	defer rows.Close()

	perms := make([]PermissionBrief, 0, 16)
	for rows.Next() {
		var pb PermissionBrief
		if err := rows.Scan(&pb.ID, &pb.Name, &pb.Description); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, pb)
	}
	out.Permissions = perms
	return &out, nil
}

func (r *PostgresRoleRepository) GetDefaultEmployeeRole(ctx context.Context, orgID int64) (*Role, error) {
	// Upsert: create the Employee role if it doesn't exist for this org
	sql := `
		INSERT INTO roles (name, organization_id, is_assignable, created_at, updated_at)
		VALUES ('Employee', $1, true, NOW(), NOW())
		ON CONFLICT (name, organization_id) DO UPDATE SET is_assignable = true
		RETURNING id, name, organization_id
	`
	row := r.db.QueryRow(ctx, sql, orgID)

	var role Role
	if err := row.Scan(&role.ID, &role.Name, &role.OrganizationID); err != nil {
		return nil, fmt.Errorf("could not get or create default employee role for organization %d: %v", orgID, err)
	}
	return &role, nil
}
