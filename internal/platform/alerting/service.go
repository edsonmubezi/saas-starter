package alerting

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/edsonmubezi/myapp/internal/config"
	"github.com/edsonmubezi/myapp/internal/platform/alerting/channels"
)

// Service handles alert dispatching
type Service struct {
	channels     map[ChannelType]Channel
	enabled      bool
	cooldowns    map[string]time.Time // Track cooldowns by config ID + event type
	cooldownLock sync.RWMutex
}

// NewService creates a new alerting service
func NewService(cfg config.AlertingConfig, emailCfg config.EmailConfig) *Service {
	svc := &Service{
		channels:  make(map[ChannelType]Channel),
		enabled:   cfg.Enabled,
		cooldowns: make(map[string]time.Time),
	}

	if !cfg.Enabled {
		log.Println("Alerting service is disabled")
		return svc
	}

	// Initialize email channel
	if emailCfg.Enabled {
		emailChannel := channels.NewEmailChannel(channels.EmailConfig{
			SMTPHost:     emailCfg.SMTPHost,
			SMTPPort:     emailCfg.SMTPPort,
			SMTPUser:     emailCfg.SMTPUser,
			SMTPPassword: emailCfg.SMTPPassword,
			FromAddress:  emailCfg.FromAddress,
			FromName:     emailCfg.FromName,
		})
		svc.channels[ChannelEmail] = emailChannel
		log.Println("✓ Email alert channel initialized")
	}

	// Initialize Slack channel
	if cfg.SlackWebhookURL != "" {
		slackChannel := channels.NewSlackChannel(cfg.SlackWebhookURL)
		svc.channels[ChannelSlack] = slackChannel
		log.Println("✓ Slack alert channel initialized")
	}

	// Initialize SMS channel (if configured)
	if cfg.SMSProvider != "" && cfg.SMSAPIKey != "" {
		smsChannel := channels.NewSMSChannel(channels.SMSConfig{
			Provider: cfg.SMSProvider,
			APIKey:   cfg.SMSAPIKey,
			Username: cfg.SMSUsername,
			SenderID: cfg.SMSSenderID,
		})
		svc.channels[ChannelSMS] = smsChannel
		log.Println("✓ SMS alert channel initialized")
	}

	// Initialize Microsoft Teams channel
	if cfg.TeamsWebhookURL != "" {
		teamsChannel := channels.NewTeamsChannel(cfg.TeamsWebhookURL)
		svc.channels[ChannelTeams] = teamsChannel
		log.Println("✓ Microsoft Teams alert channel initialized")
	}

	// Initialize generic Webhook channel
	if cfg.WebhookURL != "" {
		webhookChannel := channels.NewWebhookChannel(channels.WebhookConfig{
			URL:    cfg.WebhookURL,
			Secret: cfg.WebhookSecret,
		})
		svc.channels[ChannelWebhook] = webhookChannel
		log.Println("✓ Webhook alert channel initialized")
	}

	return svc
}

// Send dispatches an alert through specified channels
func (s *Service) Send(ctx context.Context, alert Alert, channelTypes []ChannelType) []ChannelResult {
	if !s.enabled {
		return nil
	}

	results := make([]ChannelResult, 0, len(channelTypes))

	for _, ct := range channelTypes {
		channel, exists := s.channels[ct]
		if !exists {
			results = append(results, ChannelResult{
				Channel: ct,
				Success: false,
				Error:   "channel not configured",
			})
			continue
		}

		if !channel.IsConfigured() {
			results = append(results, ChannelResult{
				Channel: ct,
				Success: false,
				Error:   "channel not properly configured",
			})
			continue
		}

		err := channel.Send(alert)
		if err != nil {
			results = append(results, ChannelResult{
				Channel: ct,
				Success: false,
				Error:   err.Error(),
			})
			log.Printf("Failed to send alert via %s: %v", ct, err)
		} else {
			results = append(results, ChannelResult{
				Channel: ct,
				Success: true,
			})
		}
	}

	return results
}

// SendAsync dispatches an alert asynchronously
func (s *Service) SendAsync(alert Alert, channelTypes []ChannelType) {
	if !s.enabled {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		s.Send(ctx, alert, channelTypes)
	}()
}

// CheckCooldown checks if an alert is within cooldown period
func (s *Service) CheckCooldown(configID int64, eventType string, cooldownSeconds int) bool {
	if cooldownSeconds <= 0 {
		return false // No cooldown
	}

	key := fmt.Sprintf("%d:%s", configID, eventType)

	s.cooldownLock.RLock()
	lastSent, exists := s.cooldowns[key]
	s.cooldownLock.RUnlock()

	if !exists {
		return false // No previous alert
	}

	return time.Since(lastSent) < time.Duration(cooldownSeconds)*time.Second
}

