package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

func main() {
	// Input and output paths
	inputPath := "docs/ARCHITECTURE_REPORT.md"
	outputPath := "docs/Microfinance_Architecture_Report.pdf"

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	// Add fonts
	pdf.AddPage()

	// Title page
	createTitlePage(pdf)

	// Read markdown file
	content, err := readMarkdownFile(inputPath)
	if err != nil {
		log.Fatalf("Error reading markdown file: %v", err)
	}

	// Process markdown content
	processMarkdownContent(pdf, content)

	// Save PDF
	err = pdf.OutputFileAndClose(outputPath)
	if err != nil {
		log.Fatalf("Error creating PDF: %v", err)
	}

	fmt.Printf("✅ Architecture report generated: %s\n", outputPath)
}

func createTitlePage(pdf *gofpdf.Fpdf) {
	// Title
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(41, 128, 185) // Blue color
	pdf.CellFormat(0, 15, "Microfinance", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(0, 0, 0) // Black
	pdf.CellFormat(0, 10, "System Architecture Report", "", 1, "C", false, 0, "")

	pdf.Ln(10)

	// Subtitle
	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(100, 100, 100) // Gray
	pdf.CellFormat(0, 8, "SaaS Human Resource Management Platform", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 8, "Version 1.0.1", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 8, "September 2025", "", 1, "C", false, 0, "")

	pdf.Ln(20)

	// Add architecture diagram placeholder
	pdf.SetFont("Arial", "I", 10)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(0, 6, "Built with Go • PostgreSQL • Clean Architecture", "", 1, "C", false, 0, "")

	pdf.AddPage()
}

func readMarkdownFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func processMarkdownContent(pdf *gofpdf.Fpdf, lines []string) {
	inCodeBlock := false
	inTable := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines in code blocks
		if inCodeBlock && line == "" {
			continue
		}

		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				pdf.Ln(3)
				pdf.SetFillColor(245, 245, 245)
				pdf.SetFont("Courier", "", 9)
			} else {
				pdf.Ln(3)
				pdf.SetFont("Arial", "", 11)
			}
			continue
		}

		if inCodeBlock {
			addCodeLine(pdf, line)
			continue
		}

		// Handle table detection
		if strings.Contains(line, "|") && !strings.HasPrefix(line, "#") {
			if !inTable {
				inTable = true
				pdf.Ln(3)
			}
			addTableRow(pdf, line)
			continue
		} else if inTable {
			inTable = false
			pdf.Ln(3)
		}

		// Handle headers
		if strings.HasPrefix(line, "# ") {
			addHeader1(pdf, strings.TrimPrefix(line, "# "))
		} else if strings.HasPrefix(line, "## ") {
			addHeader2(pdf, strings.TrimPrefix(line, "## "))
		} else if strings.HasPrefix(line, "### ") {
			addHeader3(pdf, strings.TrimPrefix(line, "### "))
		} else if strings.HasPrefix(line, "#### ") {
			addHeader4(pdf, strings.TrimPrefix(line, "#### "))
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			// Handle bullet points
			addBulletPoint(pdf, strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
		} else if line == "---" {
			// Handle horizontal rules
			addHorizontalRule(pdf)
		} else if line != "" {
			// Handle regular text
			addParagraph(pdf, line)
		} else {
			// Empty line
			pdf.Ln(2)
		}
	}
}

