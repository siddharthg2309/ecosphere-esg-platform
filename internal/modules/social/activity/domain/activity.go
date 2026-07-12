package domain

import (
	"strings"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

type CSRActivity struct {
	ID               id.ID     `json:"id"`
	Title            string    `json:"title"`
	CategoryID       id.ID     `json:"categoryId"`
	Description      string    `json:"description"`
	Points           int       `json:"points"`
	EvidenceRequired bool      `json:"evidenceRequired"`
	Status           Status    `json:"status"`
	ActivityDate     *time.Time `json:"activityDate,omitempty"`
	JoinedCount      int       `json:"joinedCount,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func New(title string, categoryID id.ID, description string, points int, evidenceRequired bool, activityDate *time.Time, now time.Time) (*CSRActivity, error) {
	a := &CSRActivity{
		ID:               id.New(),
		Title:            strings.TrimSpace(title),
		CategoryID:       categoryID,
		Description:      strings.TrimSpace(description),
		Points:           points,
		EvidenceRequired: evidenceRequired,
		Status:           StatusActive,
		ActivityDate:     activityDate,
		CreatedAt:        now.UTC(),
		UpdatedAt:        now.UTC(),
	}
	if err := a.Validate(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *CSRActivity) Validate() error {
	fields := map[string]string{}
	if a.Title == "" || len(a.Title) > 200 {
		fields["title"] = "Title is required and must be at most 200 characters"
	}
	if a.CategoryID == "" {
		fields["categoryId"] = "Category is required"
	}
	if a.Points < 0 {
		fields["points"] = "Points cannot be negative"
	}
	if a.Status != StatusActive && a.Status != StatusInactive {
		fields["status"] = "Status must be active or inactive"
	}
	if len(fields) > 0 {
		return errs.Invalid("invalid_csr_activity", "CSR activity details are invalid", fields)
	}
	return nil
}
