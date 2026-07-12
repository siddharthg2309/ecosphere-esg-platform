package port

import (
	"context"

	activity "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/activity/domain"
	participation "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/participation/domain"
	training "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/training/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type ActivityRepo interface {
	Create(ctx context.Context, a *activity.CSRActivity) error
	ByID(ctx context.Context, id id.ID) (*activity.CSRActivity, error)
	List(ctx context.Context, p page.Page) (page.Result[activity.CSRActivity], error)
}

type ParticipationRepo interface {
	Create(ctx context.Context, p *participation.EmployeeParticipation) error
	ByID(ctx context.Context, id id.ID) (*participation.EmployeeParticipation, error)
	Save(ctx context.Context, p *participation.EmployeeParticipation) error
	List(ctx context.Context, p page.Page, approval string) (page.Result[participation.EmployeeParticipation], error)
	ListByEmployee(ctx context.Context, employeeID id.ID) ([]participation.EmployeeParticipation, error)
}

type TrainingRepo interface {
	Create(ctx context.Context, t *training.Training) error
	List(ctx context.Context) ([]training.Training, error)
	Complete(ctx context.Context, c *training.Completion) error
}

type UserPoints interface {
	AddPoints(ctx context.Context, userID id.ID, points int) error
}

type Flags interface {
	IsEnabled(ctx context.Context, key string) bool
}

type DiversityMetrics struct {
	GenderWomenPct        float64 `json:"genderWomenPct"`
	GenderMenPct          float64 `json:"genderMenPct"`
	GenderNonBinaryPct    float64 `json:"genderNonBinaryPct"`
	LeadershipWomenPct    float64 `json:"leadershipWomenPct"`
	DiverseLeadersPct     float64 `json:"diverseLeadersPct"`
	LeadershipTargetPct   float64 `json:"leadershipTargetPct"`
	TrainingCompletionPct float64 `json:"trainingCompletionPct"`
	CSRParticipationPct   float64 `json:"csrParticipationPct"`
}

type DiversityReader interface {
	Metrics(ctx context.Context) (DiversityMetrics, error)
}
