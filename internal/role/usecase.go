// internal/role/usecase.go
package role

import (
	"context"
	"fmt"

	"github.com/edsonmubezi/myapp/internal/identity"
	"github.com/edsonmubezi/myapp/internal/platform/audit"
	"github.com/edsonmubezi/myapp/pkg/cache"
	"github.com/edsonmubezi/myapp/pkg/pagination"
)

// RoleUseCase defines the methods for role-related business logic
type RoleUseCase interface {
	CreateRole(ctx context.Context, name string) (*Role, error)
	GetRoleByID(ctx context.Context, id int64) (*Role, error)

	// Simple list (no pagination)
	GetAllRoles(ctx context.Context) ([]Role, error)
	GetSystemRoles(ctx context.Context) ([]Role, error)
	GetAllRolesAdmin(ctx context.Context) ([]Role, error) // all roles including system

	// Paginated list (org-scoped) — mirrors your payroll use case style
	GetAllRolesPaginated(ctx context.Context, pag pagination.Pagination) (pagination.Result[Role], error)
	GetAllRolesAdminPaginated(ctx context.Context, pag pagination.Pagination) (pagination.Result[Role], error)

	UpdateRole(ctx context.Context, id int64, name string) (*Role, error)
	DeleteRole(ctx context.Context, id int64) error
	AssignPermissionsToRole(ctx context.Context, roleID int64, permissionIDs []int64) error

	GetRoleWithPermissions(ctx context.Context, roleID int64) (*RoleWithPermissions, error)
}

type RoleUseCaseImpl struct {
	repo    RoleRepository
	auditor *audit.Recorder
}

func NewRoleUseCase(repo RoleRepository, auditService *audit.PostgresService) RoleUseCase {
	return &RoleUseCaseImpl{
		repo:    repo,
		auditor: audit.NewRecorder(auditService, audit.ResourceRole),
	}
}

// helper: extract orgID from context, bubble up errors
func orgIDFromCtx(ctx context.Context) (int64, error) {
	orgID, err := identity.OrgID(ctx)
	if err != nil {
		return 0, fmt.Errorf("could not determine organization: %w", err)
	}
	return orgID, nil
}

func (u *RoleUseCaseImpl) CreateRole(ctx context.Context, name string) (*Role, error) {
	orgID, err := orgIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	created, err := u.repo.CreateRole(ctx, orgID, name)
	if err != nil {
		return nil, err
	}

	// Audit: Log role creation
	u.auditor.Record(ctx, audit.ActionCreated, created.ID, created.Name, nil, created)

	return created, nil
}

func (u *RoleUseCaseImpl) GetRoleByID(ctx context.Context, id int64) (*Role, error) {
	orgID, err := orgIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	role, err := u.repo.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Enforce org ownership
	if role.OrganizationID != orgID {
		return nil, fmt.Errorf("role not found")
	}
	return role, nil
}

// -------- Lists --------

// Simple, non-paginated, org-scoped
func (u *RoleUseCaseImpl) GetAllRoles(ctx context.Context) ([]Role, error) {
	orgID, err := orgIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	return u.repo.GetRolesByOrganizationID(ctx, orgID)
}

// System roles (non-assignable: SuperAdmin, OrgsAdmin)
func (u *RoleUseCaseImpl) GetSystemRoles(ctx context.Context) ([]Role, error) {
	orgID, err := orgIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	return u.repo.GetSystemRolesByOrganizationID(ctx, orgID)
}

// All roles including system (for SuperAdmin context)
func (u *RoleUseCaseImpl) GetAllRolesAdmin(ctx context.Context) ([]Role, error) {
	orgID, err := orgIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	return u.repo.GetAllOrgRoles(ctx, orgID)
}

