package middleware

import (
	"context"

	"github.com/edsonmubezi/myapp/internal/identity"
	"github.com/edsonmubezi/myapp/internal/permission"
)

// Helper getters that are safe for handlers; use identity in new code.

func GetRole(ctx context.Context) (string, bool) {
	if p, ok := identity.FromContext(ctx); ok && p.Role != "" {
		return p.Role, true
	}
	return "", false
}

func ExtractPermissionNames(perms []permission.Permission) []string {
	if len(perms) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(perms))
	out := make([]string, 0, len(perms))
	for _, p := range perms {
		if p.Name == "" {
			continue
		}
		if _, ok := seen[p.Name]; ok {
			continue
		}
		seen[p.Name] = struct{}{}
		out = append(out, p.Name)
	}
	return out
}
