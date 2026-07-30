// internal/role/entity.go
package role

import (
	"github.com/edsonmubezi/myapp/internal/permission"

	"github.com/go-playground/validator/v10"
)

type Role struct {
	ID             int64                   `json:"id" secure:"encrypt_id"`
	Name           string                  `json:"name" validate:"required,min=3,max=50"`
	OrganizationID int64                   `json:"organization_id" validate:"required"`
	Permissions    []permission.Permission `json:"permissions"`
}

var Validator = validator.New()

func (u *Role) Validate() error {
	return Validator.Struct(u)
}

type PermissionBrief struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RoleWithPermissions bundles a role and its assigned permissions.
type RoleWithPermissions struct {
	ID             int64             `json:"id"`
	Name           string            `json:"name"`
	OrganizationID int64             `json:"organization_id"`
	Permissions    []PermissionBrief `json:"permissions"`
}
