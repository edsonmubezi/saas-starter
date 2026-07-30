package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/edsonmubezi/myapp/pkg/pagination"

	"github.com/edsonmubezi/myapp/internal/organization"
	"github.com/edsonmubezi/myapp/internal/permission"
	"github.com/edsonmubezi/myapp/internal/role"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// UserRepository defines the interface for user persistence logic
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	CreateUserTx(ctx context.Context, tx pgx.Tx, u *User) (*User, error)
	GetUserByID(ctx context.Context, id int64, organizationID int64) (*User, error)
	GetAllUsers(ctx context.Context, pg pagination.Pagination, organizationID int64, userType string) (pagination.Result[UserListItem], error)
	GetUserCounts(ctx context.Context) (map[string]int, error)
	GetUsersByOrganization(ctx context.Context, pg pagination.Pagination, organizationID int64, userType string) (pagination.Result[UserListItem], error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByEmailWithSecurity(ctx context.Context, email string) (*User, error)
	GetUserByIDNoOrg(ctx context.Context, id int64) (*User, error)
	SoftDeleteUser(ctx context.Context, id int64, organizationID int64) error
	HardDeleteUser(ctx context.Context, id int64, organizationID int64) error
	UpdateUserStatus(ctx context.Context, id int64, organizationID int64) error
	ChangePassword(ctx context.Context, userID int64, organizationID int64, newPassword string) error
	UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error
	GetPermissionsForRole(ctx context.Context, roleID int64) ([]permission.Permission, error)
	GetRoleForUser(ctx context.Context, roleID int64) (*role.Role, error)
	UpdateUserPartial(ctx context.Context, input *UpdateUserInput) (*User, error)

	// Security methods for account lockout
	IncrementFailedAttempts(ctx context.Context, userID int64) (int, error)
	ResetFailedAttempts(ctx context.Context, userID int64) error
	LockAccount(ctx context.Context, userID int64, reason string) error
	UnlockAccount(ctx context.Context, userID int64) error
	UpdateLastLogin(ctx context.Context, userID int64, ip string) error
	SetMustChangePassword(ctx context.Context, userID int64, mustChange bool) error
	SetEmailVerified(ctx context.Context, userID int64, verified bool) error
	CheckEmailExists(ctx context.Context, email string, excludeUserID int64) (bool, error)
	CheckPhoneExists(ctx context.Context, phone string, excludeUserID int64) (bool, error)
	UpdatePhoneNumber(ctx context.Context, userID int64, phoneNumber string) error
}

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

// NewPostgresUserRepository creates a new repository instance
func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// CreateUser inserts a new user into the database
func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *User) (*User, error) {
	sql := `INSERT INTO users (full_name, email, password, role_id, active_status, photo, organization_id, must_change_password, created_at, updated_at)
	        VALUES ($1, $2, $3, $4, $5, $6, $7, true, $8, $9)
	        RETURNING id, full_name, email, role_id, active_status, photo, organization_id, created_at, updated_at`

	now := time.Now()

	row := r.db.QueryRow(ctx, sql,
		user.FullName,
		user.Email,
		user.Password,
		user.RoleID,
		user.ActiveStatus,
		user.Photo,
		user.OrganizationID,
		now,
		now,
	)

	err := row.Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.RoleID,
		&user.ActiveStatus,
		&user.Photo,
		&user.OrganizationID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("could not insert user: %v", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) CreateUserTx(ctx context.Context, tx pgx.Tx, u *User) (*User, error) {
	query := `
		INSERT INTO users (
			full_name, email, password, role_id, active_status, photo,
			organization_id, must_change_password, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, true, $8, $9
		)
		RETURNING id, full_name, email, role_id, active_status, photo,
		          organization_id, created_at, updated_at;
	`

	now := time.Now()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = now
	}

	err := tx.QueryRow(ctx, query,
		u.FullName, u.Email, u.Password, u.RoleID, u.ActiveStatus, u.Photo,
		u.OrganizationID, u.CreatedAt, u.UpdatedAt,
	).Scan(
		&u.ID,
		&u.FullName,
		&u.Email,
		&u.RoleID,
		&u.ActiveStatus,
		&u.Photo,
		&u.OrganizationID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user in tx: %w", err)
	}
	return u, nil
}

