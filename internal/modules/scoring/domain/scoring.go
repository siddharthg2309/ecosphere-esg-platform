package domain

import (
	"math"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type EnvInputs struct {
	GoalProgressPct []float64 // each 0..100
	YoYReductionPct float64   // e.g. +18 means 18% reduction
}

type SocialInputs struct {
	CSRParticipationPct   float64
	DiversityIndex        float64 // 0..100
	TrainingCompletionPct float64
}

type GovInputs struct {
	PolicyAckPct  float64
	AuditPassPct  float64
	OpenIssues    int
	OverdueIssues int
}

type DepartmentScore struct {
	DeptID  id.ID  `json:"departmentId"`
	Env     int    `json:"environmental"`
	Social  int    `json:"social"`
	Gov     int    `json:"governance"`
	Total   int    `json:"total"`
	Period  string `json:"period"`
	Name    string `json:"name,omitempty"`
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func round(v float64) int { return int(math.Round(v)) }

func avg(vals []float64) float64 {
	if len(vals) == 0 {
		return 50 // neutral when no goals
	}
	var s float64
	for _, v := range vals {
		s += v
	}
	return s / float64(len(vals))
}

// Environmental returns 0..100 from goal attainment + YoY reduction trend.
func Environmental(in EnvInputs) int {
	goalAttainment := clamp(avg(in.GoalProgressPct), 0, 100)
	trend := clamp(50+in.YoYReductionPct, 0, 100)
	return round(0.6*goalAttainment + 0.4*trend)
}

// Social returns 0..100 from CSR participation, diversity, training.
func Social(in SocialInputs) int {
	return round(0.4*clamp(in.CSRParticipationPct, 0, 100) + 0.3*clamp(in.DiversityIndex, 0, 100) + 0.3*clamp(in.TrainingCompletionPct, 0, 100))
}

// Governance returns 0..100 with penalties for open/overdue issues.
func Governance(in GovInputs) int {
	base := 0.5*clamp(in.PolicyAckPct, 0, 100) + 0.5*clamp(in.AuditPassPct, 0, 100)
	penalty := 3*in.OpenIssues + 5*in.OverdueIssues
	return int(clamp(float64(round(base)-penalty), 0, 100))
}

// DeptTotal is pillar-weighted 0..100.
func DeptTotal(env, social, gov, wEnv, wSocial, wGov int) int {
	den := wEnv + wSocial + wGov
	if den == 0 {
		return 0
	}
	return round(float64(env*wEnv+social*wSocial+gov*wGov) / float64(den))
}

// OverallESG is headcount-weighted mean of department totals.
func OverallESG(scores []DepartmentScore, headcount map[id.ID]int) int {
	var num, den int
	for _, s := range scores {
		h := headcount[s.DeptID]
		if h == 0 {
			h = 1
		}
		num += s.Total * h
		den += h
	}
	if den == 0 {
		return 0
	}
	return round(float64(num) / float64(den))
}

// GoalProgressPct: if target is emission ceiling, progress = max(0, 100*(1 - current/target)).
func GoalProgressPct(current, target float64) float64 {
	if target <= 0 {
		return 50
	}
	if current <= target {
		return 100
	}
	// over target — scale down
	over := (current - target) / target
	return clamp(100-over*100, 0, 100)
}
