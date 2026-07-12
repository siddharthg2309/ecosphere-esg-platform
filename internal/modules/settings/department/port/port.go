package port

import (
	"context"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/department/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Repository interface {
	Create(context.Context, *domain.Department) error
	Update(context.Context, *domain.Department) error
	ByID(context.Context, id.ID) (*domain.Department, error)
	List(context.Context, page.Page) (page.Result[domain.Department], error)
	CodeExists(context.Context, string, id.ID) (bool, error)
	EligibleHead(context.Context, id.ID) (bool, error)
	Deactivate(context.Context, id.ID) (*domain.Department, error)
}
