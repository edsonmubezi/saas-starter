// Deprecated compatibility shims. Prefer using internal/identity directly.
package middleware

import (
	"context"
	"net/http"

	"github.com/edsonmubezi/myapp/internal/identity"
)

type AuthContext = identity.Auth

// For use in handlers (legacy)
func GetAuthContext(r *http.Request) AuthContext {
	return GetAuthContextFromContext(r.Context())
}

// For use in use cases (legacy)
func GetAuthContextFromContext(ctx context.Context) AuthContext {
	if p, ok := identity.FromContext(ctx); ok {
		return p
	}
	return AuthContext{} // zero
}

// Legacy setter (avoid in new code; use identity.With)
func WithAuthContext(ctx context.Context, authCtx AuthContext) context.Context {
	return identity.With(ctx, authCtx)
}
