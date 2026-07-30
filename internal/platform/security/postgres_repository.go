package security

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// PostgresRepository implements security event storage using PostgreSQL
type PostgresRepository struct {
	db *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL-backed security repository
func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Record inserts a security event into platform.security_events
func (r *PostgresRepository) Record(ctx context.Context, doc *SecurityEventDocument) error {
	geoJSON, err := marshalNullableJSON(doc.Event.GeoLocation)
	if err != nil {
		return fmt.Errorf("failed to marshal geo_location: %w", err)
	}

	detailsJSON, err := marshalNullableJSON(doc.Event.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	query := `
		INSERT INTO platform.security_events (
			timestamp, tenant_id, severity, category, event_type,
			actor_id, actor_email, ip_address, user_agent,
			geo_location, details, threat_indicators,
			alert_sent, alert_channels
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12,
			$13, $14
		) RETURNING id`

	err = r.db.QueryRow(ctx, query,
		doc.Timestamp,
		doc.Metadata.TenantID,
		string(doc.Metadata.Severity),
		string(doc.Metadata.Category),
		doc.Event.Type,
		doc.Event.ActorID,
		doc.Event.ActorEmail,
		doc.Event.IPAddress,
		doc.Event.UserAgent,
		geoJSON,
		detailsJSON,
		doc.Event.ThreatIndicators,
		doc.Event.AlertSent,
		doc.Event.AlertChannels,
	).Scan(&doc.ID)

	if err != nil {
		return fmt.Errorf("failed to insert security event: %w", err)
	}

	return nil
}

// GetEvents retrieves security events with filters and pagination
func (r *PostgresRepository) GetEvents(ctx context.Context, filters SecurityFilters, page, pageSize int) ([]SecurityEventDocument, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filters.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIdx))
		args = append(args, *filters.TenantID)
		argIdx++
	}
	if filters.Severity != nil {
		conditions = append(conditions, fmt.Sprintf("severity = $%d", argIdx))
		args = append(args, string(*filters.Severity))
		argIdx++
	}
	if filters.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, string(*filters.Category))
		argIdx++
	}
	if filters.EventType != nil {
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", argIdx))
		args = append(args, *filters.EventType)
		argIdx++
	}
	if filters.IPAddress != nil {
		conditions = append(conditions, fmt.Sprintf("ip_address = $%d", argIdx))
		args = append(args, *filters.IPAddress)
		argIdx++
	}
	if filters.ActorID != nil {
		conditions = append(conditions, fmt.Sprintf("actor_id = $%d", argIdx))
		args = append(args, *filters.ActorID)
		argIdx++
	}
	if filters.FromDate != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIdx))
		args = append(args, *filters.FromDate)
		argIdx++
	}
	if filters.ToDate != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argIdx))
		args = append(args, *filters.ToDate)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM platform.security_events %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count security events: %w", err)
	}

	if total == 0 {
		return []SecurityEventDocument{}, 0, nil
	}

	// Fetch page
	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, timestamp, tenant_id, severity, category, event_type,
		       actor_id, actor_email, ip_address, user_agent,
		       geo_location, details, threat_indicators,
		       alert_sent, alert_channels
		FROM platform.security_events
		%s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query security events: %w", err)
	}
	defer rows.Close()

	var events []SecurityEventDocument
	for rows.Next() {
		doc, err := scanSecurityEvent(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan security event: %w", err)
		}
		events = append(events, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating security events: %w", err)
	}

	return events, total, nil
}

// GetByID retrieves a single security event by its ID
func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*SecurityEventDocument, error) {
	query := `
		SELECT id, timestamp, tenant_id, severity, category, event_type,
		       actor_id, actor_email, ip_address, user_agent,
		       geo_location, details, threat_indicators,
		       alert_sent, alert_channels
		FROM platform.security_events
		WHERE id = $1`

	row := r.db.QueryRow(ctx, query, id)

	var doc SecurityEventDocument
	var geoJSON, detailsJSON []byte

	err := row.Scan(
		&doc.ID,
		&doc.Timestamp,
		&doc.Metadata.TenantID,
		&doc.Metadata.Severity,
		&doc.Metadata.Category,
		&doc.Event.Type,
		&doc.Event.ActorID,
		&doc.Event.ActorEmail,
		&doc.Event.IPAddress,
		&doc.Event.UserAgent,
		&geoJSON,
		&detailsJSON,
		&doc.Event.ThreatIndicators,
		&doc.Event.AlertSent,
		&doc.Event.AlertChannels,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find security event: %w", err)
	}

	if err := unmarshalNullableJSON(geoJSON, &doc.Event.GeoLocation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal geo_location: %w", err)
	}
	if err := unmarshalNullableJSON(detailsJSON, &doc.Event.Details); err != nil {
		return nil, fmt.Errorf("failed to unmarshal details: %w", err)
	}

	return &doc, nil
}

