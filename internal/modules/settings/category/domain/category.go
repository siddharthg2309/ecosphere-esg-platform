package domain

import (
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"strings"
	"time"
)

type Type string

const (
	TypeCSR       Type = "csr_activity"
	TypeChallenge Type = "challenge"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

type Category struct {
	ID        id.ID     `json:"id"`
	Name      string    `json:"name"`
	Type      Type      `json:"type"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func New(name string, categoryType Type, status Status, now time.Time) (Category, error) {
	c := Category{ID: id.New(), Name: strings.TrimSpace(name), Type: categoryType, Status: status, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
	if c.Status == "" {
		c.Status = StatusActive
	}
	return c, c.Validate()
}
func (c Category) Validate() error {
	fields := map[string]string{}
	if c.Name == "" {
		fields["name"] = "Name is required"
	}
	if c.Type != TypeCSR && c.Type != TypeChallenge {
		fields["type"] = "Type must be csr_activity or challenge"
	}
	if c.Status != StatusActive && c.Status != StatusInactive {
		fields["status"] = "Status must be active or inactive"
	}
	if len(fields) > 0 {
		return errs.Invalid("invalid_category", "Category details are invalid", fields)
	}
	return nil
}
