package export

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/edsonmubezi/myapp/internal/organization"
	"github.com/jung-kurt/gofpdf"
)

// PasswordResetFormData holds the data needed for generating a password reset form
type PasswordResetFormData struct {
	UserFullName     string
	UserEmail        string
	UserID           string
	EmployeeCode     string
	Department       string
	Position         string
	RequestDate      time.Time
	RequestedBy      string
	RequestedByRole  string
	TemporaryPassword string
	ExpiresAt        time.Time
	Notes            string
	FormReference    string
}

// GeneratePasswordResetForm creates a PDF password reset form
func GeneratePasswordResetForm(
	ctx context.Context,
	orgUseCase organization.OrganizationUseCase,
	orgID int64,
	data PasswordResetFormData,
) ([]byte, error) {
	// Fetch organization template
	config := DocumentConfig{
		Title:       "Password Reset Form",
		Orientation: "P",
		PageSize:    "A4",
		ShowLogo:    true,
		ShowFooter:  true,
	}

	template, err := FetchOrganizationTemplate(ctx, orgUseCase, orgID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch template: %w", err)
	}

	// Create PDF
	pdf := template.CreateStandardPDF("P", "A4")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Render header
	template.RenderHeader(pdf, "PASSWORD RESET AUTHORIZATION FORM")

	// Set starting Y position after header
	pdf.SetY(50)

	// Form Reference Number
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(0, 8, fmt.Sprintf("Form Reference: %s", data.FormReference), "1", 1, "L", true, 0, "")
	pdf.Ln(5)

	// Section 1: Employee Information
	renderSectionTitle(pdf, "SECTION A: EMPLOYEE INFORMATION")
	pdf.Ln(2)

	// Employee details table
	renderFormRow(pdf, "Full Name", data.UserFullName)
	renderFormRow(pdf, "Email Address", data.UserEmail)
	renderFormRow(pdf, "Employee ID/Code", data.EmployeeCode)
	renderFormRow(pdf, "Department", data.Department)
	renderFormRow(pdf, "Position", data.Position)
	pdf.Ln(5)

	// Section 2: Reset Request Details
	renderSectionTitle(pdf, "SECTION B: RESET REQUEST DETAILS")
	pdf.Ln(2)

	renderFormRow(pdf, "Request Date", data.RequestDate.Format("02-Jan-2006 15:04"))
	renderFormRow(pdf, "Requested By", data.RequestedBy)
	renderFormRow(pdf, "Requester Role", data.RequestedByRole)
	if data.TemporaryPassword != "" {
		renderFormRow(pdf, "Temporary Password", data.TemporaryPassword)
		renderFormRow(pdf, "Password Expires", data.ExpiresAt.Format("02-Jan-2006 15:04"))
	}
	if data.Notes != "" {
		renderFormRow(pdf, "Notes", data.Notes)
	}
	pdf.Ln(5)

	// Section 3: Instructions
	renderSectionTitle(pdf, "SECTION C: INSTRUCTIONS FOR EMPLOYEE")
	pdf.Ln(2)

	pdf.SetFont("Arial", "", 10)
	instructions := []string{
		"1. Use the temporary password provided above to login to the system.",
		"2. You will be required to change your password immediately upon login.",
		"3. Your new password must meet the following requirements:",
		"   - Minimum 8 characters long",
		"   - At least one uppercase letter",
		"   - At least one lowercase letter",
		"   - At least one number",
		"   - At least one special character (!@#$%^&*)",
		"4. Do not share your password with anyone.",
		"5. If you did not request this password reset, contact IT immediately.",
	}

	for _, instruction := range instructions {
		pdf.CellFormat(0, 6, instruction, "", 1, "L", false, 0, "")
	}
	pdf.Ln(5)

	// Section 4: Acknowledgment
	renderSectionTitle(pdf, "SECTION D: EMPLOYEE ACKNOWLEDGMENT")
	pdf.Ln(2)

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 5, "I acknowledge receipt of this password reset authorization. I understand that I must change my password immediately upon first login and that I am responsible for maintaining the confidentiality of my login credentials.", "", "L", false)
	pdf.Ln(10)

	// Signature section
	pdf.SetFont("Arial", "", 10)

	// Left column - Employee signature
	pdf.SetX(15)
	pdf.CellFormat(85, 5, "Employee Signature:", "", 0, "L", false, 0, "")

	// Right column - Witness/Admin signature
	pdf.CellFormat(85, 5, "Authorized By:", "", 1, "L", false, 0, "")

	pdf.Ln(15)

	// Signature lines
	pdf.SetX(15)
	pdf.Line(15, pdf.GetY(), 95, pdf.GetY())
	pdf.Line(105, pdf.GetY(), 185, pdf.GetY())

	pdf.Ln(3)
	pdf.SetX(15)
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(85, 5, "(Sign above)", "", 0, "C", false, 0, "")
	pdf.CellFormat(85, 5, "(Sign above)", "", 1, "C", false, 0, "")

	pdf.Ln(5)

	// Date line
	pdf.SetFont("Arial", "", 10)
	pdf.SetX(15)
	pdf.CellFormat(85, 5, "Date: _______________________", "", 0, "L", false, 0, "")
	pdf.CellFormat(85, 5, "Date: _______________________", "", 1, "L", false, 0, "")

	// Section 5: IT Use Only
	if pdf.GetY() > 220 {
		pdf.AddPage()
	}
	pdf.Ln(10)
	renderSectionTitle(pdf, "SECTION E: FOR IT USE ONLY")
	pdf.Ln(2)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 6, "Reset Completed By: _______________________________________", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, "Date/Time: ________________________________________________", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, "Comments: _________________________________________________", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, "________________________________________________", "", 1, "L", false, 0, "")

	// Footer with confidentiality notice
	pdf.SetY(-30)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(0, 5, "CONFIDENTIAL - This document contains sensitive information. Handle appropriately.", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Generated: %s | %s", time.Now().Format("02-Jan-2006 15:04:05"), template.OrgName), "", 1, "C", false, 0, "")

	// Output to bytes
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// renderSectionTitle renders a section title with background
func renderSectionTitle(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(0, 7, title, "1", 1, "L", true, 0, "")
}

// renderFormRow renders a form row with label and value
func renderFormRow(pdf *gofpdf.Fpdf, label, value string) {
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(50, 6, label+":", "LTB", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 6, value, "RTB", 1, "L", false, 0, "")
}

// GenerateSimplePasswordResetForm creates a simpler version without organization details
// This can be used when organization details are not available
func GenerateSimplePasswordResetForm(data PasswordResetFormData, companyName string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Simple header
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, companyName, "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 8, "PASSWORD RESET AUTHORIZATION FORM", "", 1, "C", false, 0, "")

	// Separator
	pdf.SetLineWidth(0.5)
	pdf.Line(15, pdf.GetY()+2, 195, pdf.GetY()+2)
	pdf.Ln(8)

	// Form Reference
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(0, 8, fmt.Sprintf("Reference: %s", data.FormReference), "1", 1, "L", true, 0, "")
	pdf.Ln(5)

	// Employee Information
	renderSectionTitle(pdf, "EMPLOYEE INFORMATION")
	pdf.Ln(2)
	renderFormRow(pdf, "Full Name", data.UserFullName)
	renderFormRow(pdf, "Email", data.UserEmail)
	if data.EmployeeCode != "" {
		renderFormRow(pdf, "Employee Code", data.EmployeeCode)
	}
	if data.Department != "" {
		renderFormRow(pdf, "Department", data.Department)
	}
	pdf.Ln(5)

	// Request Details
	renderSectionTitle(pdf, "RESET REQUEST DETAILS")
	pdf.Ln(2)
	renderFormRow(pdf, "Request Date", data.RequestDate.Format("02-Jan-2006 15:04"))
	renderFormRow(pdf, "Requested By", data.RequestedBy)
	if data.TemporaryPassword != "" {
		renderFormRow(pdf, "Temporary Password", data.TemporaryPassword)
		renderFormRow(pdf, "Expires At", data.ExpiresAt.Format("02-Jan-2006 15:04"))
	}
	pdf.Ln(5)

	// Instructions
	renderSectionTitle(pdf, "INSTRUCTIONS")
	pdf.Ln(2)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 5, "1. Login using the temporary password provided.\n2. Change your password immediately upon login.\n3. Ensure your new password is strong and unique.\n4. Never share your password with anyone.", "", "L", false)
	pdf.Ln(5)

	// Signatures
	renderSectionTitle(pdf, "ACKNOWLEDGMENT")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(85, 5, "Employee Signature:", "", 0, "L", false, 0, "")
	pdf.CellFormat(85, 5, "Admin Signature:", "", 1, "L", false, 0, "")
	pdf.Ln(12)

	pdf.Line(15, pdf.GetY(), 95, pdf.GetY())
	pdf.Line(105, pdf.GetY(), 185, pdf.GetY())

	pdf.Ln(3)
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(85, 5, "Date: _______________", "", 0, "L", false, 0, "")
	pdf.CellFormat(85, 5, "Date: _______________", "", 1, "L", false, 0, "")

	// Footer
	pdf.SetY(-20)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(0, 5, fmt.Sprintf("Generated: %s", time.Now().Format("02-Jan-2006 15:04:05")), "", 1, "C", false, 0, "")

	// Output to bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}
