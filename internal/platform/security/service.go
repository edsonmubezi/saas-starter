package security

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Service provides security event logging functionality
type Service struct {
	db              *pgxpool.Pool
	repo            *PostgresRepository
	threatDetector  *ThreatDetector
	alertingEnabled bool
}

// NewService creates a new security service
func NewService(db *pgxpool.Pool, alertingEnabled bool) *Service {
	repo := NewPostgresRepository(db)
	return &Service{
		db:              db,
		repo:            repo,
		threatDetector:  NewThreatDetector(),
		alertingEnabled: alertingEnabled,
	}
}

// EventBuilder provides a fluent interface for building security events
type EventBuilder struct {
	doc *SecurityEventDocument
	svc *Service
}

// NewEvent creates a new event builder
func (s *Service) NewEvent() *EventBuilder {
	return &EventBuilder{
		doc: &SecurityEventDocument{
			Timestamp: time.Now().UTC(),
			Metadata: SecurityMetadata{
				Severity: SeverityInfo,
				Category: CategoryAuthentication,
			},
			Event: SecurityEventDetails{
				Details: make(map[string]interface{}),
			},
		},
		svc: s,
	}
}

// Type sets the event type
func (b *EventBuilder) Type(eventType string) *EventBuilder {
	b.doc.Event.Type = eventType
	return b
}

// Tenant sets the event tenant
func (b *EventBuilder) Tenant(tenantID int64) *EventBuilder {
	b.doc.Metadata.TenantID = &tenantID
	return b
}

// Severity sets the event severity
func (b *EventBuilder) Severity(severity Severity) *EventBuilder {
	b.doc.Metadata.Severity = severity
	return b
}

// Category sets the event category
func (b *EventBuilder) Category(category Category) *EventBuilder {
	b.doc.Metadata.Category = category
	return b
}

// Actor sets the actor information
func (b *EventBuilder) Actor(actorID *int64, email *string) *EventBuilder {
	b.doc.Event.ActorID = actorID
	b.doc.Event.ActorEmail = email
	return b
}

// Request sets HTTP request context
func (b *EventBuilder) Request(ipAddress, userAgent string) *EventBuilder {
	b.doc.Event.IPAddress = ipAddress
	b.doc.Event.UserAgent = userAgent
	return b
}

// GeoLocation sets the IP geolocation
func (b *EventBuilder) GeoLocation(geo *GeoLocation) *EventBuilder {
	b.doc.Event.GeoLocation = geo
	return b
}

// Detail adds a detail to the event
func (b *EventBuilder) Detail(key string, value interface{}) *EventBuilder {
	b.doc.Event.Details[key] = value
	return b
}

// Details sets multiple details at once
func (b *EventBuilder) Details(details map[string]interface{}) *EventBuilder {
	for k, v := range details {
		b.doc.Event.Details[k] = v
	}
	return b
}

// ThreatIndicator adds a threat indicator
func (b *EventBuilder) ThreatIndicator(indicator string) *EventBuilder {
	b.doc.Event.ThreatIndicators = append(b.doc.Event.ThreatIndicators, indicator)
	return b
}

// ThreatIndicators adds multiple threat indicators
func (b *EventBuilder) ThreatIndicators(indicators []string) *EventBuilder {
	b.doc.Event.ThreatIndicators = append(b.doc.Event.ThreatIndicators, indicators...)
	return b
}

// Build finalizes and returns the event document
func (b *EventBuilder) Build() *SecurityEventDocument {
	return b.doc
}

// Record saves a security event to PostgreSQL
func (s *Service) Record(ctx context.Context, event *SecurityEventDocument) error {
	if s.repo == nil {
		return nil
	}

	// Run threat detection
	threats := s.threatDetector.Detect(ctx, event)
	event.Event.ThreatIndicators = append(event.Event.ThreatIndicators, threats...)

	// Upgrade severity if critical threats detected
	if len(threats) > 0 && event.Metadata.Severity != SeverityCritical {
		for _, t := range threats {
			if t == ThreatBulkDataAccess || t == ThreatRoleEscalation || t == ThreatImpossibleTravel {
				event.Metadata.Severity = SeverityCritical
				break
			}
		}
	}

	return s.repo.Record(ctx, event)
}

// RecordAsync saves a security event asynchronously
func (s *Service) RecordAsync(event *SecurityEventDocument) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.Record(ctx, event); err != nil {
			log.Printf("ERROR [security] Failed to record security event (type=%s): %v", event.Event.Type, err)
		}
	}()
}

