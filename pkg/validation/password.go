package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// PasswordPolicy defines the requirements for a secure password
type PasswordPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
	ForbidCommon   bool
}

// DefaultPasswordPolicy returns a strong password policy suitable for production
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:      12,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
		ForbidCommon:   true,
	}
}

// PasswordValidationError represents a password validation failure
type PasswordValidationError struct {
	Errors []string
}

func (e *PasswordValidationError) Error() string {
	return strings.Join(e.Errors, "; ")
}

// Common weak passwords that should be forbidden
var commonPasswords = map[string]bool{
	"password":     true,
	"password123":  true,
	"123456":       true,
	"12345678":     true,
	"123456789":    true,
	"qwerty":       true,
	"abc123":       true,
	"monkey":       true,
	"1234567":      true,
	"letmein":      true,
	"trustno1":     true,
	"dragon":       true,
	"baseball":     true,
	"iloveyou":     true,
	"master":       true,
	"sunshine":     true,
	"ashley":       true,
	"bailey":       true,
	"passw0rd":     true,
	"shadow":       true,
	"123123":       true,
	"654321":       true,
	"superman":     true,
	"qazwsx":       true,
	"michael":      true,
	"football":     true,
	"welcome":      true,
	"welcome123":   true,
	"admin":        true,
	"admin123":     true,
	"password1":    true,
	"password12":   true,
	"password1234": true,
}

// ValidatePassword checks if a password meets the policy requirements
func ValidatePassword(password string, policy PasswordPolicy) error {
	var errors []string

	// Check minimum length
	if len(password) < policy.MinLength {
		errors = append(errors, fmt.Sprintf("Password must be at least %d characters long", policy.MinLength))
	}

	// Check maximum length (prevent DoS via bcrypt)
	if len(password) > 72 {
		errors = append(errors, "Password must be no more than 72 characters long")
	}

	// Check for uppercase letters
	if policy.RequireUpper && !hasUppercase(password) {
		errors = append(errors, "Password must contain at least one uppercase letter")
	}

	// Check for lowercase letters
	if policy.RequireLower && !hasLowercase(password) {
		errors = append(errors, "Password must contain at least one lowercase letter")
	}

	// Check for numbers
	if policy.RequireNumber && !hasNumber(password) {
		errors = append(errors, "Password must contain at least one number")
	}

	// Check for special characters
	if policy.RequireSpecial && !hasSpecialChar(password) {
		errors = append(errors, "Password must contain at least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)")
	}

	// Check against common passwords
	if policy.ForbidCommon && isCommonPassword(password) {
		errors = append(errors, "Password is too common and easily guessable")
	}

	if len(errors) > 0 {
		return &PasswordValidationError{Errors: errors}
	}

	return nil
}

// hasUppercase checks if string contains at least one uppercase letter
func hasUppercase(s string) bool {
	for _, c := range s {
		if unicode.IsUpper(c) {
			return true
		}
	}
	return false
}

// hasLowercase checks if string contains at least one lowercase letter
func hasLowercase(s string) bool {
	for _, c := range s {
		if unicode.IsLower(c) {
			return true
		}
	}
	return false
}

// hasNumber checks if string contains at least one digit
func hasNumber(s string) bool {
	for _, c := range s {
		if unicode.IsDigit(c) {
			return true
		}
	}
	return false
}

// hasSpecialChar checks if string contains at least one special character
func hasSpecialChar(s string) bool {
	specialChars := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{}|;:,.<>?]`)
	return specialChars.MatchString(s)
}

// isCommonPassword checks if password is in the common passwords list
func isCommonPassword(password string) bool {
	lowerPassword := strings.ToLower(password)
	return commonPasswords[lowerPassword]
}

// hasSequentialChars detects sequential characters like "abc" or "123"
func hasSequentialChars(s string) bool {
	if len(s) < 3 {
		return false
	}

	for i := 0; i < len(s)-2; i++ {
		// Check for sequential ascending (abc, 123)
		if s[i]+1 == s[i+1] && s[i+1]+1 == s[i+2] {
			return true
		}
		// Check for sequential descending (cba, 321)
		if s[i]-1 == s[i+1] && s[i+1]-1 == s[i+2] {
			return true
		}
	}

	return false
}

// hasRepeatedChars detects repeated characters like "aaa" or "111"
func hasRepeatedChars(s string, maxRepeat int) bool {
	if len(s) < maxRepeat {
		return false
	}

	count := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			count++
			if count >= maxRepeat {
				return true
			}
		} else {
			count = 1
		}
	}

	return false
}

// PasswordStrength returns a score (0-5) indicating password strength
func PasswordStrength(password string) int {
	score := 0

	// Length bonus
	if len(password) >= 8 {
		score++
	}
	if len(password) >= 12 {
		score++
	}
	if len(password) >= 16 {
		score++
	}

	// Complexity bonus
	if hasUppercase(password) && hasLowercase(password) {
		score++
	}
	if hasNumber(password) {
		score++
	}
	if hasSpecialChar(password) {
		score++
	}

	// Penalties
	if isCommonPassword(password) {
		score = 0
	}
	if hasSequentialChars(password) {
		score--
	}
	if hasRepeatedChars(password, 3) {
		score--
	}

	// Ensure score is between 0 and 5
	if score < 0 {
		score = 0
	}
	if score > 5 {
		score = 5
	}

	return score
}

// ValidatePasswordWithDetails returns validation result with strength score
func ValidatePasswordWithDetails(password string, policy PasswordPolicy) (bool, int, []string) {
	err := ValidatePassword(password, policy)
	strength := PasswordStrength(password)

	if err != nil {
		if validationErr, ok := err.(*PasswordValidationError); ok {
			return false, strength, validationErr.Errors
		}
		return false, strength, []string{err.Error()}
	}

	return true, strength, nil
}
