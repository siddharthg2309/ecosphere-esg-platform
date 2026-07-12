package domain

import (
	"regexp"
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

var codePattern = regexp.MustCompile(`^[A-Z0-9]{2,12}$`)

type Department struct {
	ID            id.ID     `json:"id"`
	Name          string    `json:"name"`
	Code          string    `json:"code"`
	HeadID        *id.ID    `json:"headId,omitempty"`
	ParentID      *id.ID    `json:"parentId,omitempty"`
	EmployeeCount int       `json:"employeeCount"`
	Status        Status    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func New(name, code string, now time.Time) (*Department, error) {
	department := &Department{ID: id.New(), Name: strings.TrimSpace(name), Code: strings.ToUpper(strings.TrimSpace(code)), Status: StatusActive, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
	if err := department.Validate(); err != nil {
		return nil, err
	}
	return department, nil
}

func (d *Department) Validate() error {
	fields := map[string]string{}
	if d.Name == "" || len(d.Name) > 120 {
		fields["name"] = "Name is required and must be at most 120 characters"
	}
	if !codePattern.MatchString(d.Code) {
		fields["code"] = "Code must contain 2-12 uppercase letters or numbers"
	}
	if d.Status != StatusActive && d.Status != StatusInactive {
		fields["status"] = "Status must be active or inactive"
	}
	if d.EmployeeCount < 0 {
		fields["employeeCount"] = "Employee count cannot be negative"
	}
	if d.ParentID != nil && *d.ParentID == d.ID {
		fields["parentId"] = "A department cannot be its own parent"
	}
	if len(fields) > 0 {
		return errs.Invalid("invalid_department", "Department details are invalid", fields)
	}
	return nil
}

func (d *Department) Update(name, code string, headID, parentID *id.ID, employeeCount int, status Status, now time.Time) error {
	d.Name = strings.TrimSpace(name)
	d.Code = strings.ToUpper(strings.TrimSpace(code))
	d.HeadID = headID
	d.ParentID = parentID
	d.EmployeeCount = employeeCount
	d.Status = status
	d.UpdatedAt = now.UTC()
	return d.Validate()
}
