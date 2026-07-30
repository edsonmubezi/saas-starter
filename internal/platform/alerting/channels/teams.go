package channels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TeamsChannel sends alerts via Microsoft Teams webhook
type TeamsChannel struct {
	webhookURL string
	configured bool
	httpClient *http.Client
}

// NewTeamsChannel creates a new Microsoft Teams channel
func NewTeamsChannel(webhookURL string) *TeamsChannel {
	return &TeamsChannel{
		webhookURL: webhookURL,
		configured: webhookURL != "",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send sends an alert via Microsoft Teams
func (c *TeamsChannel) Send(alert Alert) error {
	if !c.configured {
		return fmt.Errorf("teams channel not configured")
	}

	payload := c.buildPayload(alert)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal teams payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send teams message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("teams returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

// Name returns the channel type
func (c *TeamsChannel) Name() ChannelType {
	return ChannelTeams
}

// IsConfigured returns whether the channel is properly configured
func (c *TeamsChannel) IsConfigured() bool {
	return c.configured
}

// TeamsMessageCard represents a Microsoft Teams message card (Legacy format)
type TeamsMessageCard struct {
	Type       string        `json:"@type"`
	Context    string        `json:"@context"`
	ThemeColor string        `json:"themeColor"`
	Summary    string        `json:"summary"`
	Sections   []TeamsSection `json:"sections"`
}

// TeamsSection represents a section in a Teams message
type TeamsSection struct {
	ActivityTitle    string       `json:"activityTitle,omitempty"`
	ActivitySubtitle string       `json:"activitySubtitle,omitempty"`
	ActivityImage    string       `json:"activityImage,omitempty"`
	Facts            []TeamsFact  `json:"facts,omitempty"`
	Text             string       `json:"text,omitempty"`
	Markdown         bool         `json:"markdown"`
}

// TeamsFact represents a fact in a Teams message section
type TeamsFact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (c *TeamsChannel) buildPayload(alert Alert) TeamsMessageCard {
	severityColors := map[string]string{
		"CRITICAL": "dc3545", // Red
		"HIGH":     "fd7e14", // Orange
		"WARNING":  "ffc107", // Yellow
		"INFO":     "17a2b8", // Blue
	}

	severityEmoji := map[string]string{
		"CRITICAL": "🚨",
		"HIGH":     "🔶",
		"WARNING":  "⚠️",
		"INFO":     "ℹ️",
	}

	color := severityColors[alert.Severity]
	if color == "" {
		color = "6c757d"
	}

	emoji := severityEmoji[alert.Severity]
	if emoji == "" {
		emoji = "📢"
	}

	// Build facts
	facts := []TeamsFact{
		{
			Name:  "Severity",
			Value: fmt.Sprintf("%s %s", emoji, alert.Severity),
		},
		{
			Name:  "Event Type",
			Value: alert.EventType,
		},
		{
			Name:  "Timestamp",
			Value: alert.Timestamp.Format(time.RFC1123),
		},
	}

	if alert.IPAddress != "" {
		facts = append(facts, TeamsFact{
			Name:  "IP Address",
			Value: alert.IPAddress,
		})
	}

	if alert.ActorEmail != "" {
		facts = append(facts, TeamsFact{
			Name:  "Actor",
			Value: alert.ActorEmail,
		})
	}

	// Add relevant details
	for k, v := range alert.Details {
		if k == "recipients" || k == "recipient" || k == "phone_numbers" || k == "phone_number" {
			continue
		}
		if str, ok := v.(string); ok && str != "" {
			facts = append(facts, TeamsFact{
				Name:  formatFieldName(k),
				Value: str,
			})
		}
	}

	return TeamsMessageCard{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: color,
		Summary:    alert.Title,
		Sections: []TeamsSection{
			{
				ActivityTitle:    alert.Title,
				ActivitySubtitle: "Microfinance Security Alert",
				Facts:            facts,
				Text:             alert.Message,
				Markdown:         true,
			},
		},
	}
}
