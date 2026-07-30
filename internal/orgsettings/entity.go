package orgsettings

import "time"

// Organization type constants
const (
	OrgTypeSingleCompany   = "single_company"
	OrgTypeMultipleCompany = "multiple_company"
	OrgTypeMultiBranch     = "multi_branch"
	OrgTypeOutsourcing     = "outsourcing"
)

// OrganizationSettings represents core organization configuration
type OrganizationSettings struct {
	ID                        int64     `json:"id"`
	OrganizationID            int64     `json:"organizationId"`
	OrganizationType          string    `json:"organizationType" validate:"required,oneof=single_company multiple_company multi_branch outsourcing"`
	SessionLockTimeoutMinutes int       `json:"sessionLockTimeoutMinutes"`
	CreatedAt                 time.Time `json:"createdAt"`
	UpdatedAt                 time.Time `json:"updatedAt"`
}

// OrganizationSettingsCreateInput for creating new organization settings
type OrganizationSettingsCreateInput struct {
	OrganizationID   int64  `json:"organizationId" validate:"required"`
	OrganizationType string `json:"organizationType" validate:"required,oneof=single_company multiple_company multi_branch outsourcing"`
}

// OrganizationSettingsUpdateInput for updating organization settings
type OrganizationSettingsUpdateInput struct {
	ID                        int64   `json:"id" validate:"required"`
	OrganizationType          *string `json:"organizationType,omitempty" validate:"omitempty,oneof=single_company multiple_company multi_branch outsourcing"`
	SessionLockTimeoutMinutes *int    `json:"sessionLockTimeoutMinutes,omitempty"`
}
