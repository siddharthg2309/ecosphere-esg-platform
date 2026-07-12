package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/department/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/db/sqlc"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Repository struct{ queries *sqlc.Queries }

func New(queries *sqlc.Queries) *Repository { return &Repository{queries: queries} }
func (r *Repository) Create(ctx context.Context, d *domain.Department) error {
	_, err := r.queries.CreateDepartment(ctx, sqlc.CreateDepartmentParams{ID: uuid(d.ID), Name: d.Name, Code: d.Code, HeadID: nullableUUID(d.HeadID), ParentID: nullableUUID(d.ParentID), EmployeeCount: int32(d.EmployeeCount), Status: string(d.Status)})
	return err
}
func (r *Repository) Update(ctx context.Context, d *domain.Department) error {
	row, err := r.queries.UpdateDepartment(ctx, sqlc.UpdateDepartmentParams{ID: uuid(d.ID), Name: d.Name, Code: d.Code, HeadID: nullableUUID(d.HeadID), ParentID: nullableUUID(d.ParentID), EmployeeCount: int32(d.EmployeeCount), Status: string(d.Status)})
	if err != nil {
		return err
	}
	*d = mapDepartment(row)
	return nil
}
func (r *Repository) ByID(ctx context.Context, departmentID id.ID) (*domain.Department, error) {
	row, err := r.queries.DepartmentByID(ctx, uuid(departmentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NotFound("department_not_found", "Department not found")
	}
	if err != nil {
		return nil, err
	}
	d := mapDepartment(row)
	return &d, nil
}
func (r *Repository) List(ctx context.Context, p page.Page) (page.Result[domain.Department], error) {
	rows, err := r.queries.ListDepartments(ctx, sqlc.ListDepartmentsParams{Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return page.Result[domain.Department]{}, err
	}
	total, err := r.queries.CountDepartments(ctx)
	if err != nil {
		return page.Result[domain.Department]{}, err
	}
	items := make([]domain.Department, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapDepartment(row))
	}
	return page.Result[domain.Department]{Items: items, Total: int(total)}, nil
}
func (r *Repository) CodeExists(ctx context.Context, code string, exclude id.ID) (bool, error) {
	return r.queries.DepartmentCodeExists(ctx, sqlc.DepartmentCodeExistsParams{Code: code, ID: uuid(exclude)})
}
func (r *Repository) EligibleHead(ctx context.Context, userID id.ID) (bool, error) {
	return r.queries.EligibleDepartmentHead(ctx, uuid(userID))
}
func (r *Repository) Deactivate(ctx context.Context, departmentID id.ID) (*domain.Department, error) {
	row, err := r.queries.DeactivateDepartment(ctx, uuid(departmentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NotFound("department_not_found", "Department not found")
	}
	if err != nil {
		return nil, err
	}
	d := mapDepartment(row)
	return &d, nil
}
func mapDepartment(row sqlc.Department) domain.Department {
	d := domain.Department{ID: fromUUID(row.ID), Name: row.Name, Code: row.Code, EmployeeCount: int(row.EmployeeCount), Status: domain.Status(row.Status), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}
	if row.HeadID.Valid {
		v := fromUUID(row.HeadID)
		d.HeadID = &v
	}
	if row.ParentID.Valid {
		v := fromUUID(row.ParentID)
		d.ParentID = &v
	}
	return d
}
func uuid(value id.ID) pgtype.UUID {
	var target pgtype.UUID
	_ = target.Scan(value.String())
	return target
}
func nullableUUID(value *id.ID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}
	return uuid(*value)
}
func fromUUID(value pgtype.UUID) id.ID {
	if !value.Valid {
		return ""
	}
	return id.ID(fmt.Sprintf("%x-%x-%x-%x-%x", value.Bytes[0:4], value.Bytes[4:6], value.Bytes[6:8], value.Bytes[8:10], value.Bytes[10:16]))
}
