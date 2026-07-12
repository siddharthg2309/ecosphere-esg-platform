package app

import (
	"context"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/department/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/department/port"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
	"time"
)

type CreateCommand struct {
	Name, Code       string
	HeadID, ParentID *id.ID
	EmployeeCount    int
}
type UpdateCommand struct {
	ID               id.ID
	Name, Code       string
	HeadID, ParentID *id.ID
	EmployeeCount    int
	Status           domain.Status
}
type Service struct {
	repo port.Repository
	now  func() time.Time
}

func New(repo port.Repository) *Service {
	return &Service{repo: repo, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Create(ctx context.Context, c CreateCommand) (*domain.Department, error) {
	d, err := domain.New(c.Name, c.Code, s.now())
	if err != nil {
		return nil, err
	}
	d.HeadID = c.HeadID
	d.ParentID = c.ParentID
	d.EmployeeCount = c.EmployeeCount
	if err = d.Validate(); err != nil {
		return nil, err
	}
	if err = s.validateRelations(ctx, d); err != nil {
		return nil, err
	}
	exists, err := s.repo.CodeExists(ctx, d.Code, d.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errs.Conflict("department_code_taken", "Department code is already in use")
	}
	if err = s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}
func (s *Service) Update(ctx context.Context, c UpdateCommand) (*domain.Department, error) {
	d, err := s.repo.ByID(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	if err = d.Update(c.Name, c.Code, c.HeadID, c.ParentID, c.EmployeeCount, c.Status, s.now()); err != nil {
		return nil, err
	}
	if err = s.validateRelations(ctx, d); err != nil {
		return nil, err
	}
	exists, err := s.repo.CodeExists(ctx, d.Code, d.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errs.Conflict("department_code_taken", "Department code is already in use")
	}
	if err = s.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}
func (s *Service) ByID(ctx context.Context, departmentID id.ID) (*domain.Department, error) {
	return s.repo.ByID(ctx, departmentID)
}
func (s *Service) List(ctx context.Context, p page.Page) (page.Result[domain.Department], error) {
	return s.repo.List(ctx, p)
}
func (s *Service) Deactivate(ctx context.Context, departmentID id.ID) (*domain.Department, error) {
	return s.repo.Deactivate(ctx, departmentID)
}
func (s *Service) validateRelations(ctx context.Context, d *domain.Department) error {
	if d.HeadID != nil {
		eligible, err := s.repo.EligibleHead(ctx, *d.HeadID)
		if err != nil {
			return err
		}
		if !eligible {
			return errs.Invalid("invalid_department_head", "Department head must be an admin or department head", map[string]string{"headId": "User is not eligible"})
		}
	}
	if d.ParentID != nil {
		if _, err := s.repo.ByID(ctx, *d.ParentID); err != nil {
			return errs.Invalid("invalid_parent_department", "Parent department does not exist", map[string]string{"parentId": "Department not found"})
		}
	}
	return nil
}
