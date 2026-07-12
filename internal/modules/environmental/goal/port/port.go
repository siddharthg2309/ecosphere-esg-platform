package port

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/goal/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Repository interface {
	Create(context.Context, *domain.Goal) error
	ByID(context.Context, id.ID) (*domain.Goal, error)
	Update(context.Context, *domain.Goal) error
	List(context.Context, *id.ID, page.Page) (page.Result[domain.Goal], error)
	VerifiedEmissionsThrough(context.Context, id.ID, time.Time) (decimal.Decimal, error)
	GoalsForDepartment(context.Context, id.ID) ([]domain.Goal, error)
}
