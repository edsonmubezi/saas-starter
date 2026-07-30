package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/edsonmubezi/myapp/internal/identity"
)

// Recorder provides a convenient way to record audit events from usecases
type Recorder struct {
	service      *PostgresService
	resourceType string
}

// NewRecorder creates a new audit recorder for a specific resource type
func NewRecorder(service *PostgresService, resourceType string) *Recorder {
	return &Recorder{
		service:      service,
		resourceType: resourceType,
	}
}

// Record records an audit event asynchronously
func (r *Recorder) Record(ctx context.Context, action string, resourceID int64, resourceName string, before, after interface{}) {
	if r.service == nil {
		return // Audit service not configured
	}

	p, err := identity.Require(ctx)
	if err != nil {
		return // Cannot audit without identity
	}

	var beforeMap, afterMap map[string]interface{}
	if before != nil {
		beforeMap, _ = StructToMap(before)
	}
	if after != nil {
		afterMap, _ = StructToMap(after)
	}

	doc := &AuditDocument{
		AuditID:      fmt.Sprintf("%s-%d-%d", r.resourceType, resourceID, time.Now().UnixNano()),
		Timestamp:    time.Now().UTC(),
		TenantID:     p.OrganizationID,
		ActorType:    ActorUser,
		ActorID:      &p.UserID,
		ActorEmail:   p.Email,
		ActorName:    identity.FullName(ctx),
		Action:       action,
		ResourceType: r.resourceType,
		ResourceID:   &resourceID,
		ResourceName: resourceName,
		Severity:     DetermineSeverity(action),
		RequestID:    identity.RequestID(ctx),
		IPAddress:    identity.IPAddress(ctx),
		UserAgent:    identity.UserAgent(ctx),
		BeforeState:  beforeMap,
		AfterState:   afterMap,
	}

	r.service.RecordAsync(doc)
}

// RecordWithSeverity records an audit event with explicit severity
func (r *Recorder) RecordWithSeverity(ctx context.Context, action string, resourceID int64, resourceName string, severity Severity, before, after interface{}) {
	if r.service == nil {
		return
	}

	p, err := identity.Require(ctx)
	if err != nil {
		return
	}

	var beforeMap, afterMap map[string]interface{}
	if before != nil {
		beforeMap, _ = StructToMap(before)
	}
	if after != nil {
		afterMap, _ = StructToMap(after)
	}

	doc := &AuditDocument{
		AuditID:      fmt.Sprintf("%s-%d-%d", r.resourceType, resourceID, time.Now().UnixNano()),
		Timestamp:    time.Now().UTC(),
		TenantID:     p.OrganizationID,
		ActorType:    ActorUser,
		ActorID:      &p.UserID,
		ActorEmail:   p.Email,
		ActorName:    identity.FullName(ctx),
		Action:       action,
		ResourceType: r.resourceType,
		ResourceID:   &resourceID,
		ResourceName: resourceName,
		Severity:     severity,
		RequestID:    identity.RequestID(ctx),
		IPAddress:    identity.IPAddress(ctx),
		UserAgent:    identity.UserAgent(ctx),
		BeforeState:  beforeMap,
		AfterState:   afterMap,
	}

	r.service.RecordAsync(doc)
}

// RecordSystem records an audit event from a system/service context (no user)
func (r *Recorder) RecordSystem(ctx context.Context, action string, resourceID int64, resourceName string, tenantID int64, before, after interface{}) {
	if r.service == nil {
		return
	}

	var beforeMap, afterMap map[string]interface{}
	if before != nil {
		beforeMap, _ = StructToMap(before)
	}
	if after != nil {
		afterMap, _ = StructToMap(after)
	}

	doc := &AuditDocument{
		AuditID:      fmt.Sprintf("%s-%d-%d", r.resourceType, resourceID, time.Now().UnixNano()),
		Timestamp:    time.Now().UTC(),
		TenantID:     tenantID,
		ActorType:    ActorSystem,
		Action:       action,
		ResourceType: r.resourceType,
		ResourceID:   &resourceID,
		ResourceName: resourceName,
		Severity:     DetermineSeverity(action),
		BeforeState:  beforeMap,
		AfterState:   afterMap,
	}

	r.service.RecordAsync(doc)
}
