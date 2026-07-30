package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateToken(userID int64, email string) (string, error) {
	claims := TokenClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)), // 30 minutes for security
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", jwt.ErrInvalidKey
	}

	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenStr string) (*TokenClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, jwt.ErrInvalidKey
	}

	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Ensure signing method is HS256
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	return claims, nil
}

// ValidateTokenWithBlacklist validates a token and checks if it's blacklisted
func ValidateTokenWithBlacklist(ctx context.Context, tokenStr string) (*TokenClaims, error) {
	// First, validate the token structure and signature
	claims, err := ValidateToken(tokenStr)
	if err != nil {
		return nil, err
	}

	// Check if the specific token is blacklisted
	isBlacklisted, err := IsTokenBlacklisted(ctx, tokenStr)
	if err != nil {
		// Log error but don't fail authentication if Redis is down
		// In production, you might want to fail closed (deny access) instead
		fmt.Printf("Warning: Failed to check token blacklist: %v\n", err)
	} else if isBlacklisted {
		return nil, fmt.Errorf("token has been revoked")
	}

	// Check if all user tokens are blacklisted (password change, account lockout)
	if claims.IssuedAt != nil {
		isUserBlacklisted, err := IsUserTokensBlacklisted(ctx, claims.UserID, claims.IssuedAt.Time)
		if err != nil {
			fmt.Printf("Warning: Failed to check user token blacklist: %v\n", err)
		} else if isUserBlacklisted {
			return nil, fmt.Errorf("token has been invalidated")
		}
	}

	return claims, nil
}

func GenerateAccessToken(userID int64, email string, orgID int64, role string, fullName string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if err := validateJWTSecret(secret); err != nil {
		return "", err
	}

	claims := TokenClaims{
		UserID:         userID,
		Email:          email,
		OrganizationID: orgID,
		Role:           role,
		FullName:       fullName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)), // 30 minutes for security
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "Microfinance",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(userID int64, orgID int64) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if err := validateJWTSecret(secret); err != nil {
		return "", err
	}

	claims := TokenClaims{
		UserID:         userID,
		OrganizationID: orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "Microfinance",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// validateJWTSecret ensures the JWT secret meets minimum security requirements
func validateJWTSecret(secret string) error {
	if secret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is not set")
	}
	// Use configurable minimum length (default 64)
	minLength := 64
	if minLengthEnv := os.Getenv("JWT_SECRET_MIN_LENGTH"); minLengthEnv != "" {
		fmt.Sscanf(minLengthEnv, "%d", &minLength)
	}
	if len(secret) < minLength {
		return fmt.Errorf("JWT_SECRET must be at least %d characters long (current: %d)", minLength, len(secret))
	}
	return nil
}

// ValidateJWTSecretOnStartup should be called at application startup
func ValidateJWTSecretOnStartup() error {
	secret := os.Getenv("JWT_SECRET")
	return validateJWTSecret(secret)
}

// Generate2FASessionToken generates a short-lived token for 2FA verification during login
// This token has limited claims and short expiration (5 minutes)
func Generate2FASessionToken(userID int64, orgID int64) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if err := validateJWTSecret(secret); err != nil {
		return "", err
	}

	claims := TokenClaims{
		UserID:         userID,
		OrganizationID: orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)), // Short-lived token
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "Microfinance",
			Subject:   "2fa-session",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
