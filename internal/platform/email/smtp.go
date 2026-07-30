package email

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"gopkg.in/gomail.v2"
)

// SMTPEmailService implements EmailService using SMTP
type SMTPEmailService struct {
	config EmailConfig
	dialer *gomail.Dialer
}

// NewSMTPEmailService creates a new SMTP email service
func NewSMTPEmailService(config EmailConfig) (*SMTPEmailService, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("email service is not enabled")
	}

	if config.SMTPHost == "" || config.SMTPPort == 0 {
		return nil, fmt.Errorf("SMTP host and port are required")
	}

	dialer := gomail.NewDialer(
		config.SMTPHost,
		config.SMTPPort,
		config.SMTPUser,
		config.SMTPPassword,
	)

	return &SMTPEmailService{
		config: config,
		dialer: dialer,
	}, nil
}

// SendEmail sends a plain text or HTML email
func (s *SMTPEmailService) SendEmail(ctx context.Context, to []string, subject string, body string, isHTML bool) error {
	m := gomail.NewMessage()

	// Set from address
	if s.config.FromName != "" {
		m.SetHeader("From", fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromAddress))
	} else {
		m.SetHeader("From", s.config.FromAddress)
	}

	// Set recipients
	m.SetHeader("To", to...)

	// Set subject
	m.SetHeader("Subject", subject)

	// Set body
	if isHTML {
		m.SetBody("text/html", body)
	} else {
		m.SetBody("text/plain", body)
	}

	// Send email
	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendEmailWithAttachment sends an email with an in-memory file attachment.
func (s *SMTPEmailService) SendEmailWithAttachment(ctx context.Context, to []string, subject, body string, isHTML bool, attachmentName string, attachmentData []byte) error {
	m := gomail.NewMessage()

	if s.config.FromName != "" {
		m.SetHeader("From", fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromAddress))
	} else {
		m.SetHeader("From", s.config.FromAddress)
	}

	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)

	if isHTML {
		m.SetBody("text/html", body)
	} else {
		m.SetBody("text/plain", body)
	}

	// Attach in-memory bytes (no temp file needed).
	m.Attach(attachmentName, gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := io.Copy(w, bytes.NewReader(attachmentData))
		return err
	}))

	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email with attachment: %w", err)
	}
	return nil
}

// SendTemplatedEmail sends an email using a predefined template
func (s *SMTPEmailService) SendTemplatedEmail(ctx context.Context, to []string, subject string, template Template, data interface{}) error {
	body, err := RenderTemplate(template, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.SendEmail(ctx, to, subject, body, true)
}
