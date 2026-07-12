package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/reporting/domain"
	"github.com/xuri/excelize/v2"
)

type Service struct{}

func New() *Service { return &Service{} }

func (s *Service) ExportCSV(r *domain.Report) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"report_type", string(r.Type)})
	_ = w.Write([]string{"generated_at", r.GeneratedAt.Format(timeRFC3339)})
	_ = w.Write([]string{})
	for _, sec := range r.Sections {
		_ = w.Write([]string{"section", sec.Title})
		if sec.Summary != "" {
			_ = w.Write([]string{"summary", sec.Summary})
		}
		if len(sec.Rows) > 0 {
			// header from keys of first row
			keys := make([]string, 0, len(sec.Rows[0]))
			for k := range sec.Rows[0] {
				keys = append(keys, k)
			}
			_ = w.Write(keys)
			for _, row := range sec.Rows {
				vals := make([]string, len(keys))
				for i, k := range keys {
					vals[i] = row[k]
				}
				_ = w.Write(vals)
			}
		}
		_ = w.Write([]string{})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"

func (s *Service) ExportXLSX(r *domain.Report) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Report"
	_ = f.SetSheetName("Sheet1", sheet)
	_ = f.SetCellValue(sheet, "A1", "EcoSphere ESG Report")
	_ = f.SetCellValue(sheet, "A2", string(r.Type))
	_ = f.SetCellValue(sheet, "A3", r.GeneratedAt.Format(timeRFC3339))
	row := 5
	for _, sec := range r.Sections {
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), sec.Title)
		row++
		if sec.Summary != "" {
			_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), sec.Summary)
			row++
		}
		if len(sec.Rows) > 0 {
			keys := make([]string, 0, len(sec.Rows[0]))
			for k := range sec.Rows[0] {
				keys = append(keys, k)
			}
			for i, k := range keys {
				cell, _ := excelize.CoordinatesToCellName(i+1, row)
				_ = f.SetCellValue(sheet, cell, k)
			}
			row++
			for _, data := range sec.Rows {
				for i, k := range keys {
					cell, _ := excelize.CoordinatesToCellName(i+1, row)
					_ = f.SetCellValue(sheet, cell, data[k])
				}
				row++
			}
		}
		row++
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *Service) ExportPDF(r *domain.Report) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("EcoSphere ESG Report", false)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "EcoSphere ESG Report")
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(40, 8, "Type: "+string(r.Type))
	pdf.Ln(8)
	pdf.Cell(40, 8, "Generated: "+r.GeneratedAt.Format("2006-01-02 15:04"))
	pdf.Ln(12)
	for _, sec := range r.Sections {
		pdf.SetFont("Arial", "B", 13)
		pdf.MultiCell(0, 7, sec.Title, "", "", false)
		pdf.SetFont("Arial", "", 10)
		if sec.AI {
			pdf.SetTextColor(113, 75, 103)
			pdf.MultiCell(0, 5, "[AI-generated — advisory only]", "", "", false)
			pdf.SetTextColor(0, 0, 0)
		}
		if sec.Summary != "" {
			pdf.MultiCell(0, 5, sec.Summary, "", "", false)
		}
		for _, row := range sec.Rows {
			parts := make([]string, 0, len(row))
			for k, v := range row {
				parts = append(parts, k+": "+v)
			}
			pdf.MultiCell(0, 5, strings.Join(parts, "  |  "), "", "", false)
		}
		pdf.Ln(4)
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
