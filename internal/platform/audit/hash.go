package audit

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
)

// ComputeSignature creates a canonical JSON representation and computes SHA-256
func ComputeSignature(event *AuditEvent) ([]byte, error) {
	// Create canonical representation (deterministic ordering)
	canonical := make(map[string]interface{})

	// Required fields
	canonical["audit_id"] = event.AuditID
	canonical["occurred_at"] = event.OccurredAt.UTC().Format("2006-01-02T15:04:05.999999Z")
	canonical["actor_type"] = string(event.ActorType)
	canonical["tenant_id"] = event.TenantID
	canonical["action"] = event.Action
	canonical["resource_type"] = event.ResourceType
	canonical["origin_service"] = event.OriginService
	canonical["severity"] = string(event.Severity)
	canonical["schema_version"] = event.SchemaVersion

	// Optional fields (only if present)
	if event.ActorID != nil {
		canonical["actor_id"] = *event.ActorID
	}
	if event.ResourceID != nil {
		canonical["resource_id"] = *event.ResourceID
	}
	if event.RequestID != "" {
		canonical["request_id"] = event.RequestID
	}
	if event.IP != "" {
		canonical["ip"] = event.IP
	}
	if event.BeforeJSON != nil {
		canonical["before_json"] = sortedMapKeys(event.BeforeJSON)
	}
	if event.AfterJSON != nil {
		canonical["after_json"] = sortedMapKeys(event.AfterJSON)
	}

	// Marshal with sorted keys (Go's json.Marshal sorts map keys by default)
	jsonBytes, err := json.Marshal(canonical)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal canonical JSON: %w", err)
	}

	// Compute SHA-256
	hash := sha256.Sum256(jsonBytes)
	return hash[:], nil
}

// sortedMapKeys ensures deterministic JSON representation by sorting keys
func sortedMapKeys(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}

	sorted := make(map[string]interface{})
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		value := data[k]
		// Recursively sort nested maps
		if nestedMap, ok := value.(map[string]interface{}); ok {
			sorted[k] = sortedMapKeys(nestedMap)
		} else {
			sorted[k] = value
		}
	}

	return sorted
}

// VerifySignature verifies that an audit event's signature matches its content
func VerifySignature(event *AuditEvent) (bool, error) {
	// Store original signature
	originalSig := event.SigSHA256

	// Compute expected signature
	computedSig, err := ComputeSignature(event)
	if err != nil {
		return false, err
	}

	// Compare byte by byte
	if len(originalSig) != len(computedSig) {
		return false, nil
	}

	for i := range originalSig {
		if originalSig[i] != computedSig[i] {
			return false, nil
		}
	}

	return true, nil
}
