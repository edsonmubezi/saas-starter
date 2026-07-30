package channels

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookConfig holds webhook channel configuration
type WebhookConfig struct {
	URL           string
	Secret        string            // For HMAC signature
	Headers       map[string]string // Custom headers
	Method        string            // POST or PUT
	RetryAttempts int
	RetryDelay    time.Duration
}

// WebhookChannel sends alerts via HTTP webhook
type WebhookChannel struct {
	config     WebhookConfig
	configured bool
	httpClient *http.Client
}

// NewWebhookChannel creates a new webhook channel
func NewWebhookChannel(cfg WebhookConfig) *WebhookChannel {
	if cfg.Method == "" {
		cfg.Method = "POST"
	}
	if cfg.RetryAttempts == 0 {
		cfg.RetryAttempts = 3
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = time.Second
	}

	return &WebhookChannel{
		config:     cfg,
		configured: cfg.URL != "",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Send sends an alert via webhook
func (c *WebhookChannel) Send(alert Alert) error {
	if !c.configured {
		return fmt.Errorf("webhook channel not configured")
	}

	payload := c.buildPayload(alert)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < c.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(c.config.RetryDelay * time.Duration(attempt))
		}

		err := c.sendRequest(jsonPayload)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("webhook failed after %d attempts: %w", c.config.RetryAttempts, lastErr)
}

// Name returns the channel type
func (c *WebhookChannel) Name() ChannelType {
	return ChannelWebhook
}

// IsConfigured returns whether the channel is properly configured
func (c *WebhookChannel) IsConfigured() bool {
	return c.configured
}

// WebhookPayload represents the webhook request body
type WebhookPayload struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Title        string                 `json:"title"`
	Message      string                 `json:"message"`
	Severity     string                 `json:"severity"`
	EventType    string                 `json:"event_type"`
	TenantID     int64                  `json:"tenant_id"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	ActorEmail   string                 `json:"actor_email,omitempty"`
	ResourceID   *int64                 `json:"resource_id,omitempty"`
	ResourceType string                 `json:"resource_type,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Source       string                 `json:"source"`
}

func (c *WebhookChannel) buildPayload(alert Alert) WebhookPayload {
	return WebhookPayload{
		ID:           alert.ID,
		Timestamp:    alert.Timestamp,
		Title:        alert.Title,
		Message:      alert.Message,
		Severity:     alert.Severity,
		EventType:    alert.EventType,
		TenantID:     alert.TenantID,
		IPAddress:    alert.IPAddress,
		ActorEmail:   alert.ActorEmail,
		ResourceID:   alert.ResourceID,
		ResourceType: alert.ResourceType,
		Details:      alert.Details,
		Source:       "SaaS Platform",
	}
}

func (c *WebhookChannel) sendRequest(payload []byte) error {
	req, err := http.NewRequest(c.config.Method, c.config.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "SaaS-Platform-Alerting/1.0")

	// Add custom headers
	for k, v := range c.config.Headers {
		req.Header.Set(k, v)
	}

	// Add HMAC signature if secret is configured
	if c.config.Secret != "" {
		signature := c.computeSignature(payload)
		req.Header.Set("X-Signature-256", "sha256="+signature)
		req.Header.Set("X-Hub-Signature-256", "sha256="+signature) // GitHub-style
	}

	req.Header.Set("X-Request-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status: %d", resp.StatusCode)
	}

	return nil
}

func (c *WebhookChannel) computeSignature(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(c.config.Secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
