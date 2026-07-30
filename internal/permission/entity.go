// internal/permission/entity.go
package permission

import "strings"

// Permission represents a permission in the system
type Permission struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Scope constants for permission prefixes
const (
	ScopeAdmin  = "admin"  // SuperAdmin/HQ only
	ScopeTenant = "tenant" // Organization-level (OrgAdmin)
)

// IsAdminOnly returns true if this permission is for SuperAdmin only
func (p *Permission) IsAdminOnly() bool {
	return strings.HasPrefix(p.Name, ScopeAdmin+".")
}

// IsTenantScope returns true if this permission is for tenant/org level
func (p *Permission) IsTenantScope() bool {
	return strings.HasPrefix(p.Name, ScopeTenant+".")
}

// GetScope extracts the scope prefix from permission name
func (p *Permission) GetScope() string {
	parts := strings.SplitN(p.Name, ".", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// GetResource extracts the resource from permission name (2nd segment)
func (p *Permission) GetResource() string {
	parts := strings.Split(p.Name, ".")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// GetAction extracts the action from permission name (3rd segment)
func (p *Permission) GetAction() string {
	parts := strings.Split(p.Name, ".")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

// GetVisibility returns computed visibility for backward compatibility
// Returns "ADMIN_ONLY" for admin.* permissions, "ALL" otherwise
func (p *Permission) GetVisibility() string {
	if p.IsAdminOnly() {
		return "ADMIN_ONLY"
	}
	return "ALL"
}

// PermissionResponse is the API response format with computed visibility
type PermissionResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

// ToResponse converts Permission to PermissionResponse with computed visibility
func (p *Permission) ToResponse() PermissionResponse {
	return PermissionResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Visibility:  p.GetVisibility(),
	}
}
