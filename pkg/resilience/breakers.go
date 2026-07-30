package resilience

import (
	"sync"
	"time"
)

// Service names for circuit breakers
const (
	ServiceDatabase = "database"
	ServiceRedis    = "redis"
	ServiceClamAV   = "clamav"
	ServiceS3       = "s3"
	ServiceSMTP     = "smtp"
)

var (
	once     sync.Once
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
)

// serviceConfigs holds custom configurations for each service
var serviceConfigs = map[string]*CircuitBreakerConfig{
	ServiceDatabase: {
		MaxRequests:      10,
		Interval:         time.Minute,
		Timeout:          15 * time.Second, // Shorter timeout for critical service
		FailureThreshold: 5,
		SuccessThreshold: 3,
	},
	ServiceRedis: {
		MaxRequests:      10,
		Interval:         time.Minute,
		Timeout:          10 * time.Second, // Redis should recover quickly
		FailureThreshold: 5,
		SuccessThreshold: 2,
	},
	ServiceClamAV: {
		MaxRequests:      5,
		Interval:         2 * time.Minute,
		Timeout:          60 * time.Second, // ClamAV may take time to restart
		FailureThreshold: 3,
		SuccessThreshold: 2,
	},
	ServiceS3: {
		MaxRequests:      5,
		Interval:         time.Minute,
		Timeout:          30 * time.Second,
		FailureThreshold: 5,
		SuccessThreshold: 3,
	},
	ServiceSMTP: {
		MaxRequests:      3,
		Interval:         2 * time.Minute,
		Timeout:          60 * time.Second, // Email services may have rate limits
		FailureThreshold: 3,
		SuccessThreshold: 2,
	},
}

// InitializeCircuitBreakers sets up circuit breakers for all services
// This should be called once during application startup
func InitializeCircuitBreakers() {
	once.Do(func() {
		breakers = make(map[string]*CircuitBreaker)

		for service, config := range serviceConfigs {
			breakers[service] = NewCircuitBreaker(service, config)
		}
	})
}

// GetBreaker returns a circuit breaker for the given service
// Returns nil if the service is not registered
func GetBreaker(service string) *CircuitBreaker {
	mu.RLock()
	defer mu.RUnlock()

	if breakers == nil {
		InitializeCircuitBreakers()
	}

	return breakers[service]
}

// GetOrCreateBreaker returns an existing circuit breaker or creates a new one
func GetOrCreateBreaker(service string, config *CircuitBreakerConfig) *CircuitBreaker {
	mu.Lock()
	defer mu.Unlock()

	if breakers == nil {
		breakers = make(map[string]*CircuitBreaker)
	}

	if cb, exists := breakers[service]; exists {
		return cb
	}

	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	cb := NewCircuitBreaker(service, config)
	breakers[service] = cb
	return cb
}

// GetAllBreakers returns a copy of all circuit breakers for monitoring
func GetAllBreakers() map[string]*CircuitBreaker {
	mu.RLock()
	defer mu.RUnlock()

	if breakers == nil {
		return nil
	}

	result := make(map[string]*CircuitBreaker, len(breakers))
	for k, v := range breakers {
		result[k] = v
	}
	return result
}

// GetBreakerStates returns the current state of all circuit breakers
func GetBreakerStates() map[string]BreakerStatus {
	mu.RLock()
	defer mu.RUnlock()

	if breakers == nil {
		return nil
	}

	result := make(map[string]BreakerStatus, len(breakers))
	for name, cb := range breakers {
		counts := cb.GetCounts()
		result[name] = BreakerStatus{
			Name:                 name,
			State:                cb.GetState().String(),
			TotalSuccesses:       counts.TotalSuccesses,
			TotalFailures:        counts.TotalFailures,
			ConsecutiveSuccesses: counts.ConsecutiveSuccesses,
			ConsecutiveFailures:  counts.ConsecutiveFailures,
			IsAvailable:          cb.IsAvailable(),
		}
	}
	return result
}

// BreakerStatus represents the status of a circuit breaker for monitoring
type BreakerStatus struct {
	Name                 string `json:"name"`
	State                string `json:"state"`
	TotalSuccesses       uint32 `json:"total_successes"`
	TotalFailures        uint32 `json:"total_failures"`
	ConsecutiveSuccesses uint32 `json:"consecutive_successes"`
	ConsecutiveFailures  uint32 `json:"consecutive_failures"`
	IsAvailable          bool   `json:"is_available"`
}

// ResetBreaker resets a specific circuit breaker
func ResetBreaker(service string) bool {
	mu.RLock()
	defer mu.RUnlock()

	if cb, exists := breakers[service]; exists {
		cb.Reset()
		return true
	}
	return false
}

// ResetAllBreakers resets all circuit breakers
func ResetAllBreakers() {
	mu.RLock()
	defer mu.RUnlock()

	for _, cb := range breakers {
		cb.Reset()
	}
}
