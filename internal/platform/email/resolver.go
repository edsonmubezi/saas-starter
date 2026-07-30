package email

import (
	"context"
	"log"

	"github.com/edsonmubezi/myapp/internal/emailconfig"
)

// Resolver provides org-specific email services with a global fallback.
// It checks the organization_email_configs table first; if no config is found
// (or it's disabled), it falls back to the global service from .env.
type Resolver struct {
	globalSvc    EmailService
	emailCfgRepo emailconfig.Repository
}

// NewResolver creates a resolver that checks org DB config before falling back to the global service.
func NewResolver(globalSvc EmailService, emailCfgRepo emailconfig.Repository) *Resolver {
	return &Resolver{
		globalSvc:    globalSvc,
		emailCfgRepo: emailCfgRepo,
	}
}

// ForOrg returns an org-specific SMTP service if configured and enabled,
// otherwise returns the global fallback.
func (r *Resolver) ForOrg(ctx context.Context, orgID int64) EmailService {
	if orgID == 0 {
		return r.globalSvc
	}

	orgCfg, err := r.emailCfgRepo.GetByOrgID(ctx, orgID)
	if err != nil || orgCfg == nil || !orgCfg.SMTPEnabled {
		return r.globalSvc
	}

	svc, err := NewSMTPEmailService(EmailConfig{
		Enabled:      true,
		SMTPHost:     orgCfg.SMTPHost,
		SMTPPort:     orgCfg.SMTPPort,
		SMTPUser:     orgCfg.SMTPUser,
		SMTPPassword: orgCfg.SMTPPassword,
		FromAddress:  orgCfg.FromAddress,
		FromName:     orgCfg.FromName,
	})
	if err != nil {
		log.Printf("Failed to create org SMTP service for org %d, using global: %v", orgID, err)
		return r.globalSvc
	}
	return svc
}

// Global returns the global fallback email service directly.
func (r *Resolver) Global() EmailService {
	return r.globalSvc
}
