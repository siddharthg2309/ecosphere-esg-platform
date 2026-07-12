package app

import (
	"context"
	"time"

	activity "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/activity/domain"
	participation "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/participation/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/port"
	training "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/training/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Service struct {
	activities     port.ActivityRepo
	participations port.ParticipationRepo
	trainings      port.TrainingRepo
	users          port.UserPoints
	flags          port.Flags
	diversity      port.DiversityReader
	bus            events.Bus
	now            func() time.Time
}

func New(
	activities port.ActivityRepo,
	participations port.ParticipationRepo,
	trainings port.TrainingRepo,
	users port.UserPoints,
	flags port.Flags,
	diversity port.DiversityReader,
	bus events.Bus,
) *Service {
	return &Service{
		activities: activities, participations: participations, trainings: trainings,
		users: users, flags: flags, diversity: diversity, bus: bus,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type CreateActivityCmd struct {
	Title            string
	CategoryID       id.ID
	Description      string
	Points           int
	EvidenceRequired bool
	ActivityDate     *time.Time
}

func (s *Service) CreateActivity(ctx context.Context, cmd CreateActivityCmd) (*activity.CSRActivity, error) {
	a, err := activity.New(cmd.Title, cmd.CategoryID, cmd.Description, cmd.Points, cmd.EvidenceRequired, cmd.ActivityDate, s.now())
	if err != nil {
		return nil, err
	}
	if err = s.activities.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) ListActivities(ctx context.Context, p page.Page) (page.Result[activity.CSRActivity], error) {
	return s.activities.List(ctx, p)
}

func (s *Service) GetActivity(ctx context.Context, activityID id.ID) (*activity.CSRActivity, error) {
	return s.activities.ByID(ctx, activityID)
}

type JoinActivityCmd struct {
	EmployeeID id.ID
	ActivityID id.ID
	ProofURL   string
	Notes      string
}

func (s *Service) JoinActivity(ctx context.Context, cmd JoinActivityCmd) (*participation.EmployeeParticipation, error) {
	if _, err := s.activities.ByID(ctx, cmd.ActivityID); err != nil {
		return nil, err
	}
	p, err := participation.New(cmd.EmployeeID, cmd.ActivityID, cmd.ProofURL, cmd.Notes, s.now())
	if err != nil {
		return nil, err
	}
	if err = s.participations.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) ListParticipations(ctx context.Context, p page.Page, approval string) (page.Result[participation.EmployeeParticipation], error) {
	return s.participations.List(ctx, p, approval)
}

func (s *Service) ApproveParticipation(ctx context.Context, participationID, by id.ID) (*participation.EmployeeParticipation, error) {
	_ = by
	p, err := s.participations.ByID(ctx, participationID)
	if err != nil {
		return nil, err
	}
	act, err := s.activities.ByID(ctx, p.ActivityID)
	if err != nil {
		return nil, err
	}
	requireEvidence := act.EvidenceRequired || (s.flags != nil && s.flags.IsEnabled(ctx, "require_csr_evidence"))
	if err = p.Approve(act.Points, requireEvidence, s.now()); err != nil {
		return nil, err
	}
	if err = s.participations.Save(ctx, p); err != nil {
		return nil, err
	}
	if err = s.users.AddPoints(ctx, p.EmployeeID, p.PointsEarned); err != nil {
		return nil, err
	}
	_ = s.bus.Publish(ctx, events.ParticipationDecided{
		Kind: "csr", EmployeeID: p.EmployeeID, Approved: true, Points: p.PointsEarned, XP: 0,
	})
	return p, nil
}

func (s *Service) RejectParticipation(ctx context.Context, participationID, by id.ID) (*participation.EmployeeParticipation, error) {
	_ = by
	p, err := s.participations.ByID(ctx, participationID)
	if err != nil {
		return nil, err
	}
	if err = p.Reject(); err != nil {
		return nil, err
	}
	if err = s.participations.Save(ctx, p); err != nil {
		return nil, err
	}
	_ = s.bus.Publish(ctx, events.ParticipationDecided{
		Kind: "csr", EmployeeID: p.EmployeeID, Approved: false, Points: 0, XP: 0,
	})
	return p, nil
}

func (s *Service) Diversity(ctx context.Context) (port.DiversityMetrics, error) {
	if s.diversity == nil {
		return port.DiversityMetrics{}, errs.NotFound("diversity_unavailable", "Diversity metrics unavailable")
	}
	return s.diversity.Metrics(ctx)
}

func (s *Service) CreateTraining(ctx context.Context, name, assignedTo string) (*training.Training, error) {
	t, err := training.New(name, assignedTo, s.now())
	if err != nil {
		return nil, err
	}
	if err = s.trainings.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) ListTrainings(ctx context.Context) ([]training.Training, error) {
	return s.trainings.List(ctx)
}

func (s *Service) CompleteTraining(ctx context.Context, trainingID, employeeID id.ID) error {
	c := &training.Completion{
		ID: id.New(), EmployeeID: employeeID, TrainingID: trainingID, CompletedAt: s.now(),
	}
	return s.trainings.Complete(ctx, c)
}
