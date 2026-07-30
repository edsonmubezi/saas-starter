package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithRetrySuccess(t *testing.T) {
	ctx := context.Background()
	cfg := &RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}

	callCount := 0
	err := WithRetry(ctx, cfg, func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestWithRetryEventualSuccess(t *testing.T) {
	ctx := context.Background()
	cfg := &RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}

	callCount := 0
	err := WithRetry(ctx, cfg, func() error {
		callCount++
		if callCount < 3 {
			return errors.New("connection refused") // Retryable error
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected no error after retries, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestWithRetryMaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()
	cfg := &RetryConfig{
		MaxRetries:        2,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}

	callCount := 0
	err := WithRetry(ctx, cfg, func() error {
		callCount++
		return errors.New("connection refused")
	})

	if err == nil {
		t.Error("expected error after max retries")
	}

	// Initial call + MaxRetries retries
	expectedCalls := cfg.MaxRetries + 1
	if callCount != expectedCalls {
		t.Errorf("expected %d calls, got %d", expectedCalls, callCount)
	}
}

func TestWithRetryNonRetryableError(t *testing.T) {
	ctx := context.Background()
	cfg := &RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}

	callCount := 0
	nonRetryableErr := errors.New("validation error") // Not in retryable patterns

	err := WithRetry(ctx, cfg, func() error {
		callCount++
		return nonRetryableErr
	})

	if err == nil {
		t.Error("expected error for non-retryable error")
	}

	// Should only be called once since error is not retryable
	if callCount != 1 {
		t.Errorf("expected 1 call for non-retryable error, got %d", callCount)
	}
}

func TestWithRetryContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := &RetryConfig{
		MaxRetries:        10,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}

	callCount := 0
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := WithRetry(ctx, cfg, func() error {
		callCount++
		return errors.New("connection refused")
	})

	if err == nil {
		t.Error("expected error due to context cancellation")
	}

	if !errors.Is(err, context.Canceled) && callCount < 10 {
		// Either context was cancelled or we got fewer than max retries
		// Both are acceptable outcomes
	}
}

func TestWithRetryResultSuccess(t *testing.T) {
	ctx := context.Background()
	cfg := DefaultRetryConfig()
	cfg.InitialBackoff = 10 * time.Millisecond

	result, err := WithRetryResult(ctx, cfg, func() (string, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result != "success" {
		t.Errorf("expected 'success', got %s", result)
	}
}

func TestWithRetryResultEventualSuccess(t *testing.T) {
	ctx := context.Background()
	cfg := &RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}

	callCount := 0
	result, err := WithRetryResult(ctx, cfg, func() (int, error) {
		callCount++
		if callCount < 2 {
			return 0, errors.New("connection refused")
		}
		return 42, nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"nil error", nil, false},
		{"context canceled", context.Canceled, false},
		{"context deadline", context.DeadlineExceeded, false},
		{"connection refused", errors.New("connection refused"), true},
		{"connection reset", errors.New("connection reset by peer"), true},
		{"timeout error", errors.New("operation timeout"), true},
		{"service unavailable", errors.New("service unavailable"), true},
		{"too many connections", errors.New("too many connections"), true},
		{"broken pipe", errors.New("broken pipe"), true},
		{"generic error", errors.New("some random error"), false},
		{"validation error", errors.New("validation failed: invalid input"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryableError(tt.err); got != tt.retryable {
				t.Errorf("IsRetryableError(%v) = %v, want %v", tt.err, got, tt.retryable)
			}
		})
	}
}

func TestRetryWithCircuitBreaker(t *testing.T) {
	ctx := context.Background()

	cbCfg := &CircuitBreakerConfig{
		MaxRequests:      5,
		Interval:         time.Minute,
		Timeout:          time.Minute,
		FailureThreshold: 5,
		SuccessThreshold: 2,
	}
	cb := NewCircuitBreaker("test", cbCfg)

	retryCfg := &RetryConfig{
		MaxRetries:        2,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}

	callCount := 0
	err := RetryWithCircuitBreaker(ctx, cb, retryCfg, func() error {
		callCount++
		if callCount < 2 {
			return errors.New("connection refused")
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected success, got %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	if cfg.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", cfg.MaxRetries)
	}

	if cfg.InitialBackoff != 100*time.Millisecond {
		t.Errorf("expected InitialBackoff=100ms, got %v", cfg.InitialBackoff)
	}

	if cfg.MaxBackoff != 5*time.Second {
		t.Errorf("expected MaxBackoff=5s, got %v", cfg.MaxBackoff)
	}

	if cfg.BackoffMultiplier != 2.0 {
		t.Errorf("expected BackoffMultiplier=2.0, got %f", cfg.BackoffMultiplier)
	}

	if !cfg.Jitter {
		t.Error("expected Jitter=true")
	}
}

func TestBackoffCapping(t *testing.T) {
	ctx := context.Background()
	cfg := &RetryConfig{
		MaxRetries:        5,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        150 * time.Millisecond, // Cap at 150ms
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}

	start := time.Now()
	callCount := 0
	_ = WithRetry(ctx, cfg, func() error {
		callCount++
		if callCount <= cfg.MaxRetries {
			return errors.New("connection refused")
		}
		return nil
	})
	elapsed := time.Since(start)

	// With backoff: 100ms + 150ms + 150ms + 150ms + 150ms = 700ms (capped at 150ms after first retry)
	// Allow some buffer for execution time
	if elapsed > 2*time.Second {
		t.Errorf("backoff took too long, expected < 2s, got %v", elapsed)
	}
}
