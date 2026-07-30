package channels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SMSConfig holds SMS channel configuration
type SMSConfig struct {
	Provider string // "africastalking", "twilio", "nexmo"
	APIKey   string
	Username string // For Africa's Talking
	SenderID string
	// Twilio specific
	AccountSID string
	AuthToken  string
}

// SMSChannel sends alerts via SMS
type SMSChannel struct {
	config     SMSConfig
	configured bool
	httpClient *http.Client
}

// NewSMSChannel creates a new SMS channel
func NewSMSChannel(cfg SMSConfig) *SMSChannel {
	configured := false
	switch cfg.Provider {
	case "africastalking":
		configured = cfg.APIKey != "" && cfg.Username != ""
	case "twilio":
		configured = cfg.AccountSID != "" && cfg.AuthToken != ""
	case "nexmo":
		configured = cfg.APIKey != ""
	}

	return &SMSChannel{
		config:     cfg,
		configured: configured,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Send sends an alert via SMS
func (c *SMSChannel) Send(alert Alert) error {
	if !c.configured {
		return fmt.Errorf("sms channel not configured")
	}

	recipients := c.getRecipients(alert)
	if len(recipients) == 0 {
		return fmt.Errorf("no phone numbers specified for SMS alert")
	}

	message := c.buildMessage(alert)

	var err error
	switch c.config.Provider {
	case "africastalking":
		err = c.sendViaAfricasTalking(recipients, message)
	case "twilio":
		err = c.sendViaTwilio(recipients, message)
	case "nexmo":
		err = c.sendViaNexmo(recipients, message)
	default:
		err = fmt.Errorf("unknown SMS provider: %s", c.config.Provider)
	}

	return err
}

// Name returns the channel type
func (c *SMSChannel) Name() ChannelType {
	return ChannelSMS
}

// IsConfigured returns whether the channel is properly configured
func (c *SMSChannel) IsConfigured() bool {
	return c.configured
}

func (c *SMSChannel) getRecipients(alert Alert) []string {
	if phones, ok := alert.Details["phone_numbers"].([]string); ok {
		return phones
	}
	if phone, ok := alert.Details["phone_number"].(string); ok {
		return []string{phone}
	}
	return nil
}

func (c *SMSChannel) buildMessage(alert Alert) string {
	severityEmoji := map[string]string{
		"CRITICAL": "🚨",
		"HIGH":     "🔶",
		"WARNING":  "⚠️",
		"INFO":     "ℹ️",
	}

	emoji := severityEmoji[alert.Severity]
	if emoji == "" {
		emoji = "📢"
	}

	// SMS has character limits, keep it concise
	msg := fmt.Sprintf("%s [%s] %s\n%s",
		emoji,
		alert.Severity,
		alert.Title,
		truncateString(alert.Message, 120))

	if alert.IPAddress != "" {
		msg += fmt.Sprintf("\nIP: %s", alert.IPAddress)
	}

	return msg
}

// Africa's Talking implementation
func (c *SMSChannel) sendViaAfricasTalking(recipients []string, message string) error {
	apiURL := "https://api.africastalking.com/version1/messaging"
	if c.config.Username == "sandbox" {
		apiURL = "https://api.sandbox.africastalking.com/version1/messaging"
	}

	data := url.Values{}
	data.Set("username", c.config.Username)
	data.Set("to", strings.Join(recipients, ","))
	data.Set("message", message)
	if c.config.SenderID != "" {
		data.Set("from", c.config.SenderID)
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apiKey", c.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Africa's Talking returned status: %d", resp.StatusCode)
	}

	return nil
}

// Twilio implementation
func (c *SMSChannel) sendViaTwilio(recipients []string, message string) error {
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", c.config.AccountSID)

	for _, recipient := range recipients {
		data := url.Values{}
		data.Set("To", recipient)
		data.Set("From", c.config.SenderID)
		data.Set("Body", message)

		req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.SetBasicAuth(c.config.AccountSID, c.config.AuthToken)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send SMS to %s: %w", recipient, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Twilio returned status: %d for %s", resp.StatusCode, recipient)
		}
	}

	return nil
}

// Nexmo/Vonage implementation
func (c *SMSChannel) sendViaNexmo(recipients []string, message string) error {
	apiURL := "https://rest.nexmo.com/sms/json"

	for _, recipient := range recipients {
		payload := map[string]string{
			"api_key":    c.config.APIKey,
			"api_secret": c.config.Username, // Nexmo uses api_secret
			"to":         recipient,
			"from":       c.config.SenderID,
			"text":       message,
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send SMS to %s: %w", recipient, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Nexmo returned status: %d for %s", resp.StatusCode, recipient)
		}
	}

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
