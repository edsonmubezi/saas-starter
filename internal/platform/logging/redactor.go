package logging

import (
	"strings"
)

// SensitiveFields is the list of field names that should never be logged
// This extends the work from SECURITY-LOGGING-FIXES.md
var SensitiveFields = map[string]bool{
	"password":            true,
	"current_password":    true,
	"new_password":        true,
	"old_password":        true,
	"token":               true,
	"access_token":        true,
	"refresh_token":       true,
	"authorization":       true,
	"secret":              true,
	"api_key":             true,
	"apikey":              true,
	"private_key":         true,
	"privatekey":          true,
	"credit_card":         true,
	"creditcard":          true,
	"card_number":         true,
	"cardnumber":          true,
	"cvv":                 true,
	"cvc":                 true,
	"ssn":                 true,
	"social_security":     true,
	"tax_id":              true,
	"taxid":               true,
	"bank_account":        true,
	"bankaccount":         true,
	"routing_number":      true,
	"routingnumber":       true,
}

// IsSensitiveField checks if a field name is sensitive
func IsSensitiveField(fieldName string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(fieldName, "_", ""))
	normalized = strings.ReplaceAll(normalized, "-", "")

	// Check both original and normalized
	if SensitiveFields[strings.ToLower(fieldName)] {
		return true
	}
	return SensitiveFields[normalized]
}

// RedactValue replaces sensitive values with [REDACTED]
func RedactValue(fieldName string, value interface{}) interface{} {
	if IsSensitiveField(fieldName) {
		return "[REDACTED]"
	}
	return value
}

// RedactMap redacts sensitive fields in a map
func RedactMap(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}

	redacted := make(map[string]interface{})
	for key, value := range data {
		// Recursively handle nested maps
		if nestedMap, ok := value.(map[string]interface{}); ok {
			redacted[key] = RedactMap(nestedMap)
		} else {
			redacted[key] = RedactValue(key, value)
		}
	}
	return redacted
}

// SanitizeForLogging removes sensitive data before logging
func SanitizeForLogging(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		return RedactMap(v)
	case map[string]string:
		result := make(map[string]string)
		for key, value := range v {
			if IsSensitiveField(key) {
				result[key] = "[REDACTED]"
			} else {
				result[key] = value
			}
		}
		return result
	default:
		return data
	}
}
