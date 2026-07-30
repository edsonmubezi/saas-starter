package channels

import "time"

// ChannelType represents the type of alert channel
type ChannelType string

const (
	ChannelEmail   ChannelType = "email"
	ChannelSMS     ChannelType = "sms"
	ChannelSlack   ChannelType = "slack"
	ChannelTeams   ChannelType = "teams"
	ChannelWebhook ChannelType = "webhook"
)

// Alert represents an alert to be sent
type Alert struct {
	ID           string                 `json:"id"`
	Title        string                 `json:"title"`
	Message      string                 `json:"message"`
	Severity     string                 `json:"severity"`
	EventType    string                 `json:"event_type"`
	TenantID     int64                  `json:"tenant_id"`
	Timestamp    time.Time              `json:"timestamp"`
	Details      map[string]interface{} `json:"details"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	ActorEmail   string                 `json:"actor_email,omitempty"`
	ResourceID   *int64                 `json:"resource_id,omitempty"`
	ResourceType string                 `json:"resource_type,omitempty"`
}

// Channel is the interface for alert channels
type Channel interface {
	// Send sends an alert through the channel
	Send(alert Alert) error

	// Name returns the channel type name
	Name() ChannelType

	// IsConfigured returns whether the channel is properly configured
	IsConfigured() bool
}

// ChannelResult represents the result of sending through a channel
type ChannelResult struct {
	Channel ChannelType `json:"channel"`
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
}
