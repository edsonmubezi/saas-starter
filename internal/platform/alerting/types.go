package alerting

import (
	"time"

	"github.com/edsonmubezi/myapp/internal/platform/alerting/channels"
)

// Re-export channel types for external use
type ChannelType = channels.ChannelType
type Alert = channels.Alert
type Channel = channels.Channel
type ChannelResult = channels.ChannelResult

const (
	ChannelEmail   = channels.ChannelEmail
	ChannelSMS     = channels.ChannelSMS
	ChannelSlack   = channels.ChannelSlack
	ChannelTeams   = channels.ChannelTeams
	ChannelWebhook = channels.ChannelWebhook
)

// AlertConfig represents an alert configuration
type AlertConfig struct {
	ID              int64                  `json:"id" db:"id"`
	TenantID        int64                  `json:"tenant_id" db:"tenant_id"`
	Name            string                 `json:"name" db:"name"`
	Enabled         bool                   `json:"enabled" db:"enabled"`
	EventTypes      []string               `json:"event_types" db:"event_types"`
	Severities      []string               `json:"severities" db:"severities"`
	Conditions      map[string]interface{} `json:"conditions" db:"conditions"`
	CooldownSeconds int                    `json:"cooldown_seconds" db:"cooldown_seconds"`
	ThresholdCount  int                    `json:"threshold_count" db:"threshold_count"`
	WindowSeconds   int                    `json:"window_seconds" db:"window_seconds"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// AlertChannel represents a channel configuration
type AlertChannel struct {
	ID            int64             `json:"id" db:"id"`
	AlertConfigID int64             `json:"alert_config_id" db:"alert_config_id"`
	ChannelType   ChannelType       `json:"channel_type" db:"channel_type"`
	Config        map[string]string `json:"config" db:"config"`
	Enabled       bool              `json:"enabled" db:"enabled"`
	CreatedAt     time.Time         `json:"created_at" db:"created_at"`
}

// AlertHistory records sent alerts
type AlertHistory struct {
	ID              int64             `json:"id" db:"id"`
	AlertConfigID   int64             `json:"alert_config_id" db:"alert_config_id"`
	TenantID        int64             `json:"tenant_id" db:"tenant_id"`
	EventID         string            `json:"event_id" db:"event_id"`
	EventType       string            `json:"event_type" db:"event_type"`
	Severity        string            `json:"severity" db:"severity"`
	ChannelsNotified []string         `json:"channels_notified" db:"channels_notified"`
	DeliveryStatus  map[string]string `json:"delivery_status" db:"delivery_status"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
}


// CreateAlertConfigInput for creating alert configs
type CreateAlertConfigInput struct {
	Name            string                 `json:"name" validate:"required,min=1,max=100"`
	Enabled         bool                   `json:"enabled"`
	EventTypes      []string               `json:"event_types"`
	Severities      []string               `json:"severities" validate:"required"`
	Conditions      map[string]interface{} `json:"conditions"`
	CooldownSeconds int                    `json:"cooldown_seconds"`
	ThresholdCount  int                    `json:"threshold_count"`
	WindowSeconds   int                    `json:"window_seconds"`
	Channels        []CreateChannelInput   `json:"channels"`
}

// CreateChannelInput for creating channels
type CreateChannelInput struct {
	ChannelType ChannelType       `json:"channel_type" validate:"required"`
	Config      map[string]string `json:"config" validate:"required"`
	Enabled     bool              `json:"enabled"`
}

// UpdateAlertConfigInput for updating alert configs
type UpdateAlertConfigInput struct {
	Name            *string                 `json:"name"`
	Enabled         *bool                   `json:"enabled"`
	EventTypes      []string                `json:"event_types"`
	Severities      []string                `json:"severities"`
	Conditions      map[string]interface{}  `json:"conditions"`
	CooldownSeconds *int                    `json:"cooldown_seconds"`
	ThresholdCount  *int                    `json:"threshold_count"`
	WindowSeconds   *int                    `json:"window_seconds"`
}
