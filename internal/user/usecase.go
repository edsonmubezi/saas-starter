package user

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmubezi/myapp/internal/identity"
	"github.com/edsonmubezi/myapp/internal/platform/audit"
	"github.com/edsonmubezi/myapp/pkg/cache"
	"github.com/edsonmubezi/myapp/pkg/pagination"
)

// UserUseCase defines the interface for user-related business logic
type UserUseCase interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByID(ctx context.Context, id int64, organizationID int64) (*User, error)
	GetUserByIDNoOrg(ctx context.Context, id int64) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)

	// User listing - separate methods for SuperAdmin and Org Admin
	GetAllUsersForAdmin(ctx context.Context, pg pagination.Pagination, organizationID int64, userType string) (pagination.Result[UserListItem], error)
	GetUserCounts(ctx context.Context) (map[string]int, error)
	GetUsersForOrganization(ctx context.Context, pg pagination.Pagination, userType string) (pagination.Result[UserListItem], error)

	SoftDeleteUser(ctx context.Context, id int64, organizationID int64) error
	UpdateUserStatus(ctx context.Context, id int64, organizationID int64) error
	ChangePassword(ctx context.Context, userID int64, organizationID int64, newPassword string) error
	UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error
	HardDeleteUser(ctx context.Context, id int64, organizationID int64) error
	UpdateUserPartial(ctx context.Context, input *UpdateUserInput) (*User, error)
}

// UserUseCaseImpl is the implementation of the UserUseCase interface
type UserUseCaseImpl struct {
	repo     UserRepository
	auditor  *audit.Recorder
}

// NewUserUseCase creates a new UserUseCaseImpl instance
func NewUserUseCase(repo UserRepository, auditService *audit.PostgresService) UserUseCase {
	return &UserUseCaseImpl{
		repo:    repo,
		auditor: audit.NewRecorder(auditService, audit.ResourceUser),
	}
}

// CreateUser creates a new user after checking email uniqueness
func (u *UserUseCaseImpl) CreateUser(ctx context.Context, user *User) (*User, error) {
	// Check if email already exists
	existingUser, _ := u.repo.GetUserByEmail(ctx, user.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("email is already taken")
	}

	user.CreatedAt = time.Now()
	orgID, err := identity.OrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not determine organization from token: %w", err)
	}

	if orgID == 1 {
		// HQ user → trust the payload (can be 1 or any other org ID)
		if user.OrganizationID == 0 {
			return nil, fmt.Errorf("organization_id must be provided for HQ users")
		}
	} else {
		// Non-HQ user → force org ID from context
		user.OrganizationID = orgID
	}

	user.ActiveStatus = true

	createdUser, err := u.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("could not create user: %v", err)
	}

	// Audit: Log user creation
	u.auditor.Record(ctx, audit.ActionCreated, createdUser.ID, createdUser.Email, nil, createdUser)

	return createdUser, nil
}

// GetUserByID retrieves a user by ID
func (u *UserUseCaseImpl) GetUserByID(ctx context.Context, id int64, organizationID int64) (*User, error) {
	user, err := u.repo.GetUserByID(ctx, id, organizationID)
	if err != nil {
		return nil, fmt.Errorf("could not find user: %v", err)
	}
	return user, nil
}

