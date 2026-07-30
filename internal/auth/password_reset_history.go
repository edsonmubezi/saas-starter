package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// PasswordResetMethod represents the method used to reset password
type PasswordResetMethod string

const (
	ResetMethodEmail  PasswordResetMethod = "email"
	ResetMethodForm   PasswordResetMethod = "form"
	ResetMethodAdmin  PasswordResetMethod = "admin_reset"
	ResetMethodSelf   PasswordResetMethod = "self_service"
	ResetMethodForced PasswordResetMethod = "forced"
)

// PasswordResetStatus represents the status of the reset
type PasswordResetStatus string

const (
	ResetStatusPending   PasswordResetStatus = "pending"
	ResetStatusCompleted PasswordResetStatus = "completed"
	ResetStatusExpired   PasswordResetStatus = "expired"
	ResetStatusCancelled PasswordResetStatus = "cancelled"
)

// PasswordResetHistory represents a password reset history record
type PasswordResetHistory struct {
	ID               int64               `json:"id"`
	UserID           int64               `json:"user_id"`
	InitiatedByID    *int64              `json:"initiated_by_id,omitempty"`
	InitiatedByName  *string             `json:"initiated_by_name,omitempty"`
	ResetMethod      PasswordResetMethod `json:"reset_method"`
	Status           PasswordResetStatus `json:"status"`
	IPAddress        string              `json:"ip_address"`
	UserAgent        string              `json:"user_agent"`
	FormAttached     bool                `json:"form_attached"`
	FormURL          *string             `json:"form_url,omitempty"`
	Notes            *string             `json:"notes,omitempty"`
	TokenExpiresAt   *time.Time          `json:"token_expires_at,omitempty"`
	CompletedAt      *time.Time          `json:"completed_at,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	OrganizationID   *int64              `json:"organization_id,omitempty"`
	OrganizationName *string             `json:"organization_name,omitempty"`
	UserEmail        *string             `json:"user_email,omitempty"`
	UserFullName     *string             `json:"user_fullname,omitempty"`
}

// PasswordResetHistoryInput represents input for creating a reset history record
type PasswordResetHistoryInput struct {
	UserID         int64
	InitiatedByID  *int64
	ResetMethod    PasswordResetMethod
	IPAddress      string
	UserAgent      string
	FormAttached   bool
	FormURL        *string
	Notes          *string
	TokenExpiresAt *time.Time
	OrganizationID *int64
}

// PasswordResetHistoryRepository defines the interface for password reset history operations
type PasswordResetHistoryRepository interface {
	Create(ctx context.Context, input PasswordResetHistoryInput) (int64, error)
	UpdateStatus(ctx context.Context, id int64, status PasswordResetStatus) error
	MarkCompleted(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*PasswordResetHistory, error)
	GetByUserID(ctx context.Context, userID int64, limit int) ([]PasswordResetHistory, error)
	GetByOrganizationID(ctx context.Context, orgID int64, limit, offset int) ([]PasswordResetHistory, error)
	GetPendingByUserID(ctx context.Context, userID int64) (*PasswordResetHistory, error)
	GetAll(ctx context.Context, limit, offset int) ([]PasswordResetHistory, error)
	CountByOrganization(ctx context.Context, orgID int64) (int, error)
	CountAll(ctx context.Context) (int, error)
}

// PostgresPasswordResetHistoryRepository implements PasswordResetHistoryRepository using PostgreSQL
type PostgresPasswordResetHistoryRepository struct {
	db *pgxpool.Pool
}

// NewPasswordResetHistoryRepository creates a new PostgreSQL password reset history repository
func NewPasswordResetHistoryRepository(db *pgxpool.Pool) *PostgresPasswordResetHistoryRepository {
	return &PostgresPasswordResetHistoryRepository{db: db}
}

// Create creates a new password reset history record
func (r *PostgresPasswordResetHistoryRepository) Create(ctx context.Context, input PasswordResetHistoryInput) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO password_reset_history (
			user_id, initiated_by_id, reset_method, status, ip_address, user_agent,
			form_attached, form_url, notes, token_expires_at, organization_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`,
		input.UserID,
		input.InitiatedByID,
		input.ResetMethod,
		ResetStatusPending,
		input.IPAddress,
		input.UserAgent,
		input.FormAttached,
		input.FormURL,
		input.Notes,
		input.TokenExpiresAt,
		input.OrganizationID,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create password reset history: %w", err)
	}

	return id, nil
}