// RecordLoginSuccess records a successful login
func (s *Service) RecordLoginSuccess(ctx context.Context, tenantID, userID int64, email, ipAddress, userAgent string) error {
	event := s.NewEvent().
		Type(EventLoginSuccess).
		Tenant(tenantID).
		Category(CategoryAuthentication).
		Severity(SeverityInfo).
		Actor(&userID, &email).
		Request(ipAddress, userAgent).
		Build()

	return s.Record(ctx, event)
}

// RecordLoginFailed records a failed login attempt
func (s *Service) RecordLoginFailed(ctx context.Context, tenantID *int64, email, reason, ipAddress, userAgent string) error {
	builder := s.NewEvent().
		Type(EventLoginFailed).
		Category(CategoryAuthentication).
		Severity(SeverityWarning).
		Actor(nil, &email).
		Request(ipAddress, userAgent).
		Detail("reason", reason)

	if tenantID != nil {
		builder.Tenant(*tenantID)
	}

	event := builder.Build()

	// Check for brute force pattern
	count, _ := s.CountRecentFailedLogins(ctx, ipAddress, 5*time.Minute)
	if count >= 3 {
		event.Event.ThreatIndicators = append(event.Event.ThreatIndicators, ThreatMultipleFailedAttempts)
		event.Event.Details["attempt_count"] = count + 1
		event.Metadata.Severity = SeverityHigh
	}

	return s.Record(ctx, event)
}

// RecordAccountLocked records an account lock event
func (s *Service) RecordAccountLocked(ctx context.Context, tenantID, userID int64, email, reason, ipAddress string) error {
	event := s.NewEvent().
		Type(EventAccountLocked).
		Tenant(tenantID).
		Category(CategoryAuthentication).
		Severity(SeverityHigh).
		Actor(&userID, &email).
		Request(ipAddress, "").
		Detail("reason", reason).
		ThreatIndicator(ThreatMultipleFailedAttempts).
		Build()

	return s.Record(ctx, event)
}

// RecordPermissionDenied records an authorization failure
func (s *Service) RecordPermissionDenied(ctx context.Context, tenantID, userID int64, email, resource, action, ipAddress string) error {
	event := s.NewEvent().
		Type(EventPermissionDenied).
		Tenant(tenantID).
		Category(CategoryAuthorization).
		Severity(SeverityWarning).
		Actor(&userID, &email).
		Request(ipAddress, "").
		Detail("resource", resource).
		Detail("action", action).
		Build()

	return s.Record(ctx, event)
}

// RecordCrossTenantAttempt records a cross-tenant access attempt
func (s *Service) RecordCrossTenantAttempt(ctx context.Context, userTenantID, attemptedTenantID, userID int64, email, ipAddress string) error {
	event := s.NewEvent().
		Type(EventCrossTenantAttempt).
		Tenant(userTenantID).
		Category(CategoryAuthorization).
		Severity(SeverityCritical).
		Actor(&userID, &email).
		Request(ipAddress, "").
		Detail("attempted_tenant_id", attemptedTenantID).
		ThreatIndicator(ThreatCrossTenantAccess).
		Build()

	return s.Record(ctx, event)
}

// RecordBulkExport records a bulk data export
func (s *Service) RecordBulkExport(ctx context.Context, tenantID, userID int64, email, dataType string, recordCount int, ipAddress string) error {
	severity := SeverityInfo
	var indicators []string

	if recordCount > 1000 {
		severity = SeverityHigh
		indicators = append(indicators, ThreatBulkDataAccess)
	}

	event := s.NewEvent().
		Type(EventBulkExport).
		Tenant(tenantID).
		Category(CategoryDataAccess).
		Severity(severity).
		Actor(&userID, &email).
		Request(ipAddress, "").
		Detail("data_type", dataType).
		Detail("record_count", recordCount).
		ThreatIndicators(indicators).
		Build()

	return s.Record(ctx, event)
}

// RecordRateLimitHit records a rate limit hit
func (s *Service) RecordRateLimitHit(ctx context.Context, tenantID *int64, email, endpoint, ipAddress string) error {
	builder := s.NewEvent().
		Type(EventRateLimitHit).
		Category(CategoryAnomaly).
		Severity(SeverityWarning).
		Request(ipAddress, "").
		Detail("endpoint", endpoint)

	if email != "" {
		builder.Actor(nil, &email)
	}
	if tenantID != nil {
		builder.Tenant(*tenantID)
	}

	return s.Record(ctx, builder.Build())
}

