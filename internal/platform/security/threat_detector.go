package security

import (
	"context"
	"strings"
	"time"
)

// ThreatDetector analyzes security events for potential threats
type ThreatDetector struct {
	rules []ThreatRule
}

// ThreatRule defines a threat detection rule
type ThreatRule struct {
	Name        string
	Description string
	Condition   func(event *SecurityEventDocument) bool
	Severity    Severity
	Indicator   string
}

// NewThreatDetector creates a new threat detector with default rules
func NewThreatDetector() *ThreatDetector {
	return &ThreatDetector{
		rules: DefaultThreatRules,
	}
}

// Detect checks an event against all rules and returns detected threat indicators
func (d *ThreatDetector) Detect(ctx context.Context, event *SecurityEventDocument) []string {
	var indicators []string

	for _, rule := range d.rules {
		if rule.Condition(event) {
			indicators = append(indicators, rule.Indicator)
		}
	}

	return indicators
}

// DefaultThreatRules contains the default threat detection rules
var DefaultThreatRules = []ThreatRule{
	{
		Name:        "unusual_user_agent",
		Description: "Detects requests with suspicious user agents",
		Indicator:   ThreatUnusualUserAgent,
		Condition: func(e *SecurityEventDocument) bool {
			ua := strings.ToLower(e.Event.UserAgent)

			// Empty user agent
			if ua == "" {
				return true
			}

			// Known automation tools
			suspiciousPatterns := []string{
				"curl/",
				"wget/",
				"python-requests",
				"go-http-client",
				"java/",
				"libwww-perl",
				"scrapy",
				"bot",
				"crawler",
				"spider",
			}

			for _, pattern := range suspiciousPatterns {
				if strings.Contains(ua, pattern) {
					return true
				}
			}

			return false
		},
	},
	{
		Name:        "after_hours_access",
		Description: "Detects access outside normal business hours",
		Indicator:   ThreatAfterHoursAccess,
		Condition: func(e *SecurityEventDocument) bool {
			// Only for sensitive events
			sensitiveEvents := map[string]bool{
				EventSensitiveDataView: true,
				EventBulkExport:        true,
				EventPIIAccess:         true,
				EventDataDownload:      true,
			}

			if !sensitiveEvents[e.Event.Type] {
				return false
			}

			hour := e.Timestamp.Hour()
			// Consider 22:00 - 06:00 as after hours
			return hour < 6 || hour >= 22
		},
	},
	{
		Name:        "suspicious_bot",
		Description: "Detects bot-like behavior patterns",
		Indicator:   ThreatSuspiciousBot,
		Condition: func(e *SecurityEventDocument) bool {
			ua := strings.ToLower(e.Event.UserAgent)

			// Check for headless browsers
			headlessPatterns := []string{
				"headlesschrome",
				"phantomjs",
				"selenium",
				"puppeteer",
				"playwright",
			}

			for _, pattern := range headlessPatterns {
				if strings.Contains(ua, pattern) {
					return true
				}
			}

			return false
		},
	},
	{
		Name:        "bulk_data_access",
		Description: "Detects large data exports",
		Indicator:   ThreatBulkDataAccess,
		Condition: func(e *SecurityEventDocument) bool {
			if e.Event.Type != EventBulkExport {
				return false
			}

			if count, ok := e.Event.Details["record_count"].(int); ok {
				return count > 1000
			}
			if count, ok := e.Event.Details["record_count"].(int64); ok {
				return count > 1000
			}
			if count, ok := e.Event.Details["record_count"].(float64); ok {
				return count > 1000
			}

			return false
		},
	},
	{
		Name:        "role_escalation",
		Description: "Detects privilege escalation attempts",
		Indicator:   ThreatRoleEscalation,
		Condition: func(e *SecurityEventDocument) bool {
			if e.Event.Type != EventRoleEscalation {
				return false
			}

			// Check if escalating to admin/superadmin
			if newRole, ok := e.Event.Details["new_role"].(string); ok {
				newRoleLower := strings.ToLower(newRole)
				return strings.Contains(newRoleLower, "admin") || strings.Contains(newRoleLower, "super")
			}

			return false
		},
	},
}

// AddRule adds a custom threat rule
func (d *ThreatDetector) AddRule(rule ThreatRule) {
	d.rules = append(d.rules, rule)
}

// AnalyzeLoginPattern analyzes login patterns for a user
func (d *ThreatDetector) AnalyzeLoginPattern(recentLogins []SecurityEventDocument) []string {
	var indicators []string

	if len(recentLogins) < 2 {
		return indicators
	}

	// Check for impossible travel
	for i := 1; i < len(recentLogins); i++ {
		prev := recentLogins[i-1]
		curr := recentLogins[i]

		if prev.Event.GeoLocation == nil || curr.Event.GeoLocation == nil {
			continue
		}

		// Calculate time difference
		timeDiff := prev.Timestamp.Sub(curr.Timestamp)
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}

		// If logins are less than 1 hour apart
		if timeDiff < time.Hour {
			// Check if locations are significantly different
			if isImpossibleTravel(prev.Event.GeoLocation, curr.Event.GeoLocation) {
				indicators = append(indicators, ThreatImpossibleTravel)
				break
			}
		}
	}

	return indicators
}

// isImpossibleTravel checks if two locations are impossibly far apart for the time window
func isImpossibleTravel(loc1, loc2 *GeoLocation) bool {
	// Simple check: if countries are different, consider it suspicious
	if loc1.Country != loc2.Country && loc1.Country != "" && loc2.Country != "" {
		return true
	}

	// For more accurate detection, you would calculate actual distance
	// and compare against reasonable travel speed (e.g., 900 km/h for flights)

	return false
}

// GetThreatSeverity returns the highest severity for given threat indicators
func GetThreatSeverity(indicators []string) Severity {
	if len(indicators) == 0 {
		return SeverityInfo
	}

	criticalThreats := map[string]bool{
		ThreatImpossibleTravel:  true,
		ThreatRoleEscalation:    true,
		ThreatCrossTenantAccess: true,
	}

	highThreats := map[string]bool{
		ThreatMultipleFailedAttempts: true,
		ThreatBulkDataAccess:         true,
	}

	for _, indicator := range indicators {
		if criticalThreats[indicator] {
			return SeverityCritical
		}
	}

	for _, indicator := range indicators {
		if highThreats[indicator] {
			return SeverityHigh
		}
	}

	return SeverityWarning
}