// RecordAlertSent records that an alert was sent (for cooldown tracking)
func (s *Service) RecordAlertSent(configID int64, eventType string) {
	key := fmt.Sprintf("%d:%s", configID, eventType)

	s.cooldownLock.Lock()
	s.cooldowns[key] = time.Now()
	s.cooldownLock.Unlock()
}

// GetChannel returns a specific channel
func (s *Service) GetChannel(channelType ChannelType) (Channel, bool) {
	channel, exists := s.channels[channelType]
	return channel, exists
}

// IsEnabled returns whether alerting is enabled
func (s *Service) IsEnabled() bool {
	return s.enabled
}

// GetConfiguredChannels returns list of configured channel types
func (s *Service) GetConfiguredChannels() []ChannelType {
	var types []ChannelType
	for t, ch := range s.channels {
		if ch.IsConfigured() {
			types = append(types, t)
		}
	}
	return types
}

// CreateSecurityAlert creates a standardized security alert
func CreateSecurityAlert(eventType, severity string, tenantID int64, details map[string]interface{}) Alert {
	title := fmt.Sprintf("Security Alert: %s", eventType)
	message := formatSecurityMessage(eventType, severity, details)

	return Alert{
		ID:        fmt.Sprintf("sec-%d", time.Now().UnixNano()),
		Title:     title,
		Message:   message,
		Severity:  severity,
		EventType: eventType,
		TenantID:  tenantID,
		Timestamp: time.Now().UTC(),
		Details:   details,
	}
}

func formatSecurityMessage(eventType, severity string, details map[string]interface{}) string {
	message := fmt.Sprintf("A %s security event has been detected.\n\n", severity)

	switch eventType {
	case "LOGIN_FAILED":
		if email, ok := details["attempted_email"].(string); ok {
			message += fmt.Sprintf("Failed login attempt for: %s\n", email)
		}
		if reason, ok := details["reason"].(string); ok {
			message += fmt.Sprintf("Reason: %s\n", reason)
		}
		if count, ok := details["attempt_count"].(int); ok {
			message += fmt.Sprintf("Failed attempts: %d\n", count)
		}

	case "ACCOUNT_LOCKED":
		if email, ok := details["email"].(string); ok {
			message += fmt.Sprintf("Account locked: %s\n", email)
		}
		if reason, ok := details["reason"].(string); ok {
			message += fmt.Sprintf("Reason: %s\n", reason)
		}

	case "PERMISSION_DENIED":
		if resource, ok := details["resource"].(string); ok {
			message += fmt.Sprintf("Resource: %s\n", resource)
		}
		if action, ok := details["action"].(string); ok {
			message += fmt.Sprintf("Action: %s\n", action)
		}

	case "CROSS_TENANT_ATTEMPT":
		if attemptedTenant, ok := details["attempted_tenant_id"].(int64); ok {
			message += fmt.Sprintf("Attempted access to tenant ID: %d\n", attemptedTenant)
		}

	case "BULK_EXPORT":
		if dataType, ok := details["data_type"].(string); ok {
			message += fmt.Sprintf("Data type: %s\n", dataType)
		}
		if count, ok := details["record_count"].(int); ok {
			message += fmt.Sprintf("Records exported: %d\n", count)
		}

	default:
		// Generic details formatting
		for k, v := range details {
			message += fmt.Sprintf("%s: %v\n", k, v)
		}
	}

	if ip, ok := details["ip_address"].(string); ok {
		message += fmt.Sprintf("\nIP Address: %s\n", ip)
	}

	message += "\nPlease review this event in the security dashboard."

	return message
}

// CreateAuditAlert creates a standardized audit alert
func CreateAuditAlert(action, resourceType string, resourceID *int64, tenantID int64, actorName string, changes []map[string]interface{}) Alert {
	title := fmt.Sprintf("Audit Alert: %s on %s", action, resourceType)

	message := fmt.Sprintf("%s performed %s on %s", actorName, action, resourceType)
	if resourceID != nil {
		message += fmt.Sprintf(" (ID: %d)", *resourceID)
	}
	message += ".\n\n"

	if len(changes) > 0 {
		message += "Changes:\n"
		for _, change := range changes {
			if field, ok := change["field"].(string); ok {
				message += fmt.Sprintf("- %s: %v → %v\n", field, change["old"], change["new"])
			}
		}
	}

	return Alert{
		ID:           fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		Title:        title,
		Message:      message,
		Severity:     "MEDIUM",
		EventType:    action,
		TenantID:     tenantID,
		Timestamp:    time.Now().UTC(),
		ResourceID:   resourceID,
		ResourceType: resourceType,
		Details: map[string]interface{}{
			"actor_name": actorName,
			"changes":    changes,
		},
	}
}
