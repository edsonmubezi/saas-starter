package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	// TOTPSecretLength is the length of the TOTP secret in bytes
	TOTPSecretLength = 20

	// TOTPDigits is the number of digits in the TOTP code
	TOTPDigits = 6

	// TOTPPeriod is the time step in seconds
	TOTPPeriod = 30

	// TOTPWindow is the number of periods to check before/after current time
	TOTPWindow = 1

	// BackupCodeCount is the number of backup codes to generate
	BackupCodeCount = 10

	// BackupCodeLength is the length of each backup code
	BackupCodeLength = 8
)

// TwoFactorMethod represents the 2FA method type
type TwoFactorMethod string

const (
	TwoFactorMethodTOTP  TwoFactorMethod = "totp"
	TwoFactorMethodEmail TwoFactorMethod = "email"
	TwoFactorMethodSMS   TwoFactorMethod = "sms"
)

// TOTPService handles TOTP (Time-based One-Time Password) operations
type TOTPService struct {
	db *pgxpool.Pool
}

// BackupCode represents a backup code for 2FA recovery
type BackupCode struct {
	ID        int64
	UserID    int64
	CodeHash  string
	UsedAt    *time.Time
	CreatedAt time.Time
}

// OTPCode represents a one-time password (for email/SMS)
type OTPCode struct {
	ID        int64
	UserID    int64
	Code      string
	Method    TwoFactorMethod
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// NewTOTPService creates a new TOTP service
func NewTOTPService(db *pgxpool.Pool) *TOTPService {
	return &TOTPService{db: db}
}

// GenerateSecret generates a new TOTP secret
func GenerateSecret() (string, error) {
	secret := make([]byte, TOTPSecretLength)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}

// GenerateTOTP generates a TOTP code for the given secret and time
func GenerateTOTP(secret string, t time.Time) (string, error) {
	// Decode the secret
	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", fmt.Errorf("invalid secret: %w", err)
	}

	// Calculate the counter value
	counter := uint64(t.Unix()) / TOTPPeriod

	// Generate HOTP
	code := generateHOTP(secretBytes, counter)

	return fmt.Sprintf("%0*d", TOTPDigits, code), nil
}

// ValidateTOTP validates a TOTP code against the secret
func ValidateTOTP(secret string, code string, t time.Time) bool {
	// Check current period and adjacent periods (for clock drift)
	for i := -TOTPWindow; i <= TOTPWindow; i++ {
		checkTime := t.Add(time.Duration(i*TOTPPeriod) * time.Second)
		expectedCode, err := GenerateTOTP(secret, checkTime)
		if err != nil {
			continue
		}
		if hmac.Equal([]byte(expectedCode), []byte(code)) {
			return true
		}
	}
	return false
}

// generateHOTP generates an HOTP value
func generateHOTP(secret []byte, counter uint64) int {
	// Convert counter to bytes (big-endian)
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, counter)

	// Calculate HMAC-SHA1
	h := hmac.New(sha1.New, secret)
	h.Write(counterBytes)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	code := int(binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff)

	// Modulo to get the right number of digits
	mod := 1
	for i := 0; i < TOTPDigits; i++ {
		mod *= 10
	}

	return code % mod
}

// GenerateProvisioningURI generates a URI for QR code provisioning
func GenerateProvisioningURI(secret, email, issuer string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&digits=%d&period=%d",
		issuer, email, secret, issuer, TOTPDigits, TOTPPeriod)
}

// SetupTOTP sets up TOTP for a user
func (s *TOTPService) SetupTOTP(ctx context.Context, userID int64, secret string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users
		SET totp_secret = $1, two_factor_method = $2, updated_at = NOW()
		WHERE id = $3
	`, secret, TwoFactorMethodTOTP, userID)

	if err != nil {
		return fmt.Errorf("failed to setup TOTP: %w", err)
	}

	return nil
}

// EnableTOTP enables TOTP for a user after verification
func (s *TOTPService) EnableTOTP(ctx context.Context, userID int64) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users
		SET two_factor_enabled = TRUE, updated_at = NOW()
		WHERE id = $1
	`, userID)

	if err != nil {
		return fmt.Errorf("failed to enable TOTP: %w", err)
	}

	return nil
}

