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
