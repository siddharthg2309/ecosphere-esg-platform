package app

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/goal/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/goal/port"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Service struct {
	repo port.Repository
	now  func() time.Time
}

func New(repo port.Repository, bus events.Bus) *Service {
	s := &Service{repo: repo, now: func() time.Time { return time.Now().UTC() }}
	bus.Subscribe(events.NameEmissionRecorded, func(ctx context.Context, event events.Event) error {
		recorded, ok := event.(events.EmissionRecorded)
		if !ok {
			return nil
		}
		return s.RefreshDepartment(ctx, recorded.DepartmentID)
	})
	return s
}

func (s *Service) Create(ctx context.Context, name string, departmentID id.ID, target, current decimal.Decimal, deadline time.Time) (*domain.Goal, error) {
	g, err := domain.New(name, departmentID, target, current, deadline, s.now())
	if err != nil {
		return nil, err
	}
	if err = s.repo.Create(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}
func (s *Service) Update(ctx context.Context, goalID id.ID, name string, target decimal.Decimal, deadline time.Time) (*domain.Goal, error) {
	g, err := s.repo.ByID(ctx, goalID)
	if err != nil {
		return nil, err
	}
	g.Name = name
	g.TargetCO2 = target
	g.Deadline = deadline
	if err = g.Validate(); err != nil {
		return nil, err
	}
	g.Recompute(s.now())
	if err = s.repo.Update(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}
func (s *Service) List(ctx context.Context, departmentID *id.ID, p page.Page) (page.Result[domain.Goal], error) {
	return s.repo.List(ctx, departmentID, p)
}
func (s *Service) ByID(ctx context.Context, goalID id.ID) (*domain.Goal, error) {
	return s.repo.ByID(ctx, goalID)
}
func (s *Service) RefreshDepartment(ctx context.Context, departmentID id.ID) error {
	goals, err := s.repo.GoalsForDepartment(ctx, departmentID)
	if err != nil {
		return err
	}
	for i := range goals {
		current, e := s.repo.VerifiedEmissionsThrough(ctx, departmentID, goals[i].Deadline)
		if e != nil {
			return e
		}
		goals[i].CurrentCO2 = current
		goals[i].Recompute(s.now())
		if e = s.repo.Update(ctx, &goals[i]); e != nil {
			return e
		}
	}
	return nil
}