// GetUserByID fetches a user by their ID with organization_id verification
func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id int64, organizationID int64) (*User, error) {
	sql := `
		SELECT id, full_name, email, password, role_id, active_status, photo,
		       organization_id, created_at, updated_at,
		       COALESCE(must_change_password, false),
		       COALESCE(email_verified, false),
		       phone_number,
		       locked_at
		FROM users
		WHERE id = $1 AND organization_id = $2
	`

	row := r.db.QueryRow(ctx, sql, id, organizationID)

	var user User
	err := row.Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Password,
		&user.RoleID,
		&user.ActiveStatus,
		&user.Photo,
		&user.OrganizationID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.MustChangePassword,
		&user.EmailVerified,
		&user.PhoneNumber,
		&user.LockedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("could not fetch user: %v", err)
	}

	// --- Fetch Role ---
	roleSQL := `SELECT id, name FROM roles WHERE id = $1`
	var role role.Role
	err = r.db.QueryRow(ctx, roleSQL, user.RoleID).Scan(&role.ID, &role.Name)
	if err != nil {
		// Log the error and continue without the role
		fmt.Printf("could not fetch role for user %d: %v\n", user.ID, err)
	} else {
		// --- Fetch Permissions ---
		permSQL := `
			SELECT p.id, p.name
			FROM permissions p
			JOIN role_permissions rp ON rp.permission_id = p.id
			WHERE rp.role_id = $1
		`
		rows, err := r.db.Query(ctx, permSQL, role.ID)
		if err != nil {
			// Log the error and continue without permissions
			fmt.Printf("could not fetch permissions for role %d: %v\n", role.ID, err)
		} else {
			defer rows.Close()

			for rows.Next() {
				var perm permission.Permission
				if err := rows.Scan(&perm.ID, &perm.Name); err != nil {
					// Log the error and continue
					fmt.Printf("could not scan permission: %v\n", err)
				} else {
					role.Permissions = append(role.Permissions, perm)
				}
			}
		}
		user.Role = &role
	}

	// --- Fetch Organization Name only ---
	var orgName string
	err = r.db.QueryRow(ctx, `SELECT name FROM organizations WHERE id = $1`, user.OrganizationID).Scan(&orgName)
	if err != nil {
		// Log the error and continue without the organization name
		fmt.Printf("could not fetch org name for user %d: %v\n", user.ID, err)
	} else {
		user.Organization = &organization.Organization{Name: orgName}
	}

	return &user, nil
}

func (r *PostgresUserRepository) GetAllUsers(
	ctx context.Context,
	pg pagination.Pagination,
	organizationID int64,
	userType string,
) (pagination.Result[UserListItem], error) {
	args := []interface{}{}
	where := "WHERE u.deleted_at IS NULL"

	if pg.Search != "" {
		args = append(args, "%"+pg.Search+"%")
		ph := len(args)
		where += fmt.Sprintf(" AND (u.email ILIKE $%d OR u.full_name ILIKE $%d)", ph, ph)
	}

	if organizationID != 0 {
		args = append(args, organizationID)
		ph := len(args)
		where += fmt.Sprintf(" AND u.organization_id = $%d", ph)
	}

	// User type filtering for SuperAdmin tabs
	switch userType {
	case "system":
		where += " AND u.organization_id = 1"
	case "org_admin":
		where += " AND u.organization_id != 1"
	}

	sortField := "u.id"
	switch pg.SortBy {
	case "email", "full_name", "created_at":
		sortField = "u." + pg.SortBy
	}
	sortOrder := "ASC"
	if strings.EqualFold(pg.SortOrder, "desc") {
		sortOrder = "DESC"
	}

	baseFrom := `
	FROM users u
	INNER JOIN roles r ON r.id = u.role_id
	LEFT JOIN organizations o ON o.id = u.organization_id
	`

	//  Pagination params
	offset := pg.Offset()
	args = append(args, pg.PageSize, offset)
	limitPH := len(args) - 1
	offsetPH := len(args)

	//  Main query
	query := fmt.Sprintf(`
SELECT
	u.id,
	u.full_name,
	u.email,
	COALESCE(r.name, '') AS role_name,
	COALESCE(u.active_status, false) AS active_status,
	COALESCE(u.photo, '') AS photo_url,
	CASE
		WHEN COALESCE(u.organization_id, 0) = 0 THEN 'HQ'
		ELSE COALESCE(o.name, '')
	END AS organization,
	CASE WHEN u.locked_at IS NOT NULL THEN true ELSE false END AS is_locked
%s
%s
ORDER BY %s %s
LIMIT $%d OFFSET $%d
`, baseFrom, where, sortField, sortOrder, limitPH, offsetPH)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return pagination.Result[UserListItem]{}, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	items := make([]UserListItem, 0, pg.PageSize)
	for rows.Next() {
		var it UserListItem
		if err := rows.Scan(
			&it.ID,
			&it.FullName,
			&it.Email,
			&it.Role,
			&it.ActiveStatus,
			&it.PhotoURL,
			&it.Organization,
			&it.IsLocked,
		); err != nil {
			return pagination.Result[UserListItem]{}, fmt.Errorf("scan error: %w", err)
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		return pagination.Result[UserListItem]{}, fmt.Errorf("rows error: %w", rows.Err())
	}

	// Count query
	countArgs := args[:len(args)-2]
	countQuery := fmt.Sprintf(`SELECT COUNT(*) %s %s`, baseFrom, where)

	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return pagination.Result[UserListItem]{}, fmt.Errorf("count query failed: %w", err)
	}

	return pagination.Result[UserListItem]{
		Data:       items,
		Page:       pg.Page,
		PageSize:   pg.PageSize,
		TotalCount: total,
	}, nil
}

