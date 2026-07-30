// api/middleware/logging.go
package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/edsonmubezi/myapp/internal/identity"
	"github.com/edsonmubezi/myapp/internal/logs"

	"github.com/gorilla/mux"
)

// LoggingMiddleware logs important user actions to the database for audit compliance.
// It captures actions like password changes, user deletions, and other security-sensitive operations.
func LoggingMiddleware(logRepo logs.LogRepository) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First, serve the request so the identity context is populated by JWTMiddleware
			next.ServeHTTP(w, r)

			// After the request completes, log the action if applicable
			action := determineAction(r)
			if action == "" {
				return
			}

			// Get user ID from context (set by JWTMiddleware)
			userID := int64(0)
			if auth, ok := identity.FromContext(r.Context()); ok {
				userID = auth.UserID
			}

			// Skip logging if we couldn't identify the user (except for signup which creates the user)
			if userID == 0 && action != "User Signup" {
				return
			}

			ipAddress := GetClientIP(r)
			userAgent := r.UserAgent()

			logEntry := &logs.Log{
				UserID:    userID,
				Action:    action,
				IPAddress: ipAddress,
				UserAgent: userAgent,
				CreatedAt: time.Now(),
			}

			// Save log entry asynchronously to avoid slowing down the response
			go func(entry *logs.Log) {
				if err := logRepo.SaveLog(r.Context(), entry); err != nil {
					log.Printf("Error saving audit log: %v", err)
				}
			}(logEntry)
		})
	}
}

// determineAction maps request paths and methods to human-readable action descriptions
func determineAction(r *http.Request) string {
	path := r.URL.Path
	method := r.Method

	// Authentication and security actions
	switch {
	case path == "/api/register" && method == http.MethodPost:
		return "User Signup"
	case path == "/api/change-password" && method == http.MethodPut:
		return "Password Change"
	case path == "/api/logout" && method == http.MethodPost:
		return "User Logout"

	// Two-Factor Authentication actions
	case path == "/api/auth/2fa/setup" && method == http.MethodPost:
		return "2FA Setup Initiated"
	case path == "/api/auth/2fa/verify" && method == http.MethodPost:
		return "2FA Setup Verified"
	case path == "/api/auth/2fa/disable" && method == http.MethodPost:
		return "2FA Disabled"
	case path == "/api/auth/2fa/backup-codes/regenerate" && method == http.MethodPost:
		return "2FA Backup Codes Regenerated"

	// Session management actions
	case path == "/api/sessions/revoke" && method == http.MethodPost:
		return "Session Revoked"
	case path == "/api/sessions/revoke-others" && method == http.MethodPost:
		return "Other Sessions Revoked"
	case strings.HasPrefix(path, "/api/sessions/") && method == http.MethodDelete:
		return "Session Deleted"

	// Admin user management actions
	case strings.HasPrefix(path, "/api/admin/users") && method == http.MethodDelete:
		return "Admin Deleted User"
	case strings.HasPrefix(path, "/api/admin/users") && method == http.MethodPost:
		return "Admin Created User"

	// Organization management actions
	case strings.HasPrefix(path, "/api/org/employees") && method == http.MethodDelete:
		return "Employee Deleted"
	case strings.HasPrefix(path, "/api/org/employees") && method == http.MethodPost:
		return "Employee Created"

	// Payroll actions (sensitive financial data)
	case strings.HasPrefix(path, "/api/payroll/finalize") && method == http.MethodPost:
		return "Payroll Finalized"
	case strings.HasPrefix(path, "/api/payroll/adjustments") && method == http.MethodDelete:
		return "Payroll Adjustment Deleted"

	// Loan actions
	case strings.HasPrefix(path, "/api/org/loans") && strings.Contains(path, "/approve"):
		return "Loan Approved"
	case strings.HasPrefix(path, "/api/org/loans") && strings.Contains(path, "/reject"):
		return "Loan Rejected"
	}

	return ""
}

// Note: GetClientIP is defined in rate_limit.go and reused here
