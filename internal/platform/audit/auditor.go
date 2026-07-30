package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Auditor handles recording audit events
type Auditor struct {
	db *pgxpool.Pool
}

// NewAuditor creates a new Auditor instance
func NewAuditor(db *pgxpool.Pool) *Auditor {
	return &Auditor{db: db}
}

// Record writes an audit event to the outbox table within the same transaction
// This ensures atomicity: either both the business operation and audit event succeed, or both fail
func (a *Auditor) Record(ctx context.Context, tx pgx.Tx, event *AuditEvent) error {
	// Set defaults
	if event.AuditID == "" {
		event.AuditID = uuid.New().String()
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}
	if event.OriginService == "" {
		event.OriginService = "saas-api"
	}
	if event.SchemaVersion == 0 {
		event.SchemaVersion = 1
	}
	if event.Severity == "" {
		event.Severity = SeverityLow
	}

	// Serialize to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// Insert into outbox (same transaction as business logic)
	query := `
		INSERT INTO audit.outbox (event_json, created_at)
		VALUES ($1, $2)
	`

	_, err = tx.Exec(ctx, query, eventJSON, event.OccurredAt)
	if err != nil {
		return fmt.Errorf("failed to insert audit event to outbox: %w", err)
	}

	return nil
}

// RecordWithoutTx records an audit event without a transaction
// Use sparingly - prefer Record() with a transaction for atomicity
func (a *Auditor) RecordWithoutTx(ctx context.Context, event *AuditEvent) error {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := a.Record(ctx, tx, event); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Builder pattern for convenient event creation
type EventBuilder struct {
	event *AuditEvent
}

// NewEvent creates a new audit event builder
func (a *Auditor) NewEvent() *EventBuilder {
	return &EventBuilder{
		event: &AuditEvent{
			AuditID:       uuid.New().String(),
			OccurredAt:    time.Now().UTC(),
			ActorType:     ActorUser,
			OriginService: "saas-api",
			Severity:      SeverityLow,
			SchemaVersion: 1,
		},
	}
}

func (b *EventBuilder) Actor(actorType ActorType, actorID *int64) *EventBuilder {
	b.event.ActorType = actorType
	b.event.ActorID = actorID
	return b
}

func (b *EventBuilder) Tenant(tenantID int64) *EventBuilder {
	b.event.TenantID = tenantID
	return b
}

func (b *EventBuilder) Action(action string) *EventBuilder {
	b.event.Action = action
	return b
}

func (b *EventBuilder) Resource(resourceType string, resourceID *int64) *EventBuilder {
	b.event.ResourceType = resourceType
	b.event.ResourceID = resourceID
	return b
}

func (b *EventBuilder) Request(requestID, ip, userAgent string) *EventBuilder {
	b.event.RequestID = requestID
	b.event.IP = ip
	b.event.UserAgent = userAgent
	return b
}

func (b *EventBuilder) Severity(severity Severity) *EventBuilder {
	b.event.Severity = severity
	return b
}

func (b *EventBuilder) Before(data map[string]interface{}) *EventBuilder {
	b.event.BeforeJSON = data
	return b
}

func (b *EventBuilder) After(data map[string]interface{}) *EventBuilder {
	b.event.AfterJSON = data
	return b
}

func (b *EventBuilder) Build() *AuditEvent {
	return b.event
}
