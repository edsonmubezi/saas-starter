package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// RefreshTokenRepository handles refresh token database operations
type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

// RefreshToken represents a stored refresh token
type RefreshToken struct {
	ID                 int64
	UserID             int64
	TokenFingerprint   string
	IssuedAt           time.Time
	ExpiresAt          time.Time
	RevokedAt          *time.Time
	ReplacedByTokenID  *int64
	DeviceInfo         *string
	IPAddress          *string
	UserAgent          *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// StoreRefreshToken stores a new refresh token fingerprint in the database
func (r *RefreshTokenRepository) StoreRefreshToken(ctx context.Context, userID int64, tokenFingerprint string, expiresAt time.Time, deviceInfo, ipAddress, userAgent string) (int64, error) {
	query := `
		INSERT INTO refresh_tokens (
			user_id,
			token_fingerprint,
			issued_at,
			expires_at,
			device_info,
			ip_address,
			user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var tokenID int64
	err := r.db.QueryRow(ctx, query,
		userID,
		tokenFingerprint,
		time.Now(),
		expiresAt,
		deviceInfo,
		ipAddress,
		userAgent,
	).Scan(&tokenID)

	if err != nil {
		return 0, err
	}

	return tokenID, nil
}

// GetRefreshToken retrieves a refresh token by its fingerprint
func (r *RefreshTokenRepository) GetRefreshToken(ctx context.Context, tokenFingerprint string) (*RefreshToken, error) {
	query := `
		SELECT
			id,
			user_id,
			token_fingerprint,
			issued_at,
			expires_at,
			revoked_at,
			replaced_by_token_id,
			device_info,
			ip_address,
			user_agent,
			created_at,
			updated_at
		FROM refresh_tokens
		WHERE token_fingerprint = $1
	`

	token := &RefreshToken{}
	err := r.db.QueryRow(ctx, query, tokenFingerprint).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenFingerprint,
		&token.IssuedAt,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.ReplacedByTokenID,
		&token.DeviceInfo,
		&token.IPAddress,
		&token.UserAgent,
		&token.CreatedAt,
		&token.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Token not found
	}
	if err != nil {
		return nil, err
	}

	return token, nil
}

// RevokeToken marks a token as revoked and optionally links it to the replacement token
func (r *RefreshTokenRepository) RevokeToken(ctx context.Context, tokenFingerprint string, replacedByTokenID *int64) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1,
		    replaced_by_token_id = $2,
		    updated_at = $3
		WHERE token_fingerprint = $4
	`

	_, err := r.db.Exec(ctx, query,
		time.Now(),
		replacedByTokenID,
		time.Now(),
		tokenFingerprint,
	)

	return err
}

// RevokeAllUserTokens revokes all refresh tokens for a specific user (used for token reuse detection)
func (r *RefreshTokenRepository) RevokeAllUserTokens(ctx context.Context, userID int64) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1,
		    updated_at = $2
		WHERE user_id = $3
		  AND revoked_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, time.Now(), time.Now(), userID)
	return err
}

// IsTokenRevoked checks if a token has been revoked
func (r *RefreshTokenRepository) IsTokenRevoked(ctx context.Context, tokenFingerprint string) (bool, error) {
	token, err := r.GetRefreshToken(ctx, tokenFingerprint)
	if err != nil {
		return false, err
	}

	if token == nil {
		return true, nil // Token not found = consider it revoked
	}

	return token.RevokedAt != nil, nil
}

// IsTokenExpired checks if a token has expired
func (r *RefreshTokenRepository) IsTokenExpired(ctx context.Context, tokenFingerprint string) (bool, error) {
	token, err := r.GetRefreshToken(ctx, tokenFingerprint)
	if err != nil {
		return false, err
	}

	if token == nil {
		return true, nil
	}

	return token.ExpiresAt.Before(time.Now()), nil
}

// CleanupExpiredTokens removes tokens that have been expired for more than 30 days
func (r *RefreshTokenRepository) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM refresh_tokens
		WHERE (expires_at < NOW() - INTERVAL '30 days')
		   OR (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '30 days')
	`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

// ActiveSession represents an active user session with device info
type ActiveSession struct {
	ID          int64      `json:"id"`
	DeviceInfo  string     `json:"device_info"`
	IPAddress   string     `json:"ip_address"`
	UserAgent   string     `json:"user_agent"`
	Browser     string     `json:"browser"`
	OS          string     `json:"os"`
	DeviceType  string     `json:"device_type"`
	Location    string     `json:"location"`
	IssuedAt    time.Time  `json:"issued_at"`
	LastUsedAt  time.Time  `json:"last_used_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
	IsCurrent   bool       `json:"is_current"`
}

