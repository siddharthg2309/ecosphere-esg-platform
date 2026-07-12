package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/reporting/domain"
	platformdb "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/db"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Save(ctx context.Context, r *domain.Report) error {
	filters, _ := json.Marshal(r.Filters)
	result, _ := json.Marshal(r.Sections)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO reports(id,type,filters,result,generated_by,generated_at)
		VALUES($1,$2,$3,$4,$5,$6)`,
		r.ID, r.Type, filters, result, r.GeneratedBy, r.GeneratedAt)
	return platformdb.MapError(err)
}

func (s *Store) ByID(ctx context.Context, reportID id.ID) (*domain.Report, error) {
	row := s.pool.QueryRow(ctx, `SELECT id,type,filters,result,generated_by,generated_at FROM reports WHERE id=$1`, reportID)
	var r domain.Report
	var filters, result []byte
	var by *id.ID
	err := row.Scan(&r.ID, &r.Type, &filters, &result, &by, &r.GeneratedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NotFound("report_not_found", "Report not found")
	}
	if err != nil {
		return nil, platformdb.MapError(err)
	}
	_ = json.Unmarshal(filters, &r.Filters)
	_ = json.Unmarshal(result, &r.Sections)
	r.GeneratedBy = by
	return &r, nil
}

type DataSource struct{ pool *pgxpool.Pool }

func NewData(pool *pgxpool.Pool) *DataSource { return &DataSource{pool: pool} }

func (d *DataSource) EnvironmentalFigures(ctx context.Context, f domain.Filters) (domain.Section, error) {
	var total float64
	q := `SELECT COALESCE(SUM(computed_co2),0)::float8 FROM carbon_transactions WHERE status='verified'`
	args := []any{}
	if f.DepartmentID != nil {
		q += ` AND department_id=$1`
		args = append(args, *f.DepartmentID)
	}
	_ = d.pool.QueryRow(ctx, q, args...).Scan(&total)
	var goals int
	_ = d.pool.QueryRow(ctx, `SELECT COUNT(*) FROM environmental_goals`).Scan(&goals)
	return domain.Section{
		Title:   "Environmental",
		Summary: fmt.Sprintf("Verified emissions total %.2f t CO₂e across tracked transactions.", total),
		Metrics: map[string]any{"verifiedEmissions": total, "goalsTracked": goals},
		Rows:    []map[string]string{{"metric": "Verified CO2", "value": fmt.Sprintf("%.2f", total)}, {"metric": "Goals", "value": fmt.Sprint(goals)}},
	}, nil
}

func (d *DataSource) SocialFigures(ctx context.Context, f domain.Filters) (domain.Section, error) {
	var csr, train, done int
	_ = d.pool.QueryRow(ctx, `SELECT COUNT(*) FROM employee_participations WHERE approval='approved'`).Scan(&csr)
	_ = d.pool.QueryRow(ctx, `SELECT COUNT(*) FROM trainings`).Scan(&train)
	_ = d.pool.QueryRow(ctx, `SELECT COUNT(*) FROM training_completions`).Scan(&done)
	_ = f
	return domain.Section{
		Title:   "Social",
		Summary: fmt.Sprintf("%d approved CSR participations; %d training completions.", csr, done),
		Metrics: map[string]any{"csrApproved": csr, "trainings": train, "trainingCompletions": done},
		Rows: []map[string]string{
			{"metric": "CSR approvals", "value": fmt.Sprint(csr)},
			{"metric": "Training completions", "value": fmt.Sprint(done)},
		},
	}, nil
}

func (d *DataSource) GovernanceFigures(ctx context.Context, f domain.Filters) (domain.Section, error) {
	var open, overdue, acks, audits int
	_ = d.pool.QueryRow(ctx, `SELECT COUNT(*) FROM compliance_issues WHERE status IN ('open','in_progress')`).Scan(&open)
	_ = d.pool.QueryRow(ctx, `SELECT COUNT(*) FROM compliance_issues WHERE status='open' AND due_date < CURRENT_DATE`).Scan(&overdue)
	_ = d.pool.QueryRow(ctx, `SELECT COUNT(*) FROM policy_acknowledgements`).Scan(&acks)
	_ = d.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audits`).Scan(&audits)
	_ = f
	_ = time.Now()
	return domain.Section{
		Title:   "Governance",
		Summary: fmt.Sprintf("%d open issues (%d overdue); %d policy acknowledgements; %d audits.", open, overdue, acks, audits),
		Metrics: map[string]any{"openIssues": open, "overdueIssues": overdue, "acknowledgements": acks, "audits": audits},
		Rows: []map[string]string{
			{"metric": "Open issues", "value": fmt.Sprint(open)},
			{"metric": "Overdue", "value": fmt.Sprint(overdue)},
			{"metric": "Policy acks", "value": fmt.Sprint(acks)},
		},
	}, nil
}
