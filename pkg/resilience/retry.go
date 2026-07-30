package resilience

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/jackc/pgconn"
)

// RetryConfig holds configuration for retry logic
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// InitialBackoff is the initial wait time before first retry
	InitialBackoff time.Duration
	// MaxBackoff is the maximum wait time between retries
	MaxBackoff time.Duration
	// BackoffMultiplier is the factor to multiply backoff after each retry
	BackoffMultiplier float64
	// Jitter adds randomness to backoff to prevent thundering herd
	Jitter bool
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        5 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            true,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// RetryableFuncWithResult is a function that returns a result and can be retried
type RetryableFuncWithResult[T any] func() (T, error)

// IsRetryableError determines if an error is retryable
// This checks for transient errors that may succeed on retry
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Don't retry context cancellation or deadline exceeded
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for PostgreSQL-specific retryable errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40001": // serialization_failure
			return true
		case "40P01": // deadlock_detected
			return true
		case "53000": // insufficient_resources
			return true
		case "53100": // disk_full
			return false // Not retryable
		case "53200": // out_of_memory
			return true
		case "53300": // too_many_connections
			return true
		case "57P01": // admin_shutdown
			return true
		case "57P02": // crash_shutdown
			return true
		case "57P03": // cannot_connect_now
			return true
		case "08000": // connection_exception
			return true
		case "08003": // connection_does_not_exist
			return true
		case "08006": // connection_failure
			return true
		case "08001": // sqlclient_unable_to_establish_sqlconnection
			return true
		case "08004": // sqlserver_rejected_establishment_of_sqlconnection
			return true
		}
	}

	// Check for common transient error messages
	errMsg := err.Error()
	transientPatterns := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"timeout",
		"temporary failure",
		"try again",
		"service unavailable",
		"too many connections",
		"network is unreachable",
		"no route to host",
	}

	for _, pattern := range transientPatterns {
		if containsIgnoreCase(errMsg, pattern) {
			return true
		}
	}

	return false
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	return strings.Contains(sLower, substrLower)
}

// WithRetry executes a function with retry logic
func WithRetry(ctx context.Context, cfg *RetryConfig, fn RetryableFunc) error {
	if cfg == nil {
		cfg = DefaultRetryConfig()
	}

	var lastErr error
	backoff := cfg.InitialBackoff

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check context before attempting
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		// Wait before retry (skip for first attempt)
		if attempt > 0 {
			waitTime := backoff
			if cfg.Jitter {
				// Add up to 25% jitter
				jitter := time.Duration(rand.Int63n(int64(backoff / 4)))
				waitTime = backoff + jitter
			}

			log.Printf("⟳ Retry attempt %d/%d for operation after %v", attempt, cfg.MaxRetries, waitTime)

			select {
			case <-ctx.Done():
				return fmt.Errorf("retry cancelled during backoff: %w", ctx.Err())
			case <-time.After(waitTime):
			}

			// Calculate next backoff with exponential increase
			backoff = time.Duration(float64(backoff) * cfg.BackoffMultiplier)
			if backoff > cfg.MaxBackoff {
				backoff = cfg.MaxBackoff
			}
		}

		// Execute the function
		lastErr = fn()
		if lastErr == nil {
			if attempt > 0 {
				log.Printf("✓ Operation succeeded after %d retries", attempt)
			}
			return nil
		}

		// Check if error is retryable
		if !IsRetryableError(lastErr) {
			log.Printf("✗ Non-retryable error encountered: %v", lastErr)
			return lastErr
		}

		log.Printf("⚠ Retryable error (attempt %d/%d): %v", attempt+1, cfg.MaxRetries+1, lastErr)
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxRetries, lastErr)
}

// WithRetryResult executes a function that returns a result with retry logic
func WithRetryResult[T any](ctx context.Context, cfg *RetryConfig, fn RetryableFuncWithResult[T]) (T, error) {
	if cfg == nil {
		cfg = DefaultRetryConfig()
	}

	var zero T
	var lastErr error
	var result T
	backoff := cfg.InitialBackoff

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check context before attempting
		select {
		case <-ctx.Done():
			return zero, fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		// Wait before retry (skip for first attempt)
		if attempt > 0 {
			waitTime := backoff
			if cfg.Jitter {
				jitter := time.Duration(rand.Int63n(int64(backoff / 4)))
				waitTime = backoff + jitter
			}

			log.Printf("⟳ Retry attempt %d/%d after %v", attempt, cfg.MaxRetries, waitTime)

			select {
			case <-ctx.Done():
				return zero, fmt.Errorf("retry cancelled during backoff: %w", ctx.Err())
			case <-time.After(waitTime):
			}

			backoff = time.Duration(float64(backoff) * cfg.BackoffMultiplier)
			if backoff > cfg.MaxBackoff {
				backoff = cfg.MaxBackoff
			}
		}

		// Execute the function
		result, lastErr = fn()
		if lastErr == nil {
			if attempt > 0 {
				log.Printf("✓ Operation succeeded after %d retries", attempt)
			}
			return result, nil
		}

		if !IsRetryableError(lastErr) {
			return zero, lastErr
		}

		log.Printf("⚠ Retryable error (attempt %d/%d): %v", attempt+1, cfg.MaxRetries+1, lastErr)
	}

	return zero, fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxRetries, lastErr)
}

// RetryWithCircuitBreaker combines retry logic with circuit breaker
func RetryWithCircuitBreaker(ctx context.Context, cb *CircuitBreaker, retryCfg *RetryConfig, fn RetryableFunc) error {
	return WithRetry(ctx, retryCfg, func() error {
		return cb.Execute(ctx, fn)
	})
}

// RetryWithCircuitBreakerResult combines retry logic with circuit breaker for functions with results
func RetryWithCircuitBreakerResult[T any](ctx context.Context, cb *CircuitBreaker, retryCfg *RetryConfig, fn RetryableFuncWithResult[T]) (T, error) {
	return WithRetryResult(ctx, retryCfg, func() (T, error) {
		var result T
		var resultErr error
		err := cb.Execute(ctx, func() error {
			result, resultErr = fn()
			return resultErr
		})
		if err != nil && err != resultErr {
			// Circuit breaker error (open/too many requests)
			var zero T
			return zero, err
		}
		return result, resultErr
	})
}
