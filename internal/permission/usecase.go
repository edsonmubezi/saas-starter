// internal/permission/usecase.go
package permission

import (
	"context"
	"fmt"

	"github.com/edsonmubezi/myapp/internal/identity"
	"github.com/edsonmubezi/myapp/internal/platform/audit"
	"github.com/edsonmubezi/myapp/pkg/pagination"
)

// PermissionUseCase defines the methods for permission-related business logic
type PermissionUseCase interface {
	// Read single (scope-aware)
	GetPermissionByID(ctx context.Context, id int64) (*Permission, error)

	// Grouped by resource
	// scope: "admin" (admin only), "tenant" (non-admin), "" (all)
	ListPermissionResourcesPaginated(
		ctx context.Context,
		pag pagination.Pagination,
		scope string,
	) (pagination.Result[PermissionResourceRow], error)
	ListPermissionResources(ctx context.Context, scope string) ([]PermissionResourceRow, error)

	// Update (description only - admin only)
	UpdatePermissionDescription(ctx context.Context, id int64, description string) (*Permission, error)
}

type PermissionUseCaseImpl struct {
	repo    PermissionRepository
	auditor *audit.Recorder
}

func NewPermissionUseCase(repo PermissionRepository, auditService *audit.PostgresService) PermissionUseCase {
	return &PermissionUseCaseImpl{
		repo:    repo,
		auditor: audit.NewRecorder(auditService, audit.ResourcePermission),
	}
}

// isHQCaller returns true if the caller is from HQ (orgID == 1)
func isHQCaller(ctx context.Context) (bool, error) {
	orgID, err := identity.OrgID(ctx)
	if err != nil {
		return false, fmt.Errorf("could not determine organization: %w", err)
	}
	return orgID == 1, nil
}

// GetPermissionByID retrieves a permission by ID
// Admin-only permissions are only visible to HQ callers
func (u *PermissionUseCaseImpl) GetPermissionByID(ctx context.Context, id int64) (*Permission, error) {
	p, err := u.repo.GetPermissionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Admin-only permissions require HQ caller
	if p.IsAdminOnly() {
		ok, err := isHQCaller(ctx)
		if err != nil || !ok {
			return nil, fmt.Errorf("permission not found")
		}
	}

	return p, nil
}

// ListPermissionResourcesPaginated returns permissions grouped by resource with pagination
// scope: "admin" (admin only), "tenant" (non-admin), "" (all)
func (u *PermissionUseCaseImpl) ListPermissionResourcesPaginated(
	ctx context.Context,
	pag pagination.Pagination,
	scope string,
) (pagination.Result[PermissionResourceRow], error) {
	return u.repo.ListPermissionResourcesPaginated(ctx, pag, scope)
}

// ListPermissionResources returns all permissions grouped by resource (non-paginated)
// scope: "admin" (admin only), "tenant" (non-admin), "" (all)
func (u *PermissionUseCaseImpl) ListPermissionResources(ctx context.Context, scope string) ([]PermissionResourceRow, error) {
	return u.repo.ListPermissionResources(ctx, scope)
}

// UpdatePermissionDescription updates a permission's description (admin only)
func (u *PermissionUseCaseImpl) UpdatePermissionDescription(ctx context.Context, id int64, description string) (*Permission, error) {
	ok, err := isHQCaller(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("forbidden")
	}

	// Ensure the permission exists & caller can see it
	before, err := u.GetPermissionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updated, err := u.repo.UpdatePermissionDescription(ctx, id, description)
	if err != nil {
		return nil, err
	}

	// Audit: Log permission update (high severity - security sensitive)
	u.auditor.RecordWithSeverity(ctx, audit.ActionUpdated, id, before.Name, audit.SeverityHigh, before, updated)

	return updated, nil
}
