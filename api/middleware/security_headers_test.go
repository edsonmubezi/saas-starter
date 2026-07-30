package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersDefault(t *testing.T) {
	handler := SecurityHeadersDefault()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check required security headers are set
	tests := []struct {
		header   string
		expected string
	}{
		{"X-Frame-Options", "DENY"},
		{"X-Content-Type-Options", "nosniff"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := rec.Header().Get(tt.header)
			if got != tt.expected {
				t.Errorf("expected %s=%s, got %s", tt.header, tt.expected, got)
			}
		})
	}
}

func TestSecurityHeadersCSP(t *testing.T) {
	handler := SecurityHeadersDefault()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Error("expected Content-Security-Policy header to be set")
	}

	// Check for key directives
	requiredDirectives := []string{
		"default-src",
		"script-src",
		"style-src",
		"frame-ancestors",
	}

	for _, directive := range requiredDirectives {
		if !containsDirective(csp, directive) {
			t.Errorf("expected CSP to contain %s directive", directive)
		}
	}
}

func TestSecurityHeadersPermissionsPolicy(t *testing.T) {
	handler := SecurityHeadersDefault()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	pp := rec.Header().Get("Permissions-Policy")
	if pp == "" {
		t.Error("expected Permissions-Policy header to be set")
	}

	// Check that dangerous features are disabled
	disabledFeatures := []string{
		"camera=()",
		"microphone=()",
		"geolocation=()",
	}

	for _, feature := range disabledFeatures {
		if !containsDirective(pp, feature) {
			t.Errorf("expected Permissions-Policy to disable %s", feature)
		}
	}
}

func TestSecurityHeadersAPICacheControl(t *testing.T) {
	handler := SecurityHeadersDefault()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	cacheControl := rec.Header().Get("Cache-Control")
	if cacheControl == "" {
		t.Error("expected Cache-Control header for API endpoint")
	}

	if !containsDirective(cacheControl, "no-store") {
		t.Error("expected Cache-Control to include no-store for API")
	}

	pragma := rec.Header().Get("Pragma")
	if pragma != "no-cache" {
		t.Errorf("expected Pragma=no-cache, got %s", pragma)
	}
}

func TestSecurityHeadersSensitiveEndpoints(t *testing.T) {
	handler := SecurityHeadersDefault()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	sensitiveEndpoints := []string{
		"/api/login",
		"/api/logout",
		"/api/auth/reset-password",
		"/api/change-password",
		"/api/payroll/results",
		"/api/salary-slips",
		"/api/loans/123",
	}

	for _, endpoint := range sensitiveEndpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest("GET", endpoint, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			cacheControl := rec.Header().Get("Cache-Control")
			if !containsDirective(cacheControl, "no-store") {
				t.Errorf("expected no-store for sensitive endpoint %s", endpoint)
			}
			if !containsDirective(cacheControl, "private") {
				t.Errorf("expected private for sensitive endpoint %s", endpoint)
			}
		})
	}
}

func TestSecurityHeadersHSTSProduction(t *testing.T) {
	// Test with HTTPS request
	cfg := DefaultSecurityHeadersConfig()
	cfg.EnableHSTS = true

	handler := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	hsts := rec.Header().Get("Strict-Transport-Security")
	if hsts == "" {
		t.Error("expected HSTS header for HTTPS request with EnableHSTS=true")
	}

	if !containsDirective(hsts, "max-age") {
		t.Error("expected HSTS to include max-age")
	}
}

func TestSecurityHeadersNoHSTSForHTTP(t *testing.T) {
	cfg := DefaultSecurityHeadersConfig()
	cfg.EnableHSTS = true

	handler := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// HTTP request (no TLS, no forwarded proto)
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	hsts := rec.Header().Get("Strict-Transport-Security")
	if hsts != "" {
		t.Error("HSTS should not be set for HTTP requests")
	}
}

