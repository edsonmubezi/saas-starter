package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/edsonmubezi/myapp/internal/identity"
	"github.com/edsonmubezi/myapp/internal/permission"
	"github.com/edsonmubezi/myapp/internal/user"
	"github.com/edsonmubezi/myapp/pkg/auth"
	"github.com/edsonmubezi/myapp/pkg/cache"
	"github.com/edsonmubezi/myapp/pkg/database"
)

func AuthorizationMiddleware(requiredRoles []string, requiredPermissions []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Fast path: if JWTMiddleware already ran, identity + permissions are in context
			if a, ok := identity.FromContext(r.Context()); ok {
				if len(requiredRoles) > 0 && !hasRequiredRole(a.Role, requiredRoles) {
					respondWithJSON(w, http.StatusForbidden, map[string]string{
						"error":   "Forbidden",
						"message": "Insufficient role",
					})
					return
				}
				if len(requiredPermissions) > 0 {
					if cachedPerms, hasCached := cache.GetCachedPermissions(r.Context()); hasCached {
						if !hasRequiredPermissionsByName(cachedPerms, requiredPermissions) {
							respondWithJSON(w, http.StatusForbidden, map[string]string{
								"error":   "Forbidden",
								"message": "Insufficient permissions",
							})
							return
						}
					}
				}
				next.ServeHTTP(w, r)
				return
			}

			// Fallback: JWTMiddleware didn't run — full validation
			authHeader := r.Header.Get("Authorization")
			tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if tokenString == "" {
				respondWithJSON(w, http.StatusUnauthorized, map[string]string{
					"error":   "Unauthorized",
					"message": "Missing token",
				})
				return
			}

			claims, err := auth.ValidateTokenWithBlacklist(r.Context(), tokenString)
			if err != nil {
				log.Printf("Token validation failed: %v", err)
				respondWithJSON(w, http.StatusUnauthorized, map[string]string{
					"error":   "Unauthorized",
					"message": "Invalid or expired token",
				})
				return
			}

			userRepo := user.NewPostgresUserRepository(database.DB)
			u, err := userRepo.GetUserByID(r.Context(), claims.UserID, claims.OrganizationID)
			if err != nil {
				log.Printf("User fetch failed: %v", err)
				respondWithJSON(w, http.StatusUnauthorized, map[string]string{
					"error":   "Unauthorized",
					"message": "User not found",
				})
				return
			}

			if !u.ActiveStatus {
				respondWithJSON(w, http.StatusForbidden, map[string]string{
					"error":   "Forbidden",
					"message": "Account is deactivated",
				})
				return
			}

			if u.IsLocked() {
				respondWithJSON(w, http.StatusForbidden, map[string]string{
					"error":   "Forbidden",
					"message": "Account is locked. Please contact your administrator.",
				})
				return
			}

			if len(requiredRoles) > 0 && !hasRequiredRole(u.Role.Name, requiredRoles) {
				respondWithJSON(w, http.StatusForbidden, map[string]string{
					"error":   "Forbidden",
					"message": "Insufficient role",
				})
				return
			}
			if len(requiredPermissions) > 0 && !hasRequiredPermissions(u.Role.Permissions, requiredPermissions) {
				respondWithJSON(w, http.StatusForbidden, map[string]string{
					"error":   "Forbidden",
					"message": "Insufficient permissions",
				})
				return
			}

			ctx := identity.With(r.Context(), identity.Auth{
				UserID:         claims.UserID,
				Email:          claims.Email,
				Role:           claims.Role,
				OrganizationID: claims.OrganizationID,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAnyPermission creates a middleware that requires the user to have at least one of the specified permissions (OR logic)
// This is useful for routes that can be accessed by multiple roles with different permission sets
func RequireAnyPermission(permissions []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Fast path: if JWTMiddleware already ran, identity + permissions are in context
			if _, ok := identity.FromContext(r.Context()); ok {
				if len(permissions) > 0 {
					if cachedPerms, hasCached := cache.GetCachedPermissions(r.Context()); hasCached {
						if !hasAnyPermissionByName(cachedPerms, permissions) {
							respondWithJSON(w, http.StatusForbidden, map[string]string{
								"error":   "Forbidden",
								"message": "Insufficient permissions",
							})
							return
						}
					}
				}
				next.ServeHTTP(w, r)
				return
			}

			// Fallback: JWTMiddleware didn't run — full validation
			authHeader := r.Header.Get("Authorization")
			tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if tokenString == "" {
				respondWithJSON(w, http.StatusUnauthorized, map[string]string{
					"error":   "Unauthorized",
					"message": "Missing token",
				})
				return
			}

			claims, err := auth.ValidateTokenWithBlacklist(r.Context(), tokenString)
			if err != nil {
				log.Printf("Token validation failed: %v", err)
				respondWithJSON(w, http.StatusUnauthorized, map[string]string{
					"error":   "Unauthorized",
					"message": "Invalid or expired token",
				})
				return
			}

			userRepo := user.NewPostgresUserRepository(database.DB)
			u, err := userRepo.GetUserByID(r.Context(), claims.UserID, claims.OrganizationID)
			if err != nil {
				log.Printf("User fetch failed: %v", err)
				respondWithJSON(w, http.StatusUnauthorized, map[string]string{
					"error":   "Unauthorized",
					"message": "User not found",
				})
				return
			}

			if !u.ActiveStatus {
				respondWithJSON(w, http.StatusForbidden, map[string]string{
					"error":   "Forbidden",
					"message": "Account is deactivated",
				})
				return
			}

			if u.IsLocked() {
				respondWithJSON(w, http.StatusForbidden, map[string]string{
					"error":   "Forbidden",
					"message": "Account is locked. Please contact your administrator.",
				})
				return
			}

			if len(permissions) > 0 && !hasAnyPermission(u.Role.Permissions, permissions) {
				respondWithJSON(w, http.StatusForbidden, map[string]string{
					"error":   "Forbidden",
					"message": "Insufficient permissions",
				})
				return
			}

			ctx := identity.With(r.Context(), identity.Auth{
				UserID:         claims.UserID,
				Email:          claims.Email,
				Role:           claims.Role,
				OrganizationID: claims.OrganizationID,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func hasRequiredRole(userRole string, requiredRoles []string) bool {
	for _, role := range requiredRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

func hasRequiredPermissions(userPermissions []permission.Permission, required []string) bool {
	for _, need := range required {
		found := false
		for _, p := range userPermissions {
			if p.Name == need {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// hasAnyPermission checks if the user has at least one of the required permissions (OR logic)
func hasAnyPermission(userPermissions []permission.Permission, required []string) bool {
	for _, need := range required {
		for _, p := range userPermissions {
			if p.Name == need {
				return true
			}
		}
	}
	return false
}

// hasRequiredPermissionsByName checks permission names (from cached string slice) against required permissions
func hasRequiredPermissionsByName(userPerms []string, required []string) bool {
	permSet := make(map[string]struct{}, len(userPerms))
	for _, p := range userPerms {
		permSet[p] = struct{}{}
	}
	for _, need := range required {
		if _, ok := permSet[need]; !ok {
			return false
		}
	}
	return true
}

// hasAnyPermissionByName checks if user has at least one of the required permissions (from cached string slice)
func hasAnyPermissionByName(userPerms []string, required []string) bool {
	permSet := make(map[string]struct{}, len(userPerms))
	for _, p := range userPerms {
		permSet[p] = struct{}{}
	}
	for _, need := range required {
		if _, ok := permSet[need]; ok {
			return true
		}
	}
	return false
}

func respondWithJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
