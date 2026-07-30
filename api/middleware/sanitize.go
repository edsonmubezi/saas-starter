// middleware/sanitize.go
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var sanitizer = bluemonday.UGCPolicy()

// Sensitive field names that should NEVER be logged
var sensitiveFields = map[string]bool{
	"password":          true,
	"current_password":  true,
	"new_password":      true,
	"old_password":      true,
	"token":             true,
	"access_token":      true,
	"refresh_token":     true,
	"authorization":     true,
	"secret":            true,
	"api_key":           true,
	"private_key":       true,
	"credit_card":       true,
	"card_number":       true,
	"cvv":               true,
	"ssn":               true,
	"social_security":   true,
}

func SanitizeInputs(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sanitize query parameters
		r.URL.RawQuery = sanitizeAndLogValues("Query", r.URL.Query()).Encode()

		contentType := r.Header.Get("Content-Type")

		switch {
		case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"),
			strings.HasPrefix(contentType, "multipart/form-data"):
			if err := r.ParseMultipartForm(32 << 20); err == nil {
				r.Form = sanitizeAndLogValues("Form", r.Form)
				r.PostForm = sanitizeAndLogValues("PostForm", r.PostForm)

				if r.MultipartForm != nil {
					// Sanitize filenames
					for key, files := range r.MultipartForm.File {
						for _, fileHeader := range files {
							cleanName := sanitizer.Sanitize(fileHeader.Filename)
							if cleanName != fileHeader.Filename {
								log.Printf("[Sanitized Filename] Key: %s | Original: %s | Cleaned: %s", key, fileHeader.Filename, cleanName)
								fileHeader.Filename = cleanName
							}
						}
					}
				}
			}

		case strings.HasPrefix(contentType, "application/json"):
			var bodyCopy bytes.Buffer
			tee := io.TeeReader(r.Body, &bodyCopy)

			var jsonData map[string]interface{}
			if err := json.NewDecoder(tee).Decode(&jsonData); err == nil {
				sanitizeJSON(jsonData)
				sanitizedBody, _ := json.Marshal(jsonData)
				r.Body = io.NopCloser(bytes.NewReader(sanitizedBody))
				// REMOVED: Do not log entire JSON body as it may contain passwords
			} else {
				r.Body = io.NopCloser(&bodyCopy)
			}

		default:
			_ = r.ParseForm()
			r.Form = sanitizeAndLogValues("Form", r.Form)
			r.PostForm = sanitizeAndLogValues("PostForm", r.PostForm)
		}

		next.ServeHTTP(w, r)
	})
}

// isJSONValue checks if a string looks like a JSON object or array
func isJSONValue(s string) bool {
	trimmed := strings.TrimSpace(s)
	return (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"))
}

func sanitizeAndLogValues(source string, values url.Values) url.Values {
	sanitized := url.Values{}
	for key, val := range values {
		for _, v := range val {
			// Never sanitize sensitive fields (passwords, tokens, etc.)
			if isSensitiveField(key) {
				sanitized.Add(key, v)
				continue
			}
			// Skip HTML sanitization for JSON values (e.g. multipart form fields
			// containing JSON payloads like "loan_data", "claim_data", etc.)
			// bluemonday HTML-escapes quotes which corrupts JSON.
			if isJSONValue(v) {
				sanitized.Add(key, strings.TrimSpace(v))
				continue
			}
			clean := sanitizer.Sanitize(strings.TrimSpace(v))
			if clean != v {
				log.Printf("[Sanitized %s] Key: %s | Original: %s | Cleaned: %s", source, key, v, clean)
			}
			sanitized.Add(key, clean)
		}
	}
	return sanitized
}

func sanitizeJSON(data map[string]interface{}) {
	for key, val := range data {
		// Never sanitize sensitive fields — bluemonday HTML-escapes special
		// characters (& → &amp;, etc.) which corrupts passwords and tokens.
		if isSensitiveField(key) {
			continue
		}

		switch v := val.(type) {
		case string:
			cleaned := sanitizer.Sanitize(strings.TrimSpace(v))
			if cleaned != v {
				log.Printf("[Sanitized JSON Field] Key: %s | Original: %s | Cleaned: %s", key, v, cleaned)
			}
			data[key] = cleaned
		case map[string]interface{}:
			sanitizeJSON(v)
		case []interface{}:
			for i, item := range v {
				if str, ok := item.(string); ok {
					cleaned := sanitizer.Sanitize(strings.TrimSpace(str))
					if cleaned != str {
						log.Printf("[Sanitized JSON List] Key: %s[%d] | Original: %s | Cleaned: %s", key, i, str, cleaned)
					}
					v[i] = cleaned
				}
			}
		}
	}
}

// isSensitiveField checks if a field name is in the sensitive fields list
func isSensitiveField(fieldName string) bool {
	return sensitiveFields[strings.ToLower(fieldName)]
}