func addHeader1(pdf *gofpdf.Fpdf, text string) {
	pdf.Ln(8)
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(41, 128, 185) // Blue
	pdf.CellFormat(0, 10, text, "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0) // Reset to black
	pdf.Ln(3)
}

func addHeader2(pdf *gofpdf.Fpdf, text string) {
	pdf.Ln(6)
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(52, 152, 219) // Light blue
	pdf.CellFormat(0, 8, text, "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0) // Reset to black
	pdf.Ln(2)
}

func addHeader3(pdf *gofpdf.Fpdf, text string) {
	pdf.Ln(4)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(34, 34, 34) // Dark gray
	pdf.CellFormat(0, 7, text, "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0) // Reset to black
	pdf.Ln(1)
}

func addHeader4(pdf *gofpdf.Fpdf, text string) {
	pdf.Ln(3)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 6, text, "", 1, "L", false, 0, "")
	pdf.Ln(1)
}

func addParagraph(pdf *gofpdf.Fpdf, text string) {
	pdf.SetFont("Arial", "", 10)

	// Handle bold text
	if strings.Contains(text, "**") {
		parts := strings.Split(text, "**")
		x, y := pdf.GetXY()
		for i, part := range parts {
			if i%2 == 0 {
				pdf.SetFont("Arial", "", 10)
			} else {
				pdf.SetFont("Arial", "B", 10)
			}
			w := pdf.GetStringWidth(part)
			pdf.CellFormat(w, 5, part, "", 0, "L", false, 0, "")
			x += w
			pdf.SetXY(x, y)
		}
		pdf.Ln(5)
	} else {
		// Wrap long lines
		pageWidth, _ := pdf.GetPageSize()
		margins := 40.0 // Left + right margins
		maxWidth := pageWidth - margins

		pdf.SetFont("Arial", "", 10)
		words := strings.Fields(text)
		currentLine := ""

		for _, word := range words {
			testLine := currentLine
			if testLine != "" {
				testLine += " "
			}
			testLine += word

			if pdf.GetStringWidth(testLine) > maxWidth && currentLine != "" {
				pdf.CellFormat(0, 5, currentLine, "", 1, "L", false, 0, "")
				currentLine = word
			} else {
				currentLine = testLine
			}
		}

		if currentLine != "" {
			pdf.CellFormat(0, 5, currentLine, "", 1, "L", false, 0, "")
		}
	}
	pdf.Ln(1)
}

func addBulletPoint(pdf *gofpdf.Fpdf, text string) {
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(10, 5, "•", "", 0, "L", false, 0, "")

	// Wrap bullet point text
	pageWidth, _ := pdf.GetPageSize()
	margins := 50.0 // Left + right margins + bullet space
	maxWidth := pageWidth - margins

	words := strings.Fields(text)
	currentLine := ""

	for i, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if pdf.GetStringWidth(testLine) > maxWidth && currentLine != "" {
			if i == 0 {
				pdf.CellFormat(0, 5, currentLine, "", 1, "L", false, 0, "")
			} else {
				pdf.CellFormat(0, 5, currentLine, "", 1, "L", false, 0, "")
				pdf.CellFormat(10, 5, "", "", 0, "L", false, 0, "") // Indent
			}
			currentLine = word
		} else {
			currentLine = testLine
		}
	}

	if currentLine != "" {
		pdf.CellFormat(0, 5, currentLine, "", 1, "L", false, 0, "")
	}
	pdf.Ln(1)
}

func addCodeLine(pdf *gofpdf.Fpdf, text string) {
	pdf.SetFont("Courier", "", 8)
	pdf.SetFillColor(245, 245, 245)
	pdf.CellFormat(0, 4, text, "", 1, "L", true, 0, "")
}

func addTableRow(pdf *gofpdf.Fpdf, line string) {
	if strings.Contains(line, "---") {
		return // Skip separator rows
	}

	cells := strings.Split(line, "|")
	pdf.SetFont("Arial", "", 9)

	// Calculate cell width
	pageWidth, _ := pdf.GetPageSize()
	margins := 40.0
	availableWidth := pageWidth - margins
	cellWidth := availableWidth / float64(len(cells)-2) // Exclude empty first and last

	for i := 1; i < len(cells)-1; i++ {
		cell := strings.TrimSpace(cells[i])
		if strings.Contains(cell, "**") {
			cell = strings.ReplaceAll(cell, "**", "")
			pdf.SetFont("Arial", "B", 9)
		} else {
			pdf.SetFont("Arial", "", 9)
		}

		pdf.CellFormat(cellWidth, 6, cell, "1", 0, "L", false, 0, "")
	}
	pdf.Ln(6)
}

func addHorizontalRule(pdf *gofpdf.Fpdf) {
	pdf.Ln(3)
	x, y := pdf.GetXY()
	pageWidth, _ := pdf.GetPageSize()
	pdf.Line(x, y, pageWidth-20, y)
	pdf.Ln(3)
}