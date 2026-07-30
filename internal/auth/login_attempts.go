package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// LoginAttempt represents a login attempt record
type LoginAttempt struct {
	ID            int64      `json:"id"`
	UserID        *int64     `json:"user_id,omitempty"`
	Email         string     `json:"email"`
	IPAddress     string     `json:"ip_address"`
	UserAgent     string     `json:"user_agent"`
	Success       bool       `json:"success"`
	FailureReason *string    `json:"failure_reason,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// LoginAttemptInput is used when creating a new login attempt
type LoginAttemptInput struct {
	UserID        *int64
	Email         string
	IPAddress     string
	UserAgent     string
	Success       bool
	FailureReason string
}

// LoginAttemptsRepository defines the interface for login attempt persistence
type LoginAttemptsRepository interface {
	Create(ctx context.Context, input LoginAttemptInput) error
	GetRecentByUser(ctx context.Context, userID int64, limit int) ([]LoginAttempt, error)
	GetRecentByEmail(ctx context.Context, email string, limit int) ([]LoginAttempt, error)
	GetRecentByIP(ctx context.Context, ip string, limit int) ([]LoginAttempt, error)
	CountRecentFailedByIP(ctx context.Context, ip string, minutes int) (int, error)
}

// PostgresLoginAttemptsRepository implements LoginAttemptsRepository
type PostgresLoginAttemptsRepository struct {
	db *pgxpool.Pool
}

// NewLoginAttemptsRepository creates a new repository instance
func NewLoginAttemptsRepository(db *pgxpool.Pool) *PostgresLoginAttemptsRepository {
	return &PostgresLoginAttemptsRepository{db: db}
}

// Create records a new login attempt
func (r *PostgresLoginAttemptsRepository) Create(ctx context.Context, input LoginAttemptInput) error {
	query := `
		INSERT INTO login_attempts (user_id, email, ip_address, user_agent, success, failure_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`

	var failureReason *string
	if input.FailureReason != "" {
		failureReason = &input.FailureReason
	}

	_, err := r.db.Exec(ctx, query,
		input.UserID,
		input.Email,
		input.IPAddress,
		input.UserAgent,
		input.Success,
		failureReason,
	)
	if err != nil {
		return fmt.Errorf("could not create login attempt: %v", err)
	}

	return nil
}

// GetRecentByUser retrieves recent login attempts for a user
func (r *PostgresLoginAttemptsRepository) GetRecentByUser(ctx context.Context, userID int64, limit int) ([]LoginAttempt, error) {
	query := `
		SELECT id, user_id, email, ip_address, user_agent, success, failure_reason, created_at
		FROM login_attempts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("could not query login attempts: %v", err)
	}
	defer rows.Close()

	var attempts []LoginAttempt
	for rows.Next() {
		var a LoginAttempt
		err := rows.Scan(
			&a.ID,
			&a.UserID,
			&a.Email,
			&a.IPAddress,
			&a.UserAgent,
			&a.Success,
			&a.FailureReason,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan login attempt: %v", err)
		}
		attempts = append(attempts, a)
	}

	return attempts, nil
}

// GetRecentByEmail retrieves recent login attempts for an email
func (r *PostgresLoginAttemptsRepository) GetRecentByEmail(ctx context.Context, email string, limit int) ([]LoginAttempt, error) {
	query := `
		SELECT id, user_id, email, ip_address, user_agent, success, failure_reason, created_at
		FROM login_attempts
		WHERE email = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, email, limit)
	if err != nil {
		return nil, fmt.Errorf("could not query login attempts: %v", err)
	}
	defer rows.Close()

	var attempts []LoginAttempt
	for rows.Next() {
		var a LoginAttempt
		err := rows.Scan(
			&a.ID,
			&a.UserID,
			&a.Email,
			&a.IPAddress,
			&a.UserAgent,
			&a.Success,
			&a.FailureReason,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan login attempt: %v", err)
		}
		attempts = append(attempts, a)
	}

	return attempts, nil
}

// GetRecentByIP retrieves recent login attempts from an IP address
func (r *PostgresLoginAttemptsRepository) GetRecentByIP(ctx context.Context, ip string, limit int) ([]LoginAttempt, error) {
	query := `
		SELECT id, user_id, email, ip_address, user_agent, success, failure_reason, created_at
		FROM login_attempts
		WHERE ip_address = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, ip, limit)
	if err != nil {
		return nil, fmt.Errorf("could not query login attempts: %v", err)
	}
	defer rows.Close()

	var attempts []LoginAttempt
	for rows.Next() {
		var a LoginAttempt
		err := rows.Scan(
			&a.ID,
			&a.UserID,
			&a.Email,
			&a.IPAddress,
			&a.UserAgent,
			&a.Success,
			&a.FailureReason,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan login attempt: %v", err)
		}
		attempts = append(attempts, a)
	}

	return attempts, nil
}

// CountRecentFailedByIP counts failed login attempts from an IP in the last N minutes
// Useful for IP-based rate limiting
func (r *PostgresLoginAttemptsRepository) CountRecentFailedByIP(ctx context.Context, ip string, minutes int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM login_attempts
		WHERE ip_address = $1
		  AND success = false
		  AND created_at > NOW() - ($2 || ' minutes')::INTERVAL
	`

	var count int
	err := r.db.QueryRow(ctx, query, ip, fmt.Sprintf("%d", minutes)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("could not count failed attempts: %v", err)
	}

	return count, nil
}
