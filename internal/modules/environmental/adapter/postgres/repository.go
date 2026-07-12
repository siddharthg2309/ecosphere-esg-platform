package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	carbondomain "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/carbon/domain"
	carbonport "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/carbon/port"
	goaldomain "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/goal/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/db/sqlc"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Repository struct{ q *sqlc.Queries }

func New(pool *pgxpool.Pool) *Repository { return &Repository{q: sqlc.New(pool)} }

func uuid(v id.ID) pgtype.UUID { var out pgtype.UUID; _ = out.Scan(v.String()); return out }
func nullableUUID(v *id.ID) pgtype.UUID {
	if v == nil {
		return pgtype.UUID{}
	}
	return uuid(*v)
}
func fromUUID(v pgtype.UUID) id.ID {
	if !v.Valid {
		return ""
	}
	return id.ID(fmt.Sprintf("%x-%x-%x-%x-%x", v.Bytes[0:4], v.Bytes[4:6], v.Bytes[6:8], v.Bytes[8:10], v.Bytes[10:16]))
}
func numeric(v decimal.Decimal) pgtype.Numeric {
	var out pgtype.Numeric
	_ = out.Scan(v.String())
	return out
}
func fromNumeric(v pgtype.Numeric) decimal.Decimal {
	raw, _ := v.Value()
	out, _ := decimal.NewFromString(fmt.Sprint(raw))
	return out
}
func date(v time.Time) pgtype.Date {
	if v.IsZero() {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: v, Valid: true}
}
func text(v string) pgtype.Text {
	if v == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: v, Valid: true}
}
func notFound(err error, code, message string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.NotFound(code, message)
	}
	return err
}

func (r *Repository) Create(ctx context.Context, t *carbondomain.Transaction) error {
	row, err := r.q.CreateCarbonTransaction(ctx, sqlc.CreateCarbonTransactionParams{ID: uuid(t.ID), DepartmentID: uuid(t.DepartmentID), Source: string(t.Source), Quantity: numeric(t.Quantity), EmissionFactorID: uuid(t.EmissionFactorID), FactorValue: numeric(t.FactorValue), ComputedCo2: numeric(t.ComputedCO2), TxnDate: date(t.TxnDate), EvidenceUrl: text(t.EvidenceURL), Status: string(t.Status), CreatedAt: pgtype.Timestamptz{Time: t.CreatedAt, Valid: true}})
	if err != nil {
		return err
	}
	*t = mapTransaction(row)
	return nil
}
func (r *Repository) ByID(ctx context.Context, v id.ID) (*carbondomain.Transaction, error) {
	row, err := r.q.CarbonTransactionByID(ctx, uuid(v))
	if err != nil {
		return nil, notFound(err, "carbon_transaction_not_found", "Carbon transaction not found")
	}
	t := mapTransaction(row)
	return &t, nil
}
func (r *Repository) SaveVerified(ctx context.Context, t *carbondomain.Transaction) error {
	row, err := r.q.VerifyCarbonTransaction(ctx, sqlc.VerifyCarbonTransactionParams{ID: uuid(t.ID), ComputedCo2: numeric(t.ComputedCO2), VerifiedBy: uuid(*t.VerifiedBy), VerifiedAt: pgtype.Timestamptz{Time: *t.VerifiedAt, Valid: true}})
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.Conflict("already_verified", "Verified carbon transactions are immutable")
	}
	if err != nil {
		return err
	}
	*t = mapTransaction(row)
	return nil
}
func (r *Repository) List(ctx context.Context, f carbonport.Filter) (page.Result[carbondomain.Transaction], error) {
	from, to := pgtype.Date{}, pgtype.Date{}
	if f.From != nil {
		from = date(*f.From)
	}
	if f.To != nil {
		to = date(*f.To)
	}
	source, status := "", ""
	if f.Source != nil {
		source = string(*f.Source)
	}
	if f.Status != nil {
		status = string(*f.Status)
	}
	args := sqlc.ListCarbonTransactionsParams{Column1: nullableUUID(f.DepartmentID), Column2: from, Column3: to, Column4: source, Column5: status, Limit: int32(f.Page.Limit), Offset: int32(f.Page.Offset)}
	rows, err := r.q.ListCarbonTransactions(ctx, args)
	if err != nil {
		return page.Result[carbondomain.Transaction]{}, err
	}
	total, err := r.q.CountCarbonTransactions(ctx, sqlc.CountCarbonTransactionsParams{Column1: args.Column1, Column2: from, Column3: to, Column4: source, Column5: status})
	if err != nil {
		return page.Result[carbondomain.Transaction]{}, err
	}
	items := make([]carbondomain.Transaction, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapTransaction(row))
	}
	return page.Result[carbondomain.Transaction]{Items: items, Total: int(total)}, nil
}
func (r *Repository) Summary(ctx context.Context, departmentID *id.ID, from, to time.Time) (carbonport.Summary, error) {
	rows, err := r.q.CarbonSummary(ctx, sqlc.CarbonSummaryParams{Column1: nullableUUID(departmentID), TxnDate: date(from), TxnDate_2: date(to)})
	if err != nil {
		return carbonport.Summary{}, err
	}
	result := carbonport.Summary{Total: decimal.Zero, BySource: map[carbondomain.Source]decimal.Decimal{}}
	for _, row := range rows {
		value := fromNumeric(row.Co2)
		result.BySource[carbondomain.Source(row.Source)] = value
		result.Total = result.Total.Add(value)
	}
	return result, nil
}
func (r *Repository) Factor(ctx context.Context, factorID id.ID) (string, decimal.Decimal, bool, error) {
	row, err := r.q.ActiveEmissionFactor(ctx, uuid(factorID))
	if err != nil {
		return "", decimal.Zero, false, notFound(err, "factor_not_found", "Emission factor not found")
	}
	return row.Unit, fromNumeric(row.Kgco2PerUnit), row.Status == "active", nil
}
func (r *Repository) DepartmentExists(ctx context.Context, departmentID id.ID) (bool, error) {
	return r.q.DepartmentExists(ctx, uuid(departmentID))
}
func (r *Repository) IsDepartmentHead(ctx context.Context, userID, departmentID id.ID) (bool, error) {
	return r.q.IsDepartmentHead(ctx, sqlc.IsDepartmentHeadParams{ID: uuid(departmentID), HeadID: uuid(userID)})
}

