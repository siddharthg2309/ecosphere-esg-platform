package domain

import (
	"github.com/shopspring/decimal"
	category "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/category/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"strings"
	"time"
)

type Factor struct {
	ID           id.ID           `json:"id"`
	Name         string          `json:"name"`
	CategoryID   id.ID           `json:"categoryId"`
	Unit         string          `json:"unit"`
	KgCO2PerUnit decimal.Decimal `json:"kgCo2PerUnit"`
	Status       category.Status `json:"status"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

func New(name string, categoryID id.ID, unit string, value decimal.Decimal, status category.Status, now time.Time) (Factor, error) {
	f := Factor{ID: id.New(), Name: strings.TrimSpace(name), CategoryID: categoryID, Unit: strings.TrimSpace(unit), KgCO2PerUnit: value, Status: status, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
	if f.Status == "" {
		f.Status = category.StatusActive
	}
	return f, f.Validate()
}
func (f Factor) Validate() error {
	fields := map[string]string{}
	if f.Name == "" {
		fields["name"] = "Name is required"
	}
	if f.CategoryID == "" {
		fields["categoryId"] = "Category is required"
	}
	if f.Unit == "" {
		fields["unit"] = "Unit is required"
	}
	if f.KgCO2PerUnit.LessThanOrEqual(decimal.Zero) {
		fields["kgCo2PerUnit"] = "Emission factor must be greater than zero"
	}
	if f.Status != category.StatusActive && f.Status != category.StatusInactive {
		fields["status"] = "Status is invalid"
	}
	if len(fields) > 0 {
		return errs.Invalid("invalid_emission_factor", "Emission factor details are invalid", fields)
	}
	return nil
}
