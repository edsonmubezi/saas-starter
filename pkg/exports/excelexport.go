package export

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	excelize "github.com/xuri/excelize/v2"
)

// Column struct (assuming it's defined elsewhere)
type Column struct {
	FieldName string
	Header    string
}

// ExportProtectedToExcel exports data to a protected Excel file with a signature.
// It requires a slice of structs to reflectively extract data.
func ExportProtectedToExcel(
	w http.ResponseWriter,
	filename, sheetName string,
	data interface{},
	opts ExportOptions,
) error {
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return fmt.Errorf("data must be a slice")
	}
	if val.Len() == 0 {
		return fmt.Errorf("no data to export")
	}

	f := excelize.NewFile()

	index, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create sheet: %w", err)
	}

	// Remove default Sheet1
	f.DeleteSheet("Sheet1")

	rowOffset := 0

	// Determine branding colors
	fillColor := "#4F81BD"
	textColor := "#FFFFFF"
	if opts.Branding != nil {
		if opts.Branding.PrimaryColor != "" {
			fillColor = opts.Branding.PrimaryColor
		}
		if opts.Branding.HeaderTextColor != "" {
			textColor = opts.Branding.HeaderTextColor
		}
	}

	styleJSON := fmt.Sprintf(`{
		"font": {
			"bold": true,
			"color": "%s",
			"size": 12
		},
		"fill": {
			"type": "pattern",
			"color": ["%s"],
			"pattern": 1
		},
		"alignment": {
			"horizontal": "center",
			"vertical": "center"
		},
		"border": [
			{"type": "left", "color": "#000000", "style": 1},
			{"type": "right", "color": "#000000", "style": 1},
			{"type": "top", "color": "#000000", "style": 1},
			{"type": "bottom", "color": "#000000", "style": 1}
		]
	}`, textColor, fillColor)

	var headerStyle excelize.Style
	if err := json.Unmarshal([]byte(styleJSON), &headerStyle); err != nil {
		return fmt.Errorf("failed to unmarshal style JSON: %w", err)
	}

	headerStyleID, err := f.NewStyle(&headerStyle)
	if err != nil {
		return fmt.Errorf("failed to create header style: %w", err)
	}

	if len(opts.HeaderRows) > 0 {
		// Get the number of columns for merging
		numCols := len(opts.Columns)
		if numCols == 0 {
			numCols = 1
		}
		lastColName, _ := excelize.ColumnNumberToName(numCols)

		for i, row := range opts.HeaderRows {
			// Set the value in the first cell
			if len(row) > 0 && row[0] != "" {
				cell, _ := excelize.CoordinatesToCellName(1, i+1)
				f.SetCellValue(sheetName, cell, row[0])
				f.SetCellStyle(sheetName, cell, cell, headerStyleID)

				// Merge cells across all columns for header rows
				endCell := fmt.Sprintf("%s%d", lastColName, i+1)
				f.MergeCell(sheetName, cell, endCell)
			}
		}
		rowOffset = len(opts.HeaderRows)
	}

	// Write column headers with style
	for i, col := range opts.Columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, rowOffset+1)
		f.SetCellValue(sheetName, cell, col.Header)
		f.SetCellStyle(sheetName, cell, cell, headerStyleID)
	}

	// Write data rows
	for row := 0; row < val.Len(); row++ {
		item := val.Index(row)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}
		for colIdx, col := range opts.Columns {
			field := item.FieldByName(col.FieldName)
			if !field.IsValid() {
				continue
			}
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row+rowOffset+2)
			f.SetCellValue(sheetName, cell, field.Interface())
		}
	}

	// Autofit columns based on column headers and data only (not merged header rows)
	maxWidths := make([]int, len(opts.Columns))

	// Check column headers widths
	for i, col := range opts.Columns {
		if len(col.Header) > maxWidths[i] {
			maxWidths[i] = len(col.Header)
		}
	}

	// Check data rows widths
	for row := 0; row < val.Len(); row++ {
		item := val.Index(row)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}
		for colIdx, col := range opts.Columns {
			field := item.FieldByName(col.FieldName)
			if !field.IsValid() {
				continue
			}
			cellValue := fmt.Sprintf("%v", field.Interface())
			if len(cellValue) > maxWidths[colIdx] {
				maxWidths[colIdx] = len(cellValue)
			}
		}
	}

	// Set column widths with reasonable minimums
	for i, width := range maxWidths {
		colLetter, _ := excelize.ColumnNumberToName(i + 1)
		// Set minimum width of 6 for index column (#), otherwise use calculated width + padding
		colWidth := float64(width) + 2
		if i == 0 && width < 4 {
			colWidth = 6 // Minimum width for # column
		}
		f.SetColWidth(sheetName, colLetter, colLetter, colWidth)
	}

	// Freeze panes below header rows
	err = f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      rowOffset + 1,
		TopLeftCell: fmt.Sprintf("A%d", rowOffset+2),
		ActivePane:  "bottomLeft",
	})
	if err != nil {
		return fmt.Errorf("failed to set panes: %w", err)
	}

	// Lock sheet if requested
	if opts.Lock {
		protect := excelize.SheetProtectionOptions{
			SelectLockedCells:   true,
			SelectUnlockedCells: true,
			Password:            "secret123",
		}
		if err := f.ProtectSheet(sheetName, &protect); err != nil {
			return fmt.Errorf("failed to lock sheet: %w", err)
		}
	}

	// Compute signature (optional)
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return fmt.Errorf("failed to write excel to buffer: %w", err)
	}
	content := buf.Bytes()

	if opts.SignKey != "" {
		fileHash := sha256.Sum256(content)
		combined := append(fileHash[:], []byte(opts.SignKey)...)
		sigHash := sha256.Sum256(combined)
		signature := hex.EncodeToString(sigHash[:])

		f.SetDocProps(&excelize.DocProperties{
			Title:       filename,
			Creator:     "HRM-System",
			Subject:     "Secure Export",
			Keywords:    "export,secure",
			Description: "Signature:" + signature,
		})
	}

	f.SetActiveSheet(index)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.xlsx"`, filename))

	return f.Write(w)
}

func VerifyExcelSignature(filePath, signKey string) (bool, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open excel: %w", err)
	}
	props, err := f.GetDocProps()
	if err != nil {
		return false, fmt.Errorf("failed to read metadata: %w", err)
	}

	// Extract signature stored in description
	var storedSignature string
	if strings.HasPrefix(props.Description, "Signature:") {
		storedSignature = strings.TrimPrefix(props.Description, "Signature:")
	} else {
		return false, fmt.Errorf("no signature found in metadata")
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	fileHash := sha256.Sum256(content)
	combined := append(fileHash[:], []byte(signKey)...)
	sigHash := sha256.Sum256(combined)
	expectedSignature := hex.EncodeToString(sigHash[:])

	return storedSignature == expectedSignature, nil
}