func TestAPISecurityHeadersConfig(t *testing.T) {
	cfg := APISecurityHeadersConfig()

	if cfg.ContentSecurityPolicy == "" {
		t.Error("expected CSP to be set for API config")
	}

	// API CSP should be more restrictive
	if !containsDirective(cfg.ContentSecurityPolicy, "default-src 'none'") {
		t.Error("expected API CSP to have restrictive default-src")
	}
}

func TestSecureAPIResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	SecureAPIResponse(rec)

	tests := []struct {
		header   string
		contains string
	}{
		{"Content-Type", "application/json"},
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"Cache-Control", "no-store"},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := rec.Header().Get(tt.header)
			if !containsDirective(got, tt.contains) {
				t.Errorf("expected %s to contain %s, got %s", tt.header, tt.contains, got)
			}
		})
	}
}

func TestSetDownloadHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	SetDownloadHeaders(rec, "report.pdf", "application/pdf")

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/pdf" {
		t.Errorf("expected Content-Type=application/pdf, got %s", contentType)
	}

	disposition := rec.Header().Get("Content-Disposition")
	if !containsDirective(disposition, "attachment") {
		t.Error("expected Content-Disposition to be attachment")
	}
	if !containsDirective(disposition, "report.pdf") {
		t.Error("expected filename in Content-Disposition")
	}

	nosniff := rec.Header().Get("X-Content-Type-Options")
	if nosniff != "nosniff" {
		t.Error("expected X-Content-Type-Options=nosniff for downloads")
	}
}

func TestSetInlineViewHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	SetInlineViewHeaders(rec, "document.pdf", "application/pdf")

	disposition := rec.Header().Get("Content-Disposition")
	if !containsDirective(disposition, "inline") {
		t.Error("expected Content-Disposition to be inline")
	}

	frameOptions := rec.Header().Get("X-Frame-Options")
	if frameOptions != "SAMEORIGIN" {
		t.Errorf("expected X-Frame-Options=SAMEORIGIN for inline view, got %s", frameOptions)
	}
}

func TestIsSensitiveEndpoint(t *testing.T) {
	tests := []struct {
		path      string
		sensitive bool
	}{
		{"/api/login", true},
		{"/api/logout", true},
		{"/api/auth/reset-password", true},
		{"/api/change-password", true},
		{"/api/users", true},
		{"/api/users/123", true},
		{"/api/self/profile", true},
		{"/api/payroll/results", true},
		{"/api/salary-slips", true},
		{"/api/loans/123", true},
		{"/api/departments", false},
		{"/api/designations", false},
		{"/health", false},
		{"/swagger/", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isSensitiveEndpoint(tt.path)
			if got != tt.sensitive {
				t.Errorf("isSensitiveEndpoint(%s) = %v, want %v", tt.path, got, tt.sensitive)
			}
		})
	}
}

func TestDefaultSecurityHeadersConfig(t *testing.T) {
	cfg := DefaultSecurityHeadersConfig()

	if cfg.XFrameOptions != "DENY" {
		t.Errorf("expected XFrameOptions=DENY, got %s", cfg.XFrameOptions)
	}

	if cfg.XContentTypeOptions != "nosniff" {
		t.Errorf("expected XContentTypeOptions=nosniff, got %s", cfg.XContentTypeOptions)
	}

	if cfg.XSSProtection != "1; mode=block" {
		t.Errorf("expected XSSProtection='1; mode=block', got %s", cfg.XSSProtection)
	}

	if cfg.ReferrerPolicy != "strict-origin-when-cross-origin" {
		t.Errorf("expected ReferrerPolicy='strict-origin-when-cross-origin', got %s", cfg.ReferrerPolicy)
	}

	if cfg.EnableHSTS {
		t.Error("HSTS should be disabled by default")
	}
}

// Helper function to check if a string contains a directive
func containsDirective(s, directive string) bool {
	return len(s) > 0 && (s == directive || len(directive) > 0 &&
		(len(s) >= len(directive) &&
			(s[:len(directive)] == directive ||
				containsSubstring(s, directive))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
