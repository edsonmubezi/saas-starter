package organization

import (
	"time"
)

type Organization struct {
	ID                 int64      `json:"id" secure:"encrypt_id"`
	Name               string     `json:"name" validate:"required,min=3,max=100"`
	PhoneNumber        string     `json:"phone_number" validate:"required"`
	Address            string     `json:"address" validate:"required,min=6"`
	ContactPerson      string     `json:"contact_person" validate:"required"`
	LogoURL            *string    `json:"logo_url,omitempty"`
	Email              *string    `json:"email,omitempty" validate:"omitempty,email"`
	TIN                *string    `json:"tin,omitempty"`
	RegistrationNumber *string    `json:"registration_number,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

type DocumentBrandingSettings struct {
	ID                int64  `json:"id"`
	OrganizationID    int64  `json:"organization_id"`
	ShowLogo          bool   `json:"show_logo"`
	ShowOrgName       bool   `json:"show_org_name"`
	ShowAddress       bool   `json:"show_address"`
	ShowContact       bool   `json:"show_contact"`
	ShowTIN           bool   `json:"show_tin"`
	ShowRegNumber     bool   `json:"show_reg_number"`
	ShowFooter        bool   `json:"show_footer"`
	FooterText        string `json:"footer_text"`
	ShowPageNumbers   bool   `json:"show_page_numbers"`
	ShowGeneratedDate bool   `json:"show_generated_date"`
	ShowWatermark     bool   `json:"show_watermark"`
	WatermarkText     string `json:"watermark_text"`
	WatermarkType     string `json:"watermark_type"`       // "text" or "image"
	WatermarkImagePath string `json:"watermark_image_path"`
	PrimaryColor      string `json:"primary_color"`
	HeaderTextColor   string `json:"header_text_color"`
	FontFamily        string `json:"font_family"`
	// Header info overrides (empty = use org's actual values)
	HeaderOrgName string `json:"header_org_name"`
	HeaderAddress string `json:"header_address"`
	HeaderPhone   string `json:"header_phone"`
	HeaderEmail   string `json:"header_email"`
	HeaderTIN     string `json:"header_tin"`
	FooterOrgName string `json:"footer_org_name"`
}

type EmailBrandingSettings struct {
	ID              int64  `json:"id"`
	OrganizationID  int64  `json:"organization_id"`
	PrimaryColor    string `json:"primary_color"`
	HeaderTextColor string `json:"header_text_color"`
	AccentColor     string `json:"accent_color"`
	FontFamily      string `json:"font_family"`
	ShowLogo        bool   `json:"show_logo"`
	FooterText      string `json:"footer_text"`
	SignOff         string `json:"sign_off"`
}

func DefaultEmailBrandingSettings(orgID int64) *EmailBrandingSettings {
	return &EmailBrandingSettings{
		OrganizationID:  orgID,
		PrimaryColor:    "#4F46E5",
		HeaderTextColor: "#FFFFFF",
		AccentColor:     "#4F46E5",
		FontFamily:      "Arial, sans-serif",
		ShowLogo:        true,
		FooterText:      "",
		SignOff:         "",
	}
}

func DefaultDocumentBrandingSettings(orgID int64) *DocumentBrandingSettings {
	return &DocumentBrandingSettings{
		OrganizationID:    orgID,
		ShowLogo:          true,
		ShowOrgName:       true,
		ShowAddress:       true,
		ShowContact:       true,
		ShowTIN:           false,
		ShowRegNumber:     false,
		ShowFooter:        true,
		FooterText:        "",
		ShowPageNumbers:   true,
		ShowGeneratedDate: true,
		ShowWatermark:     false,
		WatermarkText:     "",
		WatermarkType:     "text",
		WatermarkImagePath: "",
		PrimaryColor:      "#1a365d",
		HeaderTextColor:   "#FFFFFF",
		FontFamily:        "Arial",
	}
}