func (r *PostgresUserRepository) GetUserCounts(ctx context.Context) (map[string]int, error) {
	query := `
	SELECT
		COUNT(*) FILTER (WHERE organization_id = 1) AS system_count,
		COUNT(*) FILTER (WHERE organization_id != 1) AS org_admin_count
	FROM users
	WHERE deleted_at IS NULL
	`
	var system, orgAdmin int
	err := r.db.QueryRow(ctx, query).Scan(&system, &orgAdmin)
	if err != nil {
		return nil, fmt.Errorf("user counts query failed: %w", err)
	}
	return map[string]int{
		"system":    system,
		"org_admin": orgAdmin,
	}, nil
}

// GetUsersByOrganization retrieves users ONLY within a specific organization.
// Organization ID is MANDATORY - this enforces strict tenant isolation.
// Used by org-admin routes for data isolation.
func (r *PostgresUserRepository) GetUsersByOrganization(
	ctx context.Context,
	pg pagination.Pagination,
	organizationID int64,
	userType string,
) (pagination.Result[UserListItem], error) {
	// CRITICAL: Organization ID must be provided for tenant isolation
	if organizationID == 0 {
		return pagination.Result[UserListItem]{}, fmt.Errorf("organization_id is required for tenant user queries")
	}

	// Build query with organization as FIRST parameter (always $1)
	args := []interface{}{organizationID}
	where := "WHERE u.deleted_at IS NULL AND u.organization_id = $1"

	// Search filter (optional)
	if pg.Search != "" {
		searchPattern := "%" + pg.Search + "%"
		args = append(args, searchPattern)
		ph := len(args)
		where += fmt.Sprintf(" AND (u.email ILIKE $%d OR u.full_name ILIKE $%d)", ph, ph)
	}

	// User type filter based on role permissions
	// "admin" = role has tenant.* permissions (HR/admin roles)
	// "employee" = role has NO tenant.* permissions (pure employee roles)
	switch userType {
	case "admin":
		where += " AND EXISTS (SELECT 1 FROM role_permissions rp JOIN permissions p ON p.id = rp.permission_id WHERE rp.role_id = u.role_id AND p.name LIKE 'tenant.%')"
	case "employee":
		where += " AND NOT EXISTS (SELECT 1 FROM role_permissions rp JOIN permissions p ON p.id = rp.permission_id WHERE rp.role_id = u.role_id AND p.name LIKE 'tenant.%')"
	}

	// Sorting
	sortField := "u.id"
	switch pg.SortBy {
	case "email", "full_name", "created_at":
		sortField = "u." + pg.SortBy
	}
	sortOrder := "ASC"
	if strings.EqualFold(pg.SortOrder, "desc") {
		sortOrder = "DESC"
	}

	baseFrom := `
	FROM users u
	INNER JOIN roles r ON r.id = u.role_id
	LEFT JOIN organizations o ON o.id = u.organization_id
	`

	// Pagination
	offset := pg.Offset()
	args = append(args, pg.PageSize, offset)
	limitPH := len(args) - 1
	offsetPH := len(args)

	query := fmt.Sprintf(`
SELECT
	u.id,
	u.full_name,
	u.email,
	COALESCE(r.name, '') AS role_name,
	COALESCE(u.active_status, false) AS active_status,
	COALESCE(u.photo, '') AS photo_url,
	COALESCE(o.name, '') AS organization,
	CASE WHEN u.locked_at IS NOT NULL THEN true ELSE false END AS is_locked
%s
%s
ORDER BY %s %s
LIMIT $%d OFFSET $%d
`, baseFrom, where, sortField, sortOrder, limitPH, offsetPH)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return pagination.Result[UserListItem]{}, fmt.Errorf("org users query failed: %w", err)
	}
	defer rows.Close()

	items := make([]UserListItem, 0, pg.PageSize)
	for rows.Next() {
		var it UserListItem
		if err := rows.Scan(
			&it.ID,
			&it.FullName,
			&it.Email,
			&it.Role,
			&it.ActiveStatus,
			&it.PhotoURL,
			&it.Organization,
			&it.IsLocked,
		); err != nil {
			return pagination.Result[UserListItem]{}, fmt.Errorf("scan error: %w", err)
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		return pagination.Result[UserListItem]{}, fmt.Errorf("rows error: %w", rows.Err())
	}

	// Count query (exclude pagination args)
	countArgs := args[:len(args)-2]
	countQuery := fmt.Sprintf(`SELECT COUNT(*) %s %s`, baseFrom, where)

	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return pagination.Result[UserListItem]{}, fmt.Errorf("count query failed: %w", err)
	}

	return pagination.Result[UserListItem]{
		Data:       items,
		Page:       pg.Page,
		PageSize:   pg.PageSize,
		TotalCount: total,
	}, nil
}