type GoalRepository struct{ q *sqlc.Queries }

func NewGoals(pool *pgxpool.Pool) *GoalRepository { return &GoalRepository{q: sqlc.New(pool)} }
func (r *GoalRepository) Create(ctx context.Context, g *goaldomain.Goal) error {
	row, err := r.q.CreateEnvironmentalGoal(ctx, sqlc.CreateEnvironmentalGoalParams{ID: uuid(g.ID), Name: g.Name, DepartmentID: uuid(g.DepartmentID), TargetCo2: numeric(g.TargetCO2), CurrentCo2: numeric(g.CurrentCO2), Deadline: date(g.Deadline), Status: string(g.Status), CreatedAt: pgtype.Timestamptz{Time: g.CreatedAt, Valid: true}, UpdatedAt: pgtype.Timestamptz{Time: g.UpdatedAt, Valid: true}})
	if err != nil {
		return err
	}
	*g = mapGoal(row)
	return nil
}
func (r *GoalRepository) ByID(ctx context.Context, v id.ID) (*goaldomain.Goal, error) {
	row, err := r.q.EnvironmentalGoalByID(ctx, uuid(v))
	if err != nil {
		return nil, notFound(err, "environmental_goal_not_found", "Environmental goal not found")
	}
	g := mapGoal(row)
	return &g, nil
}
func (r *GoalRepository) Update(ctx context.Context, g *goaldomain.Goal) error {
	row, err := r.q.UpdateEnvironmentalGoal(ctx, sqlc.UpdateEnvironmentalGoalParams{ID: uuid(g.ID), Name: g.Name, TargetCo2: numeric(g.TargetCO2), CurrentCo2: numeric(g.CurrentCO2), Deadline: date(g.Deadline), Status: string(g.Status), UpdatedAt: pgtype.Timestamptz{Time: g.UpdatedAt, Valid: true}})
	if err != nil {
		return notFound(err, "environmental_goal_not_found", "Environmental goal not found")
	}
	*g = mapGoal(row)
	return nil
}
func (r *GoalRepository) List(ctx context.Context, departmentID *id.ID, p page.Page) (page.Result[goaldomain.Goal], error) {
	rows, err := r.q.ListEnvironmentalGoals(ctx, sqlc.ListEnvironmentalGoalsParams{Column1: nullableUUID(departmentID), Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return page.Result[goaldomain.Goal]{}, err
	}
	total, err := r.q.CountEnvironmentalGoals(ctx, nullableUUID(departmentID))
	if err != nil {
		return page.Result[goaldomain.Goal]{}, err
	}
	items := make([]goaldomain.Goal, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapGoal(row))
	}
	return page.Result[goaldomain.Goal]{Items: items, Total: int(total)}, nil
}
func (r *GoalRepository) VerifiedEmissionsThrough(ctx context.Context, departmentID id.ID, deadline time.Time) (decimal.Decimal, error) {
	v, err := r.q.VerifiedEmissionsThrough(ctx, sqlc.VerifiedEmissionsThroughParams{DepartmentID: uuid(departmentID), TxnDate: date(deadline)})
	if err != nil {
		return decimal.Zero, err
	}
	return fromNumeric(v), nil
}
func (r *GoalRepository) GoalsForDepartment(ctx context.Context, departmentID id.ID) ([]goaldomain.Goal, error) {
	rows, err := r.q.GoalsForDepartment(ctx, uuid(departmentID))
	if err != nil {
		return nil, err
	}
	items := make([]goaldomain.Goal, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapGoal(row))
	}
	return items, nil
}

func mapTransaction(v sqlc.CarbonTransaction) carbondomain.Transaction {
	t := carbondomain.Transaction{ID: fromUUID(v.ID), DepartmentID: fromUUID(v.DepartmentID), Source: carbondomain.Source(v.Source), Quantity: fromNumeric(v.Quantity), EmissionFactorID: fromUUID(v.EmissionFactorID), FactorValue: fromNumeric(v.FactorValue), ComputedCO2: fromNumeric(v.ComputedCo2), TxnDate: v.TxnDate.Time, Status: carbondomain.Status(v.Status), CreatedAt: v.CreatedAt.Time}
	if v.EvidenceUrl.Valid {
		t.EvidenceURL = v.EvidenceUrl.String
	}
	if v.VerifiedBy.Valid {
		x := fromUUID(v.VerifiedBy)
		t.VerifiedBy = &x
	}
	if v.VerifiedAt.Valid {
		x := v.VerifiedAt.Time
		t.VerifiedAt = &x
	}
	return t
}
func mapGoal(v sqlc.EnvironmentalGoal) goaldomain.Goal {
	return goaldomain.Goal{ID: fromUUID(v.ID), Name: v.Name, DepartmentID: fromUUID(v.DepartmentID), TargetCO2: fromNumeric(v.TargetCo2), CurrentCO2: fromNumeric(v.CurrentCo2), Deadline: v.Deadline.Time, Status: goaldomain.Status(v.Status), CreatedAt: v.CreatedAt.Time, UpdatedAt: v.UpdatedAt.Time}
}
