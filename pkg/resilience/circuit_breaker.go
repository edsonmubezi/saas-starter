package resilience

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is in open state
	ErrCircuitOpen = errors.New("circuit breaker is open")
	// ErrTooManyRequests is returned when too many requests are made in half-open state
	ErrTooManyRequests = errors.New("circuit breaker: too many requests in half-open state")
)

// State represents the circuit breaker state
type State int

const (
	// StateClosed allows requests to pass through
	StateClosed State = iota
	// StateHalfOpen allows limited requests to test if service recovered
	StateHalfOpen
	// StateOpen blocks all requests
	StateOpen
)

// String returns the string representation of the state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF-OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

// Counts holds the statistics for circuit breaker
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// CircuitBreakerConfig holds configuration for a circuit breaker
type CircuitBreakerConfig struct {
	// MaxRequests is the maximum number of requests allowed in half-open state
	MaxRequests uint32
	// Interval is the cyclic period for clearing counts in closed state
	Interval time.Duration
	// Timeout is the period of open state before transitioning to half-open
	Timeout time.Duration
	// FailureThreshold is the number of consecutive failures to open the circuit
	FailureThreshold uint32
	// SuccessThreshold is the number of consecutive successes to close from half-open
	SuccessThreshold uint32
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		MaxRequests:      5,
		Interval:         time.Minute,
		Timeout:          30 * time.Second,
		FailureThreshold: 5,
		SuccessThreshold: 3,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name   string
	config *CircuitBreakerConfig

	mu     sync.RWMutex
	state  State
	counts *Counts
	expiry time.Time

	// Callbacks for monitoring
	onStateChange func(name string, from, to State)
}

// NewCircuitBreaker creates a new circuit breaker with the given name
func NewCircuitBreaker(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	cb := &CircuitBreaker{
		name:   name,
		config: config,
		state:  StateClosed,
		counts: &Counts{},
	}

	return cb
}

// SetOnStateChange sets a callback for state changes
func (cb *CircuitBreaker) SetOnStateChange(fn func(name string, from, to State)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStateChange = fn
}

// Name returns the circuit breaker name
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// Execute runs the given function if the circuit breaker allows it
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if we can proceed
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Execute the function
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Record the result
	cb.afterRequest(err, duration)

	return err
}

// beforeRequest checks if the request can proceed
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state := cb.currentState(now)

	switch state {
	case StateOpen:
		return ErrCircuitOpen
	case StateHalfOpen:
		if cb.counts.Requests >= cb.config.MaxRequests {
			return ErrTooManyRequests
		}
	}

	cb.counts.Requests++
	return nil
}

// afterRequest records the result of a request
func (cb *CircuitBreaker) afterRequest(err error, duration time.Duration) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state := cb.currentState(now)

	if err != nil {
		cb.onFailure(state, now)
	} else {
		cb.onSuccess(state, now)
	}
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess(state State, now time.Time) {
	cb.counts.TotalSuccesses++
	cb.counts.ConsecutiveSuccesses++
	cb.counts.ConsecutiveFailures = 0

	if state == StateHalfOpen {
		if cb.counts.ConsecutiveSuccesses >= cb.config.SuccessThreshold {
			cb.setState(StateClosed, now)
		}
	}
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure(state State, now time.Time) {
	cb.counts.TotalFailures++
	cb.counts.ConsecutiveFailures++
	cb.counts.ConsecutiveSuccesses = 0

	switch state {
	case StateClosed:
		if cb.counts.ConsecutiveFailures >= cb.config.FailureThreshold {
			cb.setState(StateOpen, now)
		}
	case StateHalfOpen:
		cb.setState(StateOpen, now)
	}
}

// currentState returns the current state, handling time-based transitions
func (cb *CircuitBreaker) currentState(now time.Time) State {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.resetCounts()
			cb.expiry = now.Add(cb.config.Interval)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state
}

// setState transitions to a new state
func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state
	cb.resetCounts()

	switch state {
	case StateClosed:
		cb.expiry = now.Add(cb.config.Interval)
		log.Printf("🟢 Circuit breaker [%s]: %s → %s (recovered)", cb.name, prev, state)
	case StateOpen:
		cb.expiry = now.Add(cb.config.Timeout)
		log.Printf("🔴 Circuit breaker [%s]: %s → %s (failures: %d)",
			cb.name, prev, state, cb.counts.ConsecutiveFailures)
	case StateHalfOpen:
		cb.expiry = time.Time{}
		log.Printf("🟡 Circuit breaker [%s]: %s → %s (testing recovery)", cb.name, prev, state)
	}

	// Call the state change callback if set
	if cb.onStateChange != nil {
		go cb.onStateChange(cb.name, prev, state)
	}
}

// resetCounts resets the statistics
func (cb *CircuitBreaker) resetCounts() {
	cb.counts = &Counts{}
}

// GetState returns the current circuit breaker state (thread-safe)
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.currentState(time.Now())
}

// GetCounts returns a copy of current counts (thread-safe)
func (cb *CircuitBreaker) GetCounts() Counts {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return *cb.counts
}

// IsAvailable returns true if the circuit breaker allows requests
func (cb *CircuitBreaker) IsAvailable() bool {
	state := cb.GetState()
	return state != StateOpen
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.resetCounts()
	cb.expiry = time.Now().Add(cb.config.Interval)
	log.Printf("🔄 Circuit breaker [%s]: manually reset to CLOSED", cb.name)
}