// DisableTOTP disables TOTP for a user
func (s *TOTPService) DisableTOTP(ctx context.Context, userID int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Disable 2FA and clear secret
	_, err = tx.Exec(ctx, `
		UPDATE users
		SET two_factor_enabled = FALSE, totp_secret = NULL, two_factor_method = NULL, updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to disable TOTP: %w", err)
	}

	// Delete backup codes
	_, err = tx.Exec(ctx, `DELETE FROM two_factor_backup_codes WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete backup codes: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTOTPSecret gets the TOTP secret for a user
func (s *TOTPService) GetTOTPSecret(ctx context.Context, userID int64) (string, error) {
	var secret *string
	err := s.db.QueryRow(ctx, `
		SELECT totp_secret FROM users WHERE id = $1
	`, userID).Scan(&secret)

	if err != nil {
		return "", fmt.Errorf("failed to get TOTP secret: %w", err)
	}

	if secret == nil {
		return "", fmt.Errorf("TOTP not configured for user")
	}

	return *secret, nil
}

// GenerateBackupCodes generates backup codes for a user
func (s *TOTPService) GenerateBackupCodes(ctx context.Context, userID int64) ([]string, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete existing backup codes
	_, err = tx.Exec(ctx, `DELETE FROM two_factor_backup_codes WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing backup codes: %w", err)
	}

	// Generate new backup codes
	codes := make([]string, BackupCodeCount)
	for i := 0; i < BackupCodeCount; i++ {
		code, err := generateBackupCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		codes[i] = code

		// Hash the code before storing
		codeHash := HashToken(code)

		_, err = tx.Exec(ctx, `
			INSERT INTO two_factor_backup_codes (user_id, code_hash)
			VALUES ($1, $2)
		`, userID, codeHash)
		if err != nil {
			return nil, fmt.Errorf("failed to insert backup code: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return codes, nil
}

// generateBackupCode generates a single backup code
func generateBackupCode() (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Removed similar-looking characters
	code := make([]byte, BackupCodeLength)
	randomBytes := make([]byte, BackupCodeLength)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	for i := 0; i < BackupCodeLength; i++ {
		code[i] = charset[int(randomBytes[i])%len(charset)]
	}

	// Format as XXXX-XXXX
	return fmt.Sprintf("%s-%s", string(code[:4]), string(code[4:])), nil
}

// ValidateBackupCode validates and consumes a backup code
func (s *TOTPService) ValidateBackupCode(ctx context.Context, userID int64, code string) (bool, error) {
	// Remove any dashes from the code
	cleanCode := strings.ReplaceAll(code, "-", "")
	codeHash := HashToken(cleanCode)

	// Try to mark the code as used
	result, err := s.db.Exec(ctx, `
		UPDATE two_factor_backup_codes
		SET used_at = NOW()
		WHERE user_id = $1 AND code_hash = $2 AND used_at IS NULL
	`, userID, codeHash)

	if err != nil {
		return false, fmt.Errorf("failed to validate backup code: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

// GetRemainingBackupCodes returns the count of remaining (unused) backup codes
func (s *TOTPService) GetRemainingBackupCodes(ctx context.Context, userID int64) (int, error) {
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM two_factor_backup_codes
		WHERE user_id = $1 AND used_at IS NULL
	`, userID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count backup codes: %w", err)
	}

	return count, nil
}

// GenerateOTPCode generates a one-time password for email/SMS
func (s *TOTPService) GenerateOTPCode(ctx context.Context, userID int64, method TwoFactorMethod, expiryMinutes int) (string, error) {
	// Generate 6-digit code
	code := make([]byte, 6)
	for i := range code {
		b := make([]byte, 1)
		_, _ = rand.Read(b)
		code[i] = '0' + (b[0] % 10)
	}
	codeStr := string(code)

	// Calculate expiry
	expiresAt := time.Now().Add(time.Duration(expiryMinutes) * time.Minute)

	// Invalidate any existing unused codes for this user and method
	_, err := s.db.Exec(ctx, `
		UPDATE two_factor_otp_codes
		SET used_at = NOW()
		WHERE user_id = $1 AND method = $2 AND used_at IS NULL
	`, userID, method)
	if err != nil {
		return "", fmt.Errorf("failed to invalidate existing codes: %w", err)
	}

	// Insert the new code
	_, err = s.db.Exec(ctx, `
		INSERT INTO two_factor_otp_codes (user_id, code, method, expires_at)
		VALUES ($1, $2, $3, $4)
	`, userID, codeStr, method, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create OTP code: %w", err)
	}

	return codeStr, nil
}

// ValidateOTPCode validates a one-time password
func (s *TOTPService) ValidateOTPCode(ctx context.Context, userID int64, method TwoFactorMethod, code string) (bool, error) {
	// Try to mark the code as used
	result, err := s.db.Exec(ctx, `
		UPDATE two_factor_otp_codes
		SET used_at = NOW()
		WHERE user_id = $1 AND method = $2 AND code = $3 AND used_at IS NULL AND expires_at > NOW()
	`, userID, method, code)

	if err != nil {
		return false, fmt.Errorf("failed to validate OTP code: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

// Is2FAEnabled checks if 2FA is enabled for a user
func (s *TOTPService) Is2FAEnabled(ctx context.Context, userID int64) (bool, TwoFactorMethod, error) {
	var enabled bool
	var method *string

	err := s.db.QueryRow(ctx, `
		SELECT two_factor_enabled, two_factor_method
		FROM users WHERE id = $1
	`, userID).Scan(&enabled, &method)

	if err != nil {
		return false, "", fmt.Errorf("failed to check 2FA status: %w", err)
	}

	if method == nil {
		return enabled, "", nil
	}

	return enabled, TwoFactorMethod(*method), nil
}

// IsRoleRequires2FA checks if the user's role requires 2FA
func (s *TOTPService) IsRoleRequires2FA(ctx context.Context, userID int64) (bool, error) {
	var requires2FA bool

	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(r.requires_2fa, FALSE)
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`, userID).Scan(&requires2FA)

	if err != nil {
		return false, fmt.Errorf("failed to check role 2FA requirement: %w", err)
	}

	return requires2FA, nil
}

// Set2FAMethod sets the 2FA method for a user (without enabling)
func (s *TOTPService) Set2FAMethod(ctx context.Context, userID int64, method TwoFactorMethod) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users
		SET two_factor_method = $1, updated_at = NOW()
		WHERE id = $2
	`, method, userID)

	if err != nil {
		return fmt.Errorf("failed to set 2FA method: %w", err)
	}

	return nil
}
