package channels

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"
)

// EmailConfig holds email channel configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromAddress  string
	FromName     string
	UseTLS       bool
}

// EmailChannel sends alerts via email
type EmailChannel struct {
	config     EmailConfig
	configured bool
}

// NewEmailChannel creates a new email channel
func NewEmailChannel(cfg EmailConfig) *EmailChannel {
	return &EmailChannel{
		config:     cfg,
		configured: cfg.SMTPHost != "" && cfg.FromAddress != "",
	}
}

// Send sends an alert via email
func (c *EmailChannel) Send(alert Alert) error {
	if !c.configured {
		return fmt.Errorf("email channel not configured")
	}

	// Get recipients from alert details
	recipients := c.getRecipients(alert)
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified for email alert")
	}

	// Build email content
	subject := c.buildSubject(alert)
	body, err := c.buildHTMLBody(alert)
	if err != nil {
		return fmt.Errorf("failed to build email body: %w", err)
	}

	// Send to each recipient
	for _, recipient := range recipients {
		if err := c.sendEmail(recipient, subject, body); err != nil {
			return fmt.Errorf("failed to send email to %s: %w", recipient, err)
		}
	}

	return nil
}

// Name returns the channel type
func (c *EmailChannel) Name() ChannelType {
	return ChannelEmail
}

// IsConfigured returns whether the channel is properly configured
func (c *EmailChannel) IsConfigured() bool {
	return c.configured
}

func (c *EmailChannel) getRecipients(alert Alert) []string {
	if recipients, ok := alert.Details["recipients"].([]string); ok {
		return recipients
	}
	if recipient, ok := alert.Details["recipient"].(string); ok {
		return []string{recipient}
	}
	return nil
}

func (c *EmailChannel) buildSubject(alert Alert) string {
	severityEmoji := map[string]string{
		"INFO":     "ℹ️",
		"WARNING":  "⚠️",
		"HIGH":     "🔶",
		"CRITICAL": "🚨",
	}

	emoji := severityEmoji[alert.Severity]
	if emoji == "" {
		emoji = "📢"
	}

	return fmt.Sprintf("%s [%s] %s", emoji, alert.Severity, alert.Title)
}

func (c *EmailChannel) buildHTMLBody(alert Alert) (string, error) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: {{.HeaderColor}}; color: white; padding: 20px; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 20px; border: 1px solid #ddd; border-top: none; }
        .footer { background: #eee; padding: 15px; text-align: center; font-size: 12px; color: #666; border-radius: 0 0 8px 8px; }
        .severity { display: inline-block; padding: 4px 12px; border-radius: 4px; font-weight: bold; }
        .severity-critical { background: #dc3545; color: white; }
        .severity-high { background: #fd7e14; color: white; }
        .severity-warning { background: #ffc107; color: #333; }
        .severity-info { background: #17a2b8; color: white; }
        .detail-table { width: 100%; border-collapse: collapse; margin: 15px 0; }
        .detail-table td { padding: 8px; border-bottom: 1px solid #ddd; }
        .detail-table td:first-child { font-weight: bold; width: 30%; }
        pre { background: #2d2d2d; color: #f8f8f2; padding: 15px; border-radius: 4px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2 style="margin: 0;">{{.Title}}</h2>
            <span class="severity severity-{{.SeverityClass}}">{{.Severity}}</span>
        </div>
        <div class="content">
            <p>{{.Message}}</p>

            <table class="detail-table">
                <tr>
                    <td>Event Type</td>
                    <td>{{.EventType}}</td>
                </tr>
                <tr>
                    <td>Timestamp</td>
                    <td>{{.Timestamp}}</td>
                </tr>
                {{if .IPAddress}}
                <tr>
                    <td>IP Address</td>
                    <td>{{.IPAddress}}</td>
                </tr>
                {{end}}
                {{if .ActorEmail}}
                <tr>
                    <td>Actor</td>
                    <td>{{.ActorEmail}}</td>
                </tr>
                {{end}}
            </table>

            {{if .HasDetails}}
            <h4>Additional Details</h4>
            <pre>{{.DetailsJSON}}</pre>
            {{end}}
        </div>
        <div class="footer">
            <p>This is an automated alert from Microfinance Security System</p>
            <p>Alert ID: {{.AlertID}}</p>
        </div>
    </div>
</body>
</html>`

	headerColors := map[string]string{
		"CRITICAL": "#dc3545",
		"HIGH":     "#fd7e14",
		"WARNING":  "#ffc107",
		"INFO":     "#17a2b8",
	}

	headerColor := headerColors[alert.Severity]
	if headerColor == "" {
		headerColor = "#6c757d"
	}

	data := map[string]interface{}{
		"Title":         alert.Title,
		"Message":       alert.Message,
		"Severity":      alert.Severity,
		"SeverityClass": strings.ToLower(alert.Severity),
		"HeaderColor":   headerColor,
		"EventType":     alert.EventType,
		"Timestamp":     alert.Timestamp.Format(time.RFC1123),
		"IPAddress":     alert.IPAddress,
		"ActorEmail":    alert.ActorEmail,
		"AlertID":       alert.ID,
		"HasDetails":    len(alert.Details) > 0,
		"DetailsJSON":   formatDetails(alert.Details),
	}

	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (c *EmailChannel) sendEmail(to, subject, htmlBody string) error {
	from := c.config.FromAddress
	if c.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", c.config.FromName, c.config.FromAddress)
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	addr := fmt.Sprintf("%s:%d", c.config.SMTPHost, c.config.SMTPPort)

	var auth smtp.Auth
	if c.config.SMTPUser != "" && c.config.SMTPPassword != "" {
		auth = smtp.PlainAuth("", c.config.SMTPUser, c.config.SMTPPassword, c.config.SMTPHost)
	}

	if c.config.UseTLS || c.config.SMTPPort == 465 {
		return c.sendWithTLS(addr, auth, c.config.FromAddress, to, msg.Bytes())
	}

	return smtp.SendMail(addr, auth, c.config.FromAddress, []string{to}, msg.Bytes())
}

func (c *EmailChannel) sendWithTLS(addr string, auth smtp.Auth, from, to string, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName: c.config.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, c.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL command failed: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT command failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return client.Quit()
}

func formatDetails(details map[string]interface{}) string {
	if len(details) == 0 {
		return ""
	}

	var lines []string
	for k, v := range details {
		// Skip internal fields
		if k == "recipients" || k == "recipient" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %v", k, v))
	}
	return strings.Join(lines, "\n")
}
