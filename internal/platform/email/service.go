package email

import (
	"context"
	"fmt"
)

// EmailService defines the interface for sending emails
type EmailService interface {
	SendEmail(ctx context.Context, to []string, subject string, body string, isHTML bool) error
	SendTemplatedEmail(ctx context.Context, to []string, subject string, template Template, data interface{}) error
	SendEmailWithAttachment(ctx context.Context, to []string, subject, body string, isHTML bool, attachmentName string, attachmentData []byte) error
}

// Template represents an email template
type Template string

const (
	TemplatePasswordReset            Template = "password_reset"
	TemplatePasswordResetConfirmation Template = "password_reset_confirmation"
	TemplateWelcome                   Template = "welcome"
	TemplateEmailVerification         Template = "email_verification"
	TemplateSalarySlip                Template = "salary_slip"
	TemplateLeaveRequest              Template = "leave_request"
	TemplateLeaveApprovedTier1        Template = "leave_approved_tier1"
	TemplateLeaveApproved             Template = "leave_approved"
	TemplateLeaveRejected             Template = "leave_rejected"
	TemplateContractExpiring              Template = "contract_expiring"
	TemplateProbationEnding               Template = "probation_ending"
	TemplateLoanApplicationSubmitted      Template = "loan_application_submitted"
	TemplateDailyPendingAlert             Template = "daily_pending_alert"
)

// EmailConfig holds the configuration for email service
type EmailConfig struct {
	Enabled      bool
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromAddress  string
	FromName     string
}

// MockEmailService is a mock implementation for testing
type MockEmailService struct {
	SentEmails []SentEmail
}

type SentEmail struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		SentEmails: make([]SentEmail, 0),
	}
}

func (m *MockEmailService) SendEmail(ctx context.Context, to []string, subject string, body string, isHTML bool) error {
	m.SentEmails = append(m.SentEmails, SentEmail{
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  isHTML,
	})
	return nil
}

func (m *MockEmailService) SendTemplatedEmail(ctx context.Context, to []string, subject string, template Template, data interface{}) error {
	body, err := RenderTemplate(template, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}
	return m.SendEmail(ctx, to, subject, body, true)
}

func (m *MockEmailService) SendEmailWithAttachment(ctx context.Context, to []string, subject, body string, isHTML bool, attachmentName string, attachmentData []byte) error {
	return m.SendEmail(ctx, to, subject, body, isHTML)
}
