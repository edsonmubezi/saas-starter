package email

import (
	"bytes"
	"fmt"
	"html/template"
)

// EmailBranding holds customizable branding for email templates.
// Callers that have org context should fetch this from the DB; callers without
// org context (e.g. password reset) simply omit it and defaults are used.
type EmailBranding struct {
	PrimaryColor    string
	HeaderTextColor string
	AccentColor     string
	FontFamily      string
	FooterText      string
	SignOff         string
}

// DefaultBranding returns the default email branding (matches original hardcoded values).
func DefaultBranding() EmailBranding {
	return EmailBranding{
		PrimaryColor:    "#4F46E5",
		HeaderTextColor: "#FFFFFF",
		AccentColor:     "#4F46E5",
		FontFamily:      "Arial, sans-serif",
		FooterText:      "",
		SignOff:         "",
	}
}

// PasswordResetData holds data for password reset email
type PasswordResetData struct {
	UserName    string
	ResetLink   string
	ExpiresIn   string
	CompanyName string
}

// PasswordResetConfirmationData holds data for password reset confirmation email
type PasswordResetConfirmationData struct {
	UserName    string
	CompanyName string
}

// WelcomeData holds data for welcome email
type WelcomeData struct {
	UserName    string
	LoginLink   string
	CompanyName string
	Email       string // optional: shown in credentials block
	TempPassword string // optional: shown in credentials block
}

// EmailVerificationData holds data for email verification email
type EmailVerificationData struct {
	UserName         string
	VerificationLink string
	ExpiresIn        string
	CompanyName      string
}

// SalarySlipData holds data for salary slip email
type SalarySlipData struct {
	EmployeeName  string
	PayrollPeriod string
	CompanyName   string
	CustomMessage string
}

// LeaveRequestData holds data for leave request notification email
type LeaveRequestData struct {
	EmployeeName string
	LeaveType    string
	StartDate    string
	EndDate      string
	Days         int
	Reason       string
	CompanyName  string
}

// LeaveApprovedTier1Data holds data for tier-1 approval notification email
type LeaveApprovedTier1Data struct {
	EmployeeName string
	ApproverName string
	LeaveType    string
	StartDate    string
	EndDate      string
	Days         int
	Remarks      string
	CompanyName  string
}

// LeaveApprovedData holds data for final (HR) approval notification email
type LeaveApprovedData struct {
	EmployeeName string
	ApproverName string
	LeaveType    string
	StartDate    string
	EndDate      string
	Days         int
	Remarks      string
	CompanyName  string
}

// LeaveRejectedData holds data for leave rejection notification email
type LeaveRejectedData struct {
	EmployeeName string
	RejectedBy   string
	LeaveType    string
	StartDate    string
	EndDate      string
	Days         int
	Reason       string
	CompanyName  string
}

// ContractExpiringData holds data for contract expiration warning emails
type ContractExpiringData struct {
	EmployeeName    string
	ContractEndDate string
	DaysRemaining   int
	CompanyName     string
}

// ProbationEndingData holds data for probation end warning emails
type ProbationEndingData struct {
	EmployeeName     string
	ProbationEndDate string
	DaysRemaining    int
	CompanyName      string
}

// LoanApplicationSubmittedData holds data for loan application notification email
type LoanApplicationSubmittedData struct {
	EmployeeName string
	LoanAmount   string
	Duration     int
	InterestRate float64
	InterestType string
	CompanyName  string
	IsApplicant  bool // true = sent to the employee themselves
}

// DailyPendingAlertData holds data for the daily pending alert digest email.
type DailyPendingAlertData struct {
	UserName   string
	OrgName    string
	Items      []DailyPendingAlertItem
	TotalCount int
}

// DailyPendingAlertItem is one row in the digest.
type DailyPendingAlertItem struct {
	Label string
	Count int
}

