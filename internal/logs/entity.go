// internal/logs/model.go
package logs

import (
	"time"
)

// Log represents a system or user activity log in the system
type Log struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`    // ID of the user performing the action (if available)
	Action    string    `json:"action"`     // Description of the action (e.g., "User logged in")
	IPAddress string    `json:"ip_address"` // IP address of the user
	UserAgent string    `json:"user_agent"` // User-Agent (browser details)
	CreatedAt time.Time `json:"created_at"` // Timestamp when the action occurred
}
