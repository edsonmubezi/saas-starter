package identity

import (
	"context"
	"errors"
)

// Auth is the normalized auth payload carried through context.
// Keep this minimal to avoid import cycles.
type Auth struct {
	UserID         int64
	Email          string
	Role           string
	OrganizationID int64
}

type ctxKey string

const key ctxKey = "identity.auth"

// With returns a new context that carries the given Auth.
func With(ctx context.Context, a Auth) context.Context {
	return context.WithValue(ctx, key, a)
}

// FromContext extracts Auth from context. ok=false if missing.
func FromContext(ctx context.Context) (Auth, bool) {
	if ctx == nil {
		return Auth{}, false
	}
	val := ctx.Value(key)
	if val == nil {
		return Auth{}, false
	}
	if a, ok := val.(Auth); ok {
		return a, true
	}
	return Auth{}, false
}

// EnforceOrgOwnership returns an error if the caller's org doesn't match ownerOrgID.
// Use this instead of importing api/middleware from internal packages.
func EnforceOrgOwnership(ctx context.Context, ownerOrgID int64) error {
	a, err := Require(ctx)
	if err != nil {
		return err
	}
	if ownerOrgID != 0 && a.OrganizationID != 0 && a.OrganizationID != ownerOrgID {
		return errors.New("unauthorized")
	}
	return nil
}
