package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	// TokenLength is the length of the reset token in bytes (32 bytes = 256 bits)
	TokenLength = 32

	// DefaultTokenExpiry is the default expiration time for reset tokens (30 minutes)
	DefaultTokenExpiry = 30 * time.Minute
)

// ResetTokenService handles password reset token operations
type ResetTokenService struct {
	db *pgxpool.Pool
}

// ResetToken represents a password reset token
type ResetToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	IPAddress string
	UserAgent string
	CreatedAt time.Time
}

// NewResetTokenService creates a new reset token service
func NewResetTokenService(db *pgxpool.Pool) *ResetTokenService {
	return &ResetTokenService{
		db: db,
	}
}

// GenerateToken generates a cryptographically secure random token
func GenerateToken() (string, error) {
	bytes := make([]byte, TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode to base64 URL-safe format
	token := base64.URLEncoding.EncodeToString(bytes)
	return token, nil
}

// HashToken creates a SHA-256 hash of the token for storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}

// CreateResetToken creates a new password reset token for a user
func (s *ResetTokenService) CreateResetToken(
	ctx context.Context,
	userID int64,
	ipAddress string,
	userAgent string,
	expiry time.Duration,
) (string, error) {
	// Generate a new token
	token, err := GenerateToken()
	if err != nil {
		return "", err
	}

	// Hash the token for storage
	tokenHash := HashToken(token)

	// Calculate expiry time
	if expiry == 0 {
		expiry = DefaultTokenExpiry
	}
	expiresAt := time.Now().Add(expiry)

	// Begin transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Invalidate any existing unused tokens for this user
	_, err = tx.Exec(ctx, `
		UPDATE password_reset_tokens
		SET used_at = NOW()
		WHERE user_id = $1 AND used_at IS NULL
	`, userID)
	if err != nil {
		return "", fmt.Errorf("failed to invalidate existing tokens: %w", err)
	}

	// Insert the new token
	_, err = tx.Exec(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, tokenHash, expiresAt, ipAddress, userAgent)
	if err != nil {
		return "", fmt.Errorf("failed to create reset token: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return token, nil
}

// ValidateResetToken validates a reset token and returns the user ID if valid
func (s *ResetTokenService) ValidateResetToken(ctx context.Context, token string) (int64, error) {
	tokenHash := HashToken(token)

	var resetToken ResetToken
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(
		&resetToken.ID,
		&resetToken.UserID,
		&resetToken.TokenHash,
		&resetToken.ExpiresAt,
		&resetToken.UsedAt,
		&resetToken.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return 0, fmt.Errorf("invalid or expired token")
	}
	if err != nil {
		return 0, fmt.Errorf("failed to query token: %w", err)
	}

	// Check if token has already been used
	if resetToken.UsedAt != nil {
		return 0, fmt.Errorf("token has already been used")
	}

	// Check if token has expired
	if time.Now().After(resetToken.ExpiresAt) {
		return 0, fmt.Errorf("token has expired")
	}

	return resetToken.UserID, nil
}

// MarkTokenAsUsed marks a reset token as used
func (s *ResetTokenService) MarkTokenAsUsed(ctx context.Context, token string) error {
	tokenHash := HashToken(token)

	result, err := s.db.Exec(ctx, `
		UPDATE password_reset_tokens
		SET used_at = NOW()
		WHERE token_hash = $1 AND used_at IS NULL
	`, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("token not found or already used")
	}

	return nil
}

// CleanupExpiredTokens removes expired tokens older than the specified duration
func (s *ResetTokenService) CleanupExpiredTokens(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)

	result, err := s.db.Exec(ctx, `
		DELETE FROM password_reset_tokens
		WHERE expires_at < $1
	`, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	return result.RowsAffected(), nil
}

// GetTokenInfo retrieves information about a reset token (for admin purposes)
func (s *ResetTokenService) GetTokenInfo(ctx context.Context, token string) (*ResetToken, error) {
	tokenHash := HashToken(token)

	var resetToken ResetToken
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, used_at, ip_address, user_agent, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(
		&resetToken.ID,
		&resetToken.UserID,
		&resetToken.TokenHash,
		&resetToken.ExpiresAt,
		&resetToken.UsedAt,
		&resetToken.IPAddress,
		&resetToken.UserAgent,
		&resetToken.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query token: %w", err)
	}

	return &resetToken, nil
}