// RecordMaliciousFileUpload records a virus/malware detection event during file upload
func (s *Service) RecordMaliciousFileUpload(ctx context.Context, tenantID *int64, userID *int64, email, filename, virusName, ipAddress, userAgent string) error {
	builder := s.NewEvent().
		Type(EventMaliciousFileUpload).
		Category(CategoryAnomaly).
		Severity(SeverityHigh).
		Actor(userID, &email).
		Request(ipAddress, userAgent).
		Detail("filename", filename).
		Detail("virus_name", virusName).
		ThreatIndicator(ThreatMaliciousFile)

	if tenantID != nil {
		builder.Tenant(*tenantID)
	}

	return s.Record(ctx, builder.Build())
}

// CountMaliciousFileUploads returns how many malicious file upload events exist for a user since the given time
func (s *Service) CountMaliciousFileUploads(ctx context.Context, userID int64, since time.Time) (int64, error) {
	if s.repo == nil {
		return 0, nil
	}
	return s.repo.CountMaliciousFileUploads(ctx, userID, since)
}

// GetEvents retrieves security events with filters and pagination
func (s *Service) GetEvents(ctx context.Context, filters SecurityFilters, page, pageSize int) ([]SecurityEventDocument, int64, error) {
	if s.repo == nil {
		return []SecurityEventDocument{}, 0, nil
	}

	return s.repo.GetEvents(ctx, filters, page, pageSize)
}

// GetByID retrieves a single security event by ID
func (s *Service) GetByID(ctx context.Context, id int64) (*SecurityEventDocument, error) {
	if s.repo == nil {
		return nil, nil
	}

	return s.repo.GetByID(ctx, id)
}

// CountRecentFailedLogins counts failed login attempts from an IP in a time window
func (s *Service) CountRecentFailedLogins(ctx context.Context, ipAddress string, window time.Duration) (int64, error) {
	if s.repo == nil {
		return 0, nil
	}

	since := time.Now().Add(-window)
	return s.repo.CountRecentFailedLogins(ctx, ipAddress, since)
}

// GetSecurityDashboard returns security metrics for dashboard
func (s *Service) GetSecurityDashboard(ctx context.Context, tenantID int64, hours int) (map[string]interface{}, error) {
	if s.repo == nil {
		return map[string]interface{}{
			"total_events":          int64(0),
			"events_by_severity":    map[string]int64{},
			"events_by_category":    map[string]int64{},
			"recent_threats":        []SecurityEventResponse{},
			"top_ip_addresses":      []map[string]interface{}{},
			"failed_logins":         int64(0),
			"suspicious_activities": int64(0),
			"hours_analyzed":        hours,
		}, nil
	}

	dashboard, err := s.repo.GetSecurityDashboard(ctx, &tenantID, hours)
	if err != nil {
		return nil, err
	}

	// Get recent threats (high/critical events)
	threatFilters := SecurityFilters{TenantID: &tenantID}
	highSev := SeverityHigh
	threatFilters.Severity = &highSev
	recentThreats := []SecurityEventResponse{}
	if threats, _, thErr := s.repo.GetEvents(ctx, threatFilters, 1, 5); thErr == nil {
		for _, t := range threats {
			recentThreats = append(recentThreats, t.ToResponse())
		}
	}

	// Derive failed logins and suspicious activities from event types
	failedLogins := int64(0)
	suspiciousActivities := int64(0)
	for _, et := range dashboard.TopEventTypes {
		if evtType, ok := et["event_type"].(string); ok {
			count, _ := et["count"].(int64)
			if evtType == EventLoginFailed {
				failedLogins = count
			}
			if evtType == EventBruteForce || evtType == EventSuspiciousPattern || evtType == EventUnusualActivity || evtType == EventRateLimitHit {
				suspiciousActivities += count
			}
		}
	}

	accountLockouts := int64(0)
	for _, et := range dashboard.TopEventTypes {
		if evtType, ok := et["event_type"].(string); ok && evtType == EventAccountLocked {
			count, _ := et["count"].(int64)
			accountLockouts = count
		}
	}

	return map[string]interface{}{
		"total_events":          dashboard.TotalEvents,
		"critical_count":        dashboard.CriticalCount,
		"high_count":            dashboard.HighCount,
		"events_by_severity":    dashboard.BySeverity,
		"events_by_category":    map[string]int64{},
		"top_event_types":       dashboard.TopEventTypes,
		"recent_threats":        recentThreats,
		"top_ip_addresses":      dashboard.TopIPAddresses,
		"failed_logins":         failedLogins,
		"suspicious_activities": suspiciousActivities,
		"account_lockouts":      accountLockouts,
		"hours_analyzed":        hours,
	}, nil
}

// GetIPActors returns the actors (users) associated with a given IP address
func (s *Service) GetIPActors(ctx context.Context, tenantID *int64, ip string) ([]IPActorRecord, error) {
	if s.repo == nil {
		return []IPActorRecord{}, nil
	}
	return s.repo.GetIPActors(ctx, tenantID, ip)
}
