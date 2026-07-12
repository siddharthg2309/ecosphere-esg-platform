package domain

import (
	category "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/category/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"strings"
	"time"
)

type Reward struct {
	ID             id.ID           `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	PointsRequired int             `json:"pointsRequired"`
	Stock          int             `json:"stock"`
	Status         category.Status `json:"status"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

func New(name, description string, points, stock int, status category.Status, now time.Time) (Reward, error) {
	r := Reward{ID: id.New(), Name: strings.TrimSpace(name), Description: description, PointsRequired: points, Stock: stock, Status: status, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
	if r.Status == "" {
		r.Status = category.StatusActive
	}
	if r.Name == "" || points < 0 || stock < 0 {
		return Reward{}, errs.Invalid("invalid_reward", "Reward name, points and stock are invalid", nil)
	}
	return r, nil
}
func (r Reward) CanRedeem() bool { return r.Status == category.StatusActive && r.Stock > 0 }

// Balance is the redeemable points held by an employee (users.points).
type Balance struct {
	UserID id.ID
	Points int
}

// Redeem mutates reward stock and employee points in memory; persistence runs in a DB tx.
func (r *Reward) Redeem(emp *Balance) error {
	if emp == nil {
		return errs.Invalid("invalid_employee", "Employee balance is required", nil)
	}
	if !r.CanRedeem() {
		return errs.Conflict("out_of_stock", "reward unavailable")
	}
	if emp.Points < r.PointsRequired {
		return errs.Invalid("insufficient_points", "not enough points", nil)
	}
	emp.Points -= r.PointsRequired
	r.Stock--
	return nil
}

