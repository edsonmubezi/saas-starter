package organization

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmubezi/myapp/pkg/pagination"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type OrganizationRepository interface {
	CreateOrganization(ctx context.Context, Organization *Organization) (*Organization, error)
	GetOrganizationByID(ctx context.Context, id int64) (*Organization, error)
	UpdateOrganization(ctx context.Context, org *Organization) (*Organization, error)
	GetAllOrganizations(ctx context.Context) ([]Organization, error)
	SoftDeleteOrganization(ctx context.Context, id int64) error
	GetOrganizations(ctx context.Context, pag pagination.Pagination) (pagination.Result[Organization], error)
	CreateOrganizationTx(ctx context.Context, tx pgx.Tx, org *Organization) (*Organization, error)
	WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error
	GetDocumentBranding(ctx context.Context, orgID int64) (*DocumentBrandingSettings, error)
	UpsertDocumentBranding(ctx context.Context, settings *DocumentBrandingSettings) error
	GetEmailBranding(ctx context.Context, orgID int64) (*EmailBrandingSettings, error)
	UpsertEmailBranding(ctx context.Context, settings *EmailBrandingSettings) error
}

type PostgresOrganizationRepository struct {
	db *pgxpool.Pool
}

func NewPostgresOrganizationRepository(db *pgxpool.Pool) *PostgresOrganizationRepository {
	return &PostgresOrganizationRepository{db: db}
}

