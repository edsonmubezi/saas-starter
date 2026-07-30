package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/edsonmubezi/myapp/internal/identity"
	"github.com/edsonmubezi/myapp/internal/user"
	"github.com/edsonmubezi/myapp/pkg/auth"
	"github.com/edsonmubezi/myapp/pkg/cache"
	"github.com/edsonmubezi/myapp/pkg/database"
	"github.com/google/uuid"
)

type ctxKey string

const userClaimsKey ctxKey = "userClaims"

// JWTMiddleware validates a Bearer token and puts both TokenClaims and identity.Auth into context.
// It also checks if the user account is locked or deactivated.
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		// Validate token AND check blacklist
		claims, err := auth.ValidateTokenWithBlacklist(r.Context(), tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Try Redis cache first for user auth data
		cacheKey := cache.AuthUserKey(claims.UserID, claims.OrganizationID)
		cached, cacheErr := cache.Get[cache.CachedAuthData](r.Context(), cacheKey)
		if cacheErr == nil && cached != nil {
			// Cache hit — check status and build context without DB queries
			if !cached.ActiveStatus {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "Forbidden",
					"message": "Account is deactivated",
				})
				return
			}
			if cached.IsLocked {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "Forbidden",
					"message": "Account is locked. Please contact your administrator.",
				})
				return
			}

			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			ctx = identity.With(ctx, identity.Auth{
				UserID:         claims.UserID,
				Email:          claims.Email,
				Role:           claims.Role,
				OrganizationID: claims.OrganizationID,
			})
			ctx = cache.WithCachedPermissions(ctx, cached.PermissionNames)

			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}
			ctx = identity.WithRequestMeta(ctx, requestID, GetClientIP(r), r.UserAgent(), claims.FullName)

			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Cache miss — fetch from DB
		userRepo := user.NewPostgresUserRepository(database.DB)
		u, err := userRepo.GetUserByID(r.Context(), claims.UserID, claims.OrganizationID)
		if err != nil {
			log.Printf("JWTMiddleware: User fetch failed for user %d: %v", claims.UserID, err)
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		// Check if account is deactivated
		if !u.ActiveStatus {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "Forbidden",
				"message": "Account is deactivated",
			})
			return
		}

		// Check if account is locked
		if u.IsLocked() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "Forbidden",
				"message": "Account is locked. Please contact your administrator.",
			})
			return
		}

		// Cache the auth data for subsequent requests
		authData := cache.CachedAuthData{
			ActiveStatus: u.ActiveStatus,
			IsLocked:     u.IsLocked(),
		}
		if u.Role != nil {
			authData.RoleName = u.Role.Name
			for _, p := range u.Role.Permissions {
				authData.PermissionNames = append(authData.PermissionNames, p.Name)
			}
		}
		if u.Organization != nil {
			authData.OrgName = u.Organization.Name
		}
		cache.Set(r.Context(), cacheKey, authData, 15*time.Minute)

		// Stash raw claims (optional)
		ctx := context.WithValue(r.Context(), userClaimsKey, claims)

		// Also attach a normalized identity.Auth used everywhere in internal/*
		ctx = identity.With(ctx, identity.Auth{
			UserID:         claims.UserID,
			Email:          claims.Email,
			Role:           claims.Role,
			OrganizationID: claims.OrganizationID,
		})
		ctx = cache.WithCachedPermissions(ctx, authData.PermissionNames)

		// Add request metadata for audit logging
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx = identity.WithRequestMeta(ctx, requestID, GetClientIP(r), r.UserAgent(), claims.FullName)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Optional helpers (kept for handlers that still want raw claims)
func GetUserClaims(ctx context.Context) (*auth.TokenClaims, bool) {
	val := ctx.Value(userClaimsKey)
	claims, ok := val.(*auth.TokenClaims)
	return claims, ok
}

func GetUserID(ctx context.Context) (int64, bool) {
	if c, ok := GetUserClaims(ctx); ok {
		return c.UserID, true
	}
	return 0, false
}

func GetEmail(ctx context.Context) (string, bool) {
	if c, ok := GetUserClaims(ctx); ok {
		return c.Email, true
	}
	return "", false
}
