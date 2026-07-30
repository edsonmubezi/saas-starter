package emailconfig

import "time"

// OrganizationEmailConfig stores SMTP email configuration per organization.
type OrganizationEmailConfig struct {
	ID             int64     `json:"id"`
	OrganizationID int64     `json:"organization_id"`
	SMTPEnabled    bool      `json:"smtp_enabled"`
	SMTPHost       string    `json:"smtp_host"`
	SMTPPort       int       `json:"smtp_port"`
	SMTPUser       string    `json:"smtp_user"`
	SMTPPassword   string    `json:"smtp_password"`
	FromAddress    string    `json:"from_address"`
	FromName       string    `json:"from_name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UpsertEmailConfigRequest is the input for creating or updating email config.
type UpsertEmailConfigRequest struct {
	SMTPEnabled  *bool  `json:"smtp_enabled"`
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password"`
	FromAddress  string `json:"from_address"`
	FromName     string `json:"from_name"`
}
