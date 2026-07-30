package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/edsonmubezi/myapp/pkg/database"
	redisclient "github.com/edsonmubezi/myapp/pkg/redis"
	"github.com/edsonmubezi/myapp/pkg/resilience"
)

// ServiceHealth represents the health status of a single service
type ServiceHealth struct {
	Status         string       `json:"status"`
	ResponseTimeMs int64        `json:"response_time_ms"`
	Error          string       `json:"error,omitempty"`
	CircuitBreaker *BreakerInfo `json:"circuit_breaker,omitempty"`
}

// BreakerInfo contains circuit breaker information
type BreakerInfo struct {
	State       string `json:"state"`
	IsAvailable bool   `json:"is_available"`
}

// PoolStats contains database pool statistics
type PoolStats struct {
	TotalConns    int32 `json:"total_conns"`
	AcquiredConns int32 `json:"acquired_conns"`
	IdleConns     int32 `json:"idle_conns"`
	MaxConns      int32 `json:"max_conns"`
}

// SystemInfo contains system resource information
type SystemInfo struct {
	NumGoroutines int    `json:"num_goroutines"`
	NumCPU        int    `json:"num_cpu"`
	GoVersion     string `json:"go_version"`
	MemAllocMB    uint64 `json:"mem_alloc_mb"`
	MemSysMB      uint64 `json:"mem_sys_mb"`
}

// AdvancedHealthResponse is the response for advanced health check
type AdvancedHealthResponse struct {
	Status          string                              `json:"status"`
	Timestamp       time.Time                           `json:"timestamp"`
	Services        map[string]ServiceHealth            `json:"services"`
	CircuitBreakers map[string]resilience.BreakerStatus `json:"circuit_breakers,omitempty"`
	DBPool          *PoolStats                          `json:"db_pool,omitempty"`
	System          *SystemInfo                         `json:"system,omitempty"`
}

// @Summary      Advanced Health Check
// @Description  Comprehensive health check with service statuses, circuit breaker states, and system metrics
// @Tags         System
// @Produce      json
// @Param        detailed query bool false "Include detailed metrics (db pool stats, system info)"
// @Success      200 {object} AdvancedHealthResponse "All services healthy"
// @Failure      503 {object} AdvancedHealthResponse "One or more critical services unhealthy"
// @Router       /health/detailed [get]
func AdvancedHealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	includeDetailed := r.URL.Query().Get("detailed") == "true"

	response := AdvancedHealthResponse{
		Timestamp: time.Now().UTC(),
		Services:  make(map[string]ServiceHealth),
	}

	overallHealthy := true

	// Check database health
	dbHealth := checkDatabaseHealth(ctx)
	response.Services["database"] = dbHealth
	if dbHealth.Status != "healthy" {
		overallHealthy = false
	}

	// Check Redis health (optional service)
	redisHealth := checkRedisHealth(ctx)
	response.Services["redis"] = redisHealth
	// Redis is optional, don't fail overall health if unavailable

	// Get circuit breaker states
	response.CircuitBreakers = resilience.GetBreakerStates()

	// Include detailed metrics if requested
	if includeDetailed {
		// Database pool stats
		if stats := database.GetPoolStats(); stats != nil {
			response.DBPool = &PoolStats{
				TotalConns:    stats.TotalConns(),
				AcquiredConns: stats.AcquiredConns(),
				IdleConns:     stats.IdleConns(),
				MaxConns:      stats.MaxConns(),
			}
		}

		// System info
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		response.System = &SystemInfo{
			NumGoroutines: runtime.NumGoroutine(),
			NumCPU:        runtime.NumCPU(),
			GoVersion:     runtime.Version(),
			MemAllocMB:    memStats.Alloc / 1024 / 1024,
			MemSysMB:      memStats.Sys / 1024 / 1024,
		}
	}

	// Set overall status
	if overallHealthy {
		response.Status = "healthy"
		w.WriteHeader(http.StatusOK)
	} else {
		response.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// checkDatabaseHealth checks the database connection health
func checkDatabaseHealth(ctx context.Context) ServiceHealth {
	health := ServiceHealth{
		Status: "healthy",
	}

	// Get circuit breaker info
	if cb := resilience.GetBreaker(resilience.ServiceDatabase); cb != nil {
		health.CircuitBreaker = &BreakerInfo{
			State:       cb.GetState().String(),
			IsAvailable: cb.IsAvailable(),
		}

		// If circuit breaker is open, report unhealthy
		if !cb.IsAvailable() {
			health.Status = "unhealthy"
			health.Error = "circuit breaker open"
			return health
		}
	}

	if database.DB == nil {
		health.Status = "unhealthy"
		health.Error = "database connection not initialized"
		return health
	}

	start := time.Now()
	err := database.DB.Ping(ctx)
	health.ResponseTimeMs = time.Since(start).Milliseconds()

	if err != nil {
		health.Status = "unhealthy"
		health.Error = "database ping failed"
		return health
	}

	return health
}

// checkRedisHealth checks the Redis connection health
func checkRedisHealth(ctx context.Context) ServiceHealth {
	health := ServiceHealth{
		Status: "healthy",
	}

	// Get circuit breaker info
	if cb := resilience.GetBreaker(resilience.ServiceRedis); cb != nil {
		health.CircuitBreaker = &BreakerInfo{
			State:       cb.GetState().String(),
			IsAvailable: cb.IsAvailable(),
		}
	}

	client := redisclient.GetClient()
	if client == nil {
		health.Status = "unavailable"
		health.Error = "redis client not initialized"
		return health
	}

	start := time.Now()
	_, err := client.Ping(ctx).Result()
	health.ResponseTimeMs = time.Since(start).Milliseconds()

	if err != nil {
		health.Status = "unhealthy"
		health.Error = "redis ping failed"
		return health
	}

	return health
}

// @Summary      Liveness Check
// @Description  Simple liveness probe for Kubernetes/container orchestration. Returns 200 if the application is running.
// @Tags         System
// @Produce      json
// @Success      200 {object} map[string]string "Application is alive"
// @Router       /health/live [get]
func LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "alive",
	})
}

