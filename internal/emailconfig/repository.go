package emailconfig

import (
	"context"
	"errors"
	"fmt"

	secureid "github.com/edsonmubezi/myapp/pkg/encrypt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Repository defines the interface for email config data access.
type Repository interface {
	GetByOrgID(ctx context.Context, orgID int64) (*OrganizationEmailConfig, error)
	Upsert(ctx context.Context, orgID int64, req *UpsertEmailConfigRequest) (*OrganizationEmailConfig, error)
}

type repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new email config repository.
func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

// GetByOrgID fetches email config for an organization. Returns nil, nil if not found.
func (r *repository) GetByOrgID(ctx context.Context, orgID int64) (*OrganizationEmailConfig, error) {
	query := `
		SELECT id, organization_id, smtp_enabled, smtp_host, smtp_port,
		       smtp_user, smtp_password, from_address, from_name,
		       created_at, updated_at
		FROM organization_email_configs
		WHERE organization_id = $1
	`

	cfg := &OrganizationEmailConfig{}
	var encryptedPassword string

	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&cfg.ID,
		&cfg.OrganizationID,
		&cfg.SMTPEnabled,
		&cfg.SMTPHost,
		&cfg.SMTPPort,
		&cfg.SMTPUser,
		&encryptedPassword,
		&cfg.FromAddress,
		&cfg.FromName,
		&cfg.CreatedAt,
		&cfg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get email config: %w", err)
	}

	// Decrypt password
	decrypted, err := secureid.DecryptID(encryptedPassword)
	if err != nil {
		// If decryption fails, the password may have been stored unencrypted (migration edge case)
		cfg.SMTPPassword = encryptedPassword
	} else {
		cfg.SMTPPassword = decrypted
	}

	return cfg, nil
}

// Upsert creates or updates email config for an organization.
func (r *repository) Upsert(ctx context.Context, orgID int64, req *UpsertEmailConfigRequest) (*OrganizationEmailConfig, error) {
	// Encrypt password before storing
	encryptedPassword, err := secureid.EncryptID(req.SMTPPassword)
	if err != nil {
		return nil, fmt.Errorf("encrypt password: %w", err)
	}

	enabled := true
	if req.SMTPEnabled != nil {
		enabled = *req.SMTPEnabled
	}

	query := `
		INSERT INTO organization_email_configs (
			organization_id, smtp_enabled, smtp_host, smtp_port,
			smtp_user, smtp_password, from_address, from_name,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (organization_id) DO UPDATE SET
			smtp_enabled = EXCLUDED.smtp_enabled,
			smtp_host = EXCLUDED.smtp_host,
			smtp_port = EXCLUDED.smtp_port,
			smtp_user = EXCLUDED.smtp_user,
			smtp_password = EXCLUDED.smtp_password,
			from_address = EXCLUDED.from_address,
			from_name = EXCLUDED.from_name,
			updated_at = NOW()
		RETURNING id, organization_id, smtp_enabled, smtp_host, smtp_port,
		          smtp_user, from_address, from_name, created_at, updated_at
	`

	cfg := &OrganizationEmailConfig{}
	err = r.db.QueryRow(ctx, query,
		orgID, enabled, req.SMTPHost, req.SMTPPort,
		req.SMTPUser, encryptedPassword, req.FromAddress, req.FromName,
	).Scan(
		&cfg.ID,
		&cfg.OrganizationID,
		&cfg.SMTPEnabled,
		&cfg.SMTPHost,
		&cfg.SMTPPort,
		&cfg.SMTPUser,
		&cfg.FromAddress,
		&cfg.FromName,
		&cfg.CreatedAt,
		&cfg.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert email config: %w", err)
	}

	cfg.SMTPPassword = req.SMTPPassword // Return plaintext (not encrypted)
	return cfg, nil
}
