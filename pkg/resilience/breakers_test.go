package resilience

import (
	"sync"
	"testing"
)

// resetBreakersForTest resets the package-level state for testing
func resetBreakersForTest() {
	mu.Lock()
	defer mu.Unlock()
	breakers = nil
	once = sync.Once{} // Reset the once to allow reinitialization
}

func TestInitializeCircuitBreakers(t *testing.T) {
	resetBreakersForTest()
	InitializeCircuitBreakers()

	// Verify all expected breakers are created
	expectedServices := []string{
		ServiceDatabase,
		ServiceRedis,
		ServiceClamAV,
		ServiceS3,
		ServiceSMTP,
	}

	for _, service := range expectedServices {
		cb := GetBreaker(service)
		if cb == nil {
			t.Errorf("expected breaker for service %s, got nil", service)
		}
		if cb != nil && cb.Name() != service {
			t.Errorf("expected breaker name %s, got %s", service, cb.Name())
		}
	}
}

func TestGetBreaker(t *testing.T) {
	resetBreakersForTest()
	InitializeCircuitBreakers()

	// Should return initialized breaker
	cb := GetBreaker(ServiceDatabase)
	if cb == nil {
		t.Error("expected database breaker to be auto-initialized")
	}

	// Non-existent service should return nil
	cb = GetBreaker("non-existent")
	if cb != nil {
		t.Error("expected nil for non-existent service")
	}
}

func TestGetOrCreateBreaker(t *testing.T) {
	resetBreakersForTest()

	// Create new breaker
	cb := GetOrCreateBreaker("custom-service", nil)
	if cb == nil {
		t.Error("expected custom breaker to be created")
	}
	if cb.Name() != "custom-service" {
		t.Errorf("expected name 'custom-service', got %s", cb.Name())
	}

	// Get existing breaker
	cb2 := GetOrCreateBreaker("custom-service", nil)
	if cb != cb2 {
		t.Error("expected same breaker instance for same service name")
	}
}

func TestGetAllBreakers(t *testing.T) {
	resetBreakersForTest()

	// Should return nil before initialization
	all := GetAllBreakers()
	if all != nil {
		t.Error("expected nil before initialization")
	}

	resetBreakersForTest()
	InitializeCircuitBreakers()

	all = GetAllBreakers()
	if all == nil {
		t.Fatal("expected non-nil after initialization")
	}

	if len(all) != 5 {
		t.Errorf("expected 5 breakers, got %d", len(all))
	}
}

func TestGetBreakerStates(t *testing.T) {
	resetBreakersForTest()
	InitializeCircuitBreakers()

	states := GetBreakerStates()
	if states == nil {
		t.Fatal("expected non-nil states")
	}

	// Verify state structure
	for name, status := range states {
		if status.Name != name {
			t.Errorf("expected status name %s, got %s", name, status.Name)
		}
		if status.State != "CLOSED" {
			t.Errorf("expected initial state CLOSED for %s, got %s", name, status.State)
		}
		if !status.IsAvailable {
			t.Errorf("expected %s to be available initially", name)
		}
	}
}

func TestResetBreaker(t *testing.T) {
	resetBreakersForTest()
	InitializeCircuitBreakers()

	// Reset existing breaker
	if !ResetBreaker(ServiceDatabase) {
		t.Error("expected ResetBreaker to return true for existing service")
	}

	// Reset non-existent breaker
	if ResetBreaker("non-existent") {
		t.Error("expected ResetBreaker to return false for non-existent service")
	}
}

func TestResetAllBreakers(t *testing.T) {
	resetBreakersForTest()
	InitializeCircuitBreakers()

	// Should not panic
	ResetAllBreakers()

	// Verify all breakers are in closed state
	states := GetBreakerStates()
	for name, status := range states {
		if status.State != "CLOSED" {
			t.Errorf("expected %s to be CLOSED after reset, got %s", name, status.State)
		}
	}
}

func TestServiceConstants(t *testing.T) {
	// Verify service constants are defined correctly
	if ServiceDatabase != "database" {
		t.Errorf("expected ServiceDatabase='database', got %s", ServiceDatabase)
	}
	if ServiceRedis != "redis" {
		t.Errorf("expected ServiceRedis='redis', got %s", ServiceRedis)
	}
	if ServiceClamAV != "clamav" {
		t.Errorf("expected ServiceClamAV='clamav', got %s", ServiceClamAV)
	}
	if ServiceS3 != "s3" {
		t.Errorf("expected ServiceS3='s3', got %s", ServiceS3)
	}
	if ServiceSMTP != "smtp" {
		t.Errorf("expected ServiceSMTP='smtp', got %s", ServiceSMTP)
	}
}

func TestBreakerStatusFields(t *testing.T) {
	resetBreakersForTest()
	InitializeCircuitBreakers()

	states := GetBreakerStates()
	status := states[ServiceDatabase]

	// Verify all fields are present and have expected types
	if status.Name == "" {
		t.Error("expected Name to be non-empty")
	}
	if status.State == "" {
		t.Error("expected State to be non-empty")
	}
	// TotalSuccesses and TotalFailures should be 0 initially
	if status.TotalSuccesses != 0 {
		t.Errorf("expected TotalSuccesses=0, got %d", status.TotalSuccesses)
	}
	if status.TotalFailures != 0 {
		t.Errorf("expected TotalFailures=0, got %d", status.TotalFailures)
	}
}
