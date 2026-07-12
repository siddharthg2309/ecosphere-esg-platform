package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/reporting/domain"
	scoringapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/scoring/app"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type Store interface {
	Save(ctx context.Context, r *domain.Report) error
	ByID(ctx context.Context, id id.ID) (*domain.Report, error)
}

type DataSource interface {
	EnvironmentalFigures(ctx context.Context, f domain.Filters) (domain.Section, error)
	SocialFigures(ctx context.Context, f domain.Filters) (domain.Section, error)
	GovernanceFigures(ctx context.Context, f domain.Filters) (domain.Section, error)
}

type Narrator interface {
	Summarize(ctx context.Context, figures map[string]any) (string, error)
}

type Exporter interface {
	ExportCSV(r *domain.Report) ([]byte, error)
	ExportXLSX(r *domain.Report) ([]byte, error)
	ExportPDF(r *domain.Report) ([]byte, error)
}

type Service struct {
	store   Store
	data    DataSource
	scores  *scoringapp.Service
	narrate Narrator
	export  Exporter
	now     func() time.Time
}

func New(store Store, data DataSource, scores *scoringapp.Service, narrate Narrator, export Exporter) *Service {
	return &Service{store: store, data: data, scores: scores, narrate: narrate, export: export, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Generate(ctx context.Context, typ domain.ReportType, filters domain.Filters, by *id.ID) (*domain.Report, error) {
	report := &domain.Report{ID: id.New(), Type: typ, Filters: filters, GeneratedAt: s.now(), GeneratedBy: by, Sections: []domain.Section{}}

	switch typ {
	case domain.TypeEnvironmental:
		sec, err := s.data.EnvironmentalFigures(ctx, filters)
		if err != nil {
			return nil, err
		}
		report.Sections = append(report.Sections, sec)
	case domain.TypeSocial:
		sec, err := s.data.SocialFigures(ctx, filters)
		if err != nil {
			return nil, err
		}
		report.Sections = append(report.Sections, sec)
	case domain.TypeGovernance:
		sec, err := s.data.GovernanceFigures(ctx, filters)
		if err != nil {
			return nil, err
		}
		report.Sections = append(report.Sections, sec)
	case domain.TypeESGSummary, domain.TypeCustom:
		overall, depts, weights, err := s.scores.Overall(ctx, "")
		if err != nil {
			return nil, err
		}
		envSec, _ := s.data.EnvironmentalFigures(ctx, filters)
		socSec, _ := s.data.SocialFigures(ctx, filters)
		govSec, _ := s.data.GovernanceFigures(ctx, filters)
		metrics := map[string]any{
			"overall": overall, "weights": weights,
			"departments": depts,
			"environmental": envSec.Metrics, "social": socSec.Metrics, "governance": govSec.Metrics,
		}
		report.Sections = append(report.Sections,
			domain.Section{Title: "Overall ESG Score", Metrics: map[string]any{"overall": overall, "weights": weights}},
			domain.Section{Title: "Department ranking", Rows: deptRows(depts)},
			envSec, socSec, govSec,
		)
		if s.narrate != nil && (typ == domain.TypeESGSummary || typ == domain.TypeCustom) {
			prose, err := s.narrate.Summarize(ctx, metrics)
			if err == nil && prose != "" {
				report.Sections = append([]domain.Section{{
					Title: "Executive summary (AI-generated)", Summary: prose, AI: true,
				}}, report.Sections...)
			}
		}
	default:
		return nil, fmt.Errorf("unknown report type")
	}

	if err := s.store.Save(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

func deptRows(depts interface{}) []map[string]string {
	// scores come as []domain.DepartmentScore via encoding — use type assert via json
	raw, _ := json.Marshal(depts)
	var list []struct {
		Name   string `json:"name"`
		Env    int    `json:"environmental"`
		Social int    `json:"social"`
		Gov    int    `json:"governance"`
		Total  int    `json:"total"`
	}
	_ = json.Unmarshal(raw, &list)
	rows := make([]map[string]string, 0, len(list))
	for _, d := range list {
		rows = append(rows, map[string]string{
			"department": d.Name, "env": fmt.Sprint(d.Env), "social": fmt.Sprint(d.Social),
			"gov": fmt.Sprint(d.Gov), "total": fmt.Sprint(d.Total),
		})
	}
	return rows
}

func (s *Service) Get(ctx context.Context, reportID id.ID) (*domain.Report, error) {
	return s.store.ByID(ctx, reportID)
}

func (s *Service) Export(ctx context.Context, reportID id.ID, fmtName string) ([]byte, string, error) {
	r, err := s.store.ByID(ctx, reportID)
	if err != nil {
		return nil, "", err
	}
	switch fmtName {
	case "csv":
		b, err := s.export.ExportCSV(r)
		return b, "text/csv", err
	case "xlsx":
		b, err := s.export.ExportXLSX(r)
		return b, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", err
	case "pdf":
		b, err := s.export.ExportPDF(r)
		return b, "application/pdf", err
	default:
		return nil, "", fmt.Errorf("unsupported format")
	}
}
