package export

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/edsonmubezi/myapp/internal/organization"
	"github.com/jung-kurt/gofpdf"
)

// hexToRGB converts a hex color string (#RRGGBB) to RGB values
func hexToRGB(hex string) (int, int, int) {
	if len(hex) != 7 || hex[0] != '#' {
		return 200, 200, 200 // fallback gray
	}
	r, _ := strconv.ParseInt(hex[1:3], 16, 64)
	g, _ := strconv.ParseInt(hex[3:5], 16, 64)
	b, _ := strconv.ParseInt(hex[5:7], 16, 64)
	return int(r), int(g), int(b)
}

// Fpdf is a type alias for gofpdf.Fpdf to make it accessible from handlers
type Fpdf = gofpdf.Fpdf

// DocumentTemplate holds organization branding information for PDF generation
type DocumentTemplate struct {
	OrgID              int64
	OrgName            string
	OrgAddress         string
	OrgPhone           string
	OrgEmail           string
	OrgTIN             string
	OrgRegistrationNum string
	LogoPath           string // Path to logo in storage or local filesystem

	// Document metadata
	Title       string
	Author      string
	Subject     string
	Creator     string
	Producer    string
	Keywords    []string

	// Layout settings
	ShowLogo      bool
	ShowWatermark bool
	ShowFooter    bool

	// Branding settings (loaded from document_branding_settings)
	ShowOrgName       bool
	ShowAddress       bool
	ShowContact       bool
	ShowTIN           bool
	ShowRegNumber     bool
	FooterText        string
	ShowPageNumbers   bool
	ShowGeneratedDate bool
	WatermarkText      string
	WatermarkType      string // "text" or "image"
	WatermarkImagePath string
	PrimaryColor       string // Hex color for separator line / Excel header bg
	HeaderTextColor   string // Hex color for Excel header text
	FontFamily        string // Arial, Helvetica, Times, Courier
}

// DocumentConfig holds configuration for creating a standard PDF
type DocumentConfig struct {
	Title        string
	Orientation  string // "P" for Portrait, "L" for Landscape
	PageSize     string // "A4", "Letter", etc.
	ShowLogo     bool
	ShowWatermark bool
	ShowFooter   bool
}

// FetchOrganizationTemplate retrieves organization details and creates a DocumentTemplate
func FetchOrganizationTemplate(
	ctx context.Context,
	orgUseCase organization.OrganizationUseCase,
	orgID int64,
	config DocumentConfig,
) (*DocumentTemplate, error) {
	// Fetch organization details from database
	org, err := orgUseCase.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organization: %w", err)
	}

	// Fetch branding settings (returns defaults if none saved)
	branding, err := orgUseCase.GetDocumentBranding(ctx, orgID)
	if err != nil {
		// Fall back to defaults on error
		branding = organization.DefaultDocumentBrandingSettings(orgID)
	}

	// Apply header overrides from branding settings
	orgName := org.Name
	if branding.HeaderOrgName != "" {
		orgName = branding.HeaderOrgName
	}
	orgAddress := org.Address
	if branding.HeaderAddress != "" {
		orgAddress = branding.HeaderAddress
	}
	orgPhone := org.PhoneNumber
	if branding.HeaderPhone != "" {
		orgPhone = branding.HeaderPhone
	}
	orgEmail := safeString(org.Email)
	if branding.HeaderEmail != "" {
		orgEmail = branding.HeaderEmail
	}
	orgTIN := safeString(org.TIN)
	if branding.HeaderTIN != "" {
		orgTIN = branding.HeaderTIN
	}

	template := &DocumentTemplate{
		OrgID:      org.ID,
		OrgName:    orgName,
		OrgAddress: orgAddress,
		OrgPhone:   orgPhone,

		// Optional fields
		OrgEmail:           orgEmail,
		OrgTIN:             orgTIN,
		OrgRegistrationNum: safeString(org.RegistrationNumber),
		LogoPath:           safeString(org.LogoURL),

		// Document metadata
		Title:    config.Title,
		Author:   org.Name,
		Subject:  config.Title,
		Creator:  "HRM System",
		Producer: "HRM PDF Generator",
		Keywords: []string{"hrm", "document", "official"},

		// Layout settings from config (backward compatible)
		ShowLogo:      config.ShowLogo && branding.ShowLogo,
		ShowWatermark: config.ShowWatermark || branding.ShowWatermark,
		ShowFooter:    config.ShowFooter && branding.ShowFooter,

		// Branding settings
		ShowOrgName:       branding.ShowOrgName,
		ShowAddress:       branding.ShowAddress,
		ShowContact:       branding.ShowContact,
		ShowTIN:           branding.ShowTIN,
		ShowRegNumber:     branding.ShowRegNumber,
		FooterText:        branding.FooterText,
		ShowPageNumbers:   branding.ShowPageNumbers,
		ShowGeneratedDate: branding.ShowGeneratedDate,
		WatermarkText:      branding.WatermarkText,
		WatermarkType:      branding.WatermarkType,
		WatermarkImagePath: branding.WatermarkImagePath,
		PrimaryColor:       branding.PrimaryColor,
		HeaderTextColor:   branding.HeaderTextColor,
		FontFamily:        branding.FontFamily,
	}

	return template, nil
}

