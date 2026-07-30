package organization

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/edsonmubezi/myapp/internal/orgsettings"
	"github.com/edsonmubezi/myapp/internal/platform/audit"
	"github.com/edsonmubezi/myapp/internal/seeder"
	"github.com/edsonmubezi/myapp/pkg/pagination"
	"github.com/jackc/pgx/v4"
)

type OrganizationUseCase interface {
	CreateOrganization(ctx context.Context, organization *Organization) (*Organization, error)
	UpdateOrganization(ctx context.Context, organization *Organization) (*Organization, error)
	GetOrganizationByID(ctx context.Context, id int64) (*Organization, error)
	GetAllOrganizations(ctx context.Context) ([]Organization, error)
	SoftDeleteOrganization(ctx context.Context, id int64) error
	GetOrganizations(ctx context.Context, pag pagination.Pagination) (pagination.Result[Organization], error)
	GetDocumentBranding(ctx context.Context, orgID int64) (*DocumentBrandingSettings, error)
	SaveDocumentBranding(ctx context.Context, settings *DocumentBrandingSettings) error
	GetEmailBranding(ctx context.Context, orgID int64) (*EmailBrandingSettings, error)
	SaveEmailBranding(ctx context.Context, settings *EmailBrandingSettings) error
}

type OrganizationImpl struct {
	repo            OrganizationRepository
	coreOrgSettings orgsettings.OrganizationSettingsRepository
	orgSeeder       seeder.OrgDefaultsSeeder
	auditor         *audit.Recorder
}

func NewOrganizationUseCase(
	repo OrganizationRepository,
	coreOrgSettingsRepo orgsettings.OrganizationSettingsRepository,
	auditService *audit.PostgresService,
) OrganizationUseCase {
	return &OrganizationImpl{
		repo:            repo,
		coreOrgSettings: coreOrgSettingsRepo,
		orgSeeder:       nil, // Optional - will be set via SetOrgSeeder
		auditor:         audit.NewRecorder(auditService, audit.ResourceOrganization),
	}
}

// SetOrgSeeder sets the organization defaults seeder (optional dependency)
func (u *OrganizationImpl) SetOrgSeeder(s seeder.OrgDefaultsSeeder) {
	u.orgSeeder = s
}

func (u *OrganizationImpl) CreateOrganization(ctx context.Context, organization *Organization) (*Organization, error) {
	now := time.Now()
	organization.CreatedAt = now
	organization.UpdatedAt = now

	var created *Organization

	err := u.repo.WithTransaction(ctx, func(tx pgx.Tx) error {
		// 1) Create organization
		org, err := u.repo.CreateOrganizationTx(ctx, tx, organization)
		if err != nil {
			return fmt.Errorf("create organization: %w", err)
		}
		created = org

		// Only create default settings for organizations with ID >= 2 (exclude HQ)
		if org.ID >= 2 {
			// 2) Default core organization settings
			coreSettings := &orgsettings.OrganizationSettings{
				OrganizationID:   org.ID,
				OrganizationType: orgsettings.OrgTypeSingleCompany, // default to single company
			}
			if err := u.coreOrgSettings.CreateOrgSettingsTx(ctx, tx, coreSettings); err != nil {
				return fmt.Errorf("create default core organization settings: %w", err)
			}

			// 3) Seed organization defaults (departments, designations, positions, etc.)
			if u.orgSeeder != nil {
				if err := u.orgSeeder.SeedOrgDefaults(ctx, tx, org.ID); err != nil {
					log.Printf("Warning: failed to seed org defaults for org %d: %v", org.ID, err)
					// Continue without failing - org defaults are optional
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Audit: Log organization creation (system-level, high severity)
	u.auditor.RecordWithSeverity(ctx, audit.ActionCreated, created.ID, created.Name, audit.SeverityHigh, nil, created)

	return created, nil
}

func (u *OrganizationImpl) GetOrganizationByID(ctx context.Context, id int64) (*Organization, error) {
	org, err := u.repo.GetOrganizationByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("could not find organization: %v", err)
	}
	return org, nil
}

func (u *OrganizationImpl) GetAllOrganizations(ctx context.Context) ([]Organization, error) {
	orgs, err := u.repo.GetAllOrganizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all organizations: %w", err)
	}
	return orgs, nil
}

func (u *OrganizationImpl) SoftDeleteOrganization(ctx context.Context, id int64) error {
	// Get organization before deletion for audit
	org, _ := u.repo.GetOrganizationByID(ctx, id)

	if err := u.repo.SoftDeleteOrganization(ctx, id); err != nil {
		return err
	}

	// Audit: Log organization deletion (critical severity)
	if org != nil {
		u.auditor.RecordWithSeverity(ctx, audit.ActionDeleted, id, org.Name, audit.SeverityCritical, org, nil)
	}

	return nil
}

func (u *OrganizationImpl) UpdateOrganization(ctx context.Context, org *Organization) (*Organization, error) {
	// Get organization before update for audit
	before, _ := u.repo.GetOrganizationByID(ctx, org.ID)

	org.UpdatedAt = time.Now()

	updatedOrg, err := u.repo.UpdateOrganization(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("could not update organization: %v", err)
	}

	// Audit: Log organization update
	u.auditor.Record(ctx, audit.ActionUpdated, updatedOrg.ID, updatedOrg.Name, before, updatedOrg)

	return updatedOrg, nil
}

func (u *OrganizationImpl) GetOrganizations(ctx context.Context, pag pagination.Pagination) (pagination.Result[Organization], error) {

	return u.repo.GetOrganizations(ctx, pag)
}

func (u *OrganizationImpl) GetDocumentBranding(ctx context.Context, orgID int64) (*DocumentBrandingSettings, error) {
	return u.repo.GetDocumentBranding(ctx, orgID)
}

func (u *OrganizationImpl) SaveDocumentBranding(ctx context.Context, settings *DocumentBrandingSettings) error {
	return u.repo.UpsertDocumentBranding(ctx, settings)
}

func (u *OrganizationImpl) GetEmailBranding(ctx context.Context, orgID int64) (*EmailBrandingSettings, error) {
	return u.repo.GetEmailBranding(ctx, orgID)
}

func (u *OrganizationImpl) SaveEmailBranding(ctx context.Context, settings *EmailBrandingSettings) error {
	return u.repo.UpsertEmailBranding(ctx, settings)
}