// All roles paginated including system (for SuperAdmin context)
func (u *RoleUseCaseImpl) GetAllRolesAdminPaginated(ctx context.Context, pag pagination.Pagination) (pagination.Result[Role], error) {
	orgID, err := orgIDFromCtx(ctx)
	if err != nil {
		return pagination.Result[Role]{}, err
	}
	return u.repo.GetAllOrgRolesPaginated(ctx, orgID, pag)
}

// Paginated, org-scoped (same style as your payroll use case)
func (u *RoleUseCaseImpl) GetAllRolesPaginated(ctx context.Context, pag pagination.Pagination) (pagination.Result[Role], error) {
	orgID, err := orgIDFromCtx(ctx)
	if err != nil {
		return pagination.Result[Role]{}, err
	}
	return u.repo.GetRolesByOrganizationIDPaginated(ctx, orgID, pag)
}

// -------- Mutations --------

func (u *RoleUseCaseImpl) UpdateRole(ctx context.Context, id int64, name string) (*Role, error) {
	orgID, err := orgIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	// Ensure the role belongs to this org before updating
	role, err := u.repo.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role.OrganizationID != orgID {
		return nil, fmt.Errorf("role not found")
	}

	// Capture before state for audit
	before := *role

	updated, err := u.repo.UpdateRole(ctx, id, name)
	if err != nil {
		return nil, err
	}

	// Invalidate auth cache for all users in this org (role name changed)
	cache.DelByPattern(ctx, cache.AuthUserOrgPattern(orgID))

	// Audit: Log role update
	u.auditor.Record(ctx, audit.ActionUpdated, id, updated.Name, &before, updated)

	return updated, nil
}

func (u *RoleUseCaseImpl) DeleteRole(ctx context.Context, id int64) error {
	orgID, err := orgIDFromCtx(ctx)
	if err != nil {
		return err
	}

	// Ensure the role belongs to this org before deleting
	role, err := u.repo.GetRoleByID(ctx, id)
	if err != nil {
		return err
	}
	if role.OrganizationID != orgID {
		return fmt.Errorf("role not found")
	}

	if err := u.repo.DeleteRole(ctx, id); err != nil {
		return err
	}

	// Invalidate auth cache for all users in this org
	cache.DelByPattern(ctx, cache.AuthUserOrgPattern(orgID))

	// Audit: Log role deletion
	u.auditor.Record(ctx, audit.ActionDeleted, id, role.Name, role, nil)

	return nil
}

func (u *RoleUseCaseImpl) AssignPermissionsToRole(ctx context.Context, roleID int64, permissionIDs []int64) error {
	orgID, err := orgIDFromCtx(ctx)
	if err != nil {
		return err
	}

	// Ensure the role belongs to this org before assigning permissions
	role, err := u.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role.OrganizationID != orgID {
		return fmt.Errorf("role not found")
	}

	// Get current permissions for audit trail
	beforePerms, _ := u.repo.GetRoleWithPermissions(ctx, roleID)

	if err := u.repo.AssignPermissionsToRole(ctx, roleID, permissionIDs); err != nil {
		return err
	}

	// Invalidate auth cache for all users in this org (permissions changed)
	cache.DelByPattern(ctx, cache.AuthUserOrgPattern(orgID))

	// Get updated permissions for audit trail
	afterPerms, _ := u.repo.GetRoleWithPermissions(ctx, roleID)

	// Audit: Log permission assignment (high severity)
	u.auditor.RecordWithSeverity(ctx, audit.ActionPermGranted, roleID, role.Name, audit.SeverityHigh, beforePerms, afterPerms)

	return nil
}

// GetRoleWithPermissions returns role + its permissions (after verifying org ownership).
func (u *RoleUseCaseImpl) GetRoleWithPermissions(ctx context.Context, roleID int64) (*RoleWithPermissions, error) {
	orgID, err := orgIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	// Fetch role+perms from repo, then enforce org ownership.
	data, err := u.repo.GetRoleWithPermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if data.OrganizationID != orgID {
		return nil, fmt.Errorf("role not found")
	}
	return data, nil
}
