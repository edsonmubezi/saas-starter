package middleware

import (
	"net/http"
	"os"
	"strings"
)

// SecurityHeadersConfig holds configuration for security headers
type SecurityHeadersConfig struct {
	// ContentSecurityPolicy defines allowed sources for content
	ContentSecurityPolicy string
	// XFrameOptions prevents clickjacking (DENY, SAMEORIGIN, or ALLOW-FROM uri)
	XFrameOptions string
	// XContentTypeOptions prevents MIME type sniffing
	XContentTypeOptions string
	// XSSProtection enables XSS filtering in older browsers
	XSSProtection string
	// StrictTransportSecurity enforces HTTPS (HSTS)
	StrictTransportSecurity string
	// ReferrerPolicy controls referrer information sent with requests
	ReferrerPolicy string
	// PermissionsPolicy controls browser features
	PermissionsPolicy string
	// CacheControl for sensitive responses
	CacheControl string
	// EnableHSTS enables HSTS header (should only be true in production with HTTPS)
	EnableHSTS bool
}

// DefaultSecurityHeadersConfig returns production-ready security header defaults
func DefaultSecurityHeadersConfig() *SecurityHeadersConfig {
	return &SecurityHeadersConfig{
		// Content Security Policy - restrictive by default
		// Allows self-hosted content, inline styles (needed for some UI frameworks)
		// Adjust based on your frontend requirements
		ContentSecurityPolicy: "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self'; " +
			"frame-ancestors 'self'; " +
			"form-action 'self'; " +
			"base-uri 'self'; " +
			"object-src 'none'",

		// Prevent clickjacking
		XFrameOptions: "DENY",

		// Prevent MIME type sniffing
		XContentTypeOptions: "nosniff",

		// XSS protection for older browsers
		XSSProtection: "1; mode=block",

		// HSTS - 1 year, include subdomains, allow preload
		StrictTransportSecurity: "max-age=31536000; includeSubDomains; preload",

		// Referrer policy - send origin only for cross-origin requests
		ReferrerPolicy: "strict-origin-when-cross-origin",

		// Permissions policy - disable sensitive browser features
		PermissionsPolicy: "accelerometer=(), " +
			"camera=(), " +
			"geolocation=(), " +
			"gyroscope=(), " +
			"magnetometer=(), " +
			"microphone=(), " +
			"payment=(), " +
			"usb=()",

		// Cache control for API responses
		CacheControl: "no-store, no-cache, must-revalidate, proxy-revalidate",

		// Only enable HSTS in production with HTTPS
		EnableHSTS: false,
	}
}

// APISecurityHeadersConfig returns config optimized for API responses
func APISecurityHeadersConfig() *SecurityHeadersConfig {
	cfg := DefaultSecurityHeadersConfig()
	// APIs typically don't need CSP, but we keep it restrictive
	cfg.ContentSecurityPolicy = "default-src 'none'; frame-ancestors 'none'"
	return cfg
}

// SecurityHeaders returns a middleware that adds security headers to responses
func SecurityHeaders(cfg *SecurityHeadersConfig) func(http.Handler) http.Handler {
	if cfg == nil {
		cfg = DefaultSecurityHeadersConfig()
	}

	// Check if we're in production for HSTS
	runtimeEnv := strings.ToLower(os.Getenv("RUNTIME_ENV"))
	isProduction := runtimeEnv == "production"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent clickjacking
			if cfg.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", cfg.XFrameOptions)
			}

			// Prevent MIME type sniffing
			if cfg.XContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", cfg.XContentTypeOptions)
			}

			// XSS protection for older browsers
			if cfg.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", cfg.XSSProtection)
			}

			// Content Security Policy
			if cfg.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
			}

			// Referrer Policy
			if cfg.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
			}

			// Permissions Policy (formerly Feature-Policy)
			if cfg.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", cfg.PermissionsPolicy)
			}

			// HSTS - only in production with HTTPS
			if cfg.EnableHSTS || isProduction {
				// Check if request is over HTTPS or behind a proxy that terminates TLS
				isSecure := r.TLS != nil ||
					r.Header.Get("X-Forwarded-Proto") == "https" ||
					r.Header.Get("X-Forwarded-Ssl") == "on"

				if isSecure && cfg.StrictTransportSecurity != "" {
					w.Header().Set("Strict-Transport-Security", cfg.StrictTransportSecurity)
				}
			}

			// Cache control for API endpoints
			if strings.HasPrefix(r.URL.Path, "/api/") {
				if cfg.CacheControl != "" {
					w.Header().Set("Cache-Control", cfg.CacheControl)
					w.Header().Set("Pragma", "no-cache")
					w.Header().Set("Expires", "0")
				}
			}

			// Prevent caching of sensitive endpoints
			if isSensitiveEndpoint(r.URL.Path) {
				w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersDefault returns middleware with default security headers
func SecurityHeadersDefault() func(http.Handler) http.Handler {
	return SecurityHeaders(DefaultSecurityHeadersConfig())
}

// SecurityHeadersAPI returns middleware optimized for API responses
func SecurityHeadersAPI() func(http.Handler) http.Handler {
	return SecurityHeaders(APISecurityHeadersConfig())
}

// isSensitiveEndpoint checks if the endpoint handles sensitive data
func isSensitiveEndpoint(path string) bool {
	sensitivePatterns := []string{
		"/api/login",
		"/api/logout",
		"/api/refresh",
		"/api/auth/",
		"/api/change-password",
		"/api/users",
		"/api/self/",
		"/api/payroll",
		"/api/salary",
		"/api/loans",
	}

	for _, pattern := range sensitivePatterns {
		if strings.HasPrefix(path, pattern) {
			return true
		}
	}
	return false
}

// SecureAPIResponse adds security headers specifically for API JSON responses
// Use this for individual handlers that need extra protection
func SecureAPIResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// SetDownloadHeaders sets headers for file downloads
func SetDownloadHeaders(w http.ResponseWriter, filename, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// SetInlineViewHeaders sets headers for inline file viewing (e.g., PDFs in browser)
func SetInlineViewHeaders(w http.ResponseWriter, filename, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "inline; filename=\""+filename+"\"")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN") // Allow embedding in same origin
}