func (r *PostgresOrganizationRepository) CreateOrganization(ctx context.Context, org *Organization) (*Organization, error) {
	// now := time.Now()

	query := `
		INSERT INTO organizations (
			name, phone_number, address, contact_person, logo_url, email, tin, registration_number)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, phone_number, address, contact_person, logo_url, email, tin, registration_number, created_at, updated_at;`

	err := r.db.QueryRow(ctx, query,
		org.Name,
		org.PhoneNumber,
		org.Address,
		org.ContactPerson,
		org.LogoURL,
		org.Email,
		org.TIN,
		org.RegistrationNumber,
	).Scan(
		&org.ID,
		&org.Name,
		&org.PhoneNumber,
		&org.Address,
		&org.ContactPerson,
		&org.LogoURL,
		&org.Email,
		&org.TIN,
		&org.RegistrationNumber,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return org, nil
}

func (r *PostgresOrganizationRepository) GetOrganizationByID(ctx context.Context, id int64) (*Organization, error) {
	sql := `SELECT id, name, phone_number, address, contact_person,
	               logo_url, email, tin, registration_number,
	               created_at, updated_at
	        FROM organizations
	        WHERE id = $1`

	row := r.db.QueryRow(ctx, sql, id)

	var organization Organization
	err := row.Scan(
		&organization.ID,
		&organization.Name,
		&organization.PhoneNumber,
		&organization.Address,
		&organization.ContactPerson,
		&organization.LogoURL,
		&organization.Email,
		&organization.TIN,
		&organization.RegistrationNumber,
		&organization.CreatedAt,
		&organization.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("could not fetch user: %v", err)
	}

	return &organization, nil
}

func (r *PostgresOrganizationRepository) GetAllOrganizations(ctx context.Context) ([]Organization, error) {
	query := `
		SELECT o.id, o.name, o.phone_number, o.address, o.contact_person,
		       o.logo_url, o.email, o.tin, o.registration_number,
		       o.created_at, o.updated_at
		FROM organizations o

	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var organizations []Organization
	for rows.Next() {
		var org Organization
		if err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.PhoneNumber,
			&org.Address,
			&org.ContactPerson,
			&org.LogoURL,
			&org.Email,
			&org.TIN,
			&org.RegistrationNumber,
			&org.CreatedAt,
			&org.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		organizations = append(organizations, org)
	}

	return organizations, nil
}

func (r *PostgresOrganizationRepository) UpdateOrganization(ctx context.Context, org *Organization) (*Organization, error) {
	query := `
		UPDATE organizations
		SET name = $1, phone_number = $2, address = $3, contact_person = $4,
		    logo_url = $5, email = $6, tin = $7, registration_number = $8, updated_at = $9
		WHERE id = $10
		RETURNING id, name, phone_number, address, contact_person, logo_url, email, tin, registration_number, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query,
		org.Name,
		org.PhoneNumber,
		org.Address,
		org.ContactPerson,
		org.LogoURL,
		org.Email,
		org.TIN,
		org.RegistrationNumber,
		org.UpdatedAt,
		org.ID,
	)

	var updated Organization
	err := row.Scan(
		&updated.ID,
		&updated.Name,
		&updated.PhoneNumber,
		&updated.Address,
		&updated.ContactPerson,
		&updated.LogoURL,
		&updated.Email,
		&updated.TIN,
		&updated.RegistrationNumber,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *PostgresOrganizationRepository) SoftDeleteOrganization(ctx context.Context, id int64) error {
	sql := `UPDATE users SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, sql, time.Now(), id)
	if err != nil {
		return fmt.Errorf("could not soft delete user: %v", err)
	}
	return nil
}

func (r *PostgresOrganizationRepository) GetOrganizations(
	ctx context.Context,
	pag pagination.Pagination,
) (pagination.Result[Organization], error) {

	opts := pagination.Options[Organization]{
		Table: `(
			SELECT id, name, phone_number, address, contact_person,
			       logo_url, email, tin, registration_number,
			       created_at, updated_at
			FROM organizations

		) AS organizations`,

		Fields:     []string{"id", "name", "phone_number", "address", "contact_person", "logo_url", "email", "tin", "registration_number", "created_at", "updated_at"},
		SearchCols: []string{"name", "phone_number", "address", "email"},
		Pagination: pag,
		SortBy:     "id",

		ScanFunc: func(row pgx.Row) (Organization, error) {
			var o Organization
			err := row.Scan(
				&o.ID,
				&o.Name,
				&o.PhoneNumber,
				&o.Address,
				&o.ContactPerson,
				&o.LogoURL,
				&o.Email,
				&o.TIN,
				&o.RegistrationNumber,
				&o.CreatedAt,
				&o.UpdatedAt,
			)
			return o, err
		},
	}

	return pagination.QueryWithPagination(ctx, r.db, opts)
}
func (r *PostgresOrganizationRepository) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *PostgresOrganizationRepository) CreateOrganizationTx(ctx context.Context, tx pgx.Tx, org *Organization) (*Organization, error) {
	const q = `
		INSERT INTO organizations (name, phone_number, address, contact_person, logo_url, email, tin, registration_number)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, phone_number, address, contact_person, logo_url, email, tin, registration_number, created_at, updated_at;
	`
	err := tx.QueryRow(ctx, q,
		org.Name,
		org.PhoneNumber,
		org.Address,
		org.ContactPerson,
		org.LogoURL,
		org.Email,
		org.TIN,
		org.RegistrationNumber,
	).Scan(
		&org.ID,
		&org.Name,
		&org.PhoneNumber,
		&org.Address,
		&org.ContactPerson,
		&org.LogoURL,
		&org.Email,
		&org.TIN,
		&org.RegistrationNumber,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert organization: %w", err)
	}
	return org, nil
}

func (r *PostgresOrganizationRepository) GetDocumentBranding(ctx context.Context, orgID int64) (*DocumentBrandingSettings, error) {
	query := `
		SELECT id, organization_id,
		       show_logo, show_org_name, show_address, show_contact, show_tin, show_reg_number,
		       show_footer, footer_text, show_page_numbers, show_generated_date,
		       show_watermark, watermark_text,
		       COALESCE(watermark_type, 'text'), COALESCE(watermark_image_path, ''),
		       primary_color, header_text_color, font_family,
		       COALESCE(header_org_name, ''), COALESCE(header_address, ''),
		       COALESCE(header_phone, ''), COALESCE(header_email, ''),
		       COALESCE(header_tin, ''), COALESCE(footer_org_name, '')
		FROM document_branding_settings
		WHERE organization_id = $1`

	var s DocumentBrandingSettings
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&s.ID, &s.OrganizationID,
		&s.ShowLogo, &s.ShowOrgName, &s.ShowAddress, &s.ShowContact, &s.ShowTIN, &s.ShowRegNumber,
		&s.ShowFooter, &s.FooterText, &s.ShowPageNumbers, &s.ShowGeneratedDate,
		&s.ShowWatermark, &s.WatermarkText,
		&s.WatermarkType, &s.WatermarkImagePath,
		&s.PrimaryColor, &s.HeaderTextColor, &s.FontFamily,
		&s.HeaderOrgName, &s.HeaderAddress,
		&s.HeaderPhone, &s.HeaderEmail,
		&s.HeaderTIN, &s.FooterOrgName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return DefaultDocumentBrandingSettings(orgID), nil
		}
		return nil, fmt.Errorf("get document branding: %w", err)
	}
	return &s, nil
}

func (r *PostgresOrganizationRepository) UpsertDocumentBranding(ctx context.Context, settings *DocumentBrandingSettings) error {
	query := `
		INSERT INTO document_branding_settings (
			organization_id,
			show_logo, show_org_name, show_address, show_contact, show_tin, show_reg_number,
			show_footer, footer_text, show_page_numbers, show_generated_date,
			show_watermark, watermark_text, watermark_type, watermark_image_path,
			primary_color, header_text_color, font_family,
			header_org_name, header_address, header_phone, header_email, header_tin, footer_org_name,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, NOW())
		ON CONFLICT (organization_id) DO UPDATE SET
			show_logo = EXCLUDED.show_logo,
			show_org_name = EXCLUDED.show_org_name,
			show_address = EXCLUDED.show_address,
			show_contact = EXCLUDED.show_contact,
			show_tin = EXCLUDED.show_tin,
			show_reg_number = EXCLUDED.show_reg_number,
			show_footer = EXCLUDED.show_footer,
			footer_text = EXCLUDED.footer_text,
			show_page_numbers = EXCLUDED.show_page_numbers,
			show_generated_date = EXCLUDED.show_generated_date,
			show_watermark = EXCLUDED.show_watermark,
			watermark_text = EXCLUDED.watermark_text,
			watermark_type = EXCLUDED.watermark_type,
			watermark_image_path = EXCLUDED.watermark_image_path,
			primary_color = EXCLUDED.primary_color,
			header_text_color = EXCLUDED.header_text_color,
			font_family = EXCLUDED.font_family,
			header_org_name = EXCLUDED.header_org_name,
			header_address = EXCLUDED.header_address,
			header_phone = EXCLUDED.header_phone,
			header_email = EXCLUDED.header_email,
			header_tin = EXCLUDED.header_tin,
			footer_org_name = EXCLUDED.footer_org_name,
			updated_at = NOW()`

	_, err := r.db.Exec(ctx, query,
		settings.OrganizationID,
		settings.ShowLogo, settings.ShowOrgName, settings.ShowAddress, settings.ShowContact, settings.ShowTIN, settings.ShowRegNumber,
		settings.ShowFooter, settings.FooterText, settings.ShowPageNumbers, settings.ShowGeneratedDate,
		settings.ShowWatermark, settings.WatermarkText, settings.WatermarkType, settings.WatermarkImagePath,
		settings.PrimaryColor, settings.HeaderTextColor, settings.FontFamily,
		settings.HeaderOrgName, settings.HeaderAddress, settings.HeaderPhone, settings.HeaderEmail, settings.HeaderTIN, settings.FooterOrgName,
	)
	if err != nil {
		return fmt.Errorf("upsert document branding: %w", err)
	}
	return nil
}

func (r *PostgresOrganizationRepository) GetEmailBranding(ctx context.Context, orgID int64) (*EmailBrandingSettings, error) {
	query := `
		SELECT id, organization_id,
		       primary_color, header_text_color, accent_color,
		       font_family, show_logo, COALESCE(footer_text, ''), COALESCE(sign_off, '')
		FROM email_branding_settings
		WHERE organization_id = $1`

	var s EmailBrandingSettings
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&s.ID, &s.OrganizationID,
		&s.PrimaryColor, &s.HeaderTextColor, &s.AccentColor,
		&s.FontFamily, &s.ShowLogo, &s.FooterText, &s.SignOff,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return DefaultEmailBrandingSettings(orgID), nil
		}
		return nil, fmt.Errorf("get email branding: %w", err)
	}
	return &s, nil
}

func (r *PostgresOrganizationRepository) UpsertEmailBranding(ctx context.Context, settings *EmailBrandingSettings) error {
	query := `
		INSERT INTO email_branding_settings (
			organization_id, primary_color, header_text_color, accent_color,
			font_family, show_logo, footer_text, sign_off, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (organization_id) DO UPDATE SET
			primary_color = EXCLUDED.primary_color,
			header_text_color = EXCLUDED.header_text_color,
			accent_color = EXCLUDED.accent_color,
			font_family = EXCLUDED.font_family,
			show_logo = EXCLUDED.show_logo,
			footer_text = EXCLUDED.footer_text,
			sign_off = EXCLUDED.sign_off,
			updated_at = NOW()`

	_, err := r.db.Exec(ctx, query,
		settings.OrganizationID,
		settings.PrimaryColor, settings.HeaderTextColor, settings.AccentColor,
		settings.FontFamily, settings.ShowLogo, settings.FooterText, settings.SignOff,
	)
	if err != nil {
		return fmt.Errorf("upsert email branding: %w", err)
	}
	return nil
}
