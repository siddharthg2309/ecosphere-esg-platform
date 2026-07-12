package app

import (
	"context"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/department/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
	"testing"
)

type fakeRepo struct {
	exists bool
	saved  *domain.Department
}

func (f *fakeRepo) Create(_ context.Context, d *domain.Department) error    { f.saved = d; return nil }
func (f *fakeRepo) Update(context.Context, *domain.Department) error        { return nil }
func (f *fakeRepo) ByID(context.Context, id.ID) (*domain.Department, error) { return f.saved, nil }
func (f *fakeRepo) List(context.Context, page.Page) (page.Result[domain.Department], error) {
	return page.Result[domain.Department]{}, nil
}
func (f *fakeRepo) CodeExists(context.Context, string, id.ID) (bool, error) { return f.exists, nil }
func (f *fakeRepo) EligibleHead(context.Context, id.ID) (bool, error)       { return true, nil }
func (f *fakeRepo) Deactivate(context.Context, id.ID) (*domain.Department, error) {
	return f.saved, nil
}
func TestCreateRejectsDuplicateCode(t *testing.T) {
	service := New(&fakeRepo{exists: true})
	if _, err := service.Create(context.Background(), CreateCommand{Name: "Logistics", Code: "LOG"}); err == nil {
		t.Fatal("expected conflict")
	}
}
