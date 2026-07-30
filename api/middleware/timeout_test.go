package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeoutMiddlewareNormalRequest(t *testing.T) {
	handler := TimeoutMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))

	req := httptest.NewRequest("GET", "/api/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTimeoutMiddlewareTimeout(t *testing.T) {
	// Create config with very short timeout
	cfg := &TimeoutConfig{
		DefaultTimeout: 50 * time.Millisecond,
		RouteTimeouts:  make(map[string]time.Duration),
		ExcludedRoutes: []string{},
	}

	handler := TimeoutMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow handler
		select {
		case <-time.After(200 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			// Context cancelled
			return
		}
	}))

	req := httptest.NewRequest("GET", "/api/slow", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503 for timeout, got %d", rec.Code)
	}
}

func TestTimeoutMiddlewareExcludedRoutes(t *testing.T) {
	cfg := &TimeoutConfig{
		DefaultTimeout: 10 * time.Millisecond, // Very short timeout
		RouteTimeouts:  make(map[string]time.Duration),
		ExcludedRoutes: []string{"/swagger/"},
	}

	handlerCalled := false
	handler := TimeoutMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow handler - would timeout if not excluded
		time.Sleep(50 * time.Millisecond)
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/swagger/index.html", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("handler should have been called for excluded route")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for excluded route, got %d", rec.Code)
	}
}

func TestGetTimeoutForRouteExactMatch(t *testing.T) {
	cfg := &TimeoutConfig{
		DefaultTimeout: 30 * time.Second,
		RouteTimeouts: map[string]time.Duration{
			"/api/login":  15 * time.Second,
			"/api/logout": 10 * time.Second,
		},
	}

	timeout := getTimeoutForRoute("/api/login", cfg)
	if timeout != 15*time.Second {
		t.Errorf("expected 15s for /api/login, got %v", timeout)
	}

	timeout = getTimeoutForRoute("/api/logout", cfg)
	if timeout != 10*time.Second {
		t.Errorf("expected 10s for /api/logout, got %v", timeout)
	}
}

func TestGetTimeoutForRoutePrefixMatch(t *testing.T) {
	cfg := &TimeoutConfig{
		DefaultTimeout: 30 * time.Second,
		RouteTimeouts: map[string]time.Duration{
			"/api/employees": 60 * time.Second,
		},
	}

	// Prefix match
	timeout := getTimeoutForRoute("/api/employees/123", cfg)
	if timeout != 60*time.Second {
		t.Errorf("expected 60s for /api/employees/123, got %v", timeout)
	}
}

func TestGetTimeoutForRouteDefault(t *testing.T) {
	cfg := &TimeoutConfig{
		DefaultTimeout: 30 * time.Second,
		RouteTimeouts:  map[string]time.Duration{},
	}

	timeout := getTimeoutForRoute("/api/unknown", cfg)
	if timeout != 30*time.Second {
		t.Errorf("expected default 30s, got %v", timeout)
	}
}

func TestGetTimeoutForRouteLongestPrefixWins(t *testing.T) {
	cfg := &TimeoutConfig{
		DefaultTimeout: 30 * time.Second,
		RouteTimeouts: map[string]time.Duration{
			"/api/":             20 * time.Second,
			"/api/payroll/":     60 * time.Second,
			"/api/payroll/generate": 120 * time.Second,
		},
	}

	// Should match longest prefix
	timeout := getTimeoutForRoute("/api/payroll/generate", cfg)
	if timeout != 120*time.Second {
		t.Errorf("expected 120s for /api/payroll/generate (longest match), got %v", timeout)
	}

	timeout = getTimeoutForRoute("/api/payroll/list", cfg)
	if timeout != 60*time.Second {
		t.Errorf("expected 60s for /api/payroll/list, got %v", timeout)
	}

	timeout = getTimeoutForRoute("/api/users", cfg)
	if timeout != 20*time.Second {
		t.Errorf("expected 20s for /api/users, got %v", timeout)
	}
}

func TestDefaultTimeoutConfig(t *testing.T) {
	cfg := DefaultTimeoutConfig()

	if cfg.DefaultTimeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", cfg.DefaultTimeout)
	}

	if len(cfg.RouteTimeouts) == 0 {
		t.Error("expected route timeouts to be configured")
	}

	// Check specific route timeouts
	if cfg.RouteTimeouts["/health"] != 5*time.Second {
		t.Errorf("expected /health timeout 5s, got %v", cfg.RouteTimeouts["/health"])
	}

	if cfg.RouteTimeouts["/api/payroll/generate"] != 120*time.Second {
		t.Errorf("expected /api/payroll/generate timeout 120s, got %v", cfg.RouteTimeouts["/api/payroll/generate"])
	}

	if cfg.RouteTimeouts["/api/employees/import"] != 300*time.Second {
		t.Errorf("expected /api/employees/import timeout 300s, got %v", cfg.RouteTimeouts["/api/employees/import"])
	}
}

