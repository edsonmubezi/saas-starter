package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	redisclient "github.com/edsonmubezi/myapp/pkg/redis"
	"github.com/edsonmubezi/myapp/pkg/resilience"
)

// ErrCacheMiss signals the key was not found.
var ErrCacheMiss = errors.New("cache: miss")

// CachedAuthData stores the result of the 4-query bundle in JWTMiddleware.
type CachedAuthData struct {
	ActiveStatus    bool     `json:"active_status"`
	IsLocked        bool     `json:"is_locked"`
	RoleName        string   `json:"role_name"`
	PermissionNames []string `json:"permission_names"`
	OrgName         string   `json:"org_name"`
}

// ---- Generic cache operations ----

// Get retrieves a value from the cache, JSON-deserializing into *T.
// Returns ErrCacheMiss if the key doesn't exist.
// Callers should treat all errors as "cache miss" and fall through to DB.
func Get[T any](ctx context.Context, key string) (*T, error) {
	raw, err := redisclient.Get(ctx, key)
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, ErrCacheMiss
		}
		return nil, err
	}
	var result T
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		// Corrupted cache entry — delete it and treat as miss
		_ = redisclient.Del(ctx, key)
		return nil, ErrCacheMiss
	}
	return &result, nil
}

// Set serializes value as JSON and stores it in Redis with the given TTL.
// Errors are logged but NOT returned — caching is best-effort.
func Set(ctx context.Context, key string, value any, ttl time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("cache.Set: marshal error for key %s: %v", key, err)
		return
	}
	if err := redisclient.Set(ctx, key, string(data), ttl); err != nil {
		log.Printf("cache.Set: write error for key %s: %v", key, err)
	}
}

// Del removes one or more keys from the cache.
// Errors are logged but NOT returned.
func Del(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	if err := redisclient.Del(ctx, keys...); err != nil {
		log.Printf("cache.Del: error deleting keys: %v", err)
	}
}

// DelByPattern scans for keys matching a glob pattern and deletes them.
// Uses SCAN to avoid blocking Redis with KEYS command.
// Errors are logged but NOT returned.
func DelByPattern(ctx context.Context, pattern string) {
	client := redisclient.GetClient()
	if client == nil {
		return
	}

	cb := resilience.GetBreaker(resilience.ServiceRedis)

	deleteFn := func() error {
		var cursor uint64
		for {
			keys, nextCursor, err := client.Scan(ctx, cursor, pattern, 100).Result()
			if err != nil {
				return err
			}
			if len(keys) > 0 {
				if err := client.Del(ctx, keys...).Err(); err != nil {
					return err
				}
			}
			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}
		return nil
	}

	var err error
	if cb != nil {
		err = cb.Execute(ctx, deleteFn)
	} else {
		err = deleteFn()
	}
	if err != nil {
		log.Printf("cache.DelByPattern: error for pattern %s: %v", pattern, err)
	}
}

// ---- Key helpers ----

func AuthUserKey(userID, orgID int64) string {
	return fmt.Sprintf("auth:user:%d:%d", userID, orgID)
}

func AuthUserOrgPattern(orgID int64) string {
	return fmt.Sprintf("auth:user:*:%d", orgID)
}

func AuthUserPattern(userID int64) string {
	return fmt.Sprintf("auth:user:%d:*", userID)
}

func DepartmentsKey(orgID int64) string {
	return fmt.Sprintf("cache:departments:%d", orgID)
}

func PositionsKey(orgID int64) string {
	return fmt.Sprintf("cache:positions:%d", orgID)
}

func DesignationsKey(orgID int64) string {
	return fmt.Sprintf("cache:designations:%d", orgID)
}

func OrgSettingsKey(orgID int64) string {
	return fmt.Sprintf("cache:org_settings:%d", orgID)
}

func LeaveTypesKey(orgID int64) string {
	return fmt.Sprintf("cache:leave_types:%d", orgID)
}

// ---- Context helpers for passing cached permissions between middlewares ----

type permCtxKey string

const cachedPermsKey permCtxKey = "cachedPermissions"

// WithCachedPermissions stores permission names in context.
func WithCachedPermissions(ctx context.Context, perms []string) context.Context {
	return context.WithValue(ctx, cachedPermsKey, perms)
}

// GetCachedPermissions retrieves permission names from context.
func GetCachedPermissions(ctx context.Context) ([]string, bool) {
	val := ctx.Value(cachedPermsKey)
	if val == nil {
		return nil, false
	}
	perms, ok := val.([]string)
	return perms, ok
}