// RenderTemplate renders an email template with the given data.
// An optional EmailBranding can be passed to customize colors and fonts;
// if omitted, defaults are used (backward-compatible).
func RenderTemplate(tmpl Template, data interface{}, branding ...EmailBranding) (string, error) {
	b := DefaultBranding()
	if len(branding) > 0 {
		b = branding[0]
	}

	var templateStr string

	switch tmpl {
	case TemplatePasswordReset:
		templateStr = passwordResetTemplate
	case TemplatePasswordResetConfirmation:
		templateStr = passwordResetConfirmationTemplate
	case TemplateWelcome:
		templateStr = welcomeTemplate
	case TemplateEmailVerification:
		templateStr = emailVerificationTemplate
	case TemplateSalarySlip:
		templateStr = salarySlipTemplate
	case TemplateLeaveRequest:
		templateStr = leaveRequestTemplate
	case TemplateLeaveApprovedTier1:
		templateStr = leaveApprovedTier1Template
	case TemplateLeaveApproved:
		templateStr = leaveApprovedTemplate
	case TemplateLeaveRejected:
		templateStr = leaveRejectedTemplate
	case TemplateContractExpiring:
		templateStr = contractExpiringTemplate
	case TemplateProbationEnding:
		templateStr = probationEndingTemplate
	case TemplateLoanApplicationSubmitted:
		templateStr = loanApplicationSubmittedTemplate
	case TemplateDailyPendingAlert:
		templateStr = dailyPendingAlertTemplate
	default:
		return "", fmt.Errorf("unknown template: %s", tmpl)
	}

	funcMap := template.FuncMap{
		"brandPrimary":    func() string { return b.PrimaryColor },
		"brandHeaderText": func() string { return b.HeaderTextColor },
		"brandAccent":     func() string { return b.AccentColor },
		"brandFont":       func() string { return b.FontFamily },
		"brandFooter":     func() string { return b.FooterText },
		"brandSignOff":    func() string { return b.SignOff },
	}

	t, err := template.New(string(tmpl)).Funcs(funcMap).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

const passwordResetTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset Request</title>
    <style>
        body {
            font-family: {{brandFont}};
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: {{brandPrimary}};
            color: {{brandHeaderText}};
            padding: 20px;
            text-align: center;
            border-radius: 5px 5px 0 0;
        }
        .content {
            background-color: #f9f9f9;
            padding: 30px;
            border: 1px solid #ddd;
            border-radius: 0 0 5px 5px;
        }
        .button {
            display: inline-block;
            padding: 12px 30px;
            background-color: {{brandPrimary}};
            color: {{brandHeaderText}};
            text-decoration: none;
            border-radius: 5px;
            margin: 20px 0;
        }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 12px;
            color: #666;
        }
        .warning {
            background-color: #FEF2F2;
            border-left: 4px solid #EF4444;
            padding: 15px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Password Reset Request</h1>
    </div>
    <div class="content">
        <p>Hi {{.UserName}},</p>

        <p>We received a request to reset your password for your {{.CompanyName}} account.</p>

        <p>Click the button below to reset your password:</p>

        <a href="{{.ResetLink}}" class="button">Reset Password</a>

        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: {{brandAccent}};">{{.ResetLink}}</p>

        <div class="warning">
            <strong>Security Notice:</strong>
            <ul>
                <li>This link will expire in {{.ExpiresIn}}</li>
                <li>If you didn't request this reset, please ignore this email</li>
                <li>Never share this link with anyone</li>
            </ul>
        </div>

        <p>If you have any questions, please contact your system administrator.</p>

        {{if brandSignOff}}<p>{{brandSignOff}}</p>{{else}}<p>Best regards,<br>{{.CompanyName}} Team</p>{{end}}
    </div>
    <div class="footer">
        {{if brandFooter}}<p>{{brandFooter}}</p>{{end}}
        <p>This is an automated email. Please do not reply to this message.</p>
        <p>&copy; {{.CompanyName}}. All rights reserved.</p>
    </div>
</body>
</html>
`

const passwordResetConfirmationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset Successful</title>
    <style>
        body {
            font-family: {{brandFont}};
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: #10B981;
            color: white;
            padding: 20px;
            text-align: center;
            border-radius: 5px 5px 0 0;
        }
        .content {
            background-color: #f9f9f9;
            padding: 30px;
            border: 1px solid #ddd;
            border-radius: 0 0 5px 5px;
        }
        .success-icon {
            font-size: 48px;
            text-align: center;
            margin: 20px 0;
        }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 12px;
            color: #666;
        }
        .warning {
            background-color: #FEF2F2;
            border-left: 4px solid #EF4444;
            padding: 15px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Password Changed Successfully</h1>
    </div>
    <div class="content">
        <div class="success-icon">&#10003;</div>

        <p>Hi {{.UserName}},</p>

        <p>Your password has been successfully changed for your {{.CompanyName}} account.</p>

        <p>You can now log in with your new password.</p>

        <div class="warning">
            <strong>Didn't make this change?</strong>
            <p>If you didn't change your password, please contact your system administrator immediately.</p>
        </div>

        {{if brandSignOff}}<p>{{brandSignOff}}</p>{{else}}<p>Best regards,<br>{{.CompanyName}} Team</p>{{end}}
    </div>
    <div class="footer">
        {{if brandFooter}}<p>{{brandFooter}}</p>{{end}}
        <p>This is an automated email. Please do not reply to this message.</p>
        <p>&copy; {{.CompanyName}}. All rights reserved.</p>
    </div>
</body>
</html>
`

const welcomeTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome</title>
    <style>
        body {
            font-family: {{brandFont}};
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: {{brandPrimary}};
            color: {{brandHeaderText}};
            padding: 20px;
            text-align: center;
            border-radius: 5px 5px 0 0;
        }
        .content {
            background-color: #f9f9f9;
            padding: 30px;
            border: 1px solid #ddd;
            border-radius: 0 0 5px 5px;
        }
        .button {
            display: inline-block;
            padding: 12px 30px;
            background-color: {{brandPrimary}};
            color: {{brandHeaderText}};
            text-decoration: none;
            border-radius: 5px;
            margin: 20px 0;
        }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 12px;
            color: #666;
        }
        .credentials {
            background-color: #EEF2FF;
            border-left: 4px solid {{brandAccent}};
            padding: 15px;
            margin: 20px 0;
            font-family: monospace;
        }
        .warning {
            background-color: #FEF2F2;
            border-left: 4px solid #EF4444;
            padding: 15px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Welcome to {{.CompanyName}}!</h1>
    </div>
    <div class="content">
        <p>Hi {{.UserName}},</p>

        <p>Welcome to {{.CompanyName}}! Your account has been successfully created.</p>

        {{if .TempPassword}}
        <div class="credentials">
            <strong>Your Login Credentials:</strong><br>
            <strong>Email:</strong> {{.Email}}<br>
            <strong>Temporary Password:</strong> {{.TempPassword}}
        </div>

        <div class="warning">
            <strong>Important:</strong>
            <ul>
                <li>You will be required to change your password upon first login</li>
                <li>Do not share this password with anyone</li>
            </ul>
        </div>
        {{end}}

        <p>Click the button below to log in:</p>

        <a href="{{.LoginLink}}" class="button">Log In</a>

        <p>If you have any questions, please contact your system administrator.</p>

        {{if brandSignOff}}<p>{{brandSignOff}}</p>{{else}}<p>Best regards,<br>{{.CompanyName}} Team</p>{{end}}
    </div>
    <div class="footer">
        {{if brandFooter}}<p>{{brandFooter}}</p>{{end}}
        <p>This is an automated email. Please do not reply to this message.</p>
        <p>&copy; {{.CompanyName}}. All rights reserved.</p>
    </div>
</body>
</html>
`

const emailVerificationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Email Verification</title>
    <style>
        body {
            font-family: {{brandFont}};
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: {{brandPrimary}};
            color: {{brandHeaderText}};
            padding: 20px;
            text-align: center;
            border-radius: 5px 5px 0 0;
        }
        .content {
            background-color: #f9f9f9;
            padding: 30px;
            border: 1px solid #ddd;
            border-radius: 0 0 5px 5px;
        }
        .button {
            display: inline-block;
            padding: 12px 30px;
            background-color: {{brandPrimary}};
            color: {{brandHeaderText}};
            text-decoration: none;
            border-radius: 5px;
            margin: 20px 0;
        }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 12px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Verify Your Email</h1>
    </div>
    <div class="content">
        <p>Hi {{.UserName}},</p>

        <p>Thank you for registering with {{.CompanyName}}. Please verify your email address to activate your account.</p>

        <p>Click the button below to verify your email:</p>

        <a href="{{.VerificationLink}}" class="button">Verify Email</a>

        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: {{brandAccent}};">{{.VerificationLink}}</p>

        <p><strong>This link will expire in {{.ExpiresIn}}.</strong></p>

        <p>If you didn't create an account, please ignore this email.</p>

        {{if brandSignOff}}<p>{{brandSignOff}}</p>{{else}}<p>Best regards,<br>{{.CompanyName}} Team</p>{{end}}
    </div>
    <div class="footer">
        {{if brandFooter}}<p>{{brandFooter}}</p>{{end}}
        <p>This is an automated email. Please do not reply to this message.</p>
        <p>&copy; {{.CompanyName}}. All rights reserved.</p>
    </div>
</body>
</html>
`

const salarySlipTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Salary Slip</title>
    <style>
        body {
            font-family: {{brandFont}};
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: {{brandPrimary}};
            color: {{brandHeaderText}};
            padding: 20px;
            text-align: center;
            border-radius: 5px 5px 0 0;
        }
        .content {
            background-color: #f9f9f9;
            padding: 30px;
            border: 1px solid #ddd;
            border-radius: 0 0 5px 5px;
        }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 12px;
            color: #666;
        }
        .info-box {
            background-color: #EEF2FF;
            border-left: 4px solid {{brandAccent}};
            padding: 15px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Salary Slip - {{.PayrollPeriod}}</h1>
    </div>
    <div class="content">
        <p>Dear {{.EmployeeName}},</p>

        <p>Please find your salary slip for the period <strong>{{.PayrollPeriod}}</strong> attached to this email.</p>

        {{if .CustomMessage}}
        <div class="info-box">
            <p>{{.CustomMessage}}</p>
        </div>
        {{end}}

        <p>If you have any questions regarding your salary slip, please contact the HR department.</p>

        {{if brandSignOff}}<p>{{brandSignOff}}</p>{{else}}<p>Best regards,<br>{{.CompanyName}} Team</p>{{end}}
    </div>
    <div class="footer">
        {{if brandFooter}}<p>{{brandFooter}}</p>{{end}}
        <p>This is an automated email. Please do not reply to this message.</p>
        <p>&copy; {{.CompanyName}}. All rights reserved.</p>
    </div>
</body>
</html>
`

const leaveRequestTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Leave Request</title>
    <style>
        body { font-family: {{brandFont}}; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: {{brandPrimary}}; color: {{brandHeaderText}}; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border: 1px solid #ddd; border-radius: 0 0 5px 5px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .info-box { background-color: #EEF2FF; border-left: 4px solid {{brandAccent}}; padding: 4px 15px; margin: 20px 0; border-radius: 4px; }
        .info-table { width: 100%; border-collapse: collapse; }
        .info-table td { padding: 8px 0; border-bottom: 1px solid #eee; }
        .info-table tr:last-child td { border-bottom: none; }
        .detail-label { font-weight: bold; color: #555; padding-right: 12px; white-space: nowrap; vertical-align: top; }
    </style>
</head>
<body>
    <div class="header">
        <h1>New Leave Request</h1>
    </div>
    <div class="content">
        <p>A new leave request has been submitted and requires your attention.</p>

        <div class="info-box">
            <table class="info-table">
            <tr><td class="detail-label">Employee:</td><td>{{.EmployeeName}}</td></tr>
            <tr><td class="detail-label">Leave Type:</td><td>{{.LeaveType}}</td></tr>
            <tr><td class="detail-label">From:</td><td>{{.StartDate}}</td></tr>
            <tr><td class="detail-label">To:</td><td>{{.EndDate}}</td></tr>
            <tr><td class="detail-label">Days:</td><td>{{.Days}}</td></tr>
            <tr><td class="detail-label">Reason:</td><td>{{.Reason}}</td></tr>
            </table>
        </div>

        <p>Please log in to the system to review and take action on this request.</p>

        {{if brandSignOff}}<p>{{brandSignOff}}</p>{{else}}<p>Best regards,<br>{{.CompanyName}} Team</p>{{end}}
    </div>
    <div class="footer">
        {{if brandFooter}}<p>{{brandFooter}}</p>{{end}}
        <p>This is an automated email. Please do not reply to this message.</p>
        <p>&copy; {{.CompanyName}}. All rights reserved.</p>
    </div>
</body>
</html>
`

const leaveApprovedTier1Template = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Leave Pre-Approved by Manager</title>
    <style>
        body { font-family: {{brandFont}}; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #F59E0B; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border: 1px solid #ddd; border-radius: 0 0 5px 5px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .info-box { background-color: #FFFBEB; border-left: 4px solid #F59E0B; padding: 4px 15px; margin: 20px 0; border-radius: 4px; }
        .info-table { width: 100%; border-collapse: collapse; }
        .info-table td { padding: 8px 0; border-bottom: 1px solid #eee; }
        .info-table tr:last-child td { border-bottom: none; }
        .detail-label { font-weight: bold; color: #555; padding-right: 12px; white-space: nowrap; vertical-align: top; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Leave Pre-Approved by Manager</h1>
    </div>
    <div class="content">
        <p>A leave request has been approved at the manager level and is now awaiting final HR approval.</p>

        <div class="info-box">
            <table class="info-table">
            <tr><td class="detail-label">Employee:</td><td>{{.EmployeeName}}</td></tr>
            <tr><td class="detail-label">Approved By:</td><td>{{.ApproverName}}</td></tr>
            <tr><td class="detail-label">Leave Type:</td><td>{{.LeaveType}}</td></tr>
            <tr><td class="detail-label">From:</td><td>{{.StartDate}}</td></tr>
            <tr><td class="detail-label">To:</td><td>{{.EndDate}}</td></tr>
            <tr><td class="detail-label">Days:</td><td>{{.Days}}</td></tr>
            {{if .Remarks}}<tr><td class="detail-label">Remarks:</td><td>{{.Remarks}}</td></tr>{{end}}
            </table>
        </div>

        {{if brandSignOff}}<p>{{brandSignOff}}</p>{{else}}<p>Best regards,<br>{{.CompanyName}} Team</p>{{end}}
    </div>
    <div class="footer">
        {{if brandFooter}}<p>{{brandFooter}}</p>{{end}}
        <p>This is an automated email. Please do not reply to this message.</p>
        <p>&copy; {{.CompanyName}}. All rights reserved.</p>
    </div>
</body>
</html>
`

const leaveApprovedTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Leave Approved</title>
    <style>
        body { font-family: {{brandFont}}; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #10B981; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border: 1px solid #ddd; border-radius: 0 0 5px 5px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .info-box { background-color: #ECFDF5; border-left: 4px solid #10B981; padding: 4px 15px; margin: 20px 0; border-radius: 4px; }
        .info-table { width: 100%; border-collapse: collapse; }
        .info-table td { padding: 8px 0; border-bottom: 1px solid #eee; }
        .info-table tr:last-child td { border-bottom: none; }
        .detail-label { font-weight: bold; color: #555; padding-right: 12px; white-space: nowrap; vertical-align: top; }
        .success-icon { font-size: 48px; text-align: center; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Leave Approved</h1>
    </div>
    <div class="content">
        <div class="success-icon">&#10003;</div>

        <p>Dear {{.EmployeeName}},</p>

        <p>Your leave request has been approved.</p>

        <div class="info-box">
            <table class="info-table">
            <tr><td class="detail-label">Leave Type:</td><td>{{.LeaveType}}</td></tr>
            <tr><td class="detail-label">From:</td><td>{{.StartDate}}</td></tr>
            <tr><td class="detail-label">To:</td><td>{{.EndDate}}</td></tr>
            <tr><td class="detail-label">Days:</td><td>{{.Days}}</td></tr>
            <tr><td class="detail-label">Approved By:</td><td>{{.ApproverName}}</td></tr>
            {{if .Remarks}}<tr><td class="detail-label">Remarks:</td><td>{{.Remarks}}</td></tr>{{end}}
            </table>
        </div>

        {{if brandSignOff}}<p>{{brandSignOff}}</p>{{else}}<p>Best regards,<br>{{.CompanyName}} Team</p>{{end}}
    </div>
    <div class="footer">
        {{if brandFooter}}<p>{{brandFooter}}</p>{{end}}
        <p>This is an automated email. Please do not reply to this message.</p>
        <p>&copy; {{.CompanyName}}. All rights reserved.</p>
    </div>
</body>
</html>
`

const leaveRejectedTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Leave Rejected</title>
    <style>
        body { font-family: {{brandFont}}; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #EF4444; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border: 1px solid #ddd; border-radius: 0 0 5px 5px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .info-box { background-color: #FEF2F2; border-left: 4px solid #EF4444; padding: 4px 15px; margin: 20px 0; border-radius: 4px; }
        .info-table { width: 100%; border-collapse: collapse; }
        .info-table td { padding: 8px 0; border-bottom: 1px solid #eee; }
        .info-table tr:last-child td { border-bottom: none; }
        .detail-label { font-weight: bold; color: #555; padding-right: 12px; white-space: nowrap; vertical-align: top; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Leave Request Rejected</h1>
    </div>
    <div class="content">
        <p>Dear {{.EmployeeName}},</p>

        <p>Unfortunately, your leave request has been rejected.</p>

        <div class="info-box">
            <table class="info-table">
            <tr><td class="detail-label">Leave Type:</td><td>{{.LeaveType}}</td></tr>
            <tr><td class="detail-label">From:</td><td>{{.StartDate}}</td></tr>
            <tr><td class="detail-label">To:</td><td>{{.EndDate}}</td></tr>
            <tr><td class="detail-label">Days:</td><td>{{.Days}}</td></tr>
            <tr><td class="detail-label">Rejected By:</td><td>{{.RejectedBy}}</td></tr>
            <tr><td class="detail-label">Reason:</td><td>{{.Reason}}</td></tr>
            </table>
        </div>

        <p>If you have any questions, please contact your supervisor or the HR department.</p>

        {{if brandSignOff}}<p>{{brandSignOff}}</p>{{else}}<p>Best regards,<br>{{.CompanyName}} Team</p>{{end}}
    </div>
    <div class="footer">
        {{if brandFooter}}<p>{{brandFooter}}</p>{{end}}
        <p>This is an automated email. Please do not reply to this message.</p>
        <p>&copy; {{.CompanyName}}. All rights reserved.</p>
    </div>
</body>
</html>
`

const contractExpiringTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Contract Expiring Soon</title>
    <style>
        body { font-family: {{brandFont}}; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #F59E0B; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border: 1px solid #ddd; border-radius: 0 0 5px 5px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .info-box { background-color: #FFFBEB; border-left: 4px solid #F59E0B; padding: 4px 15px; margin: 20px 0; border-radius: 4px; }
        .info-table { width: 100%; border-collapse: collapse; }
        .info-table td { padding: 8px 0; border-bottom: 1px solid #eee; }
        .info-table tr:last-child td { border-bottom: none; }
        .detail-label { font-weight: bold; color: #555; padding-right: 12px; white-space: nowrap; vertical-align: top; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Contract Expiring Soon</h1>
    </div>
    <div class="content">
        <p>This is a reminder that the following employee's contract is expiring soon:</p>

        <div class="info-box">
            <table class="info-table">
            <tr><td class="detail-label">Employee:</td><td>{{.EmployeeName}}</td></tr>
            <tr><td class="detail-label">Contract End Date:</td><td>{{.ContractEndDate}}</td></tr>
            <tr><td class="detail-label">Days Remaining:</td><td>{{.DaysRemaining}}</td></tr>
            </table>
        </div>

        <p>Please take the necessary action — extend the contract, prepare for transition, or confirm non-renewal.</p>

        {{if brandSignOff}}<p>{{brandSignOff}}</p>{{else}}<p>Best regards,<br>{{.CompanyName}} HR System</p>{{end}}
    </div>
    <div class="footer">
        {{if brandFooter}}<p>{{brandFooter}}</p>{{end}}
        <p>This is an automated email. Please do not reply to this message.</p>
        <p>&copy; {{.CompanyName}}. All rights reserved.</p>
    </div>
</body>
</html>
`

const probationEndingTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Probation Period Ending</title>
    <style>
        body { font-family: {{brandFont}}; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #3B82F6; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border: 1px solid #ddd; border-radius: 0 0 5px 5px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .info-box { background-color: #EFF6FF; border-left: 4px solid #3B82F6; padding: 4px 15px; margin: 20px 0; border-radius: 4px; }
        .info-table { width: 100%; border-collapse: collapse; }
        .info-table td { padding: 8px 0; border-bottom: 1px solid #eee; }
        .info-table tr:last-child td { border-bottom: none; }
        .detail-label { font-weight: bold; color: #555; padding-right: 12px; white-space: nowrap; vertical-align: top; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Probation Period Ending</h1>
    </div>
    <div class="content">
        <p>This is a reminder that the following employee's probation period is ending soon:</p>

        <div class="info-box">
            <table class="info-table">
            <tr><td class="detail-label">Employee:</td><td>{{.EmployeeName}}</td></tr>
            <tr><td class="detail-label">Probation End Date:</td><td>{{.ProbationEndDate}}</td></tr>
            <tr><td class="detail-label">Days Remaining:</td><td>{{.DaysRemaining}}</td></tr>
            </table>
        </div>

        <p>Please review the employee's performance and decide whether to confirm the contract or extend the probation period.</p>

        {{if brandSignOff}}<p>{{brandSignOff}}</p>{{else}}<p>Best regards,<br>{{.CompanyName}} HR System</p>{{end}}
    </div>
    <div class="footer">
        {{if brandFooter}}<p>{{brandFooter}}</p>{{end}}
        <p>This is an automated email. Please do not reply to this message.</p>
        <p>&copy; {{.CompanyName}}. All rights reserved.</p>
    </div>
</body>
</html>
`

const loanApplicationSubmittedTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Loan Application {{if .IsApplicant}}Submitted{{else}}Received{{end}}</title>
    <style>
        body { font-family: {{brandFont}}; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: {{brandPrimary}}; color: {{brandHeaderText}}; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border: 1px solid #ddd; border-radius: 0 0 5px 5px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .info-box { background-color: #EEF2FF; border-left: 4px solid {{brandAccent}}; padding: 4px 15px; margin: 20px 0; border-radius: 4px; }
        .info-table { width: 100%; border-collapse: collapse; }
        .info-table td { padding: 8px 0; border-bottom: 1px solid #eee; }
        .info-table tr:last-child td { border-bottom: none; }
        .detail-label { font-weight: bold; color: #555; padding-right: 12px; white-space: nowrap; vertical-align: top; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Loan Application {{if .IsApplicant}}Submitted{{else}}Received{{end}}</h1>
    </div>
    <div class="content">
        {{if .IsApplicant}}
        <p>Dear {{.EmployeeName}},</p>
        <p>Your loan application has been submitted successfully and is pending review.</p>
        {{else}}
        <p>A new loan application has been submitted and requires your attention.</p>
        {{end}}

        <div class="info-box">
            <table class="info-table">
            <tr><td class="detail-label">Employee:</td><td>{{.EmployeeName}}</td></tr>
            <tr><td class="detail-label">Loan Amount:</td><td>{{.LoanAmount}}</td></tr>
            <tr><td class="detail-label">Duration:</td><td>{{.Duration}} months</td></tr>
            <tr><td class="detail-label">Interest Rate:</td><td>{{.InterestRate}}% ({{.InterestType}})</td></tr>
            </table>
        </div>

        {{if .IsApplicant}}
        <p>You will be notified once your application has been reviewed.</p>
        {{else}}
        <p>Please log in to the system to review and take action on this application.</p>
        {{end}}

        {{if brandSignOff}}<p>{{brandSignOff}}</p>{{else}}<p>Best regards,<br>{{.CompanyName}} Team</p>{{end}}
    </div>
    <div class="footer">
        {{if brandFooter}}<p>{{brandFooter}}</p>{{end}}
        <p>This is an automated email. Please do not reply to this message.</p>
        <p>&copy; {{.CompanyName}}. All rights reserved.</p>
    </div>
</body>
</html>
`

const dailyPendingAlertTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Daily Pending Items Digest</title>
    <style>
        body { font-family: {{brandFont}}; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f4f4f4; }
        .header { background: linear-gradient(135deg, #F59E0B, #D97706); color: white; padding: 25px 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .header h1 { margin: 0 0 5px; font-size: 22px; }
        .header p { margin: 0; opacity: 0.9; font-size: 14px; }
        .content { background-color: #ffffff; padding: 30px; border: 1px solid #e5e7eb; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #9ca3af; background-color: #f9fafb; border: 1px solid #e5e7eb; border-top: none; border-radius: 0 0 8px 8px; }
        .items-table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        .items-table td { padding: 12px 16px; border-bottom: 1px solid #f3f4f6; }
        .items-table tr:last-child td { border-bottom: none; }
        .item-label { color: #374151; font-size: 14px; }
        .item-count { text-align: right; font-weight: 700; color: #D97706; font-size: 18px; }
        .total-row { background-color: #FFFBEB; border-radius: 6px; margin-top: 5px; }
        .total-row td { padding: 14px 16px; font-weight: 700; color: #92400E; font-size: 15px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Daily Pending Items</h1>
        <p>{{.TotalCount}} item{{if ne .TotalCount 1}}s{{end}} requiring attention</p>
    </div>
    <div class="content">
        <p>Good morning <strong>{{.UserName}}</strong>,</p>
        <p>The following items at <strong>{{.OrgName}}</strong> have been pending for over 12 hours and require your attention:</p>

        <table class="items-table">
            {{range .Items}}
            <tr>
                <td class="item-label">{{.Label}}</td>
                <td class="item-count">{{.Count}}</td>
            </tr>
            {{end}}
        </table>

        <table class="items-table total-row">
            <tr>
                <td>Total Pending Items</td>
                <td class="item-count" style="color: #92400E;">{{.TotalCount}}</td>
            </tr>
        </table>

        <p style="margin-top: 25px;">Please log in to the system to review and take action on these items.</p>
        {{if brandSignOff}}<p>{{brandSignOff}}</p>{{else}}<p>Best regards,<br>{{.OrgName}} HR System</p>{{end}}
    </div>
    <div class="footer">
        {{if brandFooter}}<p>{{brandFooter}}</p>{{end}}
        <p>This is an automated daily digest. You can manage your alert preferences in the system settings.</p>
        <p>&copy; {{.OrgName}}. All rights reserved.</p>
    </div>
</body>
</html>
`
