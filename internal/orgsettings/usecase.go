package orgsettings

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/edsonmubezi/myapp/pkg/cache"
)

// OrganizationSettingsUseCase defines business logic for organization settings
type OrganizationSettingsUseCase interface {
	CreateOrgSettings(ctx context.Context, input *OrganizationSettingsCreateInput) (*OrganizationSettings, error)
	UpdateOrgSettings(ctx context.Context, input *OrganizationSettingsUpdateInput) (*OrganizationSettings, error)
	GetOrgSettingsByID(ctx context.Context, id int64) (*OrganizationSettings, error)
	GetOrgSettingsByOrganizationID(ctx context.Context, organizationID int64) (*OrganizationSettings, error)
}

type orgSettingsUseCase struct {
	repo OrganizationSettingsRepository
}

// NewOrganizationSettingsUseCase creates a new organization settings use case
func NewOrganizationSettingsUseCase(repo OrganizationSettingsRepository) OrganizationSettingsUseCase {
	return &orgSettingsUseCase{repo: repo}
}

// CreateOrgSettings creates new organization settings
func (u *orgSettingsUseCase) CreateOrgSettings(ctx context.Context, input *OrganizationSettingsCreateInput) (*OrganizationSettings, error) {
	// Validate organization type
	if input.OrganizationType != OrgTypeSingleCompany &&
		input.OrganizationType != OrgTypeMultipleCompany &&
		input.OrganizationType != OrgTypeMultiBranch &&
		input.OrganizationType != OrgTypeOutsourcing {
		return nil, fmt.Errorf("invalid organization type: %s", input.OrganizationType)
	}

	// Check if settings already exist
	existing, err := u.repo.GetOrgSettingsByOrganizationID(ctx, input.OrganizationID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("organization settings already exist for this organization")
	}

	settings := &OrganizationSettings{
		OrganizationID:   input.OrganizationID,
		OrganizationType: input.OrganizationType,
	}

	if err := u.repo.CreateOrgSettings(ctx, settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// UpdateOrgSettings updates organization settings
func (u *orgSettingsUseCase) UpdateOrgSettings(ctx context.Context, input *OrganizationSettingsUpdateInput) (*OrganizationSettings, error) {
	// Get existing settings
	existing, err := u.repo.GetOrgSettingsByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("organization settings not found")
	}

	// Update fields if provided
	if input.OrganizationType != nil {
		// Validate organization type
		if *input.OrganizationType != OrgTypeSingleCompany &&
			*input.OrganizationType != OrgTypeMultipleCompany &&
			*input.OrganizationType != OrgTypeMultiBranch &&
			*input.OrganizationType != OrgTypeOutsourcing {
			return nil, fmt.Errorf("invalid organization type: %s", *input.OrganizationType)
		}
		existing.OrganizationType = *input.OrganizationType
	}

	if input.SessionLockTimeoutMinutes != nil {
		existing.SessionLockTimeoutMinutes = *input.SessionLockTimeoutMinutes
	}

	if err := u.repo.UpdateOrgSettings(ctx, existing); err != nil {
		return nil, err
	}

	cache.Del(ctx, cache.OrgSettingsKey(existing.OrganizationID))

	return existing, nil
}

// GetOrgSettingsByID retrieves organization settings by ID
func (u *orgSettingsUseCase) GetOrgSettingsByID(ctx context.Context, id int64) (*OrganizationSettings, error) {
	settings, err := u.repo.GetOrgSettingsByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, errors.New("organization settings not found")
	}
	return settings, nil
}

// GetOrgSettingsByOrganizationID retrieves organization settings by organization ID
func (u *orgSettingsUseCase) GetOrgSettingsByOrganizationID(ctx context.Context, organizationID int64) (*OrganizationSettings, error) {
	cacheKey := cache.OrgSettingsKey(organizationID)
	if cached, err := cache.Get[OrganizationSettings](ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	settings, err := u.repo.GetOrgSettingsByOrganizationID(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, errors.New("organization settings not found")
	}

	cache.Set(ctx, cacheKey, settings, 1*time.Hour)
	return settings, nil
}