// GetActiveSessions returns all active (non-revoked, non-expired) sessions for a user
func (r *RefreshTokenRepository) GetActiveSessions(ctx context.Context, userID int64) ([]ActiveSession, error) {
	query := `
		SELECT
			id,
			COALESCE(device_info, ''),
			COALESCE(ip_address, ''),
			COALESCE(user_agent, ''),
			issued_at,
			COALESCE(updated_at, issued_at) as last_used_at,
			expires_at
		FROM refresh_tokens
		WHERE user_id = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
		ORDER BY issued_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []ActiveSession
	for rows.Next() {
		var s ActiveSession
		if err := rows.Scan(
			&s.ID,
			&s.DeviceInfo,
			&s.IPAddress,
			&s.UserAgent,
			&s.IssuedAt,
			&s.LastUsedAt,
			&s.ExpiresAt,
		); err != nil {
			return nil, err
		}

		// Parse user agent to get browser, OS, device type
		s.Browser, s.OS, s.DeviceType = parseUserAgent(s.UserAgent)

		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// RevokeSessionByID revokes a specific session by its ID
func (r *RefreshTokenRepository) RevokeSessionByID(ctx context.Context, userID int64, sessionID int64) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1,
		    updated_at = $2
		WHERE id = $3
		  AND user_id = $4
		  AND revoked_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, time.Now(), time.Now(), sessionID, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// RevokeOtherSessions revokes all sessions except the current one
func (r *RefreshTokenRepository) RevokeOtherSessions(ctx context.Context, userID int64, currentTokenFingerprint string) (int64, error) {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1,
		    updated_at = $2
		WHERE user_id = $3
		  AND token_fingerprint != $4
		  AND revoked_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, time.Now(), time.Now(), userID, currentTokenFingerprint)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

// GetSessionByTokenFingerprint returns a session by its token fingerprint
func (r *RefreshTokenRepository) GetSessionByTokenFingerprint(ctx context.Context, tokenFingerprint string) (*ActiveSession, error) {
	query := `
		SELECT
			id,
			COALESCE(device_info, ''),
			COALESCE(ip_address, ''),
			COALESCE(user_agent, ''),
			issued_at,
			COALESCE(updated_at, issued_at) as last_used_at,
			expires_at
		FROM refresh_tokens
		WHERE token_fingerprint = $1
		  AND revoked_at IS NULL
	`

	var s ActiveSession
	err := r.db.QueryRow(ctx, query, tokenFingerprint).Scan(
		&s.ID,
		&s.DeviceInfo,
		&s.IPAddress,
		&s.UserAgent,
		&s.IssuedAt,
		&s.LastUsedAt,
		&s.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	s.Browser, s.OS, s.DeviceType = parseUserAgent(s.UserAgent)

	return &s, nil
}

// UpdateSessionActivity updates the last activity time for a session
func (r *RefreshTokenRepository) UpdateSessionActivity(ctx context.Context, tokenFingerprint string) error {
	query := `
		UPDATE refresh_tokens
		SET updated_at = $1
		WHERE token_fingerprint = $2
		  AND revoked_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, time.Now(), tokenFingerprint)
	return err
}

// GetSessionCount returns the count of active sessions for a user
func (r *RefreshTokenRepository) GetSessionCount(ctx context.Context, userID int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM refresh_tokens
		WHERE user_id = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
	`

	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// parseUserAgent extracts browser, OS, and device type from user agent string
func parseUserAgent(ua string) (browser, os, deviceType string) {
	browser = "Unknown Browser"
	os = "Unknown OS"
	deviceType = "desktop"

	// Simple user agent parsing
	if ua == "" {
		return
	}

	// Detect browser
	switch {
	case contains(ua, "Firefox"):
		browser = "Firefox"
	case contains(ua, "Edg"):
		browser = "Microsoft Edge"
	case contains(ua, "Chrome"):
		browser = "Chrome"
	case contains(ua, "Safari"):
		browser = "Safari"
	case contains(ua, "Opera") || contains(ua, "OPR"):
		browser = "Opera"
	case contains(ua, "MSIE") || contains(ua, "Trident"):
		browser = "Internet Explorer"
	}

	// Detect OS
	switch {
	case contains(ua, "Windows"):
		os = "Windows"
	case contains(ua, "Mac OS"):
		os = "macOS"
	case contains(ua, "Linux"):
		os = "Linux"
	case contains(ua, "Android"):
		os = "Android"
		deviceType = "mobile"
	case contains(ua, "iOS") || contains(ua, "iPhone") || contains(ua, "iPad"):
		os = "iOS"
		if contains(ua, "iPad") {
			deviceType = "tablet"
		} else {
			deviceType = "mobile"
		}
	}

	// Detect mobile devices
	if contains(ua, "Mobile") || contains(ua, "Android") {
		deviceType = "mobile"
	}
	if contains(ua, "Tablet") || contains(ua, "iPad") {
		deviceType = "tablet"
	}

	return
}

// contains is a simple case-insensitive substring check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsCI(s, substr))
}

func containsCI(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if eqCI(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func eqCI(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
