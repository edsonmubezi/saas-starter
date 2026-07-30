package identity

import (
	"context"
	"errors"
	"fmt"
)

var (
	// ErrOrganizationMismatch is returned when trying to access resources from another organization
	ErrOrganizationMismatch = errors.New("access denied: resource belongs to different organization")

	// ErrSuperAdminRestricted is returned when SuperAdmin cross-org access is disabled
	ErrSuperAdminRestricted = errors.New("access denied: SuperAdmin cross-organization access is disabled")
)

// MustMatchOrg ensures the authenticated user belongs to the same organization as the resource
// Returns error if organizations don't match (except for SuperAdmin with cross-org access enabled)
func MustMatchOrg(ctx context.Context, resourceOrgID int64) error {
	return EnforceOrgOwnership(ctx, resourceOrgID)
}

// MustBeSameOrgOrSuperAdmin allows access if user is from same org OR is SuperAdmin (if enabled)
func MustBeSameOrgOrSuperAdmin(ctx context.Context, resourceOrgID int64, allowSuperAdminCrossOrg bool) error {
	auth, err := Require(ctx)
	if err != nil {
		return err
	}

	// If same organization, allow access
	if auth.OrganizationID == resourceOrgID {
		return nil
	}

	// If SuperAdmin and cross-org access is enabled, allow
	if auth.Role == "SuperAdmin" && allowSuperAdminCrossOrg {
		return nil
	}

	// If SuperAdmin but cross-org access is disabled, return specific error
	if auth.Role == "SuperAdmin" && !allowSuperAdminCrossOrg {
		return ErrSuperAdminRestricted
	}

	return ErrOrganizationMismatch
}

// MustBeRole ensures the authenticated user has one of the specified roles
func MustBeRole(ctx context.Context, allowedRoles ...string) error {
	auth, err := Require(ctx)
	if err != nil {
		return err
	}

	for _, role := range allowedRoles {
		if auth.Role == role {
			return nil
		}
	}

	return fmt.Errorf("access denied: required role %v, got %s", allowedRoles, auth.Role)
}

// MustBeSuperAdmin ensures the authenticated user is a SuperAdmin
func MustBeSuperAdmin(ctx context.Context) error {
	return MustBeRole(ctx, "SuperAdmin")
}

// MustBeOrgsAdmin ensures the authenticated user is an OrgsAdmin
func MustBeOrgsAdmin(ctx context.Context) error {
	return MustBeRole(ctx, "OrgsAdmin")
}

// MustBeAdminOrSelf ensures the authenticated user is either an admin OR the resource owner
func MustBeAdminOrSelf(ctx context.Context, resourceUserID int64) error {
	auth, err := Require(ctx)
	if err != nil {
		return err
	}

	// Allow if user is admin
	if auth.Role == "SuperAdmin" || auth.Role == "OrgsAdmin" {
		return nil
	}

	// Allow if user is accessing their own data
	if auth.UserID == resourceUserID {
		return nil
	}

	return errors.New("access denied: must be admin or resource owner")
}

// CanAccessOrganization checks if the user can access data from a specific organization
func CanAccessOrganization(ctx context.Context, targetOrgID int64, allowSuperAdminCrossOrg bool) bool {
	auth, ok := FromContext(ctx)
	if !ok {
		return false
	}

	// Same organization always allowed
	if auth.OrganizationID == targetOrgID {
		return true
	}

	// SuperAdmin with cross-org access enabled
	if auth.Role == "SuperAdmin" && allowSuperAdminCrossOrg {
		return true
	}

	return false
}

// GetOrgIDOrFail returns the organization ID from context or error
func GetOrgIDOrFail(ctx context.Context) (int64, error) {
	auth, err := Require(ctx)
	if err != nil {
		return 0, err
	}

	if auth.OrganizationID == 0 {
		return 0, errors.New("organization ID not found in context")
	}

	return auth.OrganizationID, nil
}

// GetUserIDOrFail returns the user ID from context or error
func GetUserIDOrFail(ctx context.Context) (int64, error) {
	auth, err := Require(ctx)
	if err != nil {
		return 0, err
	}

	if auth.UserID == 0 {
		return 0, errors.New("user ID not found in context")
	}

	return auth.UserID, nil
}

