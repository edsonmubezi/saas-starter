package export

import (
	"time"
)

type ColumnMapping struct {
	FieldName string
	Header    string
}

type ExportOptions struct {
	SignKey    string
	Columns   []ColumnMapping
	Lock      bool
	HeaderRows [][]string
	Branding  *DocumentTemplate // Optional: use branding colors for header styling
}

type DocProperties struct {
	Category       string
	ContentStatus  string
	Created        time.Time
	Creator        string
	Description    string
	Identifier     string
	Keywords       string
	LastModifiedBy string
	Modified       time.Time
	Revision       string
	Subject        string
	Title          string
	Version        string
}

type PdfColumnMapping struct {
	Header    string
	FieldName string
}
