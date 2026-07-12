package domain

import (
	"testing"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

func TestNewIssueRequiresOwnerAndDue(t *testing.T) {
	now := time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC)
	_, err := NewIssue(id.New(), "", SeverityHigh, "gap", time.Time{}, nil, now)
	if err == nil {
		t.Fatal("expected error")
	}
	e, ok := errs.As(err)
	if !ok {
		t.Fatal(err)
	}
	if e.Code != "owner_required" && e.Code != "owner_and_due_required" && e.Code != "due_required" {
		// both missing — either combined or first field
		if e.Fields["ownerId"] == "" && e.Fields["dueDate"] == "" {
			t.Fatalf("code=%s fields=%v", e.Code, e.Fields)
		}
	}

	_, err = NewIssue(id.New(), id.New(), SeverityHigh, "gap", time.Time{}, nil, now)
	if err == nil {
		t.Fatal("due required")
	}
	if e, _ := errs.As(err); e.Code != "due_required" {
		t.Fatalf("code=%s", e.Code)
	}

	_, err = NewIssue(id.New(), "", SeverityHigh, "gap", now, nil, now)
	if err == nil {
		t.Fatal("owner required")
	}
	if e, _ := errs.As(err); e.Code != "owner_required" {
		t.Fatalf("code=%s", e.Code)
	}
}

func TestIsOverdueBoundary(t *testing.T) {
	due := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	issue, err := NewIssue(id.New(), id.New(), SeverityMedium, "late", due, nil, due)
	if err != nil {
		t.Fatal(err)
	}
	// same calendar day — not overdue
	sameDay := time.Date(2026, 7, 10, 23, 0, 0, 0, time.UTC)
	if issue.IsOverdue(sameDay) {
		t.Fatal("due date itself should not be overdue")
	}
	// next day — overdue
	nextDay := time.Date(2026, 7, 11, 0, 0, 1, 0, time.UTC)
	if !issue.IsOverdue(nextDay) {
		t.Fatal("expected overdue after due date")
	}
	issue.Status = StatusResolved
	if issue.IsOverdue(nextDay) {
		t.Fatal("resolved issues are not overdue")
	}
}
