package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// PostgresService provides PostgreSQL-based audit logging and querying
type PostgresService struct {
	db *pgxpool.Pool
}

// NewPostgresService creates a new PostgreSQL audit service
func NewPostgresService(db *pgxpool.Pool) *PostgresService {
	s := &PostgresService{db: db}
	if db != nil {
		s.ensureColumns()
	}
	return s
}

// ensureColumns adds columns that may not exist if migrations haven't fully run.
// Safe to call repeatedly (uses IF NOT EXISTS).
func (s *PostgresService) ensureColumns() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	columns := []string{
		"ALTER TABLE audit.events ADD COLUMN IF NOT EXISTS actor_email VARCHAR(255)",
		"ALTER TABLE audit.events ADD COLUMN IF NOT EXISTS actor_name VARCHAR(255)",
		"ALTER TABLE audit.events ADD COLUMN IF NOT EXISTS resource_name VARCHAR(255)",
		"ALTER TABLE audit.events ADD COLUMN IF NOT EXISTS endpoint VARCHAR(500)",
		"ALTER TABLE audit.events ADD COLUMN IF NOT EXISTS changes JSONB",
	}

	for _, q := range columns {
		if _, err := s.db.Exec(ctx, q); err != nil {
			log.Printf("WARNING [audit] ensureColumns: %v", err)
		}
	}
}

// Record saves an audit event directly to audit.events
func (s *PostgresService) Record(ctx context.Context, doc *AuditDocument) error {
	if s.db == nil {
		return nil
	}

	if doc.BeforeState != nil || doc.AfterState != nil {
		doc.Changes = ComputeChanges(doc.BeforeState, doc.AfterState)
	}

	doc.Signature = ComputeDocSignature(doc)

	var changesJSON []byte
	if len(doc.Changes) > 0 {
		changesJSON, _ = json.Marshal(doc.Changes)
	}

	var beforeJSON, afterJSON []byte
	if doc.BeforeState != nil {
		beforeJSON, _ = json.Marshal(doc.BeforeState)
	}
	if doc.AfterState != nil {
		afterJSON, _ = json.Marshal(doc.AfterState)
	}

	query := `
		INSERT INTO audit.events (
			audit_id, occurred_at, actor_type, actor_id, tenant_id,
			action, resource_type, resource_id, request_id, ip, user_agent,
			origin_service, severity, before_json, after_json,
			schema_version, sig_sha256,
			actor_email, actor_name, resource_name, endpoint, changes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22
		)
		ON CONFLICT (audit_id) DO NOTHING
	`

	_, err := s.db.Exec(ctx, query,
		doc.AuditID, doc.Timestamp, doc.ActorType, doc.ActorID, doc.TenantID,
		doc.Action, doc.ResourceType, doc.ResourceID, doc.RequestID, doc.IPAddress,
		doc.UserAgent, "saas-api", doc.Severity, beforeJSON, afterJSON,
		1, []byte(doc.Signature),
		doc.ActorEmail, doc.ActorName, doc.ResourceName, doc.Endpoint, changesJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to insert audit event: %w", err)
	}

	return nil
}

// RecordAsync saves an audit event asynchronously
func (s *PostgresService) RecordAsync(doc *AuditDocument) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.Record(ctx, doc); err != nil {
			log.Printf("ERROR [audit] Failed to record audit event (audit_id=%s, action=%s): %v", doc.AuditID, doc.Action, err)
		}
	}()
}

// RecordFromEvent converts an AuditEvent to AuditDocument and saves it
func (s *PostgresService) RecordFromEvent(ctx context.Context, event *AuditEvent, actorEmail, actorName string) error {
	doc := &AuditDocument{
		AuditID:      event.AuditID,
		Timestamp:    event.OccurredAt,
		TenantID:     event.TenantID,
		ActorType:    event.ActorType,
		ActorID:      event.ActorID,
		ActorEmail:   actorEmail,
		ActorName:    actorName,
		Action:       event.Action,
		ResourceType: event.ResourceType,
		ResourceID:   event.ResourceID,
		Severity:     event.Severity,
		RequestID:    event.RequestID,
		IPAddress:    event.IP,
		UserAgent:    event.UserAgent,
		BeforeState:  redactSensitive(event.BeforeJSON),
		AfterState:   redactSensitive(event.AfterJSON),
	}

	return s.Record(ctx, doc)
}