// UpdateStatus updates the status of a reset record
func (r *PostgresPasswordResetHistoryRepository) UpdateStatus(ctx context.Context, id int64, status PasswordResetStatus) error {
	_, err := r.db.Exec(ctx, `
		UPDATE password_reset_history
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, status, id)

	if err != nil {
		return fmt.Errorf("failed to update password reset status: %w", err)
	}

	return nil
}

// MarkCompleted marks a reset as completed
func (r *PostgresPasswordResetHistoryRepository) MarkCompleted(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE password_reset_history
		SET status = $1, completed_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`, ResetStatusCompleted, id)

	if err != nil {
		return fmt.Errorf("failed to mark password reset as completed: %w", err)
	}

	return nil
}

// GetByID retrieves a password reset history record by ID
func (r *PostgresPasswordResetHistoryRepository) GetByID(ctx context.Context, id int64) (*PasswordResetHistory, error) {
	var record PasswordResetHistory
	err := r.db.QueryRow(ctx, `
		SELECT
			prh.id, prh.user_id, prh.initiated_by_id, prh.reset_method, prh.status,
			prh.ip_address, prh.user_agent, prh.form_attached, prh.form_url, prh.notes,
			prh.token_expires_at, prh.completed_at, prh.created_at, prh.organization_id,
			u.email as user_email, u.fullname as user_fullname,
			ib.fullname as initiated_by_name,
			o.name as organization_name
		FROM password_reset_history prh
		LEFT JOIN users u ON prh.user_id = u.id
		LEFT JOIN users ib ON prh.initiated_by_id = ib.id
		LEFT JOIN organizations o ON prh.organization_id = o.id
		WHERE prh.id = $1
	`, id).Scan(
		&record.ID, &record.UserID, &record.InitiatedByID, &record.ResetMethod, &record.Status,
		&record.IPAddress, &record.UserAgent, &record.FormAttached, &record.FormURL, &record.Notes,
		&record.TokenExpiresAt, &record.CompletedAt, &record.CreatedAt, &record.OrganizationID,
		&record.UserEmail, &record.UserFullName,
		&record.InitiatedByName,
		&record.OrganizationName,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get password reset history: %w", err)
	}

	return &record, nil
}

