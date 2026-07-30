// internal/identity/helpers.go
package identity

import (
	"context"
	"errors"
)

var ErrUnauthenticated = errors.New("unauthenticated")

// Require returns the Auth or ErrUnauthenticated.
func Require(ctx context.Context) (Auth, error) {
	p, ok := FromContext(ctx)
	if !ok {
		return Auth{}, ErrUnauthenticated
	}
	return p, nil
}

// OrgID gets the org id (common in multi-tenant usecases).
func OrgID(ctx context.Context) (int64, error) {
	p, err := Require(ctx)
	if err != nil {
		return 0, err
	}
	return p.OrganizationID, nil
}

// UserID gets the user id.
func UserID(ctx context.Context) (int64, error) {
	p, err := Require(ctx)
	if err != nil {
		return 0, err
	}
	return p.UserID, nil
}

// Role gets the user role.
func Role(ctx context.Context) (string, error) {
	p, err := Require(ctx)
	if err != nil {
		return "", err
	}
	return p.Role, nil
}

// IsSuperAdmin checks if the user is a SuperAdmin.
func IsSuperAdmin(ctx context.Context) (bool, error) {
	role, err := Role(ctx)
	if err != nil {
		return false, err
	}
	return role == "SuperAdmin", nil
}

// Request metadata context keys
type requestMetaKey string

const (
	requestIDKey  requestMetaKey = "identity.request_id"
	ipAddressKey  requestMetaKey = "identity.ip_address"
	userAgentKey  requestMetaKey = "identity.user_agent"
	fullNameKey   requestMetaKey = "identity.full_name"
)

// WithRequestMeta adds request metadata to context
func WithRequestMeta(ctx context.Context, requestID, ipAddress, userAgent, fullName string) context.Context {
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ctx = context.WithValue(ctx, ipAddressKey, ipAddress)
	ctx = context.WithValue(ctx, userAgentKey, userAgent)
	ctx = context.WithValue(ctx, fullNameKey, fullName)
	return ctx
}

// RequestID gets the request ID from context
func RequestID(ctx context.Context) string {
	if val := ctx.Value(requestIDKey); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// IPAddress gets the client IP address from context
func IPAddress(ctx context.Context) string {
	if val := ctx.Value(ipAddressKey); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// UserAgent gets the user agent from context
func UserAgent(ctx context.Context) string {
	if val := ctx.Value(userAgentKey); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// FullName gets the user's full name from context
func FullName(ctx context.Context) string {
	if val := ctx.Value(fullNameKey); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}