// CreateStandardPDF creates a new PDF with standard organization branding
func (t *DocumentTemplate) CreateStandardPDF(orientation, pageSize string) *gofpdf.Fpdf {
	pdf := gofpdf.New(orientation, "mm", pageSize, "")

	// Register custom font if needed (TTF from assets/fonts/)
	if t.FontFamily != "" {
		t.FontFamily = RegisterFont(pdf, t.FontFamily)
	}

	// Set document metadata
	pdf.SetTitle(t.Title, false)
	pdf.SetAuthor(t.Author, false)
	pdf.SetSubject(t.Subject, false)
	pdf.SetCreator(t.Creator, false)
	pdf.SetProducer(t.Producer, false)
	if len(t.Keywords) > 0 {
		keywords := ""
		for i, kw := range t.Keywords {
			if i > 0 {
				keywords += ","
			}
			keywords += kw
		}
		pdf.SetKeywords(keywords, false)
	}

	return pdf
}

// RenderHeader renders a standard organization header on the PDF
func (t *DocumentTemplate) RenderHeader(pdf *gofpdf.Fpdf, documentTitle string) {
	font := t.FontFamily
	if font == "" {
		font = "Arial"
	}

	// Company Logo (if available and enabled)
	logoX := 10.0
	logoY := 10.0
	logoWidth := 25.0
	logoHeight := 25.0
	textX := logoX + logoWidth + 5.0 // Start text after logo

	hasLogo := false
	if t.ShowLogo && t.LogoPath != "" {
		imageType := getImageType(t.LogoPath)
		if imageType != "" {
			pdf.ImageOptions(
				t.LogoPath,
				logoX, logoY,
				logoWidth, logoHeight,
				false,
				gofpdf.ImageOptions{ImageType: imageType, ReadDpi: true},
				0, "",
			)
			hasLogo = true
		}
	}

	if !hasLogo {
		textX = 10.0 // No logo, start text from left margin
	}

	currentY := 12.0

	// Company Name
	if t.ShowOrgName {
		pdf.SetFont(font, "B", 14)
		pdf.SetXY(textX, currentY)
		pdf.CellFormat(0, 6, t.OrgName, "", 1, "L", false, 0, "")
		currentY += 6
	}

	// Address
	if t.ShowAddress && t.OrgAddress != "" {
		pdf.SetFont(font, "", 9)
		pdf.SetXY(textX, currentY)
		pdf.CellFormat(0, 4, t.OrgAddress, "", 1, "L", false, 0, "")
		currentY += 4
	}

	// Contact information
	if t.ShowContact {
		parts := []string{}
		if t.OrgPhone != "" {
			parts = append(parts, fmt.Sprintf("Tel: %s", t.OrgPhone))
		}
		if t.OrgEmail != "" {
			parts = append(parts, fmt.Sprintf("Email: %s", t.OrgEmail))
		}
		if len(parts) > 0 {
			pdf.SetFont(font, "", 9)
			pdf.SetXY(textX, currentY)
			contactInfo := ""
			for i, p := range parts {
				if i > 0 {
					contactInfo += " | "
				}
				contactInfo += p
			}
			pdf.CellFormat(0, 4, contactInfo, "", 1, "L", false, 0, "")
			currentY += 4
		}
	}

	// TIN
	if t.ShowTIN && t.OrgTIN != "" {
		pdf.SetFont(font, "", 9)
		pdf.SetXY(textX, currentY)
		pdf.CellFormat(0, 4, fmt.Sprintf("TIN: %s", t.OrgTIN), "", 1, "L", false, 0, "")
		currentY += 4
	}

	// Registration Number
	if t.ShowRegNumber && t.OrgRegistrationNum != "" {
		pdf.SetFont(font, "", 9)
		pdf.SetXY(textX, currentY)
		pdf.CellFormat(0, 4, fmt.Sprintf("Reg: %s", t.OrgRegistrationNum), "", 1, "L", false, 0, "")
		currentY += 4
	}

	// Separator line with primary color
	separatorY := logoY + logoHeight + 2
	if currentY > separatorY {
		separatorY = currentY + 2
	}
	pdf.SetY(separatorY)
	pdf.SetLineWidth(0.5)
	r, g, b := hexToRGB(t.PrimaryColor)
	pdf.SetDrawColor(r, g, b)
	pageWidth, _ := pdf.GetPageSize()
	pdf.Line(10, pdf.GetY(), pageWidth-10, pdf.GetY())

	// Document Title (centered)
	if documentTitle != "" {
		pdf.SetY(pdf.GetY() + 3)
		pdf.SetFont(font, "B", 14)
		pdf.CellFormat(0, 6, documentTitle, "", 1, "C", false, 0, "")
		pdf.Ln(2)
	}
}

