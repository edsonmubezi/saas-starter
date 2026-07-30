package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/edsonmubezi/myapp/internal/config"
	"github.com/edsonmubezi/myapp/pkg/resilience"
	"github.com/go-redis/redis/v8"
)

var Client *redis.Client

// ErrRedisUnavailable is returned when Redis is not available due to circuit breaker
var ErrRedisUnavailable = errors.New("redis service unavailable")

// Initialize creates and tests a Redis client connection
func Initialize(cfg config.RedisConfig) error {
	if !cfg.Enabled {
		log.Println("Redis is disabled, skipping initialization")
		return nil
	}

	// Initialize circuit breakers if not already done
	resilience.InitializeCircuitBreakers()

	Client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
		PoolTimeout:  4 * time.Second,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("✓ Successfully connected to Redis at %s:%d", cfg.Host, cfg.Port)
	return nil
}

// GetClient returns the Redis client (for direct access when needed)
func GetClient() *redis.Client {
	return Client
}

// Close gracefully closes the Redis client connection
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// Get retrieves a value from Redis with circuit breaker protection
func Get(ctx context.Context, key string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("Redis client not initialized")
	}

	cb := resilience.GetBreaker(resilience.ServiceRedis)
	if cb == nil {
		return Client.Get(ctx, key).Result()
	}

	var result string
	err := cb.Execute(ctx, func() error {
		var innerErr error
		result, innerErr = Client.Get(ctx, key).Result()
		return innerErr
	})

	if errors.Is(err, resilience.ErrCircuitOpen) {
		return "", ErrRedisUnavailable
	}

	return result, err
}

// Set stores a value in Redis with optional expiration and circuit breaker protection
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if Client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	cb := resilience.GetBreaker(resilience.ServiceRedis)
	if cb == nil {
		return Client.Set(ctx, key, value, expiration).Err()
	}

	err := cb.Execute(ctx, func() error {
		return Client.Set(ctx, key, value, expiration).Err()
	})

	if errors.Is(err, resilience.ErrCircuitOpen) {
		return ErrRedisUnavailable
	}

	return err
}

// Del deletes a key from Redis with circuit breaker protection
func Del(ctx context.Context, keys ...string) error {
	if Client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	cb := resilience.GetBreaker(resilience.ServiceRedis)
	if cb == nil {
		return Client.Del(ctx, keys...).Err()
	}

	err := cb.Execute(ctx, func() error {
		return Client.Del(ctx, keys...).Err()
	})

	if errors.Is(err, resilience.ErrCircuitOpen) {
		return ErrRedisUnavailable
	}

	return err
}

// Exists checks if a key exists in Redis with circuit breaker protection
func Exists(ctx context.Context, key string) (bool, error) {
	if Client == nil {
		return false, fmt.Errorf("Redis client not initialized")
	}

	cb := resilience.GetBreaker(resilience.ServiceRedis)
	if cb == nil {
		result, err := Client.Exists(ctx, key).Result()
		return result > 0, err
	}

	var exists bool
	err := cb.Execute(ctx, func() error {
		result, innerErr := Client.Exists(ctx, key).Result()
		exists = result > 0
		return innerErr
	})

	if errors.Is(err, resilience.ErrCircuitOpen) {
		return false, ErrRedisUnavailable
	}

	return exists, err
}

// Incr increments the value of a key with circuit breaker protection
func Incr(ctx context.Context, key string) (int64, error) {
	if Client == nil {
		return 0, fmt.Errorf("Redis client not initialized")
	}

	cb := resilience.GetBreaker(resilience.ServiceRedis)
	if cb == nil {
		return Client.Incr(ctx, key).Result()
	}

	var result int64
	err := cb.Execute(ctx, func() error {
		var innerErr error
		result, innerErr = Client.Incr(ctx, key).Result()
		return innerErr
	})

	if errors.Is(err, resilience.ErrCircuitOpen) {
		return 0, ErrRedisUnavailable
	}

	return result, err
}

// Expire sets an expiration time on a key with circuit breaker protection
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	if Client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	cb := resilience.GetBreaker(resilience.ServiceRedis)
	if cb == nil {
		return Client.Expire(ctx, key, expiration).Err()
	}

	err := cb.Execute(ctx, func() error {
		return Client.Expire(ctx, key, expiration).Err()
	})

	if errors.Is(err, resilience.ErrCircuitOpen) {
		return ErrRedisUnavailable
	}

	return err
}

// TTL returns the remaining time to live of a key with circuit breaker protection
func TTL(ctx context.Context, key string) (time.Duration, error) {
	if Client == nil {
		return 0, fmt.Errorf("Redis client not initialized")
	}

	cb := resilience.GetBreaker(resilience.ServiceRedis)
	if cb == nil {
		return Client.TTL(ctx, key).Result()
	}

	var result time.Duration
	err := cb.Execute(ctx, func() error {
		var innerErr error
		result, innerErr = Client.TTL(ctx, key).Result()
		return innerErr
	})

	if errors.Is(err, resilience.ErrCircuitOpen) {
		return 0, ErrRedisUnavailable
	}

	return result, err
}

// ZAdd adds members to a sorted set
func ZAdd(ctx context.Context, key string, members ...*redis.Z) error {
	if Client == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	return Client.ZAdd(ctx, key, members...).Err()
}

// ZRangeByScore retrieves members from a sorted set by score range
func ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	if Client == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}
	return Client.ZRangeByScore(ctx, key, opt).Result()
}

// ZRemRangeByScore removes members from a sorted set by score range
func ZRemRangeByScore(ctx context.Context, key, min, max string) error {
	if Client == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	return Client.ZRemRangeByScore(ctx, key, min, max).Err()
}
