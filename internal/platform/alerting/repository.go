package alerting

import (
	"context"
	"fmt"
	"time"
)

// AlertConfigDocument represents an alert configuration
type AlertConfigDocument struct {
	ID              string                 `json:"id"`
	TenantID        int64                  `json:"tenant_id"`
	Name            string                 `json:"name"`
	Enabled         bool                   `json:"enabled"`
	EventTypes      []string               `json:"event_types"`
	Severities      []string               `json:"severities"`
	Conditions      map[string]interface{} `json:"conditions,omitempty"`
	CooldownSeconds int                    `json:"cooldown_seconds"`
	ThresholdCount  int                    `json:"threshold_count"`
	WindowSeconds   int                    `json:"window_seconds"`
	Channels        []AlertChannelConfig   `json:"channels"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	CreatedBy       int64                  `json:"created_by"`
}

// AlertChannelConfig represents a channel configuration within an alert
type AlertChannelConfig struct {
	ChannelType ChannelType       `json:"channel_type"`
	Config      map[string]string `json:"config"`
	Enabled     bool              `json:"enabled"`
}

// AlertHistoryDocument represents a sent alert record
type AlertHistoryDocument struct {
	ID               string          `json:"id"`
	AlertConfigID    string          `json:"alert_config_id"`
	TenantID         int64           `json:"tenant_id"`
	EventID          string          `json:"event_id"`
	EventType        string          `json:"event_type"`
	Severity         string          `json:"severity"`
	AlertTitle       string          `json:"alert_title"`
	AlertMessage     string          `json:"alert_message"`
	ChannelsNotified []string        `json:"channels_notified"`
	DeliveryResults  []ChannelResult `json:"delivery_results"`
	CreatedAt        time.Time       `json:"created_at"`
}

// Repository handles alert configuration persistence.
// This is a no-op stub — alert configs are not currently persisted.
// A future migration can wire this to a PostgreSQL table.
type Repository struct{}

// NewRepository creates a new alerting repository
func NewRepository() *Repository {
	return &Repository{}
}

// CreateConfig creates a new alert configuration (no-op)
func (r *Repository) CreateConfig(ctx context.Context, tenantID, userID int64, input *CreateAlertConfigInput) (*AlertConfigDocument, error) {
	return nil, fmt.Errorf("alert config storage not available")
}

// GetConfigByID retrieves an alert configuration by ID (no-op)
func (r *Repository) GetConfigByID(ctx context.Context, id string, tenantID int64) (*AlertConfigDocument, error) {
	return nil, nil
}

// GetConfigsByTenant retrieves all alert configurations for a tenant (no-op)
func (r *Repository) GetConfigsByTenant(ctx context.Context, tenantID int64, page, pageSize int) ([]AlertConfigDocument, int64, error) {
	return []AlertConfigDocument{}, 0, nil
}

// UpdateConfig updates an alert configuration (no-op)
func (r *Repository) UpdateConfig(ctx context.Context, id string, tenantID int64, input *UpdateAlertConfigInput) (*AlertConfigDocument, error) {
	return nil, fmt.Errorf("alert config storage not available")
}

// DeleteConfig deletes an alert configuration (no-op)
func (r *Repository) DeleteConfig(ctx context.Context, id string, tenantID int64) error {
	return fmt.Errorf("alert config storage not available")
}

// GetEnabledConfigs retrieves all enabled configurations for a tenant (no-op)
func (r *Repository) GetEnabledConfigs(ctx context.Context, tenantID int64) ([]AlertConfigDocument, error) {
	return []AlertConfigDocument{}, nil
}

// RecordAlertHistory records that an alert was sent (no-op)
func (r *Repository) RecordAlertHistory(ctx context.Context, configID string, tenantID int64, alert Alert, results []ChannelResult) error {
	return nil
}

// GetAlertHistory retrieves alert history for a tenant (no-op)
func (r *Repository) GetAlertHistory(ctx context.Context, tenantID int64, page, pageSize int) ([]AlertHistoryDocument, int64, error) {
	return []AlertHistoryDocument{}, 0, nil
}