// RenderFooter renders a standard footer with page numbers and generation date
func (t *DocumentTemplate) RenderFooter(pdf *gofpdf.Fpdf, pageNum, totalPages int) {
	if !t.ShowFooter {
		return
	}

	font := t.FontFamily
	if font == "" {
		font = "Arial"
	}

	pdf.SetY(-15)
	pdf.SetFont(font, "I", 8)
	pdf.SetTextColor(128, 128, 128)

	// Left side - Custom footer text or generated date
	pdf.SetX(10)
	leftText := ""
	if t.FooterText != "" {
		leftText = t.FooterText
	}
	if t.ShowGeneratedDate {
		genText := fmt.Sprintf("Generated: %s", time.Now().Format("02-Jan-2006 15:04"))
		if leftText != "" {
			leftText += " | " + genText
		} else {
			leftText = genText
		}
	}
	if leftText != "" {
		pdf.CellFormat(0, 10, leftText, "", 0, "L", false, 0, "")
	}

	// Right side - Page numbers
	if t.ShowPageNumbers {
		pageText := fmt.Sprintf("Page %d of %d", pageNum, totalPages)
		pdf.CellFormat(0, 10, pageText, "", 0, "R", false, 0, "")
	}

	// Reset text color
	pdf.SetTextColor(0, 0, 0)
}

