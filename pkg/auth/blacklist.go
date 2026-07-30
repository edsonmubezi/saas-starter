package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	redisclient "github.com/edsonmubezi/myapp/pkg/redis"
)

const (
	blacklistPrefix = "token:blacklist:"
	// Additional buffer to ensure expired tokens are cleaned up
	blacklistExpiryBuffer = 1 * time.Hour
)

// BlacklistToken adds a token to the blacklist
func BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error {
	// Hash the token for privacy (don't store full JWTs in Redis)
	tokenHash := hashToken(token)
	key := blacklistPrefix + tokenHash

	// Calculate TTL (time until token expires + buffer)
	ttl := time.Until(expiresAt) + blacklistExpiryBuffer
	if ttl < 0 {
		// Token already expired, no need to blacklist
		return nil
	}

	// Store in Redis with automatic expiration
	if err := redisclient.Set(ctx, key, "blacklisted", ttl); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// IsTokenBlacklisted checks if a token is in the blacklist
func IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	tokenHash := hashToken(token)
	key := blacklistPrefix + tokenHash

	exists, err := redisclient.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return exists, nil
}

// RemoveFromBlacklist removes a token from the blacklist (admin override)
func RemoveFromBlacklist(ctx context.Context, token string) error {
	tokenHash := hashToken(token)
	key := blacklistPrefix + tokenHash

	if err := redisclient.Del(ctx, key); err != nil {
		return fmt.Errorf("failed to remove token from blacklist: %w", err)
	}

	return nil
}

// BlacklistUserTokens blacklists all tokens for a specific user
// This should be called on password change, account lockout, etc.
func BlacklistUserTokens(ctx context.Context, userID int64) error {
	// Store a user-level blacklist marker
	key := fmt.Sprintf("token:user_blacklist:%d", userID)

	// Set with a long TTL (7 days - max refresh token lifetime)
	ttl := 7 * 24 * time.Hour

	if err := redisclient.Set(ctx, key, time.Now().Unix(), ttl); err != nil {
		return fmt.Errorf("failed to blacklist user tokens: %w", err)
	}

	return nil
}

// IsUserTokensBlacklisted checks if all tokens for a user are blacklisted
func IsUserTokensBlacklisted(ctx context.Context, userID int64, tokenIssuedAt time.Time) (bool, error) {
	key := fmt.Sprintf("token:user_blacklist:%d", userID)

	blacklistTimeStr, err := redisclient.Get(ctx, key)
	if err != nil {
		// Key doesn't exist, user tokens not blacklisted
		if err.Error() == "redis: nil" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check user token blacklist: %w", err)
	}

	// Parse blacklist timestamp
	var blacklistTime int64
	fmt.Sscanf(blacklistTimeStr, "%d", &blacklistTime)

	// If token was issued before the blacklist time, it's invalid
	if tokenIssuedAt.Unix() < blacklistTime {
		return true, nil
	}

	return false, nil
}

// ClearUserTokenBlacklist removes the user-level blacklist
func ClearUserTokenBlacklist(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("token:user_blacklist:%d", userID)

	if err := redisclient.Del(ctx, key); err != nil {
		return fmt.Errorf("failed to clear user token blacklist: %w", err)
	}

	return nil
}

// hashToken creates a SHA-256 hash of the token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CleanupExpiredBlacklist manually removes expired blacklist entries
// This is typically not needed as Redis handles expiration automatically
func CleanupExpiredBlacklist(ctx context.Context) error {
	// Redis automatically removes expired keys, but we can trigger
	// a scan and delete if needed for manual cleanup
	// For now, this is a no-op as Redis handles it automatically
	return nil
}

// GetBlacklistStats returns statistics about the token blacklist
func GetBlacklistStats(ctx context.Context) (map[string]interface{}, error) {
	// Note: This is expensive for large blacklists, use sparingly
	stats := make(map[string]interface{})

	// This would require a SCAN operation in Redis which can be expensive
	// For production, consider using Redis INFO command instead
	stats["warning"] = "Detailed stats not implemented to avoid performance impact"

	return stats, nil
}
