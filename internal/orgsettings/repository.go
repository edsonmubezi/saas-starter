package orgsettings

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// OrganizationSettingsRepository defines methods for organization settings data access
type OrganizationSettingsRepository interface {
	CreateOrgSettings(ctx context.Context, settings *OrganizationSettings) error
	CreateOrgSettingsTx(ctx context.Context, tx pgx.Tx, settings *OrganizationSettings) error
	UpdateOrgSettings(ctx context.Context, settings *OrganizationSettings) error
	GetOrgSettingsByOrganizationID(ctx context.Context, organizationID int64) (*OrganizationSettings, error)
	GetOrgSettingsByID(ctx context.Context, id int64) (*OrganizationSettings, error)
}

type orgSettingsRepo struct {
	db *pgxpool.Pool
}

// NewOrganizationSettingsRepository creates a new instance of organization settings repository
func NewOrganizationSettingsRepository(db *pgxpool.Pool) OrganizationSettingsRepository {
	return &orgSettingsRepo{db: db}
}

// CreateOrgSettings creates new organization settings
func (r *orgSettingsRepo) CreateOrgSettings(ctx context.Context, settings *OrganizationSettings) error {
	// Check for duplicate
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM organization_settings WHERE organization_id = $1)`
	if err := r.db.QueryRow(ctx, query, settings.OrganizationID).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check duplicate organization settings: %w", err)
	}
	if exists {
		return errors.New("organization settings already exist for this organization")
	}

	insertQuery := `
		INSERT INTO organization_settings (
			organization_id,
			organization_type,
			session_lock_timeout_minutes
		) VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(
		ctx,
		insertQuery,
		settings.OrganizationID,
		settings.OrganizationType,
		settings.SessionLockTimeoutMinutes,
	).Scan(&settings.ID, &settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert organization settings: %w", err)
	}
	return nil
}

// CreateOrgSettingsTx creates organization settings within a transaction
func (r *orgSettingsRepo) CreateOrgSettingsTx(ctx context.Context, tx pgx.Tx, settings *OrganizationSettings) error {
	// Check for duplicate
	var exists bool
	if err := tx.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM organization_settings WHERE organization_id = $1)`,
		settings.OrganizationID,
	).Scan(&exists); err != nil {
		return fmt.Errorf("duplicate check: %w", err)
	}
	if exists {
		return errors.New("organization settings already exist for this organization")
	}

	// Insert
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO organization_settings (
			organization_id,
			organization_type,
			session_lock_timeout_minutes
		) VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`,
		settings.OrganizationID,
		settings.OrganizationType,
		settings.SessionLockTimeoutMinutes,
	).Scan(&settings.ID, &settings.CreatedAt, &settings.UpdatedAt); err != nil {
		return fmt.Errorf("insert organization settings: %w", err)
	}
	return nil
}

// UpdateOrgSettings updates organization settings
func (r *orgSettingsRepo) UpdateOrgSettings(ctx context.Context, settings *OrganizationSettings) error {
	updateQuery := `
		UPDATE organization_settings
		SET organization_type = $1,
		    session_lock_timeout_minutes = $2,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING updated_at
	`
	err := r.db.QueryRow(
		ctx,
		updateQuery,
		settings.OrganizationType,
		settings.SessionLockTimeoutMinutes,
		settings.ID,
	).Scan(&settings.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("organization settings not found")
		}
		return fmt.Errorf("failed to update organization settings: %w", err)
	}
	return nil
}

// GetOrgSettingsByOrganizationID fetches organization settings by organization ID
func (r *orgSettingsRepo) GetOrgSettingsByOrganizationID(ctx context.Context, organizationID int64) (*OrganizationSettings, error) {
	query := `
		SELECT
			id,
			organization_id,
			organization_type,
			COALESCE(session_lock_timeout_minutes, 15) as session_lock_timeout_minutes,
			created_at,
			updated_at
		FROM organization_settings
		WHERE organization_id = $1
	`

	settings := &OrganizationSettings{}
	err := r.db.QueryRow(ctx, query, organizationID).Scan(
		&settings.ID,
		&settings.OrganizationID,
		&settings.OrganizationType,
		&settings.SessionLockTimeoutMinutes,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get organization settings by organization: %w", err)
	}

	return settings, nil
}

// GetOrgSettingsByID fetches organization settings by primary key ID
func (r *orgSettingsRepo) GetOrgSettingsByID(ctx context.Context, id int64) (*OrganizationSettings, error) {
	query := `
		SELECT
			id,
			organization_id,
			organization_type,
			COALESCE(session_lock_timeout_minutes, 15) as session_lock_timeout_minutes,
			created_at,
			updated_at
		FROM organization_settings
		WHERE id = $1
	`
	settings := &OrganizationSettings{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&settings.ID,
		&settings.OrganizationID,
		&settings.OrganizationType,
		&settings.SessionLockTimeoutMinutes,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get organization settings by id: %w", err)
	}
	return settings, nil
}
