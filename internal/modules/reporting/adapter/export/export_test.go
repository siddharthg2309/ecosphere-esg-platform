package export

import (
	"strings"
	"testing"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/reporting/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

func sampleReport() *domain.Report {
	return &domain.Report{
		ID:          id.New(),
		Type:        domain.TypeESGSummary,
		GeneratedAt: time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
		Sections: []domain.Section{
			{
				Title:   "Scores",
				Summary: "Overall 75",
				Rows:    []map[string]string{{"department": "Ops", "total": "80"}},
			},
			{
				Title:   "Executive summary (AI-generated)",
				Summary: "Stable performance.",
				AI:      true,
			},
		},
	}
}

func TestExportCSV(t *testing.T) {
	b, err := New().ExportCSV(sampleReport())
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "esg_summary") || !strings.Contains(s, "Ops") {
		t.Fatalf("unexpected csv: %s", s)
	}
}

func TestExportXLSX(t *testing.T) {
	b, err := New().ExportXLSX(sampleReport())
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 100 {
		t.Fatalf("xlsx too small: %d", len(b))
	}
	// PK zip header
	if b[0] != 'P' || b[1] != 'K' {
		t.Fatalf("not a zip/xlsx: %v", b[:4])
	}
}

func TestExportPDF(t *testing.T) {
	b, err := New().ExportPDF(sampleReport())
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 100 || !strings.HasPrefix(string(b), "%PDF") {
		t.Fatalf("not a pdf: len=%d head=%q", len(b), string(b[:min(20, len(b))]))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
