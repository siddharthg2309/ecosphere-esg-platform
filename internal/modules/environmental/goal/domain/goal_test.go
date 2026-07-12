package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

func TestGoalStatusTransitions(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name          string
		current       string
		deadline, now time.Time
		want          Status
	}{
		{"completed", "90", created.AddDate(1, 0, 0), created, StatusCompleted},
		{"expired", "120", created.AddDate(0, 1, 0), created.AddDate(0, 2, 0), StatusAtRisk},
		{"behind", "1000", created.AddDate(1, 0, 0), created.AddDate(0, 9, 0), StatusAtRisk},
		{"on track", "120", created.AddDate(1, 0, 0), created.AddDate(0, 1, 0), StatusOnTrack},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := Goal{ID: id.New(), Name: "Fleet target", DepartmentID: id.New(), TargetCO2: decimal.NewFromInt(100), CurrentCO2: decimal.RequireFromString(tc.current), Deadline: tc.deadline, CreatedAt: created}
			g.Recompute(tc.now)
			if g.Status != tc.want {
				t.Fatalf("status = %s, want %s", g.Status, tc.want)
			}
		})
	}
}