// RecordFromEventAsync converts and saves asynchronously
func (s *PostgresService) RecordFromEventAsync(event *AuditEvent, actorEmail, actorName string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.RecordFromEvent(ctx, event, actorEmail, actorName); err != nil {
			log.Printf("ERROR [audit] Failed to record audit event from event: %v", err)
		}
	}()
}

// GetEvents retrieves audit events with filters and pagination
func (s *PostgresService) GetEvents(ctx context.Context, filters AuditFilters, page, pageSize int) ([]AuditDocument, int64, error) {
	if s.db == nil {
		return []AuditDocument{}, 0, nil
	}

	where := "WHERE tenant_id = $1"
	args := []interface{}{filters.TenantID}
	argIdx := 2

	if filters.ResourceType != nil {
		where += fmt.Sprintf(" AND resource_type = $%d", argIdx)
		args = append(args, *filters.ResourceType)
		argIdx++
	}
	if filters.ResourceID != nil {
		where += fmt.Sprintf(" AND resource_id = $%d", argIdx)
		args = append(args, *filters.ResourceID)
		argIdx++
	}
	if filters.ActorID != nil {
		where += fmt.Sprintf(" AND actor_id = $%d", argIdx)
		args = append(args, *filters.ActorID)
		argIdx++
	}
	if filters.ActorSearch != nil && *filters.ActorSearch != "" {
		pattern := "%" + *filters.ActorSearch + "%"
		where += fmt.Sprintf(" AND (LOWER(COALESCE(actor_name,'')) LIKE LOWER($%d) OR LOWER(COALESCE(actor_email,'')) LIKE LOWER($%d))", argIdx, argIdx+1)
		args = append(args, pattern, pattern)
		argIdx += 2
	}
	if filters.Action != nil {
		where += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, *filters.Action)
		argIdx++
	}
	if filters.Severity != nil {
		where += fmt.Sprintf(" AND severity = $%d", argIdx)
		args = append(args, *filters.Severity)
		argIdx++
	}
	if filters.FromDate != nil {
		where += fmt.Sprintf(" AND occurred_at >= $%d", argIdx)
		args = append(args, *filters.FromDate)
		argIdx++
	}
	if filters.ToDate != nil {
		where += fmt.Sprintf(" AND occurred_at <= $%d", argIdx)
		args = append(args, *filters.ToDate)
		argIdx++
	}

	countQuery := "SELECT COUNT(*) FROM audit.events " + where
	var total int64
	if err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit events: %w", err)
	}

	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(`
		SELECT audit_id, occurred_at, tenant_id, actor_type, actor_id,
			action, resource_type, resource_id, request_id, ip, user_agent,
			severity, before_json, after_json,
			COALESCE(actor_email, ''), COALESCE(actor_name, ''), COALESCE(resource_name, ''),
			COALESCE(endpoint, ''), changes
		FROM audit.events %s
		ORDER BY occurred_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := s.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer rows.Close()

	events, err := scanAuditDocuments(rows)
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// GetResourceHistory retrieves all audit events for a specific resource
func (s *PostgresService) GetResourceHistory(ctx context.Context, tenantID int64, resourceType string, resourceID int64) ([]AuditDocument, error) {
	if s.db == nil {
		return []AuditDocument{}, nil
	}

	query := `
		SELECT audit_id, occurred_at, tenant_id, actor_type, actor_id,
			action, resource_type, resource_id, request_id, ip, user_agent,
			severity, before_json, after_json,
			COALESCE(actor_email, ''), COALESCE(actor_name, ''), COALESCE(resource_name, ''),
			COALESCE(endpoint, ''), changes
		FROM audit.events
		WHERE tenant_id = $1 AND resource_type = $2 AND resource_id = $3
		ORDER BY occurred_at DESC
		LIMIT 100
	`

	rows, err := s.db.Query(ctx, query, tenantID, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find resource history: %w", err)
	}
	defer rows.Close()

	return scanAuditDocuments(rows)
}

// GetUserActivity retrieves all audit events for a specific actor
func (s *PostgresService) GetUserActivity(ctx context.Context, tenantID int64, actorID int64, from, to time.Time) ([]AuditDocument, error) {
	if s.db == nil {
		return []AuditDocument{}, nil
	}

	query := `
		SELECT audit_id, occurred_at, tenant_id, actor_type, actor_id,
			action, resource_type, resource_id, request_id, ip, user_agent,
			severity, before_json, after_json,
			COALESCE(actor_email, ''), COALESCE(actor_name, ''), COALESCE(resource_name, ''),
			COALESCE(endpoint, ''), changes
		FROM audit.events
		WHERE tenant_id = $1 AND actor_id = $2 AND occurred_at >= $3 AND occurred_at <= $4
		ORDER BY occurred_at DESC
		LIMIT 500
	`

	rows, err := s.db.Query(ctx, query, tenantID, actorID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to find user activity: %w", err)
	}
	defer rows.Close()

	return scanAuditDocuments(rows)
}

// GetByID retrieves a single audit event by audit_id
func (s *PostgresService) GetByID(ctx context.Context, tenantID int64, id string) (*AuditDocument, error) {
	if s.db == nil {
		return nil, nil
	}

	query := `
		SELECT audit_id, occurred_at, tenant_id, actor_type, actor_id,
			action, resource_type, resource_id, request_id, ip, user_agent,
			severity, before_json, after_json,
			COALESCE(actor_email, ''), COALESCE(actor_name, ''), COALESCE(resource_name, ''),
			COALESCE(endpoint, ''), changes
		FROM audit.events
		WHERE audit_id = $1 AND tenant_id = $2
	`

	var doc AuditDocument
	var beforeJSON, afterJSON, changesJSON []byte
	err := s.db.QueryRow(ctx, query, id, tenantID).Scan(
		&doc.AuditID, &doc.Timestamp, &doc.TenantID, &doc.ActorType, &doc.ActorID,
		&doc.Action, &doc.ResourceType, &doc.ResourceID, &doc.RequestID, &doc.IPAddress,
		&doc.UserAgent, &doc.Severity, &beforeJSON, &afterJSON,
		&doc.ActorEmail, &doc.ActorName, &doc.ResourceName, &doc.Endpoint, &changesJSON,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find audit event: %w", err)
	}
	if beforeJSON != nil {
		json.Unmarshal(beforeJSON, &doc.BeforeState)
	}
	if afterJSON != nil {
		json.Unmarshal(afterJSON, &doc.AfterState)
	}
	if changesJSON != nil {
		json.Unmarshal(changesJSON, &doc.Changes)
	}

	return &doc, nil
}

// LogAuditAccess records who accessed the audit logs
func (s *PostgresService) LogAuditAccess(ctx context.Context, accessedBy int64, tenantID int64, queryFilter AuditFilters, resultCount int, ipAddress string) error {
	if s.db == nil {
		return nil
	}

	filterJSON, _ := json.Marshal(queryFilter)

	query := `
		INSERT INTO audit.access_log (accessed_by, accessed_at, query_filter, result_count, ip)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.db.Exec(ctx, query, accessedBy, time.Now().UTC(), filterJSON, resultCount, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to log audit access: %w", err)
	}

	return nil
}

