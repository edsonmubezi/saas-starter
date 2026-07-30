package channels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackChannel sends alerts via Slack webhook
type SlackChannel struct {
	webhookURL string
	configured bool
	httpClient *http.Client
}

// NewSlackChannel creates a new Slack channel
func NewSlackChannel(webhookURL string) *SlackChannel {
	return &SlackChannel{
		webhookURL: webhookURL,
		configured: webhookURL != "",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send sends an alert via Slack
func (c *SlackChannel) Send(alert Alert) error {
	if !c.configured {
		return fmt.Errorf("slack channel not configured")
	}

	payload := c.buildPayload(alert)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

// Name returns the channel type
func (c *SlackChannel) Name() ChannelType {
	return ChannelSlack
}

// IsConfigured returns whether the channel is properly configured
func (c *SlackChannel) IsConfigured() bool {
	return c.configured
}

// SlackPayload represents a Slack message payload
type SlackPayload struct {
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Blocks      []Block      `json:"blocks,omitempty"`
}

// Attachment represents a Slack attachment
type Attachment struct {
	Color      string  `json:"color,omitempty"`
	Title      string  `json:"title,omitempty"`
	Text       string  `json:"text,omitempty"`
	Fields     []Field `json:"fields,omitempty"`
	Footer     string  `json:"footer,omitempty"`
	FooterIcon string  `json:"footer_icon,omitempty"`
	Ts         int64   `json:"ts,omitempty"`
}

// Field represents a Slack attachment field
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Block represents a Slack block
type Block struct {
	Type     string      `json:"type"`
	Text     *TextObject `json:"text,omitempty"`
	Elements []Element   `json:"elements,omitempty"`
	Fields   []TextObject `json:"fields,omitempty"`
}

// TextObject represents a Slack text object
type TextObject struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

// Element represents a Slack block element
type Element struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

func (c *SlackChannel) buildPayload(alert Alert) SlackPayload {
	severityColors := map[string]string{
		"CRITICAL": "#dc3545", // Red
		"HIGH":     "#fd7e14", // Orange
		"WARNING":  "#ffc107", // Yellow
		"INFO":     "#17a2b8", // Blue
	}

	severityEmoji := map[string]string{
		"CRITICAL": "🚨",
		"HIGH":     "🔶",
		"WARNING":  "⚠️",
		"INFO":     "ℹ️",
	}

	color := severityColors[alert.Severity]
	if color == "" {
		color = "#6c757d"
	}

	emoji := severityEmoji[alert.Severity]
	if emoji == "" {
		emoji = "📢"
	}

	// Build fields
	fields := []Field{
		{
			Title: "Severity",
			Value: fmt.Sprintf("%s %s", emoji, alert.Severity),
			Short: true,
		},
		{
			Title: "Event Type",
			Value: alert.EventType,
			Short: true,
		},
	}

	if alert.IPAddress != "" {
		fields = append(fields, Field{
			Title: "IP Address",
			Value: alert.IPAddress,
			Short: true,
		})
	}

	if alert.ActorEmail != "" {
		fields = append(fields, Field{
			Title: "Actor",
			Value: alert.ActorEmail,
			Short: true,
		})
	}

	// Add relevant details
	for k, v := range alert.Details {
		if k == "recipients" || k == "recipient" {
			continue
		}
		if str, ok := v.(string); ok && str != "" {
			fields = append(fields, Field{
				Title: formatFieldName(k),
				Value: str,
				Short: true,
			})
		}
	}

	return SlackPayload{
		Attachments: []Attachment{
			{
				Color:  color,
				Title:  alert.Title,
				Text:   alert.Message,
				Fields: fields,
				Footer: "Microfinance Security",
				Ts:     alert.Timestamp.Unix(),
			},
		},
	}
}

func formatFieldName(name string) string {
	// Convert snake_case to Title Case
	words := bytes.Split([]byte(name), []byte("_"))
	for i, word := range words {
		if len(word) > 0 {
			word[0] = bytes.ToUpper([]byte{word[0]})[0]
		}
		words[i] = word
	}
	return string(bytes.Join(words, []byte(" ")))
}
