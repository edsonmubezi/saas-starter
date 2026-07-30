package export

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/edsonmubezi/myapp/pkg/customdatatype"
	"github.com/jung-kurt/gofpdf"
)

type HeaderRenderer func(pdf *gofpdf.Fpdf)

func ExportPDFWithColumns(
	w http.ResponseWriter,
	filename string,
	data interface{},
	columns []ColumnMapping,
	opts ExportOptions,
	renderHeader HeaderRenderer, //  pass custom header callback
) error {
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return fmt.Errorf("data must be a slice")
	}
	if val.Len() == 0 {
		return fmt.Errorf("no data to export")
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(filename, false)
	pdf.SetAuthor("HRM-System", false)
	pdf.SetSubject("Secure Export", false)
	pdf.SetKeywords("pdf,secure,export", false)
	pdf.SetCreator("HRM-System", false)

	pdf.SetFont("Arial", "", 12)
	pdf.AddPage()

	//  Custom header (like loan info) if provided
	if renderHeader != nil {
		renderHeader(pdf)
		pdf.Ln(10)
	}

	//  Table header
	pdf.SetFont("Arial", "B", 11)
	colWidth := 190.0 / float64(len(columns))
	for _, col := range columns {
		pdf.CellFormat(colWidth, 8, col.Header, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	//  Table rows
	pdf.SetFont("Arial", "", 10)
	for i := 0; i < val.Len(); i++ {
		item := val.Index(i)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		for _, col := range columns {
			fieldVal := item.FieldByName(col.FieldName)
			var strVal string
			if fieldVal.IsValid() {
				switch v := fieldVal.Interface().(type) {
				case time.Time:
					strVal = v.Format("02-Jan-2006") // dd-MMM-yyyy
				case customdatatype.DateOnly:
					strVal = v.ToTime().Format("02-Jan-2006")
				default:
					strVal = fmt.Sprintf("%v", v)
				}
			}
			pdf.CellFormat(colWidth, 8, strVal, "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}

	//  Signature
	signatureData := fmt.Sprintf("%v-%s-%d", data, opts.SignKey, time.Now().Unix())
	hash := sha256.Sum256([]byte(signatureData))
	signature := hex.EncodeToString(hash[:])
	pdf.SetProducer("Signature:"+signature, true)

	//  Response
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, filename))

	if err := pdf.Output(w); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	return nil
}