// CountRecentFailedLogins counts failed login attempts from an IP since the given time
func (r *PostgresRepository) CountRecentFailedLogins(ctx context.Context, ipAddress string, since time.Time) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM platform.security_events
		WHERE event_type = $1
		  AND ip_address = $2
		  AND timestamp >= $3`

	var count int64
	err := r.db.QueryRow(ctx, query, EventLoginFailed, ipAddress, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count failed logins: %w", err)
	}

	return count, nil
}

// DashboardData contains aggregated security metrics for the dashboard
type DashboardData struct {
	PeriodHours    int                      `json:"period_hours"`
	BySeverity     map[string]int64         `json:"by_severity"`
	TopEventTypes  []map[string]interface{} `json:"top_event_types"`
	TopIPAddresses []map[string]interface{} `json:"top_ip_addresses"`
	TotalEvents    int64                    `json:"total_events"`
	CriticalCount  int64                    `json:"critical_count"`
	HighCount      int64                    `json:"high_count"`
}

// GetSecurityDashboard returns aggregated security metrics
func (r *PostgresRepository) GetSecurityDashboard(ctx context.Context, tenantID *int64, hours int) (*DashboardData, error) {
	since := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)

	dashboard := &DashboardData{
		PeriodHours:    hours,
		BySeverity:     make(map[string]int64),
		TopEventTypes:  make([]map[string]interface{}, 0),
		TopIPAddresses: make([]map[string]interface{}, 0),
	}

	// Build tenant filter
	tenantFilter := ""
	var tenantArgs []interface{}
	argIdx := 2 // $1 is always since

	if tenantID != nil {
		tenantFilter = fmt.Sprintf("AND tenant_id = $%d", argIdx)
		tenantArgs = append(tenantArgs, *tenantID)
	}

	// Count by severity
	severityQuery := fmt.Sprintf(`
		SELECT severity, COUNT(*)
		FROM platform.security_events
		WHERE timestamp >= $1 %s
		GROUP BY severity`, tenantFilter)

	severityArgs := append([]interface{}{since}, tenantArgs...)
	rows, err := r.db.Query(ctx, severityQuery, severityArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query severity counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var severity string
		var count int64
		if err := rows.Scan(&severity, &count); err != nil {
			return nil, fmt.Errorf("failed to scan severity row: %w", err)
		}
		dashboard.BySeverity[severity] = count
		dashboard.TotalEvents += count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating severity rows: %w", err)
	}

	dashboard.CriticalCount = dashboard.BySeverity["CRITICAL"]
	dashboard.HighCount = dashboard.BySeverity["HIGH"]

	// Top event types
	typeQuery := fmt.Sprintf(`
		SELECT event_type, COUNT(*) as cnt
		FROM platform.security_events
		WHERE timestamp >= $1 %s
		GROUP BY event_type
		ORDER BY cnt DESC
		LIMIT 10`, tenantFilter)

	typeArgs := append([]interface{}{since}, tenantArgs...)
	typeRows, err := r.db.Query(ctx, typeQuery, typeArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query event type counts: %w", err)
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var eventType string
		var count int64
		if err := typeRows.Scan(&eventType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan event type row: %w", err)
		}
		dashboard.TopEventTypes = append(dashboard.TopEventTypes, map[string]interface{}{
			"event_type": eventType,
			"count":      count,
		})
	}
	if err := typeRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event type rows: %w", err)
	}

	// Top IP addresses by event count
	ipQuery := fmt.Sprintf(`
		SELECT ip_address, COUNT(*) AS cnt
		FROM platform.security_events
		WHERE timestamp >= $1
		  AND ip_address IS NOT NULL
		  AND ip_address <> '' %s
		GROUP BY ip_address
		ORDER BY cnt DESC
		LIMIT 20`, tenantFilter)

	ipArgs := append([]interface{}{since}, tenantArgs...)
	ipRows, err := r.db.Query(ctx, ipQuery, ipArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query top IP addresses: %w", err)
	}
	defer ipRows.Close()

	for ipRows.Next() {
		var ip string
		var count int64
		if err := ipRows.Scan(&ip, &count); err != nil {
			return nil, fmt.Errorf("failed to scan IP row: %w", err)
		}
		dashboard.TopIPAddresses = append(dashboard.TopIPAddresses, map[string]interface{}{
			"ip":    ip,
			"count": count,
		})
	}
	if err := ipRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating IP rows: %w", err)
	}

	return dashboard, nil
}

// IPActorRecord holds one distinct actor seen from a given IP address
type IPActorRecord struct {
	ActorID    *int64   `json:"actor_id"`
	ActorEmail string   `json:"actor_email"`
	EventCount int64    `json:"event_count"`
	LastSeen   string   `json:"last_seen"`
	EventTypes []string `json:"event_types"`
}

// GetIPActors returns distinct actors that have generated security events from the given IP address
func (r *PostgresRepository) GetIPActors(ctx context.Context, tenantID *int64, ip string) ([]IPActorRecord, error) {
	tenantClause := ""
	args := []interface{}{ip}
	if tenantID != nil {
		tenantClause = fmt.Sprintf("AND tenant_id = $%d", len(args)+1)
		args = append(args, *tenantID)
	}

	query := fmt.Sprintf(`
		SELECT
			actor_id,
			COALESCE(actor_email, '') AS actor_email,
			COUNT(*)                  AS event_count,
			MAX(timestamp)            AS last_seen,
			array_agg(DISTINCT event_type) AS event_types
		FROM platform.security_events
		WHERE ip_address = $1 %s
		GROUP BY actor_id, actor_email
		ORDER BY event_count DESC
		LIMIT 50
	`, tenantClause)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query IP actors: %w", err)
	}
	defer rows.Close()

	var records []IPActorRecord
	for rows.Next() {
		var rec IPActorRecord
		var lastSeen time.Time
		if err := rows.Scan(&rec.ActorID, &rec.ActorEmail, &rec.EventCount, &lastSeen, &rec.EventTypes); err != nil {
			return nil, fmt.Errorf("failed to scan IP actor row: %w", err)
		}
		rec.LastSeen = lastSeen.UTC().Format(time.RFC3339)
		records = append(records, rec)
	}
	if records == nil {
		records = []IPActorRecord{}
	}
	return records, rows.Err()
}

// CountMaliciousFileUploads counts how many malicious file upload events exist for a user since the given time
func (r *PostgresRepository) CountMaliciousFileUploads(ctx context.Context, userID int64, since time.Time) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM platform.security_events
		WHERE event_type = $1
		  AND actor_id = $2
		  AND timestamp >= $3`

	var count int64
	err := r.db.QueryRow(ctx, query, EventMaliciousFileUpload, userID, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count malicious uploads: %w", err)
	}
	return count, nil
}