func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	sql := `SELECT id, full_name, email, password, role_id, active_status, photo,
	               organization_id, created_at, updated_at
	        FROM users
	        WHERE email = $1`

	row := r.db.QueryRow(ctx, sql, email)

	var user User
	err := row.Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Password,
		&user.RoleID,
		&user.ActiveStatus,
		&user.Photo,
		&user.OrganizationID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {

		if err == pgx.ErrNoRows {

			return nil, fmt.Errorf("user with email '%s' not found", email)
		}
		return nil, fmt.Errorf("could not fetch user: %v", err)
	}

	return &user, nil
}

// GetUserByIDNoOrg retrieves a user by ID without organization verification
func (r *PostgresUserRepository) GetUserByIDNoOrg(ctx context.Context, id int64) (*User, error) {
	sql := `SELECT id, full_name, email, password, role_id, active_status, photo,
	               organization_id, created_at, updated_at, locked_at
	        FROM users
	        WHERE id = $1 AND deleted_at IS NULL`

	row := r.db.QueryRow(ctx, sql, id)

	var user User
	err := row.Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Password,
		&user.RoleID,
		&user.ActiveStatus,
		&user.Photo,
		&user.OrganizationID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LockedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, fmt.Errorf("could not fetch user: %v", err)
	}

	return &user, nil
}

