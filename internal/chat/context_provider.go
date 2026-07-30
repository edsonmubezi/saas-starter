package chat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/edsonmubezi/myapp/internal/orgsettings"
	"github.com/edsonmubezi/myapp/pkg/cache"
)

// OrgContextProvider fetches org-specific data and formats it as text for prompt injection.
type OrgContextProvider struct {
	orgSetRepo orgsettings.OrganizationSettingsRepository
}

// NewOrgContextProvider creates a new OrgContextProvider with the org settings repository.
func NewOrgContextProvider(
	orgSetRepo orgsettings.OrganizationSettingsRepository,
) *OrgContextProvider {
	return &OrgContextProvider{
		orgSetRepo: orgSetRepo,
	}
}

// GetOrgContext returns a formatted string describing the org's configuration.
// Results are cached in Redis for 15 minutes.
func (p *OrgContextProvider) GetOrgContext(ctx context.Context, orgID int64) string {
	cacheKey := fmt.Sprintf("chat:org_context:%d", orgID)
	if cached, err := cache.Get[string](ctx, cacheKey); err == nil && cached != nil {
		return *cached
	}

	var sections []string

	// Organization settings
	if settings, err := p.orgSetRepo.GetOrgSettingsByOrganizationID(ctx, orgID); err == nil && settings != nil {
		sections = append(sections, fmt.Sprintf("Organization Type: %s",
			settings.OrganizationType))
	}

	result := ""
	if len(sections) > 0 {
		result = "\n\n## Your Organization's Current Configuration\n" + strings.Join(sections, "\n")
	}

	cache.Set(ctx, cacheKey, &result, 15*time.Minute)
	return result
}
