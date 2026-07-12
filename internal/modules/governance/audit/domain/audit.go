package domain

import (
	"strings"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type AuditStatus string

const (
	StatusDraft       AuditStatus = "draft"
	StatusUnderReview AuditStatus = "under_review"
	StatusCompleted   AuditStatus = "completed"
)

type Audit struct {
	ID           id.ID       `json:"id"`
	Title        string      `json:"title"`
	DepartmentID id.ID       `json:"departmentId"`
	AuditorID    id.ID       `json:"auditorId"`
	AuditDate    time.Time   `json:"auditDate"`
	Findings     string      `json:"findings"`
	Status       AuditStatus `json:"status"`
	CreatedAt    time.Time   `json:"createdAt"`
	// joins
	DepartmentName string `json:"departmentName,omitempty"`
	AuditorName    string `json:"auditorName,omitempty"`
}

func New(title string, departmentID, auditorID id.ID, auditDate time.Time, findings string, status AuditStatus, now time.Time) (*Audit, error) {
	a := &Audit{
		ID: id.New(), Title: strings.TrimSpace(title), DepartmentID: departmentID, AuditorID: auditorID,
		AuditDate: auditDate.UTC(), Findings: strings.TrimSpace(findings), Status: status, CreatedAt: now.UTC(),
	}
	if a.Status == "" {
		a.Status = StatusDraft
	}
	if err := a.Validate(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Audit) Validate() error {
	fields := map[string]string{}
	if a.Title == "" {
		fields["title"] = "Title is required"
	}
	if a.DepartmentID == "" {
		fields["departmentId"] = "Department is required"
	}
	if a.AuditorID == "" {
		fields["auditorId"] = "Auditor is required"
	}
	if a.AuditDate.IsZero() {
		fields["auditDate"] = "Audit date is required"
	}
	switch a.Status {
	case StatusDraft, StatusUnderReview, StatusCompleted:
	default:
		fields["status"] = "Invalid status"
	}
	if len(fields) > 0 {
		return errs.Invalid("invalid_audit", "Audit details are invalid", fields)
	}
	return nil
}
