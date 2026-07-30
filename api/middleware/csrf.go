package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

const (
	csrfTokenHeader = "X-CSRF-Token"
	csrfTokenCookie = "csrf_token"
)

// CSRFProtection provides double-submit cookie CSRF protection.
//
// This middleware is designed for browser-facing form flows. It is automatically
// skipped when the request carries an Authorization: Bearer token, because CSRF
// attacks only affect cookie-based authentication — a cross-origin attacker
// cannot set arbitrary request headers, so JWT APIs are inherently safe.
//
// Flow for cookie-auth scenarios:
//  1. On GET requests the middleware issues a csrf_token cookie (HttpOnly: false
//     so JS can read it).
//  2. On mutating requests (POST/PUT/PATCH/DELETE) the client must echo the
//     cookie value back in the X-CSRF-Token header.
func CSRFProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip entirely when using Bearer token authentication — CSRF does not apply.
		if strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			next.ServeHTTP(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			// Safe methods: ensure the CSRF cookie is present and continue.
			ensureCSRFCookie(w, r)
			next.ServeHTTP(w, r)

		default:
			// Mutating methods: validate that the header matches the cookie.
			cookie, err := r.Cookie(csrfTokenCookie)
			if err != nil {
				http.Error(w, `{"error":{"code":"FORBIDDEN","message":"missing CSRF cookie"}}`, http.StatusForbidden)
				return
			}

			headerToken := r.Header.Get(csrfTokenHeader)
			if headerToken == "" || headerToken != cookie.Value {
				http.Error(w, `{"error":{"code":"FORBIDDEN","message":"invalid CSRF token"}}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	})
}

// ensureCSRFCookie sets the csrf_token cookie if it is not already present.
func ensureCSRFCookie(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie(csrfTokenCookie); err == nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     csrfTokenCookie,
		Value:    generateCSRFToken(),
		Path:     "/",
		HttpOnly: false,             // Must be readable by JS
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		MaxAge:   3600, // 1 hour
	})
}

func generateCSRFToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback — should never happen in practice
		return "fallback-csrf-token"
	}
	return hex.EncodeToString(b)
}
