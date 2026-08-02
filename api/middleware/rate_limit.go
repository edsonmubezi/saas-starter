package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/edsonmubezi/myapp/internal/config"
	redisclient "github.com/edsonmubezi/myapp/pkg/redis"
)

// AuthRateLimiter provides rate limiting functionality for authentication endpoints
type AuthRateLimiter struct {
	config config.SecurityConfig
}

// NewAuthRateLimiter creates a new rate limiter instance
func NewAuthRateLimiter(cfg config.SecurityConfig) *AuthRateLimiter {
	return &AuthRateLimiter{config: cfg}
}

// RateLimitMiddleware returns a middleware that enforces rate limiting
func RateLimitMiddleware(cfg config.SecurityConfig, limitType string) func(http.Handler) http.Handler {
	limiter := NewAuthRateLimiter(cfg)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip rate limiting if disabled
			if !cfg.RateLimitEnabled {
				next.ServeHTTP(w, r)
				return
			}

			// Get identifier (IP + email for login, IP for others)
			identifier := limiter.getIdentifier(r, limitType)

			// Check rate limit
			allowed, resetTime, err := limiter.checkRateLimit(r.Context(), identifier, limitType)
			if err != nil {
				// Redis is down — degrade gracefully rather than blocking all traffic
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.RateLimitLoginAttempts))
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Until(resetTime).Seconds())))

				respondWithJSON(w, http.StatusTooManyRequests, map[string]interface{}{
					"error":      "Too Many Requests",
					"message":    fmt.Sprintf("Rate limit exceeded. Try again in %d seconds", int(time.Until(resetTime).Seconds())),
					"retry_after": int(time.Until(resetTime).Seconds()),
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getIdentifier returns a unique identifier for rate limiting
func (rl *AuthRateLimiter) getIdentifier(r *http.Request, limitType string) string {
	ip := GetClientIP(r)

	switch limitType {
	case "login":
		// For login, use IP + email if available
		if r.Method == "POST" {
			// Try to extract email from body for more accurate rate limiting
			// This is a best-effort attempt, doesn't parse the body
			return fmt.Sprintf("ratelimit:login:%s", ip)
		}
		return fmt.Sprintf("ratelimit:login:%s", ip)
	case "api":
		// For general API, use IP only
		return fmt.Sprintf("ratelimit:api:%s", ip)
	default:
		return fmt.Sprintf("ratelimit:general:%s", ip)
	}
}

// checkRateLimit checks if the request should be allowed based on rate limits
func (rl *AuthRateLimiter) checkRateLimit(ctx context.Context, identifier string, limitType string) (bool, time.Time, error) {
	var (
		maxAttempts int
		window      time.Duration
	)

	// Set limits based on type
	switch limitType {
	case "login":
		maxAttempts = rl.config.RateLimitLoginAttempts
		window = rl.config.RateLimitLoginWindow
	default:
		maxAttempts = rl.config.RateLimitLoginAttempts
		window = rl.config.RateLimitLoginWindow
	}

	// Check if key exists
	exists, err := redisclient.Exists(ctx, identifier)
	if err != nil {
		return false, time.Now(), err
	}

	if !exists {
		// First request, set initial count
		if err := redisclient.Set(ctx, identifier, 1, window); err != nil {
			return false, time.Now(), err
		}
		return true, time.Now().Add(window), nil
	}

	// Get current count
	countStr, err := redisclient.Get(ctx, identifier)
	if err != nil {
		return false, time.Now(), err
	}

	// Parse count
	var count int64
	fmt.Sscanf(countStr, "%d", &count)

	// Check if limit exceeded
	if count >= int64(maxAttempts) {
		// Get TTL for reset time
		ttl, err := redisclient.TTL(ctx, identifier)
		if err != nil {
			return false, time.Now(), err
		}
		return false, time.Now().Add(ttl), nil
	}

	// Increment counter
	newCount, err := redisclient.Incr(ctx, identifier)
	if err != nil {
		return false, time.Now(), err
	}

	// If this was the first increment after creation, set expiry
	if newCount == 2 {
		if err := redisclient.Expire(ctx, identifier, window); err != nil {
			return false, time.Now(), err
		}
	}

	return true, time.Now().Add(window), nil
}

// GetClientIP extracts the real client IP from the request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (set by proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, use the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

// ClearRateLimit removes rate limit for a specific identifier (useful after successful auth)
func ClearRateLimit(ctx context.Context, identifier string) error {
	return redisclient.Del(ctx, identifier)
}

// GetRateLimitStatus returns current rate limit status for an identifier
func GetRateLimitStatus(ctx context.Context, identifier string) (int64, time.Duration, error) {
	countStr, err := redisclient.Get(ctx, identifier)
	if err != nil {
		return 0, 0, nil // No rate limit entry
	}

	var count int64
	fmt.Sscanf(countStr, "%d", &count)

	ttl, err := redisclient.TTL(ctx, identifier)
	if err != nil {
		return count, 0, err
	}

	return count, ttl, nil
}