// scannable is an interface for pgx Row and Rows
type scannable interface {
	Scan(dest ...interface{}) error
}

// scanSecurityEvent scans a row into a SecurityEventDocument
func scanSecurityEvent(row scannable) (SecurityEventDocument, error) {
	var doc SecurityEventDocument
	var geoJSON, detailsJSON []byte

	err := row.Scan(
		&doc.ID,
		&doc.Timestamp,
		&doc.Metadata.TenantID,
		&doc.Metadata.Severity,
		&doc.Metadata.Category,
		&doc.Event.Type,
		&doc.Event.ActorID,
		&doc.Event.ActorEmail,
		&doc.Event.IPAddress,
		&doc.Event.UserAgent,
		&geoJSON,
		&detailsJSON,
		&doc.Event.ThreatIndicators,
		&doc.Event.AlertSent,
		&doc.Event.AlertChannels,
	)
	if err != nil {
		return doc, err
	}

	if err := unmarshalNullableJSON(geoJSON, &doc.Event.GeoLocation); err != nil {
		return doc, fmt.Errorf("failed to unmarshal geo_location: %w", err)
	}
	if err := unmarshalNullableJSON(detailsJSON, &doc.Event.Details); err != nil {
		return doc, fmt.Errorf("failed to unmarshal details: %w", err)
	}

	return doc, nil
}

// marshalNullableJSON marshals a value to JSON bytes, returning nil for nil values
func marshalNullableJSON(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}

// unmarshalNullableJSON unmarshals JSON bytes into a target, skipping nil/empty input
func unmarshalNullableJSON(data []byte, target interface{}) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, target)
}
