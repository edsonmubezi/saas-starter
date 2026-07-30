package excel

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// ImportErrorRow represents a single import error for Excel export
type ImportErrorRow struct {
	Row          int
	EmployeeName string
	Column       string
	Value        string
	Error        string
}

// GenerateImportErrorsExcel creates a styled Excel file from import errors
func GenerateImportErrorsExcel(rows []ImportErrorRow) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Import Errors"
	f.SetSheetName("Sheet1", sheet)

	// Header style - red background, white text
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"C0392B"}},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: "000000", Style: 2},
		},
	})

	// Data style
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 11},
		Alignment: &excelize.Alignment{
			Vertical: "center",
			WrapText: true,
		},
	})

	// Error cell style (red text)
	errorStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Color: "C0392B"},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})

	// Headers
	headers := []string{"Row #", "Employee Name", "Column", "Value Provided", "Error"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// Column widths
	f.SetColWidth(sheet, "A", "A", 8)
	f.SetColWidth(sheet, "B", "B", 30)
	f.SetColWidth(sheet, "C", "C", 25)
	f.SetColWidth(sheet, "D", "D", 25)
	f.SetColWidth(sheet, "E", "E", 50)

	// Data rows
	for i, row := range rows {
		r := i + 2
		cellA, _ := excelize.CoordinatesToCellName(1, r)
		cellB, _ := excelize.CoordinatesToCellName(2, r)
		cellC, _ := excelize.CoordinatesToCellName(3, r)
		cellD, _ := excelize.CoordinatesToCellName(4, r)
		cellE, _ := excelize.CoordinatesToCellName(5, r)

		f.SetCellValue(sheet, cellA, row.Row)
		f.SetCellValue(sheet, cellB, row.EmployeeName)
		f.SetCellValue(sheet, cellC, row.Column)
		f.SetCellValue(sheet, cellD, row.Value)
		f.SetCellValue(sheet, cellE, row.Error)

		f.SetCellStyle(sheet, cellA, cellD, dataStyle)
		f.SetCellStyle(sheet, cellE, cellE, errorStyle)
	}

	// Freeze header row
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("write excel: %w", err)
	}
	return buf.Bytes(), nil
}
