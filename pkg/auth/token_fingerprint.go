package auth

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken creates a SHA-256 hash of a token for storage
// Never store plain tokens in the database - always store hashed fingerprints
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
