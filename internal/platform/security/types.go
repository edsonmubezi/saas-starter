package security

import (
	"strconv"
	"time"
)

// Severity levels for security events
type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// Category of security events
type Category string

const (
	CategoryAuthentication Category = "authentication"
	CategoryAuthorization  Category = "authorization"
	CategoryDataAccess     Category = "data_access"
	CategorySystem         Category = "system"
	CategoryAnomaly        Category = "anomaly"
)

// SecurityEventDocument represents a security event stored in PostgreSQL
type SecurityEventDocument struct {
	ID        int64                `json:"id"`
	Timestamp time.Time            `json:"timestamp"`
	Metadata  SecurityMetadata     `json:"metadata"`
	Event     SecurityEventDetails `json:"event"`
}

// SecurityMetadata contains classification information
type SecurityMetadata struct {
	TenantID *int64   `json:"tenant_id,omitempty"`
	Severity Severity `json:"severity"`
	Category Category `json:"category"`
}

// SecurityEventDetails contains the event specifics
type SecurityEventDetails struct {
	Type             string                 `json:"type"`
	ActorEmail       *string                `json:"actor_email,omitempty"`
	ActorID          *int64                 `json:"actor_id,omitempty"`
	IPAddress        string                 `json:"ip_address"`
	UserAgent        string                 `json:"user_agent,omitempty"`
	GeoLocation      *GeoLocation           `json:"geo_location,omitempty"`
	Details          map[string]interface{} `json:"details,omitempty"`
	ThreatIndicators []string               `json:"threat_indicators,omitempty"`
	AlertSent        bool                   `json:"alert_sent"`
	AlertChannels    []string               `json:"alert_channels,omitempty"`
}

