package app

import (
	"context"
	"testing"
	"time"

	ack "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/ack/domain"
	audit "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/audit/domain"
	compliance "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/compliance/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/port"
	policy "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/policy/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type memIssues struct{ items map[id.ID]*compliance.ComplianceIssue }

func (m *memIssues) Create(_ context.Context, i *compliance.ComplianceIssue) error {
	if m.items == nil {
		m.items = map[id.ID]*compliance.ComplianceIssue{}
	}
	m.items[i.ID] = i
	return nil
}
func (m *memIssues) ByID(_ context.Context, issueID id.ID) (*compliance.ComplianceIssue, error) {
	i, ok := m.items[issueID]
	if !ok {
		return nil, errs.NotFound("not_found", "not found")
	}
	cp := *i
	return &cp, nil
}
func (m *memIssues) Save(_ context.Context, i *compliance.ComplianceIssue) error {
	m.items[i.ID] = i
	return nil
}
func (m *memIssues) List(context.Context, page.Page, string, *bool, time.Time) (page.Result[compliance.ComplianceIssue], error) {
	return page.Result[compliance.ComplianceIssue]{}, nil
}
func (m *memIssues) OpenPastDue(context.Context, time.Time) ([]compliance.ComplianceIssue, error) {
	return nil, nil
}
func (m *memIssues) Stats(context.Context, time.Time) (int, int, int, error) { return 0, 0, 0, nil }

type nopAudits struct{}

func (nopAudits) Create(context.Context, *audit.Audit) error { return nil }
func (nopAudits) ByID(context.Context, id.ID) (*audit.Audit, error) {
	return nil, errs.NotFound("x", "x")
}
func (nopAudits) List(context.Context, page.Page) (page.Result[audit.Audit], error) {
	return page.Result[audit.Audit]{}, nil
}

type nopAcks struct{}

func (nopAcks) Create(context.Context, *ack.PolicyAcknowledgement) error { return nil }
func (nopAcks) List(context.Context, page.Page) (page.Result[ack.PolicyAcknowledgement], error) {
	return page.Result[ack.PolicyAcknowledgement]{}, nil
}
func (nopAcks) UnacknowledgedPolicies(context.Context, id.ID) ([]policy.Policy, error) {
	return nil, nil
}
func (nopAcks) AckRate(context.Context, id.ID, int) (int, int, error) { return 0, 0, nil }
func (nopAcks) ListByDepartment(context.Context, id.ID) ([]ack.PolicyAcknowledgement, error) {
	return nil, nil
}

type nopPolicies struct{}

func (nopPolicies) ByID(context.Context, id.ID) (*policy.Policy, error) {
	return nil, errs.NotFound("x", "x")
}
func (nopPolicies) List(context.Context) ([]policy.Policy, error) { return nil, nil }

type nopBundle struct{}

func (nopBundle) DepartmentBundle(context.Context, id.ID) (port.DepartmentBundle, error) {
	return port.DepartmentBundle{}, nil
}

func TestRaiseIssueEmitsEvent(t *testing.T) {
	bus := events.NewInProcess()
	var raised events.ComplianceIssueRaised
	bus.Subscribe(events.NameComplianceIssueRaised, func(_ context.Context, e events.Event) error {
		raised = e.(events.ComplianceIssueRaised)
		return nil
	})
	svc := New(nopAudits{}, &memIssues{}, nopAcks{}, nopPolicies{}, nopBundle{}, bus)
	svc.now = func() time.Time { return time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC) }
	owner := id.New()
	dept := id.New()
	issue, err := svc.RaiseIssue(context.Background(), RaiseIssueCmd{
		DepartmentID: dept, OwnerID: owner, Severity: compliance.SeverityHigh,
		Description: "Missing MSDS", DueDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if raised.OwnerID != owner || raised.IssueID != issue.ID {
		t.Fatalf("%+v", raised)
	}
}

func TestRaiseIssueRejectsMissingOwner(t *testing.T) {
	svc := New(nopAudits{}, &memIssues{}, nopAcks{}, nopPolicies{}, nopBundle{}, events.NewInProcess())
	_, err := svc.RaiseIssue(context.Background(), RaiseIssueCmd{
		DepartmentID: id.New(), Severity: compliance.SeverityLow, Description: "x",
		DueDate: time.Now(),
	})
	if err == nil {
		t.Fatal("expected owner_required")
	}
}