// RenderWatermark renders a watermark (text or image) in the center of the page
func (t *DocumentTemplate) RenderWatermark(pdf *gofpdf.Fpdf) {
	if !t.ShowWatermark {
		return
	}

	// Save current alpha
	currentAlpha, _ := pdf.GetAlpha()
	pageWidth, pageHeight := pdf.GetPageSize()
	centerX := pageWidth / 2
	centerY := pageHeight / 2

	if t.WatermarkType == "image" && t.WatermarkImagePath != "" {
		// Image watermark
		imageType := getImageType(t.WatermarkImagePath)
		if imageType != "" {
			pdf.SetAlpha(0.08, "Normal")
			wmWidth := 80.0
			wmHeight := 80.0
			pdf.ImageOptions(
				t.WatermarkImagePath,
				centerX-wmWidth/2, centerY-wmHeight/2,
				wmWidth, wmHeight,
				false,
				gofpdf.ImageOptions{ImageType: imageType, ReadDpi: true},
				0, "",
			)
			pdf.SetAlpha(currentAlpha, "Normal")
		}
	} else {
		// Text watermark
		font := t.FontFamily
		if font == "" {
			font = "Arial"
		}
		wmText := t.WatermarkText
		if wmText == "" {
			wmText = t.OrgName
		}

		pdf.SetAlpha(0.1, "Normal")
		pdf.SetFont(font, "B", 60)
		pdf.SetTextColor(200, 200, 200)

		pdf.SetXY(centerX-50, centerY-10)
		pdf.TransformBegin()
		pdf.TransformRotate(45, centerX, centerY)
		pdf.CellFormat(100, 20, wmText, "", 0, "C", false, 0, "")
		pdf.TransformEnd()

		pdf.SetAlpha(currentAlpha, "Normal")
		pdf.SetFont(font, "", 12)
		pdf.SetTextColor(0, 0, 0)
	}
}

// RenderDocumentInfo renders a section with document metadata (optional)
func (t *DocumentTemplate) RenderDocumentInfo(pdf *gofpdf.Fpdf, info map[string]string) {
	if len(info) == 0 {
		return
	}

	font := t.FontFamily
	if font == "" {
		font = "Arial"
	}

	pdf.SetFont(font, "B", 11)
	pdf.CellFormat(0, 6, "Document Information", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	pdf.SetFont(font, "", 10)
	for key, value := range info {
		pdf.CellFormat(50, 5, key+":", "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 5, value, "", 1, "L", false, 0, "")
	}
	pdf.Ln(5)
}

// SafeString is a helper function to safely dereference string pointers
func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// safeString is the internal version (lowercase for backward compatibility)
func safeString(s *string) string {
	return SafeString(s)
}

// Helper function to determine image type from file path
func getImageType(path string) string {
	if path == "" || len(path) < 4 {
		return ""
	}

	// Check file extension (last 4 characters)
	ext := path[len(path)-4:]
	switch ext {
	case ".png", ".PNG":
		return "PNG"
	case ".jpg", ".JPG":
		return "JPG"
	case ".gif", ".GIF":
		return "GIF"
	case "jpeg", "JPEG":
		return "JPG"
	default:
		// Try to detect from last 3 characters
		if len(path) > 3 {
			ext3 := path[len(path)-3:]
			switch ext3 {
			case "png", "PNG":
				return "PNG"
			case "jpg", "JPG":
				return "JPG"
			case "gif", "GIF":
				return "GIF"
			}
		}
		return ""
	}
}

// CreateStandardDocumentPDF creates a complete PDF document with organization branding
// This is a convenience function that combines all the template features
func CreateStandardDocumentPDF(
	ctx context.Context,
	orgUseCase organization.OrganizationUseCase,
	orgID int64,
	config DocumentConfig,
	contentRenderer func(pdf *gofpdf.Fpdf, template *DocumentTemplate) error,
) (*gofpdf.Fpdf, error) {
	// Fetch organization template
	template, err := FetchOrganizationTemplate(ctx, orgUseCase, orgID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch template: %w", err)
	}

	// Create PDF
	pdf := template.CreateStandardPDF(config.Orientation, config.PageSize)
	pdf.AddPage()

	// Render watermark (bottom layer)
	if template.ShowWatermark {
		template.RenderWatermark(pdf)
	}

	// Render header
	template.RenderHeader(pdf, config.Title)

	// Render content (provided by caller)
	if err := contentRenderer(pdf, template); err != nil {
		return nil, fmt.Errorf("failed to render content: %w", err)
	}

	// Note: Footer is typically added via PDF header/footer callback
	// For multi-page documents, use SetFooterFunc

	return pdf, nil
}
