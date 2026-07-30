package resilience

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreakerInitialState(t *testing.T) {
	cb := NewCircuitBreaker("test", nil)

	if cb.GetState() != StateClosed {
		t.Errorf("expected initial state to be CLOSED, got %s", cb.GetState())
	}

	if !cb.IsAvailable() {
		t.Error("expected circuit breaker to be available initially")
	}
}

func TestCircuitBreakerOpensAfterFailures(t *testing.T) {
	cfg := &CircuitBreakerConfig{
		MaxRequests:      5,
		Interval:         time.Minute,
		Timeout:          100 * time.Millisecond,
		FailureThreshold: 3,
		SuccessThreshold: 2,
	}
	cb := NewCircuitBreaker("test", cfg)
	ctx := context.Background()

	// Cause failures
	for i := 0; i < 3; i++ {
		_ = cb.Execute(ctx, func() error {
			return errors.New("failure")
		})
	}

	if cb.GetState() != StateOpen {
		t.Errorf("expected state to be OPEN after %d failures, got %s", cfg.FailureThreshold, cb.GetState())
	}

	if cb.IsAvailable() {
		t.Error("expected circuit breaker to be unavailable when open")
	}
}

func TestCircuitBreakerBlocksRequestsWhenOpen(t *testing.T) {
	cfg := &CircuitBreakerConfig{
		MaxRequests:      5,
		Interval:         time.Minute,
		Timeout:          100 * time.Millisecond,
		FailureThreshold: 2,
		SuccessThreshold: 2,
	}
	cb := NewCircuitBreaker("test", cfg)
	ctx := context.Background()

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func() error {
			return errors.New("failure")
		})
	}

	// Try to execute when open
	err := cb.Execute(ctx, func() error {
		return nil
	})

	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreakerTransitionsToHalfOpen(t *testing.T) {
	cfg := &CircuitBreakerConfig{
		MaxRequests:      5,
		Interval:         time.Minute,
		Timeout:          50 * time.Millisecond, // Short timeout for testing
		FailureThreshold: 2,
		SuccessThreshold: 2,
	}
	cb := NewCircuitBreaker("test", cfg)
	ctx := context.Background()

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func() error {
			return errors.New("failure")
		})
	}

	if cb.GetState() != StateOpen {
		t.Fatal("circuit should be open")
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// State should transition to half-open on next request check
	if cb.GetState() != StateHalfOpen {
		t.Errorf("expected state to be HALF-OPEN after timeout, got %s", cb.GetState())
	}
}

func TestCircuitBreakerClosesAfterSuccessInHalfOpen(t *testing.T) {
	cfg := &CircuitBreakerConfig{
		MaxRequests:      5,
		Interval:         time.Minute,
		Timeout:          50 * time.Millisecond,
		FailureThreshold: 2,
		SuccessThreshold: 2,
	}
	cb := NewCircuitBreaker("test", cfg)
	ctx := context.Background()

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func() error {
			return errors.New("failure")
		})
	}

	// Wait for half-open
	time.Sleep(60 * time.Millisecond)

	// Successful requests in half-open should close the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func() error {
			return nil
		})
	}

	if cb.GetState() != StateClosed {
		t.Errorf("expected state to be CLOSED after successful requests, got %s", cb.GetState())
	}
}

func TestCircuitBreakerReopensOnFailureInHalfOpen(t *testing.T) {
	cfg := &CircuitBreakerConfig{
		MaxRequests:      5,
		Interval:         time.Minute,
		Timeout:          50 * time.Millisecond,
		FailureThreshold: 2,
		SuccessThreshold: 2,
	}
	cb := NewCircuitBreaker("test", cfg)
	ctx := context.Background()

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func() error {
			return errors.New("failure")
		})
	}

	// Wait for half-open
	time.Sleep(60 * time.Millisecond)

	// Failure in half-open should reopen
	_ = cb.Execute(ctx, func() error {
		return errors.New("failure in half-open")
	})

	if cb.GetState() != StateOpen {
		t.Errorf("expected state to be OPEN after failure in half-open, got %s", cb.GetState())
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	cfg := &CircuitBreakerConfig{
		MaxRequests:      5,
		Interval:         time.Minute,
		Timeout:          time.Minute,
		FailureThreshold: 2,
		SuccessThreshold: 2,
	}
	cb := NewCircuitBreaker("test", cfg)
	ctx := context.Background()

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func() error {
			return errors.New("failure")
		})
	}

	if cb.GetState() != StateOpen {
		t.Fatal("circuit should be open")
	}

	// Reset the circuit
	cb.Reset()

	if cb.GetState() != StateClosed {
		t.Errorf("expected state to be CLOSED after reset, got %s", cb.GetState())
	}

	if !cb.IsAvailable() {
		t.Error("expected circuit breaker to be available after reset")
	}
}

func TestCircuitBreakerConcurrency(t *testing.T) {
	cfg := &CircuitBreakerConfig{
		MaxRequests:      100,
		Interval:         time.Minute,
		Timeout:          time.Second,
		FailureThreshold: 50,
		SuccessThreshold: 5,
	}
	cb := NewCircuitBreaker("test", cfg)
	ctx := context.Background()

	var wg sync.WaitGroup
	errCount := 0
	var mu sync.Mutex

	// Run many concurrent requests
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			err := cb.Execute(ctx, func() error {
				if n%2 == 0 {
					return errors.New("even failure")
				}
				return nil
			})
			if err != nil {
				mu.Lock()
				errCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Just verify no panics occurred and counts are reasonable
	counts := cb.GetCounts()
	if counts.TotalSuccesses+counts.TotalFailures == 0 {
		t.Error("expected some requests to be processed")
	}
}

func TestCircuitBreakerStateCallback(t *testing.T) {
	cfg := &CircuitBreakerConfig{
		MaxRequests:      5,
		Interval:         time.Minute,
		Timeout:          50 * time.Millisecond,
		FailureThreshold: 2,
		SuccessThreshold: 2,
	}
	cb := NewCircuitBreaker("test", cfg)
	ctx := context.Background()

	stateChanges := make([]State, 0)
	var mu sync.Mutex

	cb.SetOnStateChange(func(name string, from, to State) {
		mu.Lock()
		stateChanges = append(stateChanges, to)
		mu.Unlock()
	})

	// Trigger state changes
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func() error {
			return errors.New("failure")
		})
	}

	time.Sleep(60 * time.Millisecond)
	_ = cb.GetState() // Trigger half-open transition

	time.Sleep(10 * time.Millisecond) // Wait for callback

	mu.Lock()
	if len(stateChanges) == 0 {
		t.Error("expected state change callbacks to be called")
	}
	mu.Unlock()
}

func TestStateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateClosed, "CLOSED"},
		{StateOpen, "OPEN"},
		{StateHalfOpen, "HALF-OPEN"},
		{State(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("State(%d).String() = %s, want %s", tt.state, got, tt.expected)
		}
	}
}