// @Summary      Readiness Check
// @Description  Readiness probe for Kubernetes/container orchestration. Returns 200 if the application can accept traffic.
// @Tags         System
// @Produce      json
// @Success      200 {object} map[string]string "Application is ready"
// @Failure      503 {object} map[string]string "Application is not ready"
// @Router       /health/ready [get]
func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check if database is accessible
	if database.DB == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "not_ready",
			"reason": "database not initialized",
		})
		return
	}

	if err := database.DB.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "not_ready",
			"reason": "database unavailable",
		})
		return
	}

	// Check if critical circuit breakers are open
	if cb := resilience.GetBreaker(resilience.ServiceDatabase); cb != nil && !cb.IsAvailable() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "not_ready",
			"reason": "database circuit breaker open",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
	})
}

// @Summary      Circuit Breaker Status
// @Description  Returns the status of all circuit breakers
// @Tags         System
// @Produce      json
// @Success      200 {object} map[string]resilience.BreakerStatus "Circuit breaker statuses"
// @Router       /health/circuit-breakers [get]
func CircuitBreakerStatusHandler(w http.ResponseWriter, r *http.Request) {
	states := resilience.GetBreakerStates()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"circuit_breakers": states,
	})
}

// @Summary      Reset Circuit Breaker
// @Description  Manually reset a circuit breaker (admin operation)
// @Tags         System
// @Produce      json
// @Param        service path string true "Service name (database, redis, clamav, s3, smtp)"
// @Success      200 {object} map[string]string "Circuit breaker reset"
// @Failure      404 {object} map[string]string "Circuit breaker not found"
// @Router       /health/circuit-breakers/{service}/reset [post]
func ResetCircuitBreakerHandler(w http.ResponseWriter, r *http.Request) {
	// Extract service name from URL path
	segments := splitPath(r.URL.Path)
	if len(segments) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "service name required",
		})
		return
	}

	service := segments[len(segments)-2] // e.g., /health/circuit-breakers/redis/reset -> redis

	if resilience.ResetBreaker(service) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "reset",
			"service": service,
		})
		return
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "circuit breaker not found",
		"service": service,
	})
}

// splitPath splits a URL path into segments
func splitPath(path string) []string {
	var segments []string
	for _, s := range split(path, '/') {
		if s != "" {
			segments = append(segments, s)
		}
	}
	return segments
}

// split splits a string by a separator
func split(s string, sep rune) []string {
	var result []string
	var current string
	for _, r := range s {
		if r == sep {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
