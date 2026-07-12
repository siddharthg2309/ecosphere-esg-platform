package domain

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type Status string

const (
	StatusOnTrack   Status = "on_track"
	StatusAtRisk    Status = "at_risk"
	StatusCompleted Status = "completed"
)

type Goal struct {
	ID           id.ID           `json:"id"`
	Name         string          `json:"name"`
	DepartmentID id.ID           `json:"departmentId"`
	TargetCO2    decimal.Decimal `json:"targetCo2"`
	CurrentCO2   decimal.Decimal `json:"currentCo2"`
	Deadline     time.Time       `json:"deadline"`
	Status       Status          `json:"status"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

func New(name string, departmentID id.ID, target, current decimal.Decimal, deadline, now time.Time) (*Goal, error) {
	g := &Goal{ID: id.New(), Name: strings.TrimSpace(name), DepartmentID: departmentID, TargetCO2: target, CurrentCO2: current, Deadline: deadline, Status: StatusOnTrack, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
	if err := g.Validate(); err != nil {
		return nil, err
	}
	g.Recompute(now)
	return g, nil
}

func (g *Goal) Validate() error {
	fields := map[string]string{}
	if g.Name == "" {
		fields["name"] = "Name is required"
	}
	if g.DepartmentID == "" {
		fields["departmentId"] = "Department is required"
	}
	if !g.TargetCO2.GreaterThan(decimal.Zero) {
		fields["targetCo2"] = "Target must be greater than zero"
	}
	if g.CurrentCO2.IsNegative() {
		fields["currentCo2"] = "Current CO2 cannot be negative"
	}
	if g.Deadline.IsZero() {
		fields["deadline"] = "Deadline is required"
	}
	if len(fields) > 0 {
		return errs.Invalid("invalid_environmental_goal", "Environmental goal details are invalid", fields)
	}
	return nil
}

func (g *Goal) Recompute(now time.Time) {
	switch {
	case g.CurrentCO2.LessThanOrEqual(g.TargetCO2):
		g.Status = StatusCompleted
	case now.After(g.Deadline):
		g.Status = StatusAtRisk
	case g.progress().LessThan(g.expectedByNow(now)):
		g.Status = StatusAtRisk
	default:
		g.Status = StatusOnTrack
	}
	g.UpdatedAt = now.UTC()
}

func (g *Goal) progress() decimal.Decimal {
	if g.CurrentCO2.IsZero() {
		return decimal.NewFromInt(1)
	}
	return g.TargetCO2.Div(g.CurrentCO2)
}

func (g *Goal) expectedByNow(now time.Time) decimal.Decimal {
	total := g.Deadline.Sub(g.CreatedAt)
	if total <= 0 {
		return decimal.NewFromInt(1)
	}
	elapsed := now.Sub(g.CreatedAt)
	if elapsed <= 0 {
		return decimal.Zero
	}
	if elapsed >= total {
		return decimal.NewFromInt(1)
	}
	return decimal.NewFromFloat(elapsed.Seconds() / total.Seconds())
}
