package domain

import (
	"testing"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

func TestEnvironmentalFixture(t *testing.T) {
	// 100% goals + 20% YoY reduction → 0.6*100 + 0.4*70 = 88
	got := Environmental(EnvInputs{GoalProgressPct: []float64{100}, YoYReductionPct: 20})
	if got != 88 {
		t.Fatalf("got %d want 88", got)
	}
}

func TestSocialFixture(t *testing.T) {
	// 0.4*50 + 0.3*40 + 0.3*100 = 20+12+30 = 62
	got := Social(SocialInputs{CSRParticipationPct: 50, DiversityIndex: 40, TrainingCompletionPct: 100})
	if got != 62 {
		t.Fatalf("got %d want 62", got)
	}
}

func TestGovernancePenalty(t *testing.T) {
	// base 0.5*100+0.5*100=100; penalty 3*2+5*1=11 → 89
	got := Governance(GovInputs{PolicyAckPct: 100, AuditPassPct: 100, OpenIssues: 2, OverdueIssues: 1})
	if got != 89 {
		t.Fatalf("got %d want 89", got)
	}
}

func TestDeptTotal403030(t *testing.T) {
	// env 80, social 70, gov 90 → (32+21+27)=80
	got := DeptTotal(80, 70, 90, 40, 30, 30)
	if got != 80 {
		t.Fatalf("got %d want 80", got)
	}
}

func TestOverallESGHeadcount(t *testing.T) {
	a, b := id.New(), id.New()
	scores := []DepartmentScore{{DeptID: a, Total: 80}, {DeptID: b, Total: 60}}
	// headcount 3 and 1 → (240+60)/4 = 75
	got := OverallESG(scores, map[id.ID]int{a: 3, b: 1})
	if got != 75 {
		t.Fatalf("got %d want 75", got)
	}
}
