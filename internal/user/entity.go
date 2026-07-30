package user

import (
	"time"

	"github.com/edsonmubezi/myapp/internal/organization"
	"github.com/edsonmubezi/myapp/internal/role"

	"github.com/go-playground/validator/v10"
)

type User struct {
	ID             int64      `json:"id" secure:"encrypt_id"`
	FullName       string     `json:"fullname" validate:"required,min=3,max=100"`
	Email          string     `json:"email" validate:"required,email"`
	Password       string     `json:"password" validate:"omitempty,min=6"`
	RoleID         int64      `json:"role_id" validate:"required"`
	ActiveStatus   bool       `json:"active_status"`
	Photo          *string    `json:"photo" validate:"omitempty,url"`
	OrganizationID int64      `json:"organization_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`

	// Security fields
	FailedLoginAttempts int        `json:"failed_login_attempts"`
	LockedAt            *time.Time `json:"locked_at,omitempty"`
	LockReason          *string    `json:"lock_reason,omitempty"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP         *string    `json:"last_login_ip,omitempty"`
	MustChangePassword  bool       `json:"must_change_password"`

	// Two-factor authentication
	TwoFactorEnabled bool    `json:"two_factor_enabled"`
	TwoFactorMethod  *string `json:"two_factor_method,omitempty"` // 'totp', 'email', 'sms'
	TOTPSecret       *string `json:"-"`                           // never expose in JSON
	PhoneNumber      *string `json:"phone_number,omitempty"`
	EmailVerified    bool    `json:"email_verified"`

	Role *role.Role `json:"role"`

	Organization *organization.Organization `json:"organization"`
}

var Validator = validator.New()

func (u *User) Validate() error {
	return Validator.Struct(u)
}

// IsLocked returns true if the account is locked
func (u *User) IsLocked() bool {
	return u.LockedAt != nil
}

// IsActive returns true if the account is active and not locked
func (u *User) IsActive() bool {
	return u.ActiveStatus && !u.IsLocked()
}

// CanLogin returns true if the user can attempt to login
func (u *User) CanLogin() (bool, string) {
	if !u.ActiveStatus {
		return false, "account_inactive"
	}
	if u.IsLocked() {
		return false, "account_locked"
	}
	return true, ""
}

type UpdateUserInput struct {
	ID             int64   `json:"id" validate:"-"` // not required from body; set from path
	FullName       *string `json:"fullname,omitempty" validate:"omitempty,min=2,max=150"`
	Email          *string `json:"email,omitempty" validate:"omitempty,email,max=150"`
	Password       *string `json:"password,omitempty" validate:"omitempty,min=6,max=100"`
	RoleID         *int64  `json:"role_id,omitempty" validate:"omitempty,gt=0"`
	ActiveStatus   *bool   `json:"active_status,omitempty"`
	Photo          *string `json:"photo,omitempty" validate:"omitempty,url"`
	OrganizationID *int64  `json:"organization_id,omitempty" validate:"omitempty,gt=0"`
}

type UserListItem struct {
	ID           int64   `json:"id" secure:"encrypt_id"`
	FullName     string  `json:"fullname"`
	Email        string  `json:"email"`
	Role         string  `json:"role"`
	ActiveStatus bool    `json:"active_status"`
	PhotoURL     *string `json:"photo_url"`
	Organization string  `json:"organization"`
	IsLocked     bool    `json:"is_locked"`
}

// UserListItemLimited for OrgsAdmin - includes email now
type UserListItemLimited struct {
	ID           int64   `json:"id" secure:"encrypt_id"`
	FullName     string  `json:"fullname"`
	Email        string  `json:"email"`
	Role         string  `json:"role"`
	ActiveStatus bool    `json:"active_status"`
	PhotoURL     *string `json:"photo_url"`
	Organization string  `json:"organization"`
}