// DeleteByDateRange deletes all audit events for a tenant within [from, to] inclusive.
func (s *PostgresService) DeleteByDateRange(ctx context.Context, tenantID int64, from, to time.Time) (int64, error) {
	if s.db == nil {
		return 0, nil
	}
	tag, err := s.db.Exec(ctx,
		`DELETE FROM audit.events WHERE tenant_id = $1 AND occurred_at >= $2 AND occurred_at <= $3`,
		tenantID, from, to,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete audit events: %w", err)
	}
	return tag.RowsAffected(), nil
}

// AuditStats contains aggregated audit statistics for a tenant
type AuditStats struct {
	TotalEvents    int64            `json:"total_events"`
	BySeverity     map[string]int64 `json:"by_severity"`
	ByResource     map[string]int64 `json:"by_resource"`
	TopActors      []AuditActorStat `json:"top_actors"`
	RecentCritical []AuditDocument  `json:"recent_critical"`
	HoursAnalyzed  int              `json:"hours_analyzed"`
}

// AuditActorStat represents a single actor's event count
type AuditActorStat struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// GetStats returns aggregated audit statistics for a tenant within the last N hours
func (s *PostgresService) GetStats(ctx context.Context, tenantID int64, hours int) (*AuditStats, error) {
	if s.db == nil {
		return &AuditStats{
			BySeverity: map[string]int64{},
			ByResource: map[string]int64{},
			TopActors:  []AuditActorStat{},
		}, nil
	}
	since := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)

	stats := &AuditStats{
		HoursAnalyzed: hours,
		BySeverity:    make(map[string]int64),
		ByResource:    make(map[string]int64),
	}

	if err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM audit.events WHERE tenant_id = $1 AND occurred_at >= $2`,
		tenantID, since,
	).Scan(&stats.TotalEvents); err != nil {
		return nil, fmt.Errorf("audit stats total: %w", err)
	}

	sevRows, err := s.db.Query(ctx,
		`SELECT severity, COUNT(*) FROM audit.events WHERE tenant_id = $1 AND occurred_at >= $2 GROUP BY severity`,
		tenantID, since,
	)
	if err != nil {
		return nil, fmt.Errorf("audit stats by severity: %w", err)
	}
	defer sevRows.Close()
	for sevRows.Next() {
		var sev string
		var cnt int64
		if err := sevRows.Scan(&sev, &cnt); err != nil {
			return nil, fmt.Errorf("audit stats severity scan: %w", err)
		}
		stats.BySeverity[sev] = cnt
	}

	resRows, err := s.db.Query(ctx,
		`SELECT resource_type, COUNT(*) AS cnt FROM audit.events WHERE tenant_id = $1 AND occurred_at >= $2 GROUP BY resource_type ORDER BY cnt DESC LIMIT 10`,
		tenantID, since,
	)
	if err != nil {
		return nil, fmt.Errorf("audit stats by resource: %w", err)
	}
	defer resRows.Close()
	for resRows.Next() {
		var rt string
		var cnt int64
		if err := resRows.Scan(&rt, &cnt); err != nil {
			return nil, fmt.Errorf("audit stats resource scan: %w", err)
		}
		stats.ByResource[rt] = cnt
	}

	actorRows, err := s.db.Query(ctx,
		`SELECT COALESCE(actor_email, ''), COALESCE(actor_name, ''), COUNT(*) AS cnt
		 FROM audit.events
		 WHERE tenant_id = $1 AND occurred_at >= $2
		   AND actor_email IS NOT NULL AND actor_email != ''
		 GROUP BY actor_email, actor_name
		 ORDER BY cnt DESC
		 LIMIT 5`,
		tenantID, since,
	)
	if err != nil {
		return nil, fmt.Errorf("audit stats top actors: %w", err)
	}
	defer actorRows.Close()
	for actorRows.Next() {
		var a AuditActorStat
		if err := actorRows.Scan(&a.Email, &a.Name, &a.Count); err != nil {
			return nil, fmt.Errorf("audit stats actor scan: %w", err)
		}
		stats.TopActors = append(stats.TopActors, a)
	}
	if stats.TopActors == nil {
		stats.TopActors = []AuditActorStat{}
	}

	critRows, err := s.db.Query(ctx,
		`SELECT audit_id, occurred_at, tenant_id, actor_type, actor_id,
		    action, resource_type, resource_id, request_id, ip, user_agent,
		    severity, before_json, after_json,
		    COALESCE(actor_email, ''), COALESCE(actor_name, ''), COALESCE(resource_name, ''),
		    COALESCE(endpoint, ''), changes
		 FROM audit.events
		 WHERE tenant_id = $1 AND severity = 'CRITICAL'
		 ORDER BY occurred_at DESC
		 LIMIT 5`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("audit stats recent critical: %w", err)
	}
	defer critRows.Close()
	stats.RecentCritical, err = scanAuditDocuments(critRows)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// scanAuditDocuments is a helper to scan rows into AuditDocument slices
func scanAuditDocuments(rows interface {
	Next() bool
	Scan(...interface{}) error
}) ([]AuditDocument, error) {
	var events []AuditDocument
	for rows.Next() {
		var doc AuditDocument
		var beforeJSON, afterJSON, changesJSON []byte
		err := rows.Scan(
			&doc.AuditID, &doc.Timestamp, &doc.TenantID, &doc.ActorType, &doc.ActorID,
			&doc.Action, &doc.ResourceType, &doc.ResourceID, &doc.RequestID, &doc.IPAddress,
			&doc.UserAgent, &doc.Severity, &beforeJSON, &afterJSON,
			&doc.ActorEmail, &doc.ActorName, &doc.ResourceName, &doc.Endpoint, &changesJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit event: %w", err)
		}
		if beforeJSON != nil {
			json.Unmarshal(beforeJSON, &doc.BeforeState)
		}
		if afterJSON != nil {
			json.Unmarshal(afterJSON, &doc.AfterState)
		}
		if changesJSON != nil {
			json.Unmarshal(changesJSON, &doc.Changes)
		}
		events = append(events, doc)
	}
	if events == nil {
		events = []AuditDocument{}
	}
	return events, nil
}
