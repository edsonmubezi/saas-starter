package middleware

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/didip/tollbooth/v8"
)

var (
	requestCounts   = make(map[string]*requestInfo)
	requestCountsMu sync.Mutex
)

type requestInfo struct {
	count      int
	lastAccess time.Time
}

func init() {
	// Start cleanup goroutine to prevent memory leak
	go cleanupRequestCounts()
}

// cleanupRequestCounts removes stale entries from the map every 5 minutes
func cleanupRequestCounts() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		requestCountsMu.Lock()
		now := time.Now()
		for ip, info := range requestCounts {
			// Remove entries older than 10 minutes
			if now.Sub(info.lastAccess) > 10*time.Minute {
				delete(requestCounts, ip)
			}
		}
		requestCountsMu.Unlock()
	}
}

// RateLimiter returns middleware with rate limiting and request counting per IP
func RateLimiter(reqsPerSecond float64, burst int) func(http.Handler) http.Handler {
	limiter := tollbooth.NewLimiter(reqsPerSecond, nil)
	limiter.SetBurst(burst)

	log.Printf("[RateLimiter] Initialized: %.2f req/s, burst=%d", reqsPerSecond, burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractClientIP(r)

			// Increment count safely and update last access time
			requestCountsMu.Lock()
			if info, exists := requestCounts[ip]; exists {
				info.count++
				info.lastAccess = time.Now()
			} else {
				requestCounts[ip] = &requestInfo{
					count:      1,
					lastAccess: time.Now(),
				}
			}
			count := requestCounts[ip].count
			requestCountsMu.Unlock()

			// REMOVED: Excessive per-request logging
			// Only log when rate limit is exceeded

			httpError := tollbooth.LimitByRequest(limiter, w, r)
			if httpError != nil {
				// Only log when rate limit is BLOCKED
				log.Printf("[RateLimiter] BLOCKED | IP: %s | Path: %s | Count: %d",
					ip, r.URL.Path, count)

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60") // Add retry-after header
				w.WriteHeader(http.StatusTooManyRequests)
				// Reduced information disclosure - don't expose IP, count, etc to client
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "Too many requests. Please try again later.",
					"status":  429,
					"message": "Rate limit exceeded",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractClientIP returns the client IP from headers or RemoteAddr
func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	if rip := r.Header.Get("X-Real-IP"); rip != "" {
		return strings.TrimSpace(rip)
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
