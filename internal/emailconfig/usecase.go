package emailconfig

import (
	"context"
	"fmt"
)

// UseCase defines the interface for email config business logic.
type UseCase interface {
	GetEmailConfig(ctx context.Context, orgID int64) (*OrganizationEmailConfig, error)
	UpsertEmailConfig(ctx context.Context, orgID int64, req *UpsertEmailConfigRequest) (*OrganizationEmailConfig, error)
}

type useCase struct {
	repo Repository
}

// NewUseCase creates a new email config use case.
func NewUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}

func (u *useCase) GetEmailConfig(ctx context.Context, orgID int64) (*OrganizationEmailConfig, error) {
	return u.repo.GetByOrgID(ctx, orgID)
}

func (u *useCase) UpsertEmailConfig(ctx context.Context, orgID int64, req *UpsertEmailConfigRequest) (*OrganizationEmailConfig, error) {
	if req.SMTPHost == "" {
		return nil, fmt.Errorf("SMTP host is required")
	}
	if req.SMTPPort <= 0 || req.SMTPPort > 65535 {
		return nil, fmt.Errorf("SMTP port must be between 1 and 65535")
	}
	if req.SMTPUser == "" {
		return nil, fmt.Errorf("SMTP user is required")
	}
	if req.SMTPPassword == "" {
		return nil, fmt.Errorf("SMTP password is required")
	}
	if req.FromAddress == "" {
		return nil, fmt.Errorf("from address is required")
	}

	return u.repo.Upsert(ctx, orgID, req)
}