// GetUserByIDNoOrg retrieves a user by ID without organization verification
// This is used for password reset where we already validated the token
func (u *UserUseCaseImpl) GetUserByIDNoOrg(ctx context.Context, id int64) (*User, error) {
	user, err := u.repo.GetUserByIDNoOrg(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("could not find user: %v", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (u *UserUseCaseImpl) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := u.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("could not find user: %v", err)
	}
	return user, nil
}

// GetAllUsersForAdmin retrieves all users for SuperAdmin with optional organization filter.
// organizationID = 0 means all organizations, > 0 filters to specific organization.
// Used by SuperAdmin routes - no additional auth checks here (handler/middleware enforces admin access).
func (u *UserUseCaseImpl) GetAllUsersForAdmin(
	ctx context.Context,
	pg pagination.Pagination,
	organizationID int64,
	userType string,
) (pagination.Result[UserListItem], error) {
	result, err := u.repo.GetAllUsers(ctx, pg, organizationID, userType)
	if err != nil {
		return pagination.Result[UserListItem]{}, fmt.Errorf("could not fetch users for admin: %w", err)
	}
	return result, nil
}

func (u *UserUseCaseImpl) GetUserCounts(ctx context.Context) (map[string]int, error) {
	return u.repo.GetUserCounts(ctx)
}

// GetUsersForOrganization retrieves users within the caller's organization only.
// Organization ID is extracted from the JWT context at THIS level - not passed from handler.
// This ensures tenant isolation is enforced at the usecase layer.
func (u *UserUseCaseImpl) GetUsersForOrganization(
	ctx context.Context,
	pg pagination.Pagination,
	userType string,
) (pagination.Result[UserListItem], error) {
	// Extract organization ID from authenticated context
	orgID, err := identity.OrgID(ctx)
	if err != nil {
		return pagination.Result[UserListItem]{}, fmt.Errorf("could not determine organization: %w", err)
	}
	if orgID == 0 {
		return pagination.Result[UserListItem]{}, fmt.Errorf("organization context required for tenant queries")
	}

	// Call repository with mandatory org filter and optional user type filter
	result, err := u.repo.GetUsersByOrganization(ctx, pg, orgID, userType)
	if err != nil {
		return pagination.Result[UserListItem]{}, fmt.Errorf("could not fetch organization users: %w", err)
	}
	return result, nil
}

// SoftDeleteUser marks a user as deleted without removing them from the database
func (u *UserUseCaseImpl) SoftDeleteUser(ctx context.Context, id int64, organizationID int64) error {
	// Get user before deletion for audit
	user, _ := u.repo.GetUserByID(ctx, id, organizationID)

	if err := u.repo.SoftDeleteUser(ctx, id, organizationID); err != nil {
		return err
	}

	// Invalidate auth cache
	cache.Del(ctx, cache.AuthUserKey(id, organizationID))

	// Audit: Log user deletion
	if user != nil {
		u.auditor.Record(ctx, audit.ActionDeleted, id, user.Email, user, nil)
	}

	return nil
}

// HardDeleteUser permanently removes a user from the database
func (u *UserUseCaseImpl) HardDeleteUser(ctx context.Context, id int64, organizationID int64) error {
	// Get user before deletion for audit
	user, _ := u.repo.GetUserByID(ctx, id, organizationID)

	if err := u.repo.HardDeleteUser(ctx, id, organizationID); err != nil {
		return err
	}

	// Invalidate auth cache
	cache.Del(ctx, cache.AuthUserKey(id, organizationID))

	// Audit: Log permanent user deletion (high severity)
	if user != nil {
		u.auditor.RecordWithSeverity(ctx, audit.ActionDeleted, id, user.Email, audit.SeverityHigh, user, nil)
	}

	return nil
}

func (u *UserUseCaseImpl) UpdateUserStatus(ctx context.Context, id int64, organizationID int64) error {
	// Get user before update for audit
	before, _ := u.repo.GetUserByID(ctx, id, organizationID)

	if err := u.repo.UpdateUserStatus(ctx, id, organizationID); err != nil {
		return err
	}

	// Invalidate auth cache (status changed)
	cache.Del(ctx, cache.AuthUserKey(id, organizationID))

	// Get user after update
	after, _ := u.repo.GetUserByID(ctx, id, organizationID)

	// Audit: Log status change
	if before != nil {
		u.auditor.Record(ctx, audit.ActionUpdated, id, before.Email, before, after)
	}

	return nil
}

// func (u *UserUseCaseImpl) ChangePassword(ctx context.Context, id int64) error {
// 	return u.repo.ChangePassword(ctx, id)
// }

func (u *UserUseCaseImpl) ChangePassword(ctx context.Context, userID int64, organizationID int64, newPassword string) error {
	// Get user for audit
	user, _ := u.repo.GetUserByID(ctx, userID, organizationID)

	// validate, hash, and update via repo
	if err := u.repo.ChangePassword(ctx, userID, organizationID, newPassword); err != nil {
		return err
	}

	// Invalidate auth cache
	cache.Del(ctx, cache.AuthUserKey(userID, organizationID))

	// Audit: Log password change (don't include password in state)
	if user != nil {
		u.auditor.RecordWithSeverity(ctx, audit.ActionPasswordReset, userID, user.Email, audit.SeverityHigh, nil, nil)
	}

	return nil
}

// UpdatePassword updates a user's password without organization verification (for password reset)
func (u *UserUseCaseImpl) UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error {
	if err := u.repo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return err
	}
	// No orgID available here, invalidate by user pattern
	cache.DelByPattern(ctx, cache.AuthUserPattern(userID))
	return nil
}

func (u *UserUseCaseImpl) UpdateUserPartial(ctx context.Context, input *UpdateUserInput) (*User, error) {
	// Get user before update for audit (only if OrganizationID is provided)
	var before *User
	if input.OrganizationID != nil {
		before, _ = u.repo.GetUserByID(ctx, input.ID, *input.OrganizationID)
	}

	// validate, hash, and update via repo
	updated, err := u.repo.UpdateUserPartial(ctx, input)
	if err != nil {
		return nil, err
	}

	// Invalidate auth cache
	if updated != nil {
		cache.Del(ctx, cache.AuthUserKey(updated.ID, updated.OrganizationID))
	}

	// Audit: Log user update
	if before != nil {
		u.auditor.Record(ctx, audit.ActionUpdated, input.ID, before.Email, before, updated)
	}

	return updated, nil
}

