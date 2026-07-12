package domain

import (
	"strings"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type Severity string

const (
	SeverityLow    Severity = "low"
	SeverityMedium Severity = "medium"
	SeverityHigh   Severity = "high"
)

type IssueStatus string

const (
	StatusOpen       IssueStatus = "open"
	StatusInProgress IssueStatus = "in_progress"
	StatusResolved   IssueStatus = "resolved"
)

type ComplianceIssue struct {
	ID           id.ID       `json:"id"`
	AuditID      *id.ID      `json:"auditId,omitempty"`
	DepartmentID id.ID       `json:"departmentId"`
	Severity     Severity    `json:"severity"`
	Description  string      `json:"description"`
	OwnerID      id.ID       `json:"ownerId"`
	DueDate      time.Time   `json:"dueDate"`
	Status       IssueStatus `json:"status"`
	CreatedAt    time.Time   `json:"createdAt"`
	// join fields
	OwnerName      string `json:"ownerName,omitempty"`
	DepartmentName string `json:"departmentName,omitempty"`
	AuditTitle     string `json:"auditTitle,omitempty"`
	Overdue        bool   `json:"overdue"`
}

// NewIssue enforces the non-negotiable owner + due date invariant.
func NewIssue(dept, owner id.ID, sev Severity, desc string, due time.Time, auditID *id.ID, now time.Time) (*ComplianceIssue, error) {
	fields := map[string]string{}
	if dept == "" {
		fields["departmentId"] = "Department is required"
	}
	if owner == "" {
		fields["ownerId"] = "every issue needs an owner"
	}
	if due.IsZero() {
		fields["dueDate"] = "every issue needs a due date"
	}
	desc = strings.TrimSpace(desc)
	if desc == "" {
		fields["description"] = "Description is required"
	}
	switch sev {
	case SeverityLow, SeverityMedium, SeverityHigh:
	default:
		fields["severity"] = "Severity must be low, medium, or high"
	}
	if len(fields) > 0 {
		if fields["ownerId"] != "" && fields["dueDate"] != "" {
			return nil, errs.Invalid("owner_and_due_required", "every issue needs an owner and a due date", fields)
		}
		if fields["ownerId"] != "" {
			return nil, errs.Invalid("owner_required", "every issue needs an owner", fields)
		}
		if fields["dueDate"] != "" {
			return nil, errs.Invalid("due_required", "every issue needs a due date", fields)
		}
		return nil, errs.Invalid("invalid_issue", "Compliance issue is invalid", fields)
	}
	return &ComplianceIssue{
		ID: id.New(), AuditID: auditID, DepartmentID: dept, Severity: sev,
		Description: desc, OwnerID: owner, DueDate: due.UTC().Truncate(24 * time.Hour),
		Status: StatusOpen, CreatedAt: now.UTC(),
	}, nil
}

// IsOverdue: open issues past the due day (due date itself is not overdue).
func (i *ComplianceIssue) IsOverdue(now time.Time) bool {
	if i.Status != StatusOpen {
		return false
	}
	// Compare calendar days: overdue when now is after end of due date.
	dueEnd := i.DueDate.UTC().Truncate(24 * time.Hour).Add(24*time.Hour - time.Nanosecond)
	return now.UTC().After(dueEnd)
}

func (i *ComplianceIssue) UpdateStatus(status IssueStatus) error {
	switch status {
	case StatusOpen, StatusInProgress, StatusResolved:
		i.Status = status
		return nil
	default:
		return errs.Invalid("invalid_status", "Status must be open, in_progress, or resolved", nil)
	}
}
