package middleware

import (
	"encoding/json"
	"net/http"
)

// ParseJSONBody reads and decodes a JSON request body into the target struct.
// It returns an error if decoding fails or the body is empty/invalid.
func ParseJSONBody(r *http.Request, target interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // helps catch unexpected fields
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}