// GetByUserID retrieves password reset history for a user
func (r *PostgresPasswordResetHistoryRepository) GetByUserID(ctx context.Context, userID int64, limit int) ([]PasswordResetHistory, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			prh.id, prh.user_id, prh.initiated_by_id, prh.reset_method, prh.status,
			prh.ip_address, prh.user_agent, prh.form_attached, prh.form_url, prh.notes,
			prh.token_expires_at, prh.completed_at, prh.created_at, prh.organization_id,
			ib.fullname as initiated_by_name
		FROM password_reset_history prh
		LEFT JOIN users ib ON prh.initiated_by_id = ib.id
		WHERE prh.user_id = $1
		ORDER BY prh.created_at DESC
		LIMIT $2
	`, userID, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to get password reset history: %w", err)
	}
	defer rows.Close()

	var records []PasswordResetHistory
	for rows.Next() {
		var record PasswordResetHistory
		err := rows.Scan(
			&record.ID, &record.UserID, &record.InitiatedByID, &record.ResetMethod, &record.Status,
			&record.IPAddress, &record.UserAgent, &record.FormAttached, &record.FormURL, &record.Notes,
			&record.TokenExpiresAt, &record.CompletedAt, &record.CreatedAt, &record.OrganizationID,
			&record.InitiatedByName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan password reset history: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

// GetByOrganizationID retrieves password reset history for an organization
func (r *PostgresPasswordResetHistoryRepository) GetByOrganizationID(ctx context.Context, orgID int64, limit, offset int) ([]PasswordResetHistory, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			prh.id, prh.user_id, prh.initiated_by_id, prh.reset_method, prh.status,
			prh.ip_address, prh.user_agent, prh.form_attached, prh.form_url, prh.notes,
			prh.token_expires_at, prh.completed_at, prh.created_at, prh.organization_id,
			u.email as user_email, u.fullname as user_fullname,
			ib.fullname as initiated_by_name,
			o.name as organization_name
		FROM password_reset_history prh
		LEFT JOIN users u ON prh.user_id = u.id
		LEFT JOIN users ib ON prh.initiated_by_id = ib.id
		LEFT JOIN organizations o ON prh.organization_id = o.id
		WHERE prh.organization_id = $1
		ORDER BY prh.created_at DESC
		LIMIT $2 OFFSET $3
	`, orgID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to get password reset history: %w", err)
	}
	defer rows.Close()

	var records []PasswordResetHistory
	for rows.Next() {
		var record PasswordResetHistory
		err := rows.Scan(
			&record.ID, &record.UserID, &record.InitiatedByID, &record.ResetMethod, &record.Status,
			&record.IPAddress, &record.UserAgent, &record.FormAttached, &record.FormURL, &record.Notes,
			&record.TokenExpiresAt, &record.CompletedAt, &record.CreatedAt, &record.OrganizationID,
			&record.UserEmail, &record.UserFullName,
			&record.InitiatedByName,
			&record.OrganizationName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan password reset history: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

// GetPendingByUserID retrieves any pending password reset for a user
func (r *PostgresPasswordResetHistoryRepository) GetPendingByUserID(ctx context.Context, userID int64) (*PasswordResetHistory, error) {
	var record PasswordResetHistory
	err := r.db.QueryRow(ctx, `
		SELECT
			id, user_id, initiated_by_id, reset_method, status,
			ip_address, user_agent, form_attached, form_url, notes,
			token_expires_at, completed_at, created_at, organization_id
		FROM password_reset_history
		WHERE user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, userID, ResetStatusPending).Scan(
		&record.ID, &record.UserID, &record.InitiatedByID, &record.ResetMethod, &record.Status,
		&record.IPAddress, &record.UserAgent, &record.FormAttached, &record.FormURL, &record.Notes,
		&record.TokenExpiresAt, &record.CompletedAt, &record.CreatedAt, &record.OrganizationID,
	)

	if err != nil {
		return nil, err
	}

	return &record, nil
}

// GetAll retrieves all password reset history records (for super admin)
func (r *PostgresPasswordResetHistoryRepository) GetAll(ctx context.Context, limit, offset int) ([]PasswordResetHistory, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			prh.id, prh.user_id, prh.initiated_by_id, prh.reset_method, prh.status,
			prh.ip_address, prh.user_agent, prh.form_attached, prh.form_url, prh.notes,
			prh.token_expires_at, prh.completed_at, prh.created_at, prh.organization_id,
			u.email as user_email, u.fullname as user_fullname,
			ib.fullname as initiated_by_name,
			o.name as organization_name
		FROM password_reset_history prh
		LEFT JOIN users u ON prh.user_id = u.id
		LEFT JOIN users ib ON prh.initiated_by_id = ib.id
		LEFT JOIN organizations o ON prh.organization_id = o.id
		ORDER BY prh.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to get password reset history: %w", err)
	}
	defer rows.Close()

	var records []PasswordResetHistory
	for rows.Next() {
		var record PasswordResetHistory
		err := rows.Scan(
			&record.ID, &record.UserID, &record.InitiatedByID, &record.ResetMethod, &record.Status,
			&record.IPAddress, &record.UserAgent, &record.FormAttached, &record.FormURL, &record.Notes,
			&record.TokenExpiresAt, &record.CompletedAt, &record.CreatedAt, &record.OrganizationID,
			&record.UserEmail, &record.UserFullName,
			&record.InitiatedByName,
			&record.OrganizationName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan password reset history: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

// CountByOrganization counts password reset history records for an organization
func (r *PostgresPasswordResetHistoryRepository) CountByOrganization(ctx context.Context, orgID int64) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM password_reset_history WHERE organization_id = $1
	`, orgID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count password reset history: %w", err)
	}

	return count, nil
}

// CountAll counts all password reset history records
func (r *PostgresPasswordResetHistoryRepository) CountAll(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM password_reset_history`).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count password reset history: %w", err)
	}

	return count, nil
}
