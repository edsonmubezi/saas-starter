// pkg/utils/sanitize.go
package utils

import (
	"fmt"
	"html"
	"net/mail"
	"regexp"
	"strings"
)

// SanitizeString sanitizes any input string by trimming spaces, escaping HTML tags,
// removing special characters, and handling extra spaces.
func SanitizeString(input string) string {
	// Trim leading and trailing spaces
	input = strings.TrimSpace(input)

	// Remove any HTML tags (to prevent XSS)
	input = html.EscapeString(input)

	// Remove any special characters or unwanted characters using regex
	re := regexp.MustCompile(`[^a-zA-Z0-9\s-]`)
	input = re.ReplaceAllString(input, "")

	// Replace multiple spaces with a single space
	input = regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")

	return input
}

// SanitizeRoleName sanitizes the role name by removing unnecessary spaces and special characters
func SanitizeRoleName(roleName string) string {
	return SanitizeString(roleName)
}

// SanitizeFullName sanitizes the full name by cleaning up the text
func SanitizeFullName(fullName string) string {
	return SanitizeString(fullName)
}

// SanitizeEmail sanitizes the email by ensuring it's a valid email format and sanitizing its content
func SanitizeEmail(email string) (string, error) {
	email = SanitizeString(email)

	// Ensure it's a valid email
	_, err := mail.ParseAddress(email)
	if err != nil {
		return "", fmt.Errorf("invalid email format")
	}

	return email, nil
}

// SanitizePassword is not strictly necessary for sanitization (since passwords are typically hashed),
// but you can apply any necessary transformations (like trimming excess spaces).
func SanitizePassword(password string) string {
	return strings.TrimSpace(password)
}

// SanitizePhoto sanitizes the photo URL by ensuring it is a valid URL (optional sanitization logic for URLs)
func SanitizePhoto(photoURL string) string {
	// If photo URL is provided, sanitize it (just trims spaces and removes potential unsafe characters)
	return SanitizeString(photoURL)
}

// SanitizeOrganizationID sanitizes the organization ID (usually, this is a numeric value, so sanitize accordingly)
func SanitizeOrganizationID(organizationID int64) int64 {
	// In most cases, this would just be passed as is. No string sanitization is required for integers.
	// You could check if it's a valid organization ID, but we don't sanitize numbers here.
	return organizationID
}