// GetPermissionsForRole fetches all permissions associated with a role
func (r *PostgresUserRepository) GetPermissionsForRole(ctx context.Context, roleID int64) ([]permission.Permission, error) {
	sql := `SELECT p.id, p.name
			FROM permissions p
			JOIN role_permissions rp ON p.id = rp.permission_id
			WHERE rp.role_id = $1`

	rows, err := r.db.Query(ctx, sql, roleID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch permissions: %v", err)
	}
	defer rows.Close()

	var permissions []permission.Permission
	for rows.Next() {
		var perm permission.Permission
		if err := rows.Scan(&perm.ID, &perm.Name); err != nil {
			return nil, fmt.Errorf("could not scan permission: %v", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// GetRoleForUser fetches role details by role ID
func (r *PostgresUserRepository) GetRoleForUser(ctx context.Context, roleID int64) (*role.Role, error) {
	sql := `SELECT id, name FROM roles WHERE id = $1`

	row := r.db.QueryRow(ctx, sql, roleID)

	var rData role.Role
	err := row.Scan(&rData.ID, &rData.Name)
	if err != nil {
		return nil, fmt.Errorf("could not fetch role: %v", err)
	}

	return &rData, nil
}

func (r *PostgresUserRepository) UpdateUserStatus(ctx context.Context, id int64, organizationID int64) error {
	sql := `UPDATE users SET active_status = NOT active_status WHERE id = $1 AND organization_id = $2`

	_, err := r.db.Exec(ctx, sql, id, organizationID)
	if err != nil {
		return fmt.Errorf("could not toggle user status: %v", err)
	}
	return nil
}

// SoftDeleteUser marks a user as deleted by setting deleted_at with organization verification
func (r *PostgresUserRepository) SoftDeleteUser(ctx context.Context, id int64, organizationID int64) error {
	sql := `UPDATE users SET deleted_at = $1 WHERE id = $2 AND organization_id = $3 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, sql, time.Now(), id, organizationID)
	if err != nil {
		return fmt.Errorf("could not soft delete user: %v", err)
	}
	return nil
}

// HardDeleteUser permanently removes a user with organization verification
func (r *PostgresUserRepository) HardDeleteUser(ctx context.Context, id int64, organizationID int64) error {
	sql := `DELETE FROM users WHERE id = $1 AND organization_id = $2`
	_, err := r.db.Exec(ctx, sql, id, organizationID)
	if err != nil {
		return fmt.Errorf("could not hard delete user: %v", err)
	}
	return nil
}

func (r *PostgresUserRepository) ChangePassword(ctx context.Context, userID int64, organizationID int64, newPassword string) error {
	sql := `UPDATE users SET password = $1 WHERE id = $2 AND organization_id = $3`
	_, err := r.db.Exec(ctx, sql, newPassword, userID, organizationID)
	if err != nil {
		return fmt.Errorf("cant changed password%v", err)
	}
	return nil
}

// UpdatePassword updates a user's password without organization verification (for password reset)
func (r *PostgresUserRepository) UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error {
	sql := `UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`
	result, err := r.db.Exec(ctx, sql, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

func (r *PostgresUserRepository) UpdateUserPartial(ctx context.Context, input *UpdateUserInput) (*User, error) {
	setClauses := []string{}
	args := []interface{}{}
	argID := 1

	if input.FullName != nil {
		setClauses = append(setClauses, fmt.Sprintf("full_name = $%d", argID))
		args = append(args, *input.FullName)
		argID++
	}
	if input.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argID))
		args = append(args, *input.Email)
		argID++
	}
	if input.Password != nil {
		setClauses = append(setClauses, fmt.Sprintf("password = $%d", argID))
		args = append(args, *input.Password)
		argID++
	}
	if input.RoleID != nil {
		setClauses = append(setClauses, fmt.Sprintf("role_id = $%d", argID))
		args = append(args, *input.RoleID)
		argID++
	}
	if input.ActiveStatus != nil {
		setClauses = append(setClauses, fmt.Sprintf("active_status = $%d", argID))
		args = append(args, *input.ActiveStatus)
		argID++
	}
	if input.Photo != nil {
		setClauses = append(setClauses, fmt.Sprintf("photo = $%d", argID))
		args = append(args, *input.Photo)
		argID++
	}
	if input.OrganizationID != nil {
		setClauses = append(setClauses, fmt.Sprintf("organization_id = $%d", argID))
		args = append(args, *input.OrganizationID)
		argID++
	}
	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argID))
	now := time.Now()
	args = append(args, now)
	argID++

	// Add ID as final argument
	if input.ID == 0 {
		return nil, errors.New("user ID is required for update")
	}
	args = append(args, input.ID)

	query := fmt.Sprintf(`
		UPDATE users SET %s WHERE id = $%d
		RETURNING id, full_name, email, role_id, active_status, photo, organization_id, created_at, updated_at`,
		strings.Join(setClauses, ", "), argID,
	)

	row := r.db.QueryRow(ctx, query, args...)

	var updatedUser User
	err := row.Scan(
		&updatedUser.ID,
		&updatedUser.FullName,
		&updatedUser.Email,
		&updatedUser.RoleID,
		&updatedUser.ActiveStatus,
		&updatedUser.Photo,
		&updatedUser.OrganizationID,
		&updatedUser.CreatedAt,
		&updatedUser.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	return &updatedUser, nil
}

// GetOrgsAllUsersLimited retrieves users ONLY within a specific organization (limited fields).
// This method enforces STRICT tenant isolation - organizationID is MANDATORY.
// Used exclusively by tenant/org-admin routes for data isolation.
func (r *PostgresUserRepository) GetOrgsAllUsersLimited(
	ctx context.Context,
	pg pagination.Pagination,
	organizationID int64,
) (pagination.Result[UserListItemLimited], error) {
	// CRITICAL: Organization ID must be provided for tenant isolation
	if organizationID == 0 {
		return pagination.Result[UserListItemLimited]{}, fmt.Errorf("organization_id is required for tenant user queries")
	}

	// Build query with organization as FIRST parameter (always $1)
	args := []interface{}{organizationID}
	where := "WHERE u.deleted_at IS NULL AND u.organization_id = $1"

	// Search filter (optional)
	if pg.Search != "" {
		searchPattern := "%" + pg.Search + "%"
		args = append(args, searchPattern)
		searchPH := len(args)
		where += fmt.Sprintf(" AND (u.email ILIKE $%d OR u.full_name ILIKE $%d)", searchPH, searchPH)
	}

	// Sorting
	sortField := "u.id"
	switch pg.SortBy {
	case "full_name", "created_at":
		sortField = "u." + pg.SortBy
	}
	sortOrder := "ASC"
	if strings.EqualFold(pg.SortOrder, "desc") {
		sortOrder = "DESC"
	}

	baseFrom := `
FROM users u
INNER JOIN roles r ON r.id = u.role_id
LEFT JOIN organizations o ON o.id = u.organization_id
`

	// Pagination
	offset := pg.Offset()
	args = append(args, pg.PageSize, offset)
	limitPH := len(args) - 1
	offsetPH := len(args)

	query := fmt.Sprintf(`
SELECT
	u.id,
	u.full_name,
	u.email,
	COALESCE(r.name, '') AS role_name,
	COALESCE(u.active_status, false) AS active_status,
	COALESCE(u.photo, '') AS photo_url,
	COALESCE(o.name, '') AS organization
%s
%s
ORDER BY %s %s
LIMIT $%d OFFSET $%d
`, baseFrom, where, sortField, sortOrder, limitPH, offsetPH)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return pagination.Result[UserListItemLimited]{}, fmt.Errorf("org users limited query failed: %w", err)
	}
	defer rows.Close()

	items := make([]UserListItemLimited, 0, pg.PageSize)
	for rows.Next() {
		var it UserListItemLimited
		if err := rows.Scan(
			&it.ID,
			&it.FullName,
			&it.Email,
			&it.Role,
			&it.ActiveStatus,
			&it.PhotoURL,
			&it.Organization,
		); err != nil {
			return pagination.Result[UserListItemLimited]{}, fmt.Errorf("scan error: %w", err)
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		return pagination.Result[UserListItemLimited]{}, fmt.Errorf("rows error: %w", rows.Err())
	}

	// Count query (exclude pagination args)
	countArgs := args[:len(args)-2]
	countQuery := fmt.Sprintf(`SELECT COUNT(*) %s %s`, baseFrom, where)

	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return pagination.Result[UserListItemLimited]{}, fmt.Errorf("count query failed: %w", err)
	}

	return pagination.Result[UserListItemLimited]{
		Data:       items,
		Page:       pg.Page,
		PageSize:   pg.PageSize,
		TotalCount: total,
	}, nil
}

// GetUserByEmailWithSecurity fetches a user by email including all security fields
func (r *PostgresUserRepository) GetUserByEmailWithSecurity(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, full_name, email, password, role_id, active_status, photo,
		       organization_id, created_at, updated_at,
		       COALESCE(failed_login_attempts, 0), locked_at, lock_reason,
		       last_login_at, last_login_ip, COALESCE(must_change_password, false),
		       COALESCE(two_factor_enabled, false), two_factor_method, totp_secret, phone_number
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	row := r.db.QueryRow(ctx, query, email)

	var user User
	err := row.Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Password,
		&user.RoleID,
		&user.ActiveStatus,
		&user.Photo,
		&user.OrganizationID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.FailedLoginAttempts,
		&user.LockedAt,
		&user.LockReason,
		&user.LastLoginAt,
		&user.LastLoginIP,
		&user.MustChangePassword,
		&user.TwoFactorEnabled,
		&user.TwoFactorMethod,
		&user.TOTPSecret,
		&user.PhoneNumber,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user with email '%s' not found", email)
		}
		return nil, fmt.Errorf("could not fetch user: %v", err)
	}

	// Fetch role with permissions
	roleSQL := `SELECT id, name, COALESCE(requires_2fa, false) FROM roles WHERE id = $1`
	var userRole role.Role
	var requires2FA bool
	err = r.db.QueryRow(ctx, roleSQL, user.RoleID).Scan(&userRole.ID, &userRole.Name, &requires2FA)
	if err == nil {
		// Fetch permissions
		permSQL := `
			SELECT p.id, p.name
			FROM permissions p
			JOIN role_permissions rp ON rp.permission_id = p.id
			WHERE rp.role_id = $1
		`
		rows, err := r.db.Query(ctx, permSQL, userRole.ID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var perm permission.Permission
				if err := rows.Scan(&perm.ID, &perm.Name); err == nil {
					userRole.Permissions = append(userRole.Permissions, perm)
				}
			}
		}
		user.Role = &userRole
	}

	return &user, nil
}

// IncrementFailedAttempts increments the failed login attempts counter and returns the new count
func (r *PostgresUserRepository) IncrementFailedAttempts(ctx context.Context, userID int64) (int, error) {
	query := `
		UPDATE users
		SET failed_login_attempts = COALESCE(failed_login_attempts, 0) + 1,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING failed_login_attempts
	`

	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("could not increment failed attempts: %v", err)
	}

	return count, nil
}

// ResetFailedAttempts resets the failed login attempts counter to zero
func (r *PostgresUserRepository) ResetFailedAttempts(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET failed_login_attempts = 0,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("could not reset failed attempts: %v", err)
	}

	return nil
}

// LockAccount locks the user account with a reason
func (r *PostgresUserRepository) LockAccount(ctx context.Context, userID int64, reason string) error {
	query := `
		UPDATE users
		SET locked_at = NOW(),
		    lock_reason = $2,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID, reason)
	if err != nil {
		return fmt.Errorf("could not lock account: %v", err)
	}

	return nil
}

// UnlockAccount unlocks the user account and resets failed attempts
func (r *PostgresUserRepository) UnlockAccount(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET locked_at = NULL,
		    lock_reason = NULL,
		    failed_login_attempts = 0,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("could not unlock account: %v", err)
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp and IP address
func (r *PostgresUserRepository) UpdateLastLogin(ctx context.Context, userID int64, ip string) error {
	query := `
		UPDATE users
		SET last_login_at = NOW(),
		    last_login_ip = $2,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID, ip)
	if err != nil {
		return fmt.Errorf("could not update last login: %v", err)
	}

	return nil
}

// SetMustChangePassword sets the must_change_password flag
func (r *PostgresUserRepository) SetMustChangePassword(ctx context.Context, userID int64, mustChange bool) error {
	query := `
		UPDATE users
		SET must_change_password = $2,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID, mustChange)
	if err != nil {
		return fmt.Errorf("could not set must_change_password: %v", err)
	}

	return nil
}

// SetEmailVerified sets the email_verified flag for a user
func (r *PostgresUserRepository) SetEmailVerified(ctx context.Context, userID int64, verified bool) error {
	query := `
		UPDATE users
		SET email_verified = $2,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, userID, verified)
	if err != nil {
		return fmt.Errorf("could not set email_verified: %v", err)
	}
	return nil
}

// CheckEmailExists checks if an email is already in use by another user
func (r *PostgresUserRepository) CheckEmailExists(ctx context.Context, email string, excludeUserID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, email, excludeUserID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("could not check email existence: %v", err)
	}
	return exists, nil
}

// CheckPhoneExists checks if a phone number is already in use by another user
func (r *PostgresUserRepository) CheckPhoneExists(ctx context.Context, phone string, excludeUserID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE phone_number = $1 AND id != $2 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, phone, excludeUserID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("could not check phone existence: %v", err)
	}
	return exists, nil
}

// UpdatePhoneNumber updates the phone number for a user
func (r *PostgresUserRepository) UpdatePhoneNumber(ctx context.Context, userID int64, phoneNumber string) error {
	query := `UPDATE users SET phone_number = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, phoneNumber, userID)
	if err != nil {
		return fmt.Errorf("could not update phone number: %v", err)
	}
	return nil
}

