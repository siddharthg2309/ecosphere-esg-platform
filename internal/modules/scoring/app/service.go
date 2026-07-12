package app

import (
	"context"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/scoring/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type MetricsSource interface {
	// LoadInputs returns per-department inputs for the current period.
	LoadInputs(ctx context.Context, period string) (map[id.ID]PillarInputs, map[id.ID]int, map[id.ID]string, error)
	Weights(ctx context.Context) (wEnv, wSocial, wGov int, err error)
}

type PillarInputs struct {
	Env     domain.EnvInputs
	Social  domain.SocialInputs
	Gov     domain.GovInputs
}

type ScoreStore interface {
	Upsert(ctx context.Context, scores []domain.DepartmentScore) error
	List(ctx context.Context, period string) ([]domain.DepartmentScore, error)
}

type Service struct {
	metrics MetricsSource
	store   ScoreStore
	period  func() string
}

func New(metrics MetricsSource, store ScoreStore, bus events.Bus) *Service {
	s := &Service{
		metrics: metrics,
		store:   store,
		period:  func() string { return time.Now().UTC().Format("2006") }, // calendar year period
	}
	if bus != nil {
		recompute := func(ctx context.Context, _ events.Event) error {
			_, err := s.RecomputeAll(ctx)
			return err
		}
		bus.Subscribe(events.NameEmissionRecorded, recompute)
		bus.Subscribe(events.NameParticipationDecided, recompute)
		bus.Subscribe(events.NameChallengeCompleted, recompute)
		bus.Subscribe(events.NameComplianceIssueRaised, recompute)
		bus.Subscribe(events.NameComplianceOverdue, recompute)
		bus.Subscribe(events.NameESGConfigChanged, recompute)
	}
	return s
}

func (s *Service) RecomputeAll(ctx context.Context) ([]domain.DepartmentScore, error) {
	period := s.period()
	inputs, _, names, err := s.metrics.LoadInputs(ctx, period)
	if err != nil {
		return nil, err
	}
	wEnv, wSocial, wGov, err := s.metrics.Weights(ctx)
	if err != nil {
		return nil, err
	}
	scores := make([]domain.DepartmentScore, 0, len(inputs))
	for deptID, in := range inputs {
		env := domain.Environmental(in.Env)
		social := domain.Social(in.Social)
		gov := domain.Governance(in.Gov)
		total := domain.DeptTotal(env, social, gov, wEnv, wSocial, wGov)
		scores = append(scores, domain.DepartmentScore{
			DeptID: deptID, Env: env, Social: social, Gov: gov, Total: total, Period: period, Name: names[deptID],
		})
	}
	if err = s.store.Upsert(ctx, scores); err != nil {
		return nil, err
	}
	return scores, nil
}

func (s *Service) Departments(ctx context.Context, period string) ([]domain.DepartmentScore, error) {
	if period == "" {
		period = s.period()
	}
	scores, err := s.store.List(ctx, period)
	if err != nil {
		return nil, err
	}
	if len(scores) == 0 {
		return s.RecomputeAll(ctx)
	}
	return scores, nil
}

func (s *Service) Overall(ctx context.Context, period string) (int, []domain.DepartmentScore, map[string]int, error) {
	scores, err := s.Departments(ctx, period)
	if err != nil {
		return 0, nil, nil, err
	}
	_, headcount, _, err := s.metrics.LoadInputs(ctx, period)
	if err != nil {
		return 0, nil, nil, err
	}
	wEnv, wSocial, wGov, _ := s.metrics.Weights(ctx)
	overall := domain.OverallESG(scores, headcount)
	return overall, scores, map[string]int{"weightEnv": wEnv, "weightSocial": wSocial, "weightGov": wGov}, nil
}