func TestDefaultTimeoutConfigExcludedRoutes(t *testing.T) {
	cfg := DefaultTimeoutConfig()

	if len(cfg.ExcludedRoutes) == 0 {
		t.Error("expected excluded routes to be configured")
	}

	hasSwagger := false
	hasStatic := false
	for _, route := range cfg.ExcludedRoutes {
		if route == "/swagger/" {
			hasSwagger = true
		}
		if route == "/static/" {
			hasStatic = true
		}
	}

	if !hasSwagger {
		t.Error("expected /swagger/ to be excluded")
	}
	if !hasStatic {
		t.Error("expected /static/ to be excluded")
	}
}

func TestTimeoutMiddlewareSimple(t *testing.T) {
	handler := TimeoutMiddlewareSimple(1 * time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestWithTimeout(t *testing.T) {
	handler := WithTimeout(100*time.Millisecond, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestWithTimeoutExpired(t *testing.T) {
	handler := WithTimeout(50*time.Millisecond, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(200 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			return
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503 for timeout, got %d", rec.Code)
	}
}

func TestGetRequestTimeout(t *testing.T) {
	handler := TimeoutMiddleware(&TimeoutConfig{
		DefaultTimeout: 5 * time.Second,
		RouteTimeouts:  make(map[string]time.Duration),
		ExcludedRoutes: []string{},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeout := GetRequestTimeout(r)
		if timeout <= 0 || timeout > 5*time.Second {
			t.Errorf("expected timeout between 0 and 5s, got %v", timeout)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
}

func TestGetRequestTimeoutNoDeadline(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	timeout := GetRequestTimeout(req)
	if timeout != 0 {
		t.Errorf("expected 0 for request without deadline, got %v", timeout)
	}
}

func TestHasRequestDeadline(t *testing.T) {
	handler := TimeoutMiddleware(&TimeoutConfig{
		DefaultTimeout: 5 * time.Second,
		RouteTimeouts:  make(map[string]time.Duration),
		ExcludedRoutes: []string{},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !HasRequestDeadline(r) {
			t.Error("expected request to have deadline")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
}

func TestHasRequestDeadlineWithoutMiddleware(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	if HasRequestDeadline(req) {
		t.Error("expected request without middleware to have no deadline")
	}
}

func TestTimeoutResponseWriterWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	tw := &timeoutResponseWriter{
		ResponseWriter: rec,
		wroteHeader:    false,
	}

	tw.WriteHeader(http.StatusCreated)
	tw.WriteHeader(http.StatusOK) // Should be ignored

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestTimeoutResponseWriterWrite(t *testing.T) {
	rec := httptest.NewRecorder()
	tw := &timeoutResponseWriter{
		ResponseWriter: rec,
		wroteHeader:    false,
	}

	n, err := tw.Write([]byte("test"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes written, got %d", n)
	}

	// Should have auto-written 200 OK header
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTimeoutResponseWriterUnwrap(t *testing.T) {
	rec := httptest.NewRecorder()
	tw := &timeoutResponseWriter{
		ResponseWriter: rec,
		wroteHeader:    false,
	}

	unwrapped := tw.Unwrap()
	if unwrapped != rec {
		t.Error("Unwrap should return original ResponseWriter")
	}
}

func TestTimeoutMiddlewareNilConfig(t *testing.T) {
	// Should use default config when nil is passed
	handler := TimeoutMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTimeoutMiddlewareHealthEndpoint(t *testing.T) {
	cfg := DefaultTimeoutConfig()
	handler := TimeoutMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify health endpoint has short timeout
		timeout := GetRequestTimeout(r)
		if timeout > 6*time.Second {
			t.Errorf("expected health timeout <= 5s, got %v", timeout)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestTimeoutMiddlewareHeavyOperation(t *testing.T) {
	cfg := DefaultTimeoutConfig()
	handler := TimeoutMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify payroll generate has extended timeout
		timeout := GetRequestTimeout(r)
		if timeout < 100*time.Second {
			t.Errorf("expected payroll generate timeout >= 100s, got %v", timeout)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/payroll/generate", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
