package domain

import (
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"strings"
	"time"
)

type Policy struct {
	ID            id.ID     `json:"id"`
	Title         string    `json:"title"`
	Body          string    `json:"body"`
	Version       int       `json:"version"`
	EffectiveDate time.Time `json:"effectiveDate"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func New(title, body string, effective time.Time, now time.Time) (Policy, error) {
	p := Policy{ID: id.New(), Title: strings.TrimSpace(title), Body: strings.TrimSpace(body), Version: 1, EffectiveDate: effective, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
	if p.Title == "" || p.Body == "" {
		return Policy{}, errs.Invalid("invalid_policy", "Policy title and body are required", nil)
	}
	return p, nil
}
func (p *Policy) Publish(title, body string, now time.Time) error {
	p.Title = strings.TrimSpace(title)
	p.Body = strings.TrimSpace(body)
	if p.Title == "" || p.Body == "" {
		return errs.Invalid("invalid_policy", "Policy title and body are required", nil)
	}
	p.Version++
	p.EffectiveDate = now.UTC()
	p.UpdatedAt = now.UTC()
	return nil
}