// GeoLocation contains IP geolocation data
type GeoLocation struct {
	Country   string  `json:"country,omitempty"`
	City      string  `json:"city,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

// Security Event Types
const (
	// Authentication Events
	EventLoginSuccess       = "LOGIN_SUCCESS"
	EventLoginFailed        = "LOGIN_FAILED"
	EventLogout             = "LOGOUT"
	EventPasswordReset      = "PASSWORD_RESET"
	EventPasswordChanged    = "PASSWORD_CHANGED"
	EventMFAEnabled         = "MFA_ENABLED"
	EventMFADisabled        = "MFA_DISABLED"
	EventMFAChallengePassed = "MFA_CHALLENGE_PASSED"
	EventMFAChallengeFailed = "MFA_CHALLENGE_FAILED"
	EventAccountLocked      = "ACCOUNT_LOCKED"
	EventAccountUnlocked    = "ACCOUNT_UNLOCKED"
	EventSessionCreated     = "SESSION_CREATED"
	EventSessionRevoked     = "SESSION_REVOKED"
	EventTokenRefreshed     = "TOKEN_REFRESHED"

	// Authorization Events
	EventPermissionDenied     = "PERMISSION_DENIED"
	EventRoleEscalation       = "ROLE_ESCALATION"
	EventCrossTenantAttempt   = "CROSS_TENANT_ATTEMPT"
	EventUnauthorizedResource = "UNAUTHORIZED_RESOURCE"

	// Data Access Events
	EventBulkExport        = "BULK_EXPORT"
	EventSensitiveDataView = "SENSITIVE_DATA_VIEW"
	EventPIIAccess         = "PII_ACCESS"
	EventDataDownload      = "DATA_DOWNLOAD"

	// System Events
	EventConfigChanged    = "CONFIG_CHANGED"
	EventAPIKeyCreated    = "API_KEY_CREATED"
	EventAPIKeyRevoked    = "API_KEY_REVOKED"
	EventSecuritySettings = "SECURITY_SETTINGS_CHANGED"

	// Anomaly Events
	EventUnusualActivity     = "UNUSUAL_ACTIVITY"
	EventRateLimitHit        = "RATE_LIMIT_HIT"
	EventSuspiciousPattern   = "SUSPICIOUS_PATTERN"
	EventBruteForce          = "BRUTE_FORCE_DETECTED"
	EventImpossibleTravel    = "IMPOSSIBLE_TRAVEL"
	EventMaliciousFileUpload = "MALICIOUS_FILE_UPLOAD"
)

// Threat Indicators
const (
	ThreatMultipleFailedAttempts = "multiple_failed_attempts"
	ThreatUnusualUserAgent       = "unusual_user_agent"
	ThreatUnknownIP              = "unknown_ip"
	ThreatAfterHoursAccess       = "after_hours_access"
	ThreatImpossibleTravel       = "impossible_travel"
	ThreatBulkDataAccess         = "bulk_data_access"
	ThreatRoleEscalation         = "role_escalation"
	ThreatCrossTenantAccess      = "cross_tenant_access"
	ThreatSuspiciousBot          = "suspicious_bot"
	ThreatMaliciousFile          = "malicious_file_upload"
)

// SecurityFilters for querying security events
type SecurityFilters struct {
	TenantID  *int64
	Severity  *Severity
	Category  *Category
	EventType *string
	IPAddress *string
	ActorID   *int64
	FromDate  *time.Time
	ToDate    *time.Time
}

// SecurityEventResponse for API responses
type SecurityEventResponse struct {
	ID               string                 `json:"id"`
	CreatedAt        time.Time              `json:"created_at"`
	TenantID         *int64                 `json:"tenant_id,omitempty"`
	Severity         Severity               `json:"severity"`
	Category         Category               `json:"category"`
	EventType        string                 `json:"event_type"`
	Description      string                 `json:"description,omitempty"`
	ActorEmail       *string                `json:"actor_email,omitempty"`
	ActorID          *int64                 `json:"actor_id,omitempty"`
	ActorName        string                 `json:"actor_name,omitempty"`
	IPAddress        string                 `json:"ip_address"`
	UserAgent        string                 `json:"user_agent,omitempty"`
	RequestID        string                 `json:"request_id,omitempty"`
	ResourceType     string                 `json:"resource_type,omitempty"`
	ResourceID       *int64                 `json:"resource_id,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	ThreatIndicators []string               `json:"threat_indicators,omitempty"`
	RiskScore        int                    `json:"risk_score,omitempty"`
}

// ToResponse converts a document to API response format
func (d *SecurityEventDocument) ToResponse() SecurityEventResponse {
	riskScore := 0
	switch d.Metadata.Severity {
	case SeverityCritical:
		riskScore = 90
	case SeverityHigh:
		riskScore = 70
	case SeverityWarning:
		riskScore = 40
	case SeverityInfo:
		riskScore = 10
	}
	riskScore += len(d.Event.ThreatIndicators) * 10
	if riskScore > 100 {
		riskScore = 100
	}

	return SecurityEventResponse{
		ID:               strconv.FormatInt(d.ID, 10),
		CreatedAt:        d.Timestamp,
		TenantID:         d.Metadata.TenantID,
		Severity:         d.Metadata.Severity,
		Category:         d.Metadata.Category,
		EventType:        d.Event.Type,
		ActorEmail:       d.Event.ActorEmail,
		ActorID:          d.Event.ActorID,
		IPAddress:        d.Event.IPAddress,
		UserAgent:        d.Event.UserAgent,
		Metadata:         d.Event.Details,
		ThreatIndicators: d.Event.ThreatIndicators,
		RiskScore:        riskScore,
	}
}

// LoginFailureReason types
const (
	FailureReasonInvalidPassword = "invalid_password"
	FailureReasonUserNotFound    = "user_not_found"
	FailureReasonAccountLocked   = "account_locked"
	FailureReasonAccountDisabled = "account_disabled"
	FailureReason2FAFailed       = "2fa_failed"
	FailureReasonExpiredToken    = "expired_token"
	FailureReasonInvalidToken    = "invalid_token"
)
