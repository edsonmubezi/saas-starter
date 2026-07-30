package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"time"
)

// ComputeChanges compares before and after states and returns the list of changes
func ComputeChanges(before, after map[string]interface{}) []FieldChange {
	if before == nil && after == nil {
		return nil
	}

	var changes []FieldChange

	allKeys := make(map[string]bool)
	for k := range before {
		allKeys[k] = true
	}
	for k := range after {
		allKeys[k] = true
	}

	sortedKeys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		beforeVal, beforeExists := before[key]
		afterVal, afterExists := after[key]

		if beforeExists && afterExists && reflect.DeepEqual(beforeVal, afterVal) {
			continue
		}

		if !beforeExists && afterExists {
			changes = append(changes, FieldChange{Field: key, OldValue: nil, NewValue: afterVal})
			continue
		}
		if beforeExists && !afterExists {
			changes = append(changes, FieldChange{Field: key, OldValue: beforeVal, NewValue: nil})
			continue
		}
		changes = append(changes, FieldChange{Field: key, OldValue: beforeVal, NewValue: afterVal})
	}

	return changes
}

// ComputeDocSignature computes a SHA-256 signature for an AuditDocument
func ComputeDocSignature(doc *AuditDocument) string {
	canonical := map[string]interface{}{
		"audit_id":      doc.AuditID,
		"timestamp":     doc.Timestamp.UTC().Format(time.RFC3339Nano),
		"tenant_id":     doc.TenantID,
		"actor_type":    doc.ActorType,
		"action":        doc.Action,
		"resource_type": doc.ResourceType,
		"severity":      doc.Severity,
	}

	if doc.ActorID != nil {
		canonical["actor_id"] = *doc.ActorID
	}
	if doc.ResourceID != nil {
		canonical["resource_id"] = *doc.ResourceID
	}

	jsonBytes, err := json.Marshal(canonical)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(jsonBytes)
	return "sha256:" + hex.EncodeToString(hash[:])
}

// StructToMap converts any struct to map[string]interface{} for before/after JSON snapshots
func StructToMap(v interface{}) (map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// DetermineSeverity suggests a severity level based on the action
func DetermineSeverity(action string) Severity {
	criticalActions := map[string]bool{
		ActionEmployeeTerminated: true,
		ActionRoleGranted:        true,
		ActionRoleRevoked:        true,
		ActionPermGranted:        true,
		ActionPermRevoked:        true,
	}

	highActions := map[string]bool{
		ActionEmployeeCreated: true,
		ActionEmployeeDeleted: true,
		ActionLoanApproved:    true,
		ActionLoanRejected:    true,
		ActionPayrollApproved: true,
		ActionPasswordReset:   true,
		ActionMFAEnabled:      true,
		ActionMFADisabled:     true,
	}

	mediumActions := map[string]bool{
		ActionEmployeeUpdated:  true,
		ActionLoanCreated:      true,
		ActionPayrollGenerated: true,
	}

	if criticalActions[action] {
		return SeverityCritical
	}
	if highActions[action] {
		return SeverityHigh
	}
	if mediumActions[action] {
		return SeverityMedium
	}

	return SeverityLow
}

// GetResourceName extracts a display name from the state data
func GetResourceName(state map[string]interface{}) string {
	if state == nil {
		return ""
	}

	nameFields := []string{"full_name", "name", "title", "email", "email_address", "description"}

	for _, field := range nameFields {
		if val, ok := state[field]; ok {
			if str, ok := val.(string); ok && str != "" {
				return str
			}
		}
	}

	return ""
}

// sensitiveFieldList contains fields that should be redacted from audit logs
var sensitiveFieldList = map[string]bool{
	"password":          true,
	"password_hash":     true,
	"secret":            true,
	"token":             true,
	"api_key":           true,
	"totp_secret":       true,
	"two_factor_secret": true,
}

// redactSensitive replaces sensitive field values with "[REDACTED]"
func redactSensitive(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range data {
		lowerKey := strings.ToLower(k)

		if sensitiveFieldList[lowerKey] {
			result[k] = "[REDACTED]"
			continue
		}

		if strings.Contains(lowerKey, "password") ||
			strings.Contains(lowerKey, "secret") ||
			strings.Contains(lowerKey, "token") ||
			strings.Contains(lowerKey, "key") {
			result[k] = "[REDACTED]"
			continue
		}

		if nestedMap, ok := v.(map[string]interface{}); ok {
			result[k] = redactSensitive(nestedMap)
		} else {
			result[k] = v
		}
	}

	return result
}
